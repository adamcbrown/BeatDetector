package song

import (
	"github.com/youpy/go-wav"
	"os"
)

type WavSampler struct {
	reader      *wav.Reader
	samples     []wav.Sample
	file        *os.File
	sampleCount int
	sampleRate  float64
}

func NewWavSampler(filepath string) (*WavSampler, error) {
	var w WavSampler

	var err error
	w.file, err = os.Open(filepath)
	if err != nil {
		return nil, err
	}

	w.sampleCount = 0
	w.reader = wav.NewReader(w.file)
	format, err := w.reader.Format()
	w.sampleRate = float64(format.SampleRate)
	w.samples = nil
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (w *WavSampler) Next() (float64, error) {
	if w.samples == nil {
		w.sampleCount = 0
		samples, err := w.reader.ReadSamples()
		w.samples = samples
		if err != nil {
			return 0, err
		}
	}

	sample := w.samples[w.sampleCount]
	w.sampleCount += 1
	if w.sampleCount == len(w.samples) {
		w.samples = nil
	}
	return (w.reader.FloatValue(sample, 0) + w.reader.FloatValue(sample, 1)) / 2, nil
}

func (w *WavSampler) SampleRate() float64 {
	return w.sampleRate
}

func (w *WavSampler) Close() error {
	return w.file.Close()
}
