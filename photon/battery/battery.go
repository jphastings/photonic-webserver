package photon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"photon/measurements"
	"sync"
	"syscall"
	"time"

	"periph.io/x/conn/v3/gpio"
	host "periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

type photon struct {
	startPin  gpio.PinIO
	clockPin  gpio.PinIO
	dataPin   gpio.PinIO
	rpiBooted gpio.PinIO
	rpiReady  gpio.PinIO

	db            *measurements.DB
	latestReading measurements.MilliVolt
	tracking      sync.Mutex
}

// Connects to the photon, and prepares to read battery voltages.
// Will block until ready, or the timeout provided is exceeded. Recommended 5 minutes.
func Init(dbFile string, timeout time.Duration) (*photon, error) {
	db, err := measurements.OpenDB(dbFile)
	if err != nil {
		return nil, err
	}

	if _, err := host.Init(); err != nil {
		return nil, err
	}
	ph := photon{
		startPin:  rpi.P1_33, // GPIO13
		clockPin:  rpi.P1_35, // GPIO19
		dataPin:   rpi.P1_37, // GPIO26
		rpiBooted: rpi.P1_36, // GPIO16
		rpiReady:  rpi.P1_38, // GPIO20

		db: db,
	}

	if err := ph.rpiBooted.Out(gpio.High); err != nil {
		return nil, err
	}
	if err := ph.rpiReady.Out(gpio.High); err != nil {
		return nil, err
	}

	// These resistors are pulled low by the ATTiny84 on the Photon board, so leave them floating here.
	if err := ph.startPin.In(gpio.Float, gpio.FallingEdge); err != nil {
		return nil, err
	}
	if err := ph.clockPin.In(gpio.Float, gpio.RisingEdge); err != nil {
		return nil, err
	}
	if err := ph.dataPin.In(gpio.Float, gpio.NoEdge); err != nil {
		return nil, err
	}

	if !ph.startPin.WaitForEdge(timeout) {
		return nil, fmt.Errorf("the start pin didn't trigger within the %v timeout", timeout)
	}

	return &ph, nil
}

func (ph *photon) ShutdownOnTerminate() context.CancelFunc {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		if err := ph.shutdown(); err != nil {
			log.Print(err)
		}
	}()

	return stop
}

// Sets the pins such that the photon knows the device is being shut down
func (ph *photon) shutdown() error {
	if err := ph.rpiBooted.Out(gpio.Low); err != nil {
		return fmt.Errorf("unable to inform the photon we're shutting down: %w", err)
	}
	if err := ph.rpiReady.Out(gpio.High); err != nil {
		return fmt.Errorf("unable to inform the photon we're shutting down: %w", err)
	}
	return nil
}

var ErrAlreadyTracking = errors.New("the voltage is already being tracked")

// A blocking function that will check the battery voltage every
// provided period, and write it to the database
func (ph *photon) Track(period time.Duration) error {
	if !ph.tracking.TryLock() {
		return ErrAlreadyTracking
	}
	defer ph.tracking.Unlock()

	t := time.NewTicker(period)
	for {
		if err := ph.recordVoltage(); err != nil {
			return err
		}
		<-t.C
	}
}

const waitPeriod = 5 * time.Second

func (ph *photon) recordVoltage() error {
	readingTime := time.Now().UTC()

	reading, err := ph.readVoltage()
	if err != nil {
		return err
	}

	if err := ph.db.AddMeasurement(reading, readingTime); err != nil {
		return fmt.Errorf("unable to write reading to sqlite database: %w", err)
	}
	log.Printf("reading was %d mV (%.1f%%)", reading, reading.Percentage()*100)

	return nil
}

func (ph *photon) readVoltage() (measurements.MilliVolt, error) {
	// Wait for start pin to go high
	for ph.startPin.Read() == gpio.Low {
		time.Sleep(1 * time.Millisecond)
	}
	// Wait for start pin to go low again (end of 50ms pulse)
	for ph.startPin.Read() == gpio.High {
		time.Sleep(1 * time.Millisecond)
	}

	var reading uint16
	for bit := 7; bit >= 0; bit-- {
		// Wait for clock to go high (signal that data is ready)
		for ph.clockPin.Read() == gpio.High {
			time.Sleep(1 * time.Millisecond)
		}
		// Read the bit
		if ph.dataPin.Read() == gpio.High {
			reading |= 1 << bit
		}
		// Wait for clock to go low again
		for ph.clockPin.Read() == gpio.Low {
			time.Sleep(1 * time.Millisecond)
		}
	}

	// This scales the 8 bits of data sent from the ATTiny into millivolts
	return measurements.MilliVolt(reading * 40), nil
}
