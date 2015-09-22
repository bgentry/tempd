package temperature

import "math"

// KToF converts a Kelvin temperature to degrees Celsius.
func KToC(tempc float64) float64 {
	return (tempc - 273.15)
}

// KToF converts a Kelvin temperature to degrees Fahrenheit.
func KToF(tempk float64) float64 {
	return KToC(tempk)*(9.0/5) + 32
}

// SteinhartTemp returns the temperature in Kelvin of a thermistor based on its
// Steinhart-Hart equation constants a, b, and c, as well as the thermistor's
// resistance at the current temperature.
func SteinhartTemp(a, b, c, resistance float64) (kelvin float64) {
	return 1 / (a + b*math.Log(resistance) + c*(math.Pow(math.Log(resistance), 3)))
}
