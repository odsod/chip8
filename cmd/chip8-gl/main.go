package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/odsod/chip8/ui/opengl"
)

func main() {
	romFile := flag.String("rom", "roms/TETRIS", "The ROM to load")
	cpuFrequencyHz := flag.Int("cpuFrequency", 500, "The CPU frequency (Hz)")
	timerFrequencyHz := flag.Int("timerFrequency", 60, "The timer frequency (Hz)")
	scale := flag.Int("scale", 8, "The graphics upscaling coefficient")
	pixelFadeTimeMs := flag.Int("pixelFadeTime", 90, "The pixel fade time (ms)")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	ui := opengl.NewUI(opengl.Options{
		RomFile:          *romFile,
		CPUFrequencyHz:   *cpuFrequencyHz,
		TimerFrequencyHz: *timerFrequencyHz,
		Scale:            *scale,
		PixelFadeTime:    time.Duration(*pixelFadeTimeMs) * time.Millisecond,
	})

	ui.Run()
}
