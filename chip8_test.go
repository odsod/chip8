package chip8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	vm := VM{}
	vm.PC = 0x0200
	call := CALL{0xCAFE}
	call.execute(&vm)
	assert.Equal(t, vm.SP, uint8(1))
	assert.Equal(t, call.nnn, vm.PC)
	assert.Equal(t, uint16(0x200), vm.Stack[0])
}
