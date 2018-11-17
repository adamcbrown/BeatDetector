package song

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
)

type PortAudioSampler struct {
	stream      *portaudio.Stream
	samples     []float32
	sampleCount int
	sampleRate  float64
	channels    int
}

func NewPortAudioSampler(channel string) (*PortAudioSampler, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	// devices, err := portaudio.Devices()
	// if err != nil {
	// 	return nil, err
	// }

	var device *portaudio.DeviceInfo
	// for _, x := range devices {
	// 	fmt.Println(x.Name)
	// 	if x.Name == channel {
	// 		device = x
	// 		break
	// 	}
	// }

	device, _ = portaudio.DefaultInputDevice()
	fmt.Println(device.Name)

	var sampler PortAudioSampler
	sampler.channels = device.MaxInputChannels
	sampler.sampleCount = -1

	sampler.stream, err = portaudio.OpenStream(portaudio.StreamParameters{
		Output: portaudio.StreamDeviceParameters{
			Device: nil,
		},
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: sampler.channels,
		},
		SampleRate:      device.DefaultSampleRate,
		FramesPerBuffer: 1024,
	}, &sampler.samples)

	if err != nil {
		return nil, err
	}

	return &sampler, nil
}

func (s *PortAudioSampler) Next() (float64, error) {
	if s.sampleCount == -1 {
		s.sampleCount = 0
		err := s.stream.Read()
		fmt.Println("Read", err)
		if err == nil {
			return 0, err
		}
	}
	fmt.Println("Samples: ", s.samples)

	avg := 0.0
	for i := 0; i < s.channels; i++ {
		avg += float64(s.samples[s.sampleCount+i])
	}
	avg /= float64(s.channels)

	s.sampleCount += 1
	// if s.sampleCount == len(s.samples[0]) {
	// 	s.sampleCount = -1
	// }

	return float64(avg), nil
}

func (s *PortAudioSampler) SampleRate() float64 {
	return s.sampleRate
}

func (s *PortAudioSampler) Close() error {
	return s.stream.Close()
}
