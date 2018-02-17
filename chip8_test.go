package chip8

import (
	"testing"
)

func TestDecodeOps(t *testing.T) {
	for _, testCase := range []struct {
		op       EncodedOp
		expected Op // nil means decoding should panic
	}{
		{0x0000, nil},
		{0x00E0, CLS{}},
		{0x01E0, nil},
		{0x00EE, RET{}},
		{0x01EE, nil},
		{0x1123, JP{0x123}},
		{0x2123, CALL{0x123}},
		{0x3123, SEVx{x: 0x1, kk: 0x23}},
		{0x4123, SNEVx{x: 0x1, kk: 0x23}},
		{0x5120, SEVxVy{x: 0x1, y: 0x2}},
		{0x5121, nil},
		{0x6123, LDVx{x: 0x1, kk: 0x23}},
		{0x7123, ADDVx{x: 0x1, kk: 0x23}},
		{0x8120, LDVxVy{x: 0x1, y: 0x2}},
		{0x8121, ORVxVy{x: 0x1, y: 0x2}},
		{0x8122, ANDVxVy{x: 0x1, y: 0x2}},
		{0x8123, XORVxVy{x: 0x1, y: 0x2}},
		{0x8124, ADDVxVy{x: 0x1, y: 0x2}},
		{0x8125, SUBVxVy{x: 0x1, y: 0x2}},
		{0x8126, SHRVx{x: 0x1}},
		{0x8136, SHRVx{x: 0x1}},
		{0x8127, SUBNVxVy{x: 0x1, y: 0x2}},
		{0x8128, nil},
		{0x812E, SHLVx{x: 0x1}},
		{0x813E, SHLVx{x: 0x1}},
		{0x812F, nil},
		{0x9120, SNEVxVy{x: 0x1, y: 0x2}},
		{0x9121, nil},
		{0xA123, LDI{nnn: 0x123}},
		{0xB123, JPV0{nnn: 0x123}},
		{0xC123, RNDVx{x: 0x1, kk: 0x23}},
		{0xD123, DRWVxVy{x: 0x1, y: 0x2, n: 0x3}},
		{0xE19D, nil},
		{0xE18E, nil},
		{0xE19E, SKPVx{x: 0x1}},
		{0xE1A0, nil},
		{0xE1A1, SKNPVx{x: 0x1}},
		{0xE1B1, nil},
		{0xF106, nil},
		{0xF107, LDVxDT{x: 0x1}},
		{0xF108, nil},
		{0xF10A, LDVxK{x: 0x1}},
		{0xF114, nil},
		{0xF115, LDDTVx{x: 0x1}},
		{0xF116, nil},
		{0xF118, LDSTVx{x: 0x1}},
		{0xF119, nil},
		{0xF11E, ADDIVx{x: 0x1}},
		{0xF11F, nil},
		{0xF128, nil},
		{0xF129, LDFVx{x: 0x1}},
		{0xF12A, nil},
		{0xF132, nil},
		{0xF133, LDBVx{x: 0x1}},
		{0xF134, nil},
		{0xF154, nil},
		{0xF155, LDIVx{x: 0x1}},
		{0xF156, nil},
		{0xF164, nil},
		{0xF165, LDVxI{x: 0x1}},
		{0xF166, nil},
		{0xFFFF, nil},
	} {
		if testCase.expected == nil {
			// assert that decode() panics
			func() {
				defer func() { recover() }()
				actual := testCase.op.decode()
				t.Errorf(
					"(%#x).decode(): Should panic, Actual %#v",
					testCase.op, actual)
			}()
		} else {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf(
							"(%#x).decode(): Paniced, Expected %s",
							testCase.op, testCase.expected)
					}
				}()
				var actual Op = testCase.op.decode()
				if testCase.expected != actual {
					t.Errorf(
						"(%#x).decode(): Expected %#v, Actual %#v",
						testCase.op, testCase.expected, actual)
				}
			}()
		}
	}
}

func TestBCD(t *testing.T) {
	for _, testCase := range []struct {
		n, hundreds, tens, ones uint8
	}{
		{0, 0, 0, 0},
		{1, 0, 0, 1},
		{10, 0, 1, 0},
		{11, 0, 1, 1},
		{100, 1, 0, 0},
		{102, 1, 0, 2},
		{123, 1, 2, 3},
		{200, 2, 0, 0},
		{202, 2, 0, 2},
		{242, 2, 4, 2},
	} {
		hundreds, tens, ones := bcd(testCase.n)
		if hundreds != testCase.hundreds || tens != testCase.tens || ones != testCase.ones {
			t.Errorf(
				"bcd(%d) Expected: %d %d %d Actual: %d %d %d",
				testCase.n, testCase.hundreds, testCase.tens, testCase.ones,
				hundreds, tens, ones)
		}
	}
}

type Constant struct {
	c uint8
}

func (c Constant) Next() uint8 {
	return c.c
}

func TestOps(t *testing.T) {
	for _, testCase := range []struct {
		before VM
		op     Op
		after  VM
		msg    string
	}{
		/*
			00E0 - CLS
			Clear the display.
		*/
		{
			before: VM{VideoMemory: [32]uint64{0: 0x1, 31: 0x1}},
			op:     CLS{},
			after:  VM{},
		},

		/*
			00EE - RET
			Return from a subroutine.

			The interpreter sets the program counter to the address at the top of the
			stack, then subtracts 1 from the stack pointer.
		*/
		{
			before: VM{PC: 0x300, SP: 1, Stack: [16]uint16{0: 0x200}},
			op:     RET{},
			after:  VM{PC: 0x200, SP: 0, Stack: [16]uint16{0: 0x200}},
		},

		/*
			1nnn - JP addr
			Jump to location nnn.

			The interpreter sets the program counter to nnn.
		*/
		{
			before: VM{PC: 0x300},
			op:     JP{nnn: 0x400},
			after:  VM{PC: 0x400},
		},

		/*
			2nnn - CALL addr
			Call subroutine at nnn.

			The interpreter increments the stack pointer, then puts the current PC on
			the top of the stack. The PC is then set to nnn.
		*/
		{
			before: VM{PC: 0x300},
			op:     CALL{nnn: 0x400},
			after:  VM{PC: 0x400, SP: 1, Stack: [16]uint16{0: 0x300}},
		},

		/*
			3xkk - SE Vx, byte
			Skip next instruction if Vx = kk.

			The interpreter compares register Vx to kk, and if they are equal,
			increments the program counter by 2.
		*/
		{
			msg:    "should skip",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x12}},
			op:     SEVx{x: 0xA, kk: 0x12},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0x12}},
		},
		{
			msg:    "should not skip",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x12}},
			op:     SEVx{x: 0xB, kk: 0x12},
			after:  VM{PC: 0x200, V: [16]uint8{0xA: 0x12}},
		},

		/*
			4xkk - SNE Vx, byte
			Skip next instruction if Vx != kk.

			The interpreter compares register Vx to kk, and if they are not equal,
			increments the program counter by 2.
		*/
		{
			msg:    "should skip",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x13}},
			op:     SNEVx{x: 0xA, kk: 0x12},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0x13}},
		},
		{
			msg:    "should not skip",
			before: VM{PC: 0x200, V: [16]uint8{0xB: 0x12}},
			op:     SNEVx{x: 0xB, kk: 0x12},
			after:  VM{PC: 0x200, V: [16]uint8{0xB: 0x12}},
		},

		/*
			5xy0 - SE Vx, Vy
			Skip next instruction if Vx = Vy.

			The interpreter compares register Vx to register Vy, and if they are
			equal, increments the program counter by 2.
		*/
		{
			msg:    "should skip",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
			op:     SEVxVy{x: 0xA, y: 0xB},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
		},
		{
			msg:    "should not skip",
			before: VM{PC: 0x200, V: [16]uint8{0xC: 0x1, 0xD: 0x2}},
			op:     SEVxVy{x: 0xC, y: 0xD},
			after:  VM{PC: 0x200, V: [16]uint8{0xC: 0x1, 0xD: 0x2}},
		},

		/*
			6xkk - LD Vx, byte
			Set Vx = kk.

			The interpreter puts the value kk into register Vx.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x1}},
			op:     LDVx{x: 0xA, kk: 0xB},
			after:  VM{V: [16]uint8{0xA: 0xB}},
		},

		/*
			7xkk - ADD Vx, byte
			Set Vx = Vx + kk.

			Adds the value kk to the value of register Vx, then stores the result in
			Vx.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x1}},
			op:     ADDVx{x: 0xA, kk: 0x1},
			after:  VM{V: [16]uint8{0xA: 0x2}},
		},

		/*
			8xy0 - LD Vx, Vy
			Set Vx = Vy.

			Stores the value of register Vy in register Vx.
		*/
		{
			before: VM{V: [16]uint8{0xB: 0x1}},
			op:     LDVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
		},

		/*
			8xy1 - OR Vx, Vy
			Set Vx = Vx OR Vy.

			Performs a bitwise OR on the values of Vx and Vy, then stores the result
			in Vx.  A bitwise OR compares the corrseponding bits from two values, and
			if either bit is 1, then the same bit in the result is also 1.
			Otherwise, it is 0.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x0, 0xB: 0x1}},
			op:     ORVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
		},

		/*
			8xy2 - AND Vx, Vy
			Set Vx = Vx AND Vy.

			Performs a bitwise AND on the values of Vx and Vy, then stores the result
			in Vx. A bitwise AND compares the corrseponding bits from two values, and
			if both bits are 1, then the same bit in the result is also 1. Otherwise,
			it is 0.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x1, 0xB: 0x0}},
			op:     ANDVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xB: 0x0}},
		},

		/*
			8xy3 - XOR Vx, Vy
			Set Vx = Vx XOR Vy.

			Performs a bitwise exclusive OR on the values of Vx and Vy, then stores
			the result in Vx. An exclusive OR compares the corrseponding bits from
			two values, and if the bits are not both the same, then the corresponding
			bit in the result is set to 1. Otherwise, it is 0.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
			op:     XORVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xB: 0x1}},
		},

		/*
			8xy4 - ADD Vx, Vy
			Set Vx = Vx + Vy, set VF = carry.

			The values of Vx and Vy are added together. If the result is greater than
			8 bits (i.e., > 255,) VF is set to 1, otherwise 0. Only the lowest 8 bits
			of the result are kept, and stored in Vx.
		*/
		{
			msg:    "no carry",
			before: VM{V: [16]uint8{0xA: 0x1, 0xB: 0x1, 0xF: 0}},
			op:     ADDVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x2, 0xB: 0x1, 0xF: 0}},
		},
		{
			msg:    "carry",
			before: VM{V: [16]uint8{0xA: 0xFF, 0xB: 0x2, 0xF: 0}},
			op:     ADDVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x1, 0xB: 0x2, 0xF: 1}},
		},

		/*
			8xy5 - SUB Vx, Vy
			Set Vx = Vx - Vy, set VF = NOT borrow.

			If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from
			Vx, and the results stored in Vx.
		*/
		{
			msg:    "Vx > Vy",
			before: VM{V: [16]uint8{0xA: 0x2, 0xB: 0x1, 0xF: 0}},
			op:     SUBVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x1, 0xB: 0x1, 0xF: 1}},
		},
		{
			msg:    "Vx == Vy",
			before: VM{V: [16]uint8{0xA: 0x2, 0xB: 0x2, 0xF: 1}},
			op:     SUBVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xB: 0x2, 0xF: 0}},
		},
		{
			msg:    "Vx < Vy",
			before: VM{V: [16]uint8{0xA: 0x1, 0xB: 0x2, 0xF: 1}},
			op:     SUBVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0xFF, 0xB: 0x2, 0xF: 0}},
		},

		/*
			8xy6 - SHR Vx {, Vy}
			Set Vx = Vx SHR 1.

			If the least-significant bit of Vx is 1, then VF is set to 1, otherwise
			0. Then Vx is divided by 2.
		*/
		{
			msg:    "no overflow",
			before: VM{V: [16]uint8{0xA: 0x4, 0xF: 1}},
			op:     SHRVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0x2, 0xF: 0}},
		},
		{
			msg:    "overflow",
			before: VM{V: [16]uint8{0xA: 0x1, 0xF: 0}},
			op:     SHRVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xF: 1}},
		},

		/*
			8xy7 - SUBN Vx, Vy
			Set Vx = Vy - Vx, set VF = NOT borrow.

			If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from
			Vy, and the results stored in Vx.
		*/
		{
			msg:    "Vy > Vx",
			before: VM{V: [16]uint8{0xA: 0x1, 0xB: 0x3, 0xF: 0}},
			op:     SUBNVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x2, 0xB: 0x3, 0xF: 1}},
		},
		{
			msg:    "Vy == Vx",
			before: VM{V: [16]uint8{0xA: 0x2, 0xB: 0x2, 0xF: 1}},
			op:     SUBNVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xB: 0x2, 0xF: 0}},
		},
		{
			msg:    "Vy < Vx",
			before: VM{V: [16]uint8{0xA: 0x2, 0xB: 0x1, 0xF: 1}},
			op:     SUBNVxVy{x: 0xA, y: 0xB},
			after:  VM{V: [16]uint8{0xA: 0xFF, 0xB: 0x1, 0xF: 0}},
		},

		/*
			8xyE - SHL Vx {, Vy}
			Set Vx = Vx SHL 1.

			If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to
			0.  Then Vx is multiplied by 2.
		*/
		{
			msg:    "no overflow",
			before: VM{V: [16]uint8{0xA: 0x4, 0xF: 1}},
			op:     SHLVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0x8, 0xF: 0}},
		},
		{
			msg:    "overflow",
			before: VM{V: [16]uint8{0xA: 0x80, 0xF: 0}},
			op:     SHLVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0x0, 0xF: 1}},
		},

		/*
			9xy0 - SNE Vx, Vy
			Skip next instruction if Vx != Vy.

			The values of Vx and Vy are compared, and if they are not equal, the
			program counter is increased by 2.
		*/
		{
			msg:    "Vx == Vy",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
			op:     SNEVxVy{x: 0xA, y: 0xB},
			after:  VM{PC: 0x200, V: [16]uint8{0xA: 0x1, 0xB: 0x1}},
		},
		{
			msg:    "Vx != Vy",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0x1, 0xB: 0x0}},
			op:     SNEVxVy{x: 0xA, y: 0xB},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0x1, 0xB: 0x0}},
		},

		/*
			Annn - LD I, addr
			Set I = nnn.

			The value of register I is set to nnn.
		*/
		{
			before: VM{I: 0x200},
			op:     LDI{nnn: 0x400},
			after:  VM{I: 0x400},
		},

		/*
			Bnnn - JP V0, addr
			Jump to location nnn + V0.

			The program counter is set to nnn plus the value of V0.
		*/
		{
			before: VM{PC: 0x200, V: [16]uint8{0x0: 0xA}},
			op:     JPV0{nnn: 0x400},
			after:  VM{PC: 0x40A, V: [16]uint8{0x0: 0xA}},
		},

		/*
			Cxkk - RND Vx, byte
			Set Vx = random byte AND kk.

			The interpreter generates a random number from 0 to 255, which is then
			ANDed with the value kk. The results are stored in Vx. See instruction
			8xy2 for more information on AND.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0x10}, random: Constant{0x77}},
			op:     RNDVx{x: 0xA, kk: 0xF0},
			after:  VM{V: [16]uint8{0xA: 0x70}, random: Constant{0x77}},
		},

		/*
			Dxyn - DRW Vx, Vy, nibble
			Display n-byte sprite starting at memory location I at (Vx, Vy), set VF =
			collision.

			The interpreter reads n bytes from memory, starting at the address stored
			in I.  These bytes are then displayed as sprites on screen at coordinates
			(Vx, Vy).  Sprites are XORed onto the existing screen. If this causes any
			pixels to be erased, VF is set to 1, otherwise it is set to 0. If the
			sprite is positioned so part of it is outside the coordinates of the
			display, it wraps around to the opposite side of the screen. See
			instruction 8xy3 for more information on XOR, and section 2.4, Display,
			for more information on the Chip-8 screen and sprites.
		*/

		/*
			Ex9E - SKP Vx
			Skip next instruction if key with the value of Vx is pressed.

			Checks the keyboard, and if the key corresponding to the value of Vx is
			currently in the down position, PC is increased by 2.
		*/
		{
			msg:    "key is pressed",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: true}},
			op:     SKPVx{x: 0xA},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: true}},
		},
		{
			msg:    "key is not pressed",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: false}},
			op:     SKPVx{x: 0xA},
			after:  VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: false}},
		},

		/*
			ExA1 - SKNP Vx
			Skip next instruction if key with the value of Vx is not pressed.

			Checks the keyboard, and if the key corresponding to the value of Vx is
			currently in the up position, PC is increased by 2.
		*/
		{
			msg:    "key is pressed",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: true}},
			op:     SKNPVx{x: 0xA},
			after:  VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: true}},
		},
		{
			msg:    "key is not pressed",
			before: VM{PC: 0x200, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: false}},
			op:     SKNPVx{x: 0xA},
			after:  VM{PC: 0x202, V: [16]uint8{0xA: 0xB}, Keys: [16]bool{0xB: false}},
		},

		/*
			Fx07 - LD Vx, DT
			Set Vx = delay timer value.

			The value of DT is placed into Vx.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0xB}, DT: 0xC},
			op:     LDVxDT{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0xC}, DT: 0xC},
		},

		/*
			Fx0A - LD Vx, K
			Wait for a key press, store the value of the key in Vx.

			All execution stops until a key is pressed, then the value of that key is
			stored in Vx.
		*/
		{
			before: VM{},
			op:     LDVxK{x: 0xA},
			after:  VM{IsWaitingForKeyPress: true, K: 0xA},
		},

		/*
			Fx15 - LD DT, Vx
			Set delay timer = Vx.

			DT is set equal to the value of Vx.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0xB}},
			op:     LDDTVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0xB}, DT: 0xB},
		},

		/*
			Fx18 - LD ST, Vx
			Set sound timer = Vx.

			ST is set equal to the value of Vx.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0xB}},
			op:     LDSTVx{x: 0xA},
			after:  VM{V: [16]uint8{0xA: 0xB}, ST: 0xB},
		},

		/*
			Fx1E - ADD I, Vx
			Set I = I + Vx.

			The values of I and Vx are added, and the results are stored in I.
		*/
		{
			before: VM{I: 0x2, V: [16]uint8{0xA: 0xB}},
			op:     ADDIVx{x: 0xA},
			after:  VM{I: 0xD, V: [16]uint8{0xA: 0xB}},
		},

		/*
			Fx29 - LD F, Vx
			Set I = location of sprite for digit Vx.

			The value of I is set to the location for the hexadecimal sprite
			corresponding to the value of Vx. See section 2.4, Display, for more
			information on the Chip-8 hexadecimal font.
		*/
		{
			before: VM{V: [16]uint8{0xA: 0xB}},
			op:     LDFVx{x: 0xA},
			after:  VM{I: 55, V: [16]uint8{0xA: 0xB}},
		},

		/*
			Fx33 - LD B, Vx
			Store BCD representation of Vx in memory locations I, I+1, and I+2.

			The interpreter takes the decimal value of Vx, and places the hundreds
			digit in memory at location in I, the tens digit at location I+1, and
			the ones digit at location I+2.
		*/

		/*
			Fx55 - LD [I], Vx
			Store registers V0 through Vx in memory starting at location I.

			The interpreter copies the values of registers V0 through Vx into
			memory, starting at the address in I.
		*/
		{
			before: VM{I: 0x300, V: [16]uint8{0x0: 0x0, 0x1: 0x1, 0x2: 0x2, 0x3: 0x3}},
			op:     LDIVx{x: 0x3},
			after: VM{
				I:      0x300,
				V:      [16]uint8{0x0: 0x0, 0x1: 0x1, 0x2: 0x2, 0x3: 0x3},
				Memory: [4096]uint8{0x300: 0x0, 0x301: 0x1, 0x302: 0x2, 0x303: 0x3},
			},
		},

		/*
			Fx65 - LD Vx, [I]
			Read registers V0 through Vx from memory starting at location I.

			The interpreter reads values from memory starting at location I into
			registers V0 through Vx.
		*/
		{
			before: VM{
				I:      0x300,
				Memory: [4096]uint8{0x300: 0x0, 0x301: 0x1, 0x302: 0x2, 0x303: 0x3},
			},
			op: LDVxI{x: 0x3},
			after: VM{
				I:      0x300,
				Memory: [4096]uint8{0x300: 0x0, 0x301: 0x1, 0x302: 0x2, 0x303: 0x3},
				V:      [16]uint8{0x0: 0x0, 0x1: 0x1, 0x2: 0x2, 0x3: 0x3},
			},
		},
	} {
		actualAfter := testCase.before
		testCase.op.execute(&actualAfter)
		if testCase.after != actualAfter {
			t.Errorf("Unexpected VM state after executing %#v %s", testCase.op, testCase.msg)
		}
	}
}
