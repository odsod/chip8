package terminal

import (
	"time"

	termbox "github.com/nsf/termbox-go"
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

type Keyboard struct {
	keyMap       map[rune]uint8
	keyUpDelay   time.Duration
	keyUpTimes   [16]time.Time
	eventChannel chan termbox.Event
}

func NewKeyboard(keyUpDelay time.Duration, keyMap map[rune]uint8) *Keyboard {
	return &Keyboard{
		keyUpDelay:   keyUpDelay,
		keyMap:       keyMap,
		eventChannel: make(chan termbox.Event),
	}
}

func (kb *Keyboard) Listen() {
	go func() {
		for {
			kb.eventChannel <- termbox.PollEvent()
		}
	}()
}

// Check the keyboard state every emulation cycle for which keys are pressed
func (kb *Keyboard) Check(now time.Time) (keys [16]bool, exit bool) {
	select {
	case ev := <-kb.eventChannel:
		if key, ok := kb.keyMap[ev.Ch]; ok {
			kb.keyUpTimes[key] = now.Add(kb.keyUpDelay)
		} else if ev.Key == termbox.KeyEsc {
			exit = true
		}
	default: // no events
	}
	for key := 0; key <= 0xf; key++ {
		keys[key] = kb.keyUpTimes[key].After(now)
	}
	return
}
