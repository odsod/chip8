package opengl

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/odsod/chip8"
)

type display struct {
	pixelFadeTime time.Duration
	pixelLastLit  [chip8.ScreenWidth][chip8.ScreenHeight]time.Time
	buffer        *image.RGBA
}

func newDisplay(pixelFadeTime time.Duration) *display {
	return &display{
		pixelFadeTime: pixelFadeTime,
		buffer:        image.NewRGBA(image.Rect(0, 0, chip8.ScreenWidth, chip8.ScreenHeight)),
	}
}

func pixelColor(now, lastLit time.Time, fade time.Duration) color.RGBA {
	timeSinceLit := now.Sub(lastLit)
	if timeSinceLit >= fade {
		return color.RGBA{255, 255, 255, 255}
	}
	lightPercent := float64(timeSinceLit) / float64(fade)
	alpha := uint8(math.Round(lightPercent * 255))
	return color.RGBA{alpha, alpha, alpha, 255}
}

func (d *display) update(now time.Time, vm *chip8.VM) {
	for y, scanLine := range vm.VideoMemory {
		for x := 0; x < 64; x++ {
			pixel := scanLine&(0x8000000000000000>>uint(x)) > 0
			if pixel {
				d.pixelLastLit[x][y] = now
			}
			d.buffer.SetRGBA(x, y, pixelColor(now, d.pixelLastLit[x][y], d.pixelFadeTime))
		}
	}
}
