package main

import (
	"fmt"
	"github.com/adamcbrown/beat-detector/signal"
	"github.com/adamcbrown/beat-detector/song"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"math"
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

	start := time.Now()

	go func() {
		rollingAvg := 0.0
		var rollingAvgArr [1]float64
		rollingAvgIdx := 0
		last := 0.0

		ts := 0.0
		maxBass := 0.05
		// threshold := 1.0
		// shouldWait := true
		bassPowerStream := make(chan float64)
		beatCh := signal.ZScoreStream(bassPowerStream, 50, 5, 3, 0.5)

		for {
			moment, more := <-c

			ratio := moment.BassPower / maxBass

			rollingAvg -= rollingAvgArr[rollingAvgIdx] / float64(len(rollingAvgArr))
			rollingAvg += moment.BassPower / float64(len(rollingAvgArr))
			rollingAvgArr[rollingAvgIdx] = moment.BassPower

			rollingAvgIdx = (rollingAvgIdx + 1) % len(rollingAvgArr)

			bassPowerStream <- rollingAvg - last
			last = rollingAvg

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

// func main() {
// 	c = make(chan song.Moment)
// 	go func() {

// 		sampler, err := song.NewWavSampler(path)
// 		if err != nil {
// 			fmt.Printf("%s", err.Error())
// 			os.Exit(1)
// 		}

// 		err = song.ExtractFrequencies(c, sampler, INTERVAL, 4)

// 		if err != nil {
// 			fmt.Printf("%s", err.Error())
// 			os.Exit(1)
// 		}
// 	}()

// 	pixelgl.Run(run)
// }

func main() {
	c = make(chan song.Moment)
	go func() {

		sampler, err := song.NewPortAudioSampler("Built-in Output")
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}

		err = song.ExtractFrequencies(c, sampler, INTERVAL, 2, 13.75, 200)

		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
	}()

	pixelgl.Run(run)
}
