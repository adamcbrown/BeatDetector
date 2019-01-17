package signal

import (
	"math"
)

func Average(data []float64) float64 {
	total := 0.0
	for _, v := range data {
		total += v
	}
	return total / float64(len(data))
}

func Std(data []float64, avg float64) float64 {
	sigmaSq := 0.0
	for _, v := range data {
		sigmaSq += (v - avg) * (v - avg)
	}
	sigmaSq /= float64(len(data))
	return math.Sqrt(sigmaSq)
}

func ZScoreStream(sampleCh <-chan float64, numSamples, lag int, threshold, influence float64) <-chan bool {

	wentToHigh := make(chan bool)

	samples := make([]float64, numSamples)

	currIndex := 0
	filteredY := make([]float64, numSamples)
	avgFilter := make([]float64, numSamples)
	stdFilter := make([]float64, numSamples)
	wasHigh := false

	go func() {
		for {
			sample, more := <-sampleCh
			if !more {
				return
			}

			if currIndex == numSamples-1 {
				samples = samples[1:numSamples]
				samples = append(samples, sample)
			} else {
				samples[currIndex] = sample
				currIndex += 1
				wentToHigh <- false
				continue
			}

			for i, sample := range samples[0:lag] {
				filteredY[i] = sample
			}
			avgFilter[lag] = 0 //Average(samples[0:lag])
			stdFilter[lag] = Std(samples[0:lag], 0)

			for i := lag + 1; i < len(samples); i++ {

				f := float64(samples[i])

				if float64(math.Abs(samples[i]-avgFilter[i-1])) > threshold*float64(stdFilter[i-1]) {
					filteredY[i] = influence*f + (1-influence)*float64(filteredY[i-1])
					avgFilter[i] = 0 //Average(filteredY[(i - lag):i])
					stdFilter[i] = Std(filteredY[(i-lag):i], 0)
				} else {
					filteredY[i] = samples[i]
					avgFilter[i] = 0 //Average(filteredY[(i - lag):i])
					stdFilter[i] = Std(filteredY[(i-lag):i], 0)
				}
			}

			beatsThresh := math.Abs(samples[numSamples-1]-avgFilter[numSamples-2]) > threshold*float64(stdFilter[numSamples-2])
			goingUp := (samples[numSamples-1] > avgFilter[numSamples-2])

			wentToHigh <- !wasHigh && beatsThresh && goingUp
			wasHigh = beatsThresh && goingUp
		}
	}()

	return wentToHigh
}
