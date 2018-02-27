package opengl

import "github.com/go-gl/glfw/v3.1/glfw"

/*
Keyboard layout mapping to CHIP-8 keys.

	|1|2|3|4| -> |1|2|3|C|
	|q|w|e|r| -> |4|5|6|D|
	|a|s|d|f| -> |7|8|9|E|
	|z|x|c|v| -> |A|0|B|F|
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
