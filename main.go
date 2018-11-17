package main

import (
	"fmt"

	// "github.com/adamcbrown/frequency-classifier/fft"
	"github.com/adamcbrown/beat-detector/signal"
	"github.com/adamcbrown/beat-detector/song"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"math"
	// "math/rand"
	"os"
	"sync"
	"time"
)

const (
	WIDTH       = 1024
	HEIGHT      = 768
	JUMP        = 1.1
	NEED        = 1.2
	DECAY       = 0.97
	path        = "/Users/Brown/Desktop/Music Samples/Baby Driver Opening Scene (2017)  Movieclips Coming Soon.wav"
	BLOCK_WIDTH = 5
	INTERVAL    = 0.05
)

var (
	SEMITONE     float64
	c            chan song.Moment
	DECAY_FACTOR = math.Pow(1-DECAY, INTERVAL)
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, WIDTH, HEIGHT),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	mutex := &sync.Mutex{}

	imd := imdraw.New(nil)

	x := 0.0

	f, _ := os.Open(path)
	s, format, _ := wav.Decode(f)
	speaker.Init(format.SampleRate, 100000)
	speaker.Play(s)
	start := time.Now()

	go func() {
		rollingAvg := 0.0
		var rollingAvgArr [1]float64
		rollingAvgIdx := 0

		ts := 0.0
		maxBass := 0.05
		// threshold := 1.0
		// shouldWait := true
		bassPowerStream := make(chan float64)
		beatCh := signal.ZScoreStream(bassPowerStream, 20, 5, 4.2, 0.5)

		for {
			moment, more := <-c

			ratio := moment.BassPower / maxBass

			rollingAvg -= rollingAvgArr[rollingAvgIdx] / float64(len(rollingAvgArr))
			rollingAvg += moment.BassPower / float64(len(rollingAvgArr))
			rollingAvgArr[rollingAvgIdx] = moment.BassPower

			rollingAvgIdx = (rollingAvgIdx + 1) % len(rollingAvgArr)

			bassPowerStream <- rollingAvg

			maxBass = math.Max(moment.BassPower, maxBass)
			currentTime := time.Since(start)
			// fmt.Println(time.Duration(ts*1000)*time.Millisecond, currentTime)
			if currentTime.Seconds() < ts {
				time.Sleep(time.Duration(ts*1000)*time.Millisecond - currentTime)
			}

			if !more {
				break
			}

			mutex.Lock()
			h := float64(HEIGHT) / float64(len(moment.Frequencies))

			// // Draw Threshold
			// imd.Color = pixel.RGB(0, 0, 1)
			// imd.Push(pixel.V(x, 0), pixel.V(x+BLOCK_WIDTH, HEIGHT*math.Min(1, threshold*NEED)))
			// imd.Rectangle(0)

			//Draw Ratio
			imd.Color = pixel.RGB(1, 0, 0)
			imd.Push(pixel.V(x, 0), pixel.V(x+BLOCK_WIDTH, HEIGHT*ratio))
			imd.Rectangle(0)

			for i := 0; i < len(moment.Frequencies); i++ {

				y := float64(i) * h
				pow := moment.Frequencies[i]

				// imd.Color = pixel.RGB(0, 0, 1)

				imd.Color = pixel.RGB(pow, pow, pow).Mul(pixel.Alpha(0.8))
				imd.Push(pixel.V(x, y))
				imd.Push(pixel.V(x+BLOCK_WIDTH, y+h))
				imd.Rectangle(0)

			}

			//Draw Beat line
			isBeat := <-beatCh
			if isBeat && ratio > 0.1 {
				// if ratio > threshold*NEED {
				// if !shouldWait {
				imd.Color = pixel.RGB(1, 0, 1)
				imd.Push(pixel.V(x+BLOCK_WIDTH/2, 0), pixel.V(x+BLOCK_WIDTH/2, HEIGHT))
				imd.Line(1)
				// shouldWait = true
				// }
			}
			// } else {
			// shouldWait = false
			// }

			// if threshold > ratio {
			// 	threshold = (threshold*DECAY_FACTOR + ratio*(1-DECAY_FACTOR))
			// } else {
			// 	threshold = ratio * JUMP
			// }
			// threshold = math.Max(0.05, threshold)

			ts = moment.Ts

			x += BLOCK_WIDTH
			mutex.Unlock()

		}
		fmt.Println("Done!")
	}()

	for !win.Closed() {

		camPos := pixel.ZV.Add(pixel.V(math.Max(x-WIDTH, 0), 0))
		cam := pixel.IM.Moved(camPos.Scaled(-1))
		mutex.Lock()
		win.SetMatrix(cam)
		win.Clear(pixel.RGB(0, 0, 0))
		imd.Draw(win)
		mutex.Unlock()

		win.Update()
	}
}

func main() {
	c = make(chan song.Moment)
	go func() {

		sampler, err := song.NewWavSampler(path)
		if err != nil {
			fmt.Printf("%s", err.Error())
			os.Exit(1)
		}

		err = song.ExtractFrequencies(c, sampler, INTERVAL, 4)

		if err != nil {
			fmt.Printf("%s", err.Error())
			os.Exit(1)
		}
	}()

	pixelgl.Run(run)
}

// func main() {
// 	c = make(chan song.Moment)
// 	go func() {

// 		sampler, err := song.NewPortAudioSampler("Built-in Output")
// 		if err != nil {
// 			fmt.Printf("%s\n", err.Error())
// 			os.Exit(1)
// 		}

// 		err = song.ExtractFrequencies(c, sampler, INTERVAL, 2)

// 		if err != nil {
// 			fmt.Printf("%s\n", err.Error())
// 			os.Exit(1)
// 		}
// 	}()

// 	pixelgl.Run(run)
// }

// func main() {
// 	SEMITONE = math.Pow(2, 1.0/12.0)

// 	n := deep.NewNeural(&deep.Config{
// 		/* Input dimensionality */
// 		Inputs: NUM_NOTES,
// 		Layout: []int{NUM_NOTES, NUM_NOTES, NUM_NOTES},
// 		/* Activation functions: {deep.Sigmoid, deep.Tanh, deep.ReLU, deep.Linear} */
// 		Activation: deep.ActivationSigmoid,
// 		/* Determines output layer activation & loss function:
// 		ModeRegression: linear outputs with MSE loss
// 		ModeMultiClass: softmax output with Cross Entropy loss
// 		ModeMultiLabel: sigmoid output with Cross Entropy loss
// 		ModeBinary: sigmoid output with binary CE loss */
// 		Mode: deep.ModeBinary,
// 		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
// 		Weight: deep.NewNormal(1.0, 0.0),
// 		/* Apply bias */
// 		Bias: true,
// 	})

// 	if TEST_EXAMPLE {
// 		in, out := generateSimpleDatum()
// 		for i, value := range out {
// 			freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 			fmt.Printf("%v: \t\t%v\n", freq, value)
// 		}
// 		fmt.Println()

// 		var fullIn [SAMPLES + NUM_NOTES]float64
// 		copy(fullIn[:], in[:])
// 		copy(fullIn[SAMPLES:], getEstimates(in[:]))

// 		out = n.Predict(fullIn[SAMPLES:])
// 		for i, value := range out {
// 			freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 			fmt.Printf("%v: \t\t%v\n", freq, value)
// 		}
// 		return
// 	}

// 	var data training.Examples
// 	for i := 0; i < 200; i++ {
// 		in, out := generateSimpleDatum()

// 		data = append(data, training.Example{
// 			Response: out,
// 			Input:    getEstimates(in),
// 		})
// 	}

// 	optimizer := training.NewSGD(0.05, 0.1, 1e-6, true)
// 	trainer := training.NewTrainer(optimizer, 50)
// 	training, heldout := data.Split(0.5)
// 	trainer.Train(n, training, heldout, 1000)

// 	in, out := generateSimpleDatum()
// 	for i, value := range out {
// 		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 		fmt.Printf("%v: \t\t%v\n", freq, value)
// 	}
// 	fmt.Println()

// 	var fullIn [SAMPLES + NUM_NOTES]float64
// 	copy(fullIn[:], in[:])
// 	copy(fullIn[SAMPLES:], getEstimates(in[:]))

// 	out = n.Predict(fullIn[SAMPLES:])
// 	for i, value := range out {
// 		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 		fmt.Printf("%v: \t\t%v\n", freq, value)
// 	}
// }

// func getEstimates(input []float64) []float64 {
// 	fFreq := fft.GetFundementalFrequency(input, SAMPLE_RATE)
// 	extended := fft.GetExtendedWave(input, fFreq, SAMPLE_RATE, EXTENSIONS)
// 	var estimates [NUM_NOTES]float64
// 	for i := 0; i < NUM_NOTES; i++ {
// 		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 		estimates[i] = fft.GetPowerOfSamples(extended, SAMPLE_RATE, freq)
// 	}
// 	return estimates[:]
// }

// func generateSimpleDatum() ([]float64, []float64) {
// 	s1 := rand.NewSource(time.Now().UnixNano())
// 	r1 := rand.New(s1)

// 	var frequencySpace [NUM_NOTES]float64
// 	var outputSpace [SAMPLES]float64
// 	count := 0
// 	for count < 1 {
// 		i := r1.Intn(NUM_NOTES)
// 		if frequencySpace[i] != 1 {
// 			frequencySpace[i] = 1
// 			count++
// 		}

// 	}

// 	max := -1.0
// 	for i := 0; i < NUM_NOTES; i++ {
// 		// phase := r1.Float64() * 2 * math.Pi
// 		phase := 0.0

// 		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 		for j := 0; j < SAMPLES; j++ {
// 			t := float64(j) / float64(SAMPLE_RATE)
// 			outputSpace[j] += math.Sin(freq*2*math.Pi*t + phase)

// 			if i == NUM_NOTES-1 {
// 				max = math.Max(max, math.Abs(outputSpace[j]))
// 			}
// 		}
// 	}

// 	factor := 1 / max
// 	for i := 0; i < SAMPLES; i++ {
// 		outputSpace[i] = outputSpace[i] * factor
// 	}

// 	return outputSpace[:], frequencySpace[:]
// }

// func generateTestDatum(magnitude float64) ([]float64, []float64) {
// 	s1 := rand.NewSource(time.Now().UnixNano())
// 	r1 := rand.New(s1)

// 	var frequencySpace [NUM_NOTES]float64
// 	var outputSpace [SAMPLES]float64
// 	max := -1.0
// 	for i := 0; i < NUM_NOTES; i++ {
// 		frequencySpace[i] = r1.Float64()
// 		phase := r1.Float64() * 2 * math.Pi

// 		freq := BASE_FREQUENCY * math.Pow(SEMITONE, float64(i))
// 		for j := 0; j < SAMPLES; j++ {
// 			t := float64(j) / float64(SAMPLE_RATE)
// 			outputSpace[j] += math.Sin(freq*2*math.Pi*t + phase)

// 			if i == NUM_NOTES-1 {
// 				max = math.Max(max, math.Abs(outputSpace[j]))
// 			}
// 		}
// 	}

// 	factor := magnitude / max
// 	for i := 0; i < SAMPLES; i++ {
// 		outputSpace[i] = outputSpace[i] * factor
// 	}

// 	return outputSpace[:], frequencySpace[:]
// }
