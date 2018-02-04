package main

import (
	"math/rand"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/odsod/chip8"
)

type KeyMap = map[rune]uint8

/*
|1|2|3|4| -> |1|2|3|C|
|'|,|.|p| -> |4|5|6|D|
|a|o|e|u| -> |7|8|9|E|
|;|q|j|k| -> |A|0|B|F|
*/
var Dvorak KeyMap = KeyMap{
	'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
	'\'': 0x4, ',': 0x5, '.': 0x6, 'p': 0xD,
	'a': 0x7, 'o': 0x8, 'e': 0x9, 'u': 0xE,
	';': 0xA, 'q': 0x0, 'j': 0xB, 'k': 0xF,
}

/*
|1|2|3|4| -> |1|2|3|C|
|q|w|e|r| -> |4|5|6|D|
|a|s|d|f| -> |7|8|9|E|
|z|x|c|v| -> |A|0|B|F|
*/
var QWER KeyMap = KeyMap{
	'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
	'q': 0x4, 'w': 0x5, 'e': 0x6, 'r': 0xD,
	'a': 0x7, 's': 0x8, 'd': 0x9, 'f': 0xE,
	'z': 0xA, 'x': 0x0, 'c': 0xB, 'v': 0xF,
}

func readTermboxInput(keyMap KeyMap) (keyChannel chan uint8, killChannel chan bool) {
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

func Hz(x int64) time.Duration {
	return time.Second / time.Duration(x)
}

func randomUint8() uint8 {
	return uint8(rand.Uint32())
}

func simulateKeyUpEvents(vm *chip8.VM, delay time.Duration) (keyDownFn, keyUpFn func(uint8), keyUpChannel chan uint8) {
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
	keyUpFn = func(key uint8) {
		vm.SetKeyUp(key)
	}
	return
}

func pixels(scanLine uint64) [64]bool {
	var result [64]bool
	var i uint
	for i = 0; i < 64; i++ {
		result[i] = scanLine&(0x8000000000000000>>i) > 0
	}
	return result
}

func renderTermbox(vm *chip8.VM) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, scanLine := range vm.VideoMemory {
		for x, pixel := range pixels(scanLine) {
			var bg termbox.Attribute
			if pixel {
				bg = termbox.ColorWhite
			} else {
				bg = termbox.ColorDefault
			}
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, bg)
		}
	}
	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	vm := chip8.New(randomUint8)
	keyDownChannel, killChannel := readTermboxInput(Dvorak)
	keyDownFn, keyUpFn, keyUpChannel := simulateKeyUpEvents(vm, 200*time.Millisecond)
	program := []uint8{
		0x00, 0xE0, // 00E0: CLS
		0x60, 0x00, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x01, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x02, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x03, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x04, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x05, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x06, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x07, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x08, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x00, 0xE0, // 00E0: CLS
		0x60, 0x09, // 6xkk: LD Vx, kk
		0xF0, 0x29, // Fx29: LD F, Vx
		0xD1, 0x25, // Dxyn: DRW Vx, Vy, n
		0xFF, 0x0A, // Fx0A: LD Vx, K

		0x12, 0x00, // 1nnn: JP nnn
	}
	copy(vm.Memory[0x200:0x200+len(program)], program)
	cpuTicks := time.NewTicker(Hz(700))
	timerTicks := time.NewTicker(Hz(60))
	videoRefreshes := time.NewTicker(Hz(30))
Loop:
	for {
		select {
		case key := <-keyDownChannel:
			keyDownFn(key)
		case key := <-keyUpChannel:
			keyUpFn(key)
		case <-cpuTicks.C:
			vm.Step()
		case <-timerTicks.C:
			vm.TickTimers()
		case <-videoRefreshes.C:
			renderTermbox(vm)
		case <-killChannel:
			break Loop
		}
	}
}
