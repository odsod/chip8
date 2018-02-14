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
