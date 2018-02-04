package chip8

import (
	"fmt"
)

type VM struct {
	// SP is the stack pointer
	SP uint8

	// I is a pointer register
	I uint16

	// PC is the program counter
	PC uint16

	// V are the 16 general purpose registers
	V [16]uint8

	// DT is the delay timer register
	DT uint8

	// DT is the sound timer register
	ST uint8

	// Memory contains the default sprites, the ROM and the RAM
	Memory [4096]uint8

	// Stack holds up to 16 memory locations
	Stack [16]uint16

	// Keys are a list of flags (0x0 - 0xF) signfying if a key is held down or not
	Keys [16]bool

	// VideoMemory represents the 64x32 pixel screen as 64-bit scan lines
	VideoMemory [32]uint64

	// K points to the register waiting for a keypress
	K *uint8

	// random provides a random byte value
	random func() uint8
}

const (
	digitStartAddress = 0
	digitSpriteSize   = 5
)

var digitSprites = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func New(random func() uint8) *VM {
	vm := VM{}
	vm.random = random
	vm.PC = 0x200
	copy(vm.Memory[0:len(digitSprites)], digitSprites)
	return &vm
}

func (vm *VM) Step() {
	if vm.K != nil {
		// execution is suspended until next keyboard input
		return
	}
	encodedOp := vm.fetch()
	op := encodedOp.decode()
	op.execute(vm)
}

func (vm *VM) SetKeyDown(key uint8) {
	if key > 0xff {
		panic(fmt.Sprintf("Unsupported key: %#x", key))
	}
	vm.Keys[key] = true
	if vm.K != nil {
		// load the
		*vm.K = key
		vm.K = nil
	}
}

func (vm *VM) SetKeyUp(key uint8) {
	if key > 0xff {
		panic(fmt.Sprintf("Unsupported key: %#x", key))
	}
	vm.Keys[key] = false
}

func (vm *VM) TickTimers() {
	if vm.DT > 0 {
		vm.DT--
	}
	if vm.ST > 0 {
		vm.ST--
	}
}

func (vm *VM) readUint16(addr uint16) uint16 {
	return (uint16(vm.Memory[addr]) << 8) | uint16(vm.Memory[addr+1])
}

func (vm *VM) fetch() EncodedOp {
	op := EncodedOp(vm.readUint16(vm.PC))
	vm.PC += 2
	return op
}

type Op interface {
	execute(*VM)
}

type EncodedOp uint16

// .nnn
func (op EncodedOp) nnn() uint16 {
	return uint16(op & 0x0FFF)
}

// ..kk
func (op EncodedOp) kk() uint8 {
	return uint8(op & 0x00FF)
}

// ...n
func (op EncodedOp) n() uint8 {
	return uint8(op & 0x000F)
}

// .x..
func (op EncodedOp) x() uint8 {
	return uint8((op & 0x0F00) >> 8)
}

// ..y.
func (op EncodedOp) y() uint8 {
	return uint8((op & 0x00F0) >> 4)
}

func (op EncodedOp) decode() Op {
	switch op >> 12 {
	case 0x0:
		switch op {
		case 0x00E0:
			return op.decodeCLS()
		case 0x00EE:
			return op.decodeRET()
		}
	case 0x1:
		return op.decodeJP()
	case 0x2:
		return op.decodeCALL()
	case 0x3:
		return op.decodeSEVx()
	case 0x4:
		return op.decodeSNEVx()
	case 0x5:
		return op.decodeSEVxVy()
	case 0x6:
		return op.decodeLDVx()
	case 0x7:
		return op.decodeADDVx()
	case 0x8:
		switch op & 0x000F {
		case 0x0:
			return op.decodeLDVxVy()
		case 0x1:
			return op.decodeORVxVy()
		case 0x2:
			return op.decodeANDVxVy()
		case 0x3:
			return op.decodeXORVxVy()
		case 0x4:
			return op.decodeADDVxVy()
		case 0x5:
			return op.decodeSUBVxVy()
		case 0x6:
			return op.decodeSHRVx()
		case 0x7:
			return op.decodeSUBNVxVy()
		case 0xE:
			return op.decodeSHLVx()
		}
	case 0x9:
		return op.decodeSNEVxVy()
	case 0xA:
		return op.decodeLDI()
	case 0xB:
		return op.decodeJPV0()
	case 0xC:
		return op.decodeRNDVx()
	case 0xD:
		return op.decodeDRWVxVy()
	case 0xE:
		switch op & 0x00FF {
		case 0x9E:
			return op.decodeSKPVx()
		case 0xA1:
			return op.decodeSKNPVx()
		}
	case 0xF:
		switch op & 0x00FF {
		case 0x07:
			return op.decodeLDVxDT()
		case 0x0A:
			return op.decodeLDVxK()
		case 0x15:
			return op.decodeLDDTVx()
		case 0x18:
			return op.decodeLDSTVx()
		case 0x29:
			return op.decodeLDFVx()
		case 0x33:
			return op.decodeLDBVx()
		case 0x55:
			return op.decodeLDIVx()
		}
	}
	panic(fmt.Sprintf("Unsupported op: %#X", op))
}

/*
00E0 - CLS

Clear the display.
*/
type CLS struct{}

func (op EncodedOp) decodeCLS() CLS {
	return CLS{}
}

func (op CLS) execute(vm *VM) {
	for i := range vm.VideoMemory {
		vm.VideoMemory[i] = 0
	}
}

/*
00EE - RET

Return from a subroutine.

The interpreter sets the program counter to the address at the top of the
stack, then subtracts 1 from the stack pointer.
*/
type RET struct{}

func (op EncodedOp) decodeRET() RET {
	return RET{}
}

func (op RET) execute(vm *VM) {
	vm.PC = vm.Stack[vm.SP]
	vm.SP--
}

/*
1nnn - JP addr

Jump to location nnn.

The interpreter sets the program counter to nnn.
*/
type JP struct {
	nnn uint16
}

func (op EncodedOp) decodeJP() JP {
	return JP{op.nnn()}
}

func (op JP) execute(vm *VM) {
	vm.PC = op.nnn
}

/*
2nnn - CALL addr

Call subroutine at nnn.

The interpreter increments the stack pointer, then puts the current PC on the
top of the stack. The PC is then set to nnn.
*/
type CALL struct {
	nnn uint16
}

func (op EncodedOp) decodeCALL() CALL {
	return CALL{op.nnn()}
}

func (op CALL) execute(vm *VM) {
	if vm.SP >= uint8(len(vm.Stack)) {
		panic("Stack overflow")
	}
	vm.Stack[vm.SP] = vm.PC
	vm.SP += 1
	vm.PC = op.nnn
}

/*
3xkk - SE Vx, byte

Skip next instruction if Vx = kk.

The interpreter compares register Vx to kk, and if they are equal, increments
the program counter by 2.
*/
type SEVx struct {
	x, kk uint8
}

func (op EncodedOp) decodeSEVx() SEVx {
	return SEVx{op.x(), op.kk()}
}

func (op SEVx) execute(vm *VM) {
	if vm.V[op.x] == op.kk {
		vm.PC += 2
	}
}

/*
4xkk - SNE Vx, byte

Skip next instruction if Vx != kk.

The interpreter compares register Vx to kk, and if they are not equal,
increments the program counter by 2.
*/
type SNEVx struct {
	x, kk uint8
}

func (op EncodedOp) decodeSNEVx() SNEVx {
	return SNEVx{op.x(), op.kk()}
}

func (op SNEVx) execute(vm *VM) {
	if vm.V[op.x] != op.kk {
		vm.PC += 2
	}
}

/*
5xy0 - SE Vx, Vy

Skip next instruction if Vx = Vy.

The interpreter compares register Vx to register Vy, and if they are equal,
increments the program counter by 2.
*/
type SEVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeSEVxVy() SEVxVy {
	return SEVxVy{op.x(), op.y()}
}

func (op SEVxVy) execute(vm *VM) {
	if vm.V[op.x] == vm.V[op.y] {
		vm.PC += 2
	}
}

/*
6xkk - LD Vx, byte

Set Vx = kk.

The interpreter puts the value kk into register Vx.
*/
type LDVx struct {
	x, kk uint8
}

func (op EncodedOp) decodeLDVx() LDVx {
	return LDVx{op.x(), op.kk()}
}

func (op LDVx) execute(vm *VM) {
	vm.V[op.x] = op.kk
}

/*
7xkk - ADD Vx, byte

Set Vx = Vx + kk.

Adds the value kk to the value of register Vx, then stores the result in Vx.
*/
type ADDVx struct {
	x, kk uint8
}

func (op EncodedOp) decodeADDVx() ADDVx {
	return ADDVx{op.x(), op.kk()}
}

func (op ADDVx) execute(vm *VM) {
	vm.V[op.x] += op.kk
}

/*
8xy0 - LD Vx, Vy

Set Vx = Vy.

Stores the value of register Vy in register Vx.
*/
type LDVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeLDVxVy() LDVxVy {
	return LDVxVy{op.x(), op.y()}
}

func (op LDVxVy) execute(vm *VM) {
	vm.V[op.x] = vm.V[op.y]
}

/*
8xy1 - OR Vx, Vy

Set Vx = Vx OR Vy.

Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
A bitwise OR compares the corrseponding bits from two values, and if either bit
is 1, then the same bit in the result is also 1. Otherwise, it is 0.
*/
type ORVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeORVxVy() ORVxVy {
	return ORVxVy{op.x(), op.y()}
}

func (op ORVxVy) execute(vm *VM) {
	vm.V[op.x] = vm.V[op.x] | vm.V[op.y]
}

/*
8xy2 - AND Vx, Vy

Set Vx = Vx AND Vy.

Performs a bitwise AND on the values of Vx and Vy, then stores the result in
Vx. A bitwise AND compares the corrseponding bits from two values, and if both
bits are 1, then the same bit in the result is also 1. Otherwise, it is 0.
*/
type ANDVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeANDVxVy() ANDVxVy {
	return ANDVxVy{op.x(), op.y()}
}

func (op ANDVxVy) execute(vm *VM) {
	vm.V[op.x] = vm.V[op.x] & vm.V[op.y]
}

/*
8xy3 - XOR Vx, Vy

Set Vx = Vx XOR Vy.

Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the
result in Vx. An exclusive OR compares the corrseponding bits from two values,
and if the bits are not both the same, then the corresponding bit in the result
is set to 1. Otherwise, it is 0.
*/
type XORVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeXORVxVy() XORVxVy {
	return XORVxVy{op.x(), op.y()}
}

func (op XORVxVy) execute(vm *VM) {
	vm.V[op.x] = vm.V[op.x] ^ vm.V[op.y]
}

/*
8xy4 - ADD Vx, Vy

Set Vx = Vx + Vy, set VF = carry.

The values of Vx and Vy are added together. If the result is greater than 8
bits (i.e., > 255,) VF is set to 1, otherwise 0. Only the lowest 8 bits of the
result are kept, and stored in Vx.
*/
type ADDVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeADDVxVy() ADDVxVy {
	return ADDVxVy{op.x(), op.y()}
}

func (op ADDVxVy) execute(vm *VM) {
	sum := uint16(vm.V[op.x]) + uint16(vm.V[op.y])
	if sum > 255 {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
	vm.V[op.x] = uint8(sum)
}

/*
8xy5 - SUB Vx, Vy

Set Vx = Vx - Vy, set VF = NOT borrow.

If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
and the results stored in Vx.
*/
type SUBVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeSUBVxVy() SUBVxVy {
	return SUBVxVy{op.x(), op.y()}
}

func (op SUBVxVy) execute(vm *VM) {
	if vm.V[op.x] > vm.V[op.y] {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
	vm.V[op.x] = vm.V[op.x] - vm.V[op.y]
}

/*
8xy6 - SHR Vx {, Vy}

Set Vx = Vx SHR 1.

If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0. Then
Vx is divided by 2.
*/
type SHRVx struct {
	x uint8
}

func (op EncodedOp) decodeSHRVx() SHRVx {
	return SHRVx{op.x()}
}

func (op SHRVx) execute(vm *VM) {
	if vm.V[op.x]&0x01 == 1 {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
	vm.V[op.x] >>= 1
}

/*
8xy7 - SUBN Vx, Vy

Set Vx = Vy - Vx, set VF = NOT borrow.

If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
and the results stored in Vx.
*/
type SUBNVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeSUBNVxVy() SUBNVxVy {
	return SUBNVxVy{op.x(), op.y()}
}

func (op SUBNVxVy) execute(vm *VM) {
	if vm.V[op.y] > vm.V[op.x] {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
	vm.V[op.x] = vm.V[op.y] - vm.V[op.x]
}

/*
8xyE - SHL Vx {, Vy}

Set Vx = Vx SHL 1.

If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
Then Vx is multiplied by 2.
*/
type SHLVx struct {
	x uint8
}

func (op EncodedOp) decodeSHLVx() SHLVx {
	return SHLVx{op.x()}
}

func (op SHLVx) execute(vm *VM) {
	if vm.V[op.x]&0x80 == 1 {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
	vm.V[op.x] <<= 1
}

/*
9xy0 - SNE Vx, Vy

Skip next instruction if Vx != Vy.

The values of Vx and Vy are compared, and if they are not equal, the program
counter is increased by 2.
*/
type SNEVxVy struct {
	x, y uint8
}

func (op EncodedOp) decodeSNEVxVy() SNEVxVy {
	return SNEVxVy{op.x(), op.y()}
}

func (op SNEVxVy) execute(vm *VM) {
	if vm.V[op.x] != vm.V[op.y] {
		vm.PC += 2
	}
}

/*
Annn - LD I, addr

Set I = nnn.

The value of register I is set to nnn.
*/
type LDI struct {
	nnn uint16
}

func (op EncodedOp) decodeLDI() LDI {
	return LDI{op.nnn()}
}

func (op LDI) execute(vm *VM) {
	vm.I = op.nnn
}

/*
Bnnn - JP V0, addr

Jump to location nnn + V0.

The program counter is set to nnn plus the value of V0.
*/
type JPV0 struct {
	nnn uint16
}

func (op EncodedOp) decodeJPV0() JPV0 {
	return JPV0{op.nnn()}
}

func (op JPV0) execute(vm *VM) {
	vm.PC = op.nnn + uint16(vm.V[0])
}

/*
Cxkk - RND Vx, byte

Set Vx = random byte AND kk.

The interpreter generates a random number from 0 to 255, which is then ANDed
with the value kk. The results are stored in Vx. See instruction 8xy2 for more
information on AND.
*/
type RNDVx struct {
	x, kk uint8
}

func (op EncodedOp) decodeRNDVx() RNDVx {
	return RNDVx{op.x(), op.kk()}
}

func (op RNDVx) execute(vm *VM) {
	vm.V[op.x] = vm.random() & op.kk
}

/*
Dxyn - DRW Vx, Vy, nibble

Display n-byte sprite starting at memory location I at (Vx, Vy), set VF =
collision.

The interpreter reads n bytes from memory, starting at the address stored in I.
These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
Sprites are XORed onto the existing screen. If this causes any pixels to be
erased, VF is set to 1, otherwise it is set to 0. If the sprite is positioned
so part of it is outside the coordinates of the display, it wraps around to the
opposite side of the screen. See instruction 8xy3 for more information on XOR,
and section 2.4, Display, for more information on the Chip-8 screen and
sprites.
*/
type DRWVxVy struct {
	x, y, n uint8
}

func (op EncodedOp) decodeDRWVxVy() DRWVxVy {
	return DRWVxVy{op.x(), op.y(), op.n()}
}

func expandSpriteRowToScanLine(spriteRow, x uint8) uint64 {
	if x <= 56 {
		// shift sprite left
		return uint64(spriteRow) << (56 - x)
	} else {
		// shift sprite right
		return uint64(spriteRow) >> (x - 56)
	}
}

func (op DRWVxVy) execute(vm *VM) {
	collision := false
	x0 := vm.V[op.x]
	y0 := vm.V[op.y]
	sprite := vm.Memory[vm.I : vm.I+uint16(op.n)]
	for yi, spriteRow := range sprite {
		y := y0 + uint8(yi)
		oldScanLine := vm.VideoMemory[y]
		newScanLine := oldScanLine ^ expandSpriteRowToScanLine(spriteRow, x0)
		vm.VideoMemory[y] = newScanLine
		if (oldScanLine & ^newScanLine) > 0 {
			collision = true
		}
	}
	if collision {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
}

/*
Ex9E - SKP Vx

Skip next instruction if key with the value of Vx is pressed.

Checks the keyboard, and if the key corresponding to the value of Vx is
currently in the down position, PC is increased by 2.
*/
type SKPVx struct {
	x uint8
}

func (op EncodedOp) decodeSKPVx() SKPVx {
	return SKPVx{op.x()}
}

func (op SKPVx) execute(vm *VM) {
	if vm.Keys[vm.V[op.x]] {
		vm.PC += 2
	}
}

/*
ExA1 - SKNP Vx

Skip next instruction if key with the value of Vx is not pressed.

Checks the keyboard, and if the key corresponding to the value of Vx is
currently in the up position, PC is increased by 2.
*/
type SKNPVx struct {
	x uint8
}

func (op EncodedOp) decodeSKNPVx() SKNPVx {
	return SKNPVx{op.x()}
}

func (op SKNPVx) execute(vm *VM) {
	if !vm.Keys[vm.V[op.x]] {
		vm.PC += 2
	}
}

/*
Fx07 - LD Vx, DT

Set Vx = delay timer value.

The value of DT is placed into Vx.
*/
type LDVxDT struct {
	x uint8
}

func (op EncodedOp) decodeLDVxDT() LDVxDT {
	return LDVxDT{op.x()}
}

func (op LDVxDT) execute(vm *VM) {
	vm.V[op.x] = vm.DT
}

/*
Fx0A - LD Vx, K

Wait for a key press, store the value of the key in Vx.

All execution stops until a key is pressed, then the value of that key is
stored in Vx.
*/
type LDVxK struct {
	x uint8
}

func (op EncodedOp) decodeLDVxK() LDVxK {
	return LDVxK{op.x()}
}

func (op LDVxK) execute(vm *VM) {
	vm.K = &vm.V[op.x]
}

/*
Fx15 - LD DT, Vx

Set delay timer = Vx.

DT is set equal to the value of Vx.
*/
type LDDTVx struct {
	x uint8
}

func (op EncodedOp) decodeLDDTVx() LDDTVx {
	return LDDTVx{op.x()}
}

func (op LDDTVx) execute(vm *VM) {
	vm.DT = vm.V[op.x]
}

/*
Fx18 - LD ST, Vx

Set sound timer = Vx.

ST is set equal to the value of Vx.
*/
type LDSTVx struct {
	x uint8
}

func (op EncodedOp) decodeLDSTVx() LDSTVx {
	return LDSTVx{op.x()}
}

func (op LDSTVx) execute(vm *VM) {
	vm.ST = vm.V[op.x]
}

/*
Fx1E - ADD I, Vx

Set I = I + Vx.

The values of I and Vx are added, and the results are stored in I.
*/
type ADDIVx struct {
	x uint8
}

func (op EncodedOp) decodeADDIVx() ADDIVx {
	return ADDIVx{op.x()}
}

func (op ADDIVx) execute(vm *VM) {
	vm.I += uint16(vm.V[op.x])
}

/*
Fx29 - LD F, Vx

Set I = location of sprite for digit Vx.

The value of I is set to the location for the hexadecimal sprite corresponding
to the value of Vx. See section 2.4, Display, for more information on the
Chip-8 hexadecimal font.
*/
type LDFVx struct {
	x uint8
}

func (op EncodedOp) decodeLDFVx() LDFVx {
	return LDFVx{op.x()}
}

func (op LDFVx) execute(vm *VM) {
	vm.I = digitStartAddress + digitSpriteSize*uint16(vm.V[op.x])
}

/*
Fx33 - LD B, Vx

Store BCD representation of Vx in memory locations I, I+1, and I+2.

The interpreter takes the decimal value of Vx, and places the hundreds digit in
memory at location in I, the tens digit at location I+1, and the ones digit at
location I+2.
*/
type LDBVx struct {
	x uint8
}

func (op EncodedOp) decodeLDBVx() LDBVx {
	return LDBVx{op.x()}
}

func bcd(n uint8) (hundreds, tens, ones uint8) {
	hundreds = n / 100
	tens = (n - hundreds*100) / 10
	ones = n % 10
	return
}

func (op LDBVx) execute(vm *VM) {
	vm.Memory[vm.I], vm.Memory[vm.I+1], vm.Memory[vm.I+2] = bcd(vm.V[op.x])
}

/*
Fx55 - LD [I], Vx

Store registers V0 through Vx in memory starting at location I.

The interpreter copies the values of registers V0 through Vx into memory,
starting at the address in I.
*/
type LDIVx struct {
	x uint8
}

func (op EncodedOp) decodeLDIVx() LDIVx {
	return LDIVx{op.x()}
}

func (op LDIVx) execute(vm *VM) {
	for i := 0; i < int(op.x); i++ {
		vm.Memory[vm.I+uint16(i)] = vm.V[i]
	}
}

/*
Fx65 - LD Vx, [I]

Read registers V0 through Vx from memory starting at location I.

The interpreter reads values from memory starting at location I into registers
V0 through Vx.
*/
type LDVxI struct {
	x uint8
}

func (op LDVxI) execute(vm *VM) {
	for i := 0; i < int(op.x); i++ {
		vm.V[i] = vm.Memory[vm.I+uint16(i)]
	}
}
