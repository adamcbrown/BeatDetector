package song

import (
	"github.com/adamcbrown/beat-detector/fft"
	"io"
	"math"
)

const (
	BASE_FREQUENCY = 13.75
	NUM_NOTES      = 128
	SEMITONE       = 1.0594630944
)

type Moment struct {
	rawSamples  []float64
	Frequencies []float64
	maxSamples  int
	samples     int
	Ts          float64
	BassPower   float64
}

func (m *Moment) init(samplesPerMoment int, ts float64) {
	m.rawSamples = make([]float64, samplesPerMoment)
	m.Frequencies = make([]float64, NUM_NOTES)
	m.maxSamples = samplesPerMoment
	m.samples = 0
	m.Ts = ts
}

func (m *Moment) sample(sampleValue float64) bool {
	m.rawSamples[m.samples] = sampleValue
	m.samples += 1
	return m.samples == m.maxSamples
}

type Sampler interface {
	Next() (float64, error)
	SampleRate() float64
}

func ExtractFrequencies(c chan Moment, sampler Sampler, interval float64, extension int) error {
	samplesPerMoment := int(interval * float64(sampler.SampleRate()))

	var moment Moment
	moment.init(samplesPerMoment, 0)
	for {
		sample, err := sampler.Next()
		if err == io.EOF {
			break
		}

		if moment.sample(sample) {
			processMoment(&moment, sampler.SampleRate(), extension)
			c <- moment

			nextTs := moment.Ts + interval
			moment = Moment{}
			moment.init(samplesPerMoment, nextTs)
		}
	}

	close(c)
	return nil
}

func processMoment(moment *Moment, sampleRate float64, extension int) {
	//fmt.Println("\033[H\033[2JTIME: ", moment.ts)
	fFreq := fft.GetFundementalFrequency(moment.rawSamples, sampleRate)
	extended := fft.GetExtendedWave(moment.rawSamples, fFreq, sampleRate, extension)
	for i := 0; i < NUM_NOTES; i++ {
		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
		moment.Frequencies[i] = fft.GetPowerOfSamples(extended, sampleRate, freq)
	}
	moment.BassPower = 0

	highSample := int(math.Ceil(math.Log(150/'BASE_FREQUENCY) / math.Log(SEMITONE)))
	lowSample := int(math.Max(0, math.Ceil(math.Log(30/BASE_FREQUENCY)/math.Log(SEMITONE))))
	for i := lowSample; i < highSample; i++ {
		moment.BassPower += moment.Frequencies[i]
	}
	moment.BassPower /= float64(highSample - lowSample)
}
