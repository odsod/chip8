package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/odsod/chip8/ui/terminal"
)

func main() {
	romFile := flag.String("rom", "roms/TETRIS", "The ROM to load")
	keyboardLayout := flag.String("keyboard", "qwer", "They keyboard layout")
	cpuFrequencyHz := flag.Int("cpuFrequency", 500, "The CPU frequency (Hz)")
	timerFrequencyHz := flag.Int("timerFrequency", 60, "The timer frequency (Hz)")
	frameRateHz := flag.Int("frameRate", 60, "The frame rate (Hz)")
	emulatorFrequencyHz := flag.Int("emulatorFrequency", 100, "The emulator frequency (Hz)")
	keyPressDurationMs := flag.Int("keyPressDuration", 100, "The key press duration (ms)")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	ui := terminal.NewUI(terminal.Conf{
		RomFile:             *romFile,
		KeyboardLayout:      *keyboardLayout,
		CPUFrequencyHz:      *cpuFrequencyHz,
		TimerFrequencyHz:    *timerFrequencyHz,
		FrameRateHz:         *frameRateHz,
		EmulatorFrequencyHz: *emulatorFrequencyHz,
		KeyPressDuration:    time.Duration(*keyPressDurationMs) * time.Millisecond,
	})

	ui.Run()
}
