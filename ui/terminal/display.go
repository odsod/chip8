package terminal

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/odsod/chip8"
)

type Display struct {
}

func NewDisplay() *Display {
	return &Display{}
}

func (display *Display) Render(vm *chip8.VM, conf Conf) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	renderTitle(0, 0, "CHIP-8: "+conf.RomFile)
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
