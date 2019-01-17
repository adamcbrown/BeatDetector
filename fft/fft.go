package fft

import (
	"math"
)

var cosTable [100000]float64

func GetPowerOfSamples(samples []float64, sampleRate, freq float64) float64 {
	if cosTable[0] == 0 {
		for i := 0; i < len(cosTable); i++ {
			cosTable[i] = math.Cos(float64(i) / 10000)
		}
	}

	a, b := 0.0, 0.0
	period := float64(len(samples)) / sampleRate
	for i := 0; i < len(samples); i++ {
		amp := samples[i]

		temp := (freq * float64(i) / sampleRate)
		t := 2 * math.Pi * (temp - math.Floor(temp))

		index := int64(t * 10000)
		cosV := cosTable[index]
		sinV := cosTable[78540-index] // sin(x) = cos(pi/2 - x)

		// sinV, cosV := math.Sincos(t)

		// tFactor := t

		// sinV := 0.0
		// cosV := 1.0

		// sinV += tFactor
		// tFactor *= t
		// cosV -= tFactor / 2
		// tFactor *= t
		// sinV -= tFactor / 6
		// tFactor *= t
		// cosV += tFactor / 24
		// tFactor *= t
		// sinV += tFactor / 120

		sinV /= sampleRate
		cosV /= sampleRate

		a += 2 * cosV * amp
		b += 2 * sinV * amp

	}

	a *= 2 / period
	b *= 2 / period

	return math.Sqrt(a*a + b*b)
}
