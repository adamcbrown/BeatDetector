package fft

import (
	"math"
)

func GetFundementalFrequency(samples []float64, sampleRate float64) float64 {
	max := 0.0
	var prev0, next0, count0 int
	for i := 0; i < len(samples); i++ {
		max = math.Max(max, math.Abs(samples[i]))

		if math.Abs(samples[i]) < 0.03 {
			prev0 = next0
			next0 = i
			if count0 < next0-prev0 {
				count0 = next0 - prev0
			}
		}
	}

	min := math.Inf(-1)
	maxSum, avgSum := 0.0, 0.0
	minAt := 3

	sums := make([]float64, len(samples), len(samples))
	for i := count0; i < len(samples)*3/4; i++ {
		sum := 0.0
		for j := 0; j < len(samples)/4; j++ {
			diff := samples[j] - samples[i+j]
			sum += diff * diff / max
		}

		if min > sum {
			min = sum
			minAt = i
		}

		if maxSum < sum {
			maxSum = sum
		}

		sums[i] = sum
		avgSum += sum
	}
	avgSum /= float64(len(samples))

	thresh := min + (maxSum-min)*0.01

	goDown := false
	for i := count0; i < minAt; i++ {
		if sums[i] < thresh {
			goDown = true
		}

		if goDown {
			if sums[i] > sums[i-1] {
				minAt = i
				break
			}
		}
	}

	goDown = false
	highMin := minAt
	for i := len(samples)*3/4 - 2; i > minAt; i-- {
		if sums[i] < thresh {
			goDown = true
		}

		if goDown {
			if sums[i] > sums[i+1] {
				highMin = i
				break
			}
		}
	}

	harmonic := math.Round(float64(highMin) / float64(minAt))
	return sampleRate * harmonic / float64(highMin)
}

func GetExtendedWave(samples []float64, fFreq, sampleRate float64, extensions int) []float64 {
	samplesPerCycle := sampleRate / fFreq
	pull := int(math.Round(float64(len(samples))/samplesPerCycle) * samplesPerCycle)

	extended := make([]float64, pull*extensions, pull*extensions)
	for i := 0; i < extensions; i++ {
		copy(extended[i*pull:], samples[:])
	}

	return extended
}
