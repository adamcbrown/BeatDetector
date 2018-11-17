package fft

import (
	"math"
)

func GetPowerOfSamples(samples []float64, sampleRate, freq float64) float64 {
	a, b := 0.0, 0.0
	period := float64(len(samples)) / sampleRate
	for i := 0; i < len(samples); i++ {
		amp := samples[i]

		sinV, cosV := math.Sincos(2 * math.Pi * freq * float64(i) / sampleRate)
		sinV /= sampleRate
		cosV /= sampleRate

		a += 2 * cosV * amp
		b += 2 * sinV * amp
	}

	a *= 2 / period
	b *= 2 / period

	return math.Sqrt(a*a + b*b)
}
