package measurements

import (
	"encoding/json"
	"fmt"
	"io"
)

type response struct {
	Voltage    float64 `json:"voltage"`
	Percentage string  `json:"percentage"`
}

func (db *DB) WriteCurrentStats(w io.Writer) error {
	if db.latestReading == 0 {
		// No voltage reading yet, so return without writing
		return nil
	}

	return json.NewEncoder(w).Encode(response{
		Voltage:    db.latestReading.Volts(),
		Percentage: fmt.Sprintf("%.0f%%", db.latestReading.Percentage()),
	})
}
