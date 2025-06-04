package main

import (
	"time"

	photon "photon/battery"
)

func main() {
	ph, err := photon.Init("measurements.db", 5*time.Minute)
	if err != nil {
		panic(err)
	}
	ph.ShutdownOnTerminate()

	if err := ph.Track(2 * time.Minute); err != nil {
		panic(err)
	}
}
