package terminal

import (
	"fmt"
	"io/ioutil"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/odsod/chip8"
)

type Conf struct {
	RomFile             string
	KeyboardLayout      string
	CPUFrequencyHz      int
	TimerFrequencyHz    int
	FrameRateHz         int
	EmulatorFrequencyHz int
	KeyPressDuration    time.Duration
}

type UI struct {
	keyboard *Keyboard
	display  *Display
	vm       *chip8.VM
	conf     Conf
}

func NewUI(conf Conf) *UI {
	rom, err := ioutil.ReadFile(conf.RomFile)
	if err != nil {
		panic(err)
	}

	var keyMap map[rune]uint8
	switch conf.KeyboardLayout {
	case "qwer":
		keyMap = QWER
	case "dvorak":
		keyMap = Dvorak
	default:
		panic(fmt.Sprintf("Unsupported keyboard layout: %s", conf.KeyboardLayout))
	}

	return &UI{
		keyboard: NewKeyboard(conf.KeyPressDuration, keyMap),
		display:  NewDisplay(),
		vm:       chip8.New(rom),
		conf:     conf,
	}
}

func targetUpdates(runTime time.Duration, updateFrequencyHz int) int {
	updateInterval := time.Second / time.Duration(updateFrequencyHz)
	return int(runTime / updateInterval)
}

func (ui *UI) Run() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	ui.keyboard.Listen()

	startTime := time.Now()
	timerCycles := 0
	cpuCycles := 0
	frames := 0

	for {
		now := time.Now()
		keys, quit := ui.keyboard.Check(now)
		if quit {
			return
		}
		ui.vm.SetKeys(keys)
		runTime := now.Sub(startTime)
		for i := timerCycles; i < targetUpdates(runTime, ui.conf.TimerFrequencyHz); i++ {
			ui.vm.TickTimers()
			timerCycles++
		}
		for i := cpuCycles; i < targetUpdates(runTime, ui.conf.CPUFrequencyHz); i++ {
			ui.vm.Step()
			cpuCycles++
		}
		for i := frames; i < targetUpdates(runTime, ui.conf.FrameRateHz); i++ {
			ui.display.Render(ui.vm, ui.conf)
			frames++
		}
		time.Sleep(time.Second / time.Duration(ui.conf.EmulatorFrequencyHz))
	}
}
