package opengl

import (
	"image"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/odsod/chip8"
)

type Options struct {
	RomFile          string
	CPUFrequencyHz   int
	TimerFrequencyHz int
	Scale            int
	PixelFadeTime    time.Duration
}

type UI struct {
	vm      *chip8.VM
	display *display
	opts    Options
}

func NewUI(opts Options) *UI {
	rom, err := ioutil.ReadFile(opts.RomFile)
	if err != nil {
		panic(err)
	}

	return &UI{
		vm:      chip8.New(rom),
		display: newDisplay(opts.PixelFadeTime),
		opts:    opts,
	}
}

func (ui *UI) Run() {
	// OpenGL needs to run on a single OS thread
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, 1)
	glfw.WindowHint(glfw.Decorated, 1)
	window, err := glfw.CreateWindow(chip8.ScreenWidth*ui.opts.Scale, chip8.ScreenHeight*ui.opts.Scale, "chip8", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}
	gl.Enable(gl.TEXTURE_2D)

	// Create texture for chip8 screen
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	rom, err := ioutil.ReadFile("roms/TETRIS")
	if err != nil {
		panic(err)
	}

	vm := chip8.New(rom)

	startTime := time.Now()
	timerCycles := 0
	cpuCycles := 0

	for !window.ShouldClose() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT)

		now := time.Now()
		runTime := now.Sub(startTime)

		vm.SetKeys(readKeys(window))

		for i := timerCycles; i < targetUpdates(runTime, 60); i++ {
			vm.TickTimers()
			timerCycles++
		}
		for i := cpuCycles; i < targetUpdates(runTime, 500); i++ {
			vm.Step()
			cpuCycles++
		}

		ui.display.update(now, vm)

		gl.BindTexture(gl.TEXTURE_2D, texture)

		setTexture(ui.display.buffer)
		drawBuffer(window)
		gl.BindTexture(gl.TEXTURE_2D, 0)
		window.SwapBuffers()
	}
}

func setTexture(im *image.RGBA) {
	size := im.Rect.Size()
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA, int32(size.X), int32(size.Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / chip8.ScreenWidth
	s2 := float32(h) / chip8.ScreenHeight
	f := float32(1)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

/*
Keyboard layout mapping to CHIP-8 keys.

	|1|2|3|4| -> |1|2|3|C|
	|'|,|.|p| -> |4|5|6|D|
	|a|o|e|u| -> |7|8|9|E|
	|;|q|j|k| -> |A|0|B|F|
*/
var keyMap = map[glfw.Key]uint8{
	glfw.Key1: 0x1, glfw.Key2: 0x2, glfw.Key3: 0x3, glfw.Key4: 0xC,
	glfw.KeyQ: 0x4, glfw.KeyW: 0x5, glfw.KeyE: 0x6, glfw.KeyR: 0xD,
	glfw.KeyA: 0x7, glfw.KeyS: 0x8, glfw.KeyD: 0x9, glfw.KeyF: 0xE,
	glfw.KeyZ: 0xA, glfw.KeyX: 0x0, glfw.KeyC: 0xB, glfw.KeyV: 0xF,
}

func readKeys(window *glfw.Window) [16]bool {
	var result [16]bool
	for glfwKey, chip8Key := range keyMap {
		result[chip8Key] = window.GetKey(glfwKey) == glfw.Press
	}
	return result
}

func targetUpdates(runTime time.Duration, updateFrequencyHz int) int {
	updateInterval := time.Second / time.Duration(updateFrequencyHz)
	return int(runTime / updateInterval)
}
