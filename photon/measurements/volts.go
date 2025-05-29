package measurements

type MilliVolt int

func (uv MilliVolt) Volts() float64 {
	return float64(uv) / 1000
}

func (uv MilliVolt) Percentage() float64 {
	if uv == 0 {
		return 0
	}

	return (float64(uv) - float64(voltageLow)) / float64(voltageHigh-voltageLow)
}
