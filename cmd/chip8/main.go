package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/odsod/chip8"
)

/*
Dvorak keyboard layout mapping to CHIP-8 keys.

	|1|2|3|4| -> |1|2|3|C|
	|'|,|.|p| -> |4|5|6|D|
	|a|o|e|u| -> |7|8|9|E|
	|;|q|j|k| -> |A|0|B|F|
*/
var Dvorak = map[rune]uint8{
	'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
	'\'': 0x4, ',': 0x5, '.': 0x6, 'p': 0xD,
	'a': 0x7, 'o': 0x8, 'e': 0x9, 'u': 0xE,
	';': 0xA, 'q': 0x0, 'j': 0xB, 'k': 0xF,
}

/*
QWER(TY|TZ) keyboard layout mapping to CHIP-8 keys.

	|1|2|3|4| -> |1|2|3|C|
	|q|w|e|r| -> |4|5|6|D|
	|a|s|d|f| -> |7|8|9|E|
	|z|x|c|v| -> |A|0|B|F|
*/
var QWER = map[rune]uint8{
	'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
	'q': 0x4, 'w': 0x5, 'e': 0x6, 'r': 0xD,
	'a': 0x7, 's': 0x8, 'd': 0x9, 'f': 0xE,
	'z': 0xA, 'x': 0x0, 'c': 0xB, 'v': 0xF,
}

func Hz(x int) time.Duration {
	return time.Second / time.Duration(x)
}

type Random struct{}

func (r Random) Next() uint8 {
	return uint8(rand.Uint32())
}

func renderTitle(x0, y0 int, s string) {
	for xi, c := range s {
		termbox.SetCell(x0+xi, y0, c, termbox.ColorWhite, termbox.ColorDefault)
	}
}

func renderBorder(x0, y0, w, h int, borderColor termbox.Attribute) {
	x1 := x0 + w
	y1 := y0 + h
	termbox.SetCell(x0, y0, '╔', borderColor, termbox.ColorDefault)
	termbox.SetCell(x0, y1, '╚', borderColor, termbox.ColorDefault)
	termbox.SetCell(x1, y0, '╗', borderColor, termbox.ColorDefault)
	termbox.SetCell(x1, y1, '╝', borderColor, termbox.ColorDefault)
	for x := x0 + 1; x < x1; x++ {
		termbox.SetCell(x, y0, '═', borderColor, termbox.ColorDefault)
		termbox.SetCell(x, y1, '═', borderColor, termbox.ColorDefault)
	}
	for y := y0 + 1; y < y1; y++ {
		termbox.SetCell(x0, y, '║', borderColor, termbox.ColorDefault)
		termbox.SetCell(x1, y, '║', borderColor, termbox.ColorDefault)
	}
}

func render(romName string, vm *chip8.VM) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	renderTitle(0, 0, "CHIP-8: "+romName)
	renderBorder(0, 1, 65, 33, termbox.ColorWhite)
	for y, scanLine := range vm.VideoMemory {
		for x := 0; x < 64; x++ {
			pixel := scanLine&(0x8000000000000000>>uint(x)) > 0
			var bg termbox.Attribute
			if pixel {
				bg = termbox.ColorWhite
			} else {
				bg = termbox.ColorDefault
			}
			termbox.SetCell(1+x, 2+y, ' ', termbox.ColorDefault, bg)
		}
	}
	termbox.Flush()
}

func readInputFromTermbox(keyMap map[rune]uint8) (
	keyChannel chan uint8, killChannel chan bool,
) {
	keyChannel = make(chan uint8)
	killChannel = make(chan bool)
	go func() {
	Loop:
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if key, ok := keyMap[ev.Ch]; ok {
					keyChannel <- key
				} else if ev.Key == termbox.KeyEsc {
					killChannel <- true
					break Loop
				}
			case termbox.EventError:
				panic(ev.Err)
			}
		}
	}()
	return
}

func simulateKeyUpEvents(vm *chip8.VM, delay time.Duration) (
	keyDownFn func(uint8), keyUpChannel chan uint8,
) {
	keyUpChannel = make(chan uint8)
	var keyUpTimers [16]*time.Timer
	for key := 0x0; key <= 0xf; key++ {
		key := key
		keyUpTimers[key] = time.AfterFunc(delay, func() {
			keyUpChannel <- uint8(key)
		})
		keyUpTimers[key].Stop()
	}
	keyDownFn = func(key uint8) {
		keyUpTimers[key].Reset(delay)
		vm.SetKeyDown(key)
	}
	return
}

func main() {
	romFile := flag.String("rom", "roms/TETRIS", "The ROM to load")
	isDvorak := flag.Bool("dvorak", false, "Use Dvorak key map")
	cpuFrequency := flag.Int("cpuFrequency", 500, "The CPU frequency (Hz)")
	timerFrequency := flag.Int("timerFrequency", 60, "The timer frequency (Hz)")
	frameRate := flag.Int("frameRate", 60, "The frame rate (Hz)")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	rom, err := ioutil.ReadFile(*romFile)
	if err != nil {
		panic(err)
	}

	vm := chip8.New(rom, Random{})

	keyMap := QWER
	if *isDvorak {
		keyMap = Dvorak
	}
	keyDownChannel, killChannel := readInputFromTermbox(keyMap)
	keyDownFn, keyUpChannel := simulateKeyUpEvents(vm, 100*time.Millisecond)

	startTime := time.Now()
	timerCycles := 0
	cpuCycles := 0
	frames := 0

Loop:
	for {
		select {
		case key := <-keyDownChannel:
			keyDownFn(key)
		case key := <-keyUpChannel:
			vm.SetKeyUp(key)
		case <-killChannel:
			break Loop
		default:
			// do nothing
		}
		runTime := time.Now().Sub(startTime)
		for i := timerCycles; i < int(runTime/Hz(*timerFrequency)); i++ {
			vm.TickTimers()
			timerCycles++
		}
		for i := cpuCycles; i < int(runTime/Hz(*cpuFrequency)); i++ {
			vm.Step()
			cpuCycles++
		}
		for i := frames; i < int(runTime/Hz(*frameRate)); i++ {
			render(*romFile, vm)
			frames++
		}
		time.Sleep(10 * time.Millisecond)
	}
}
