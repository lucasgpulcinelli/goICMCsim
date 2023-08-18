package processor

import (
	"fmt"
)

type ALUOpFunc func(uint32, uint32) uint32

// genALUM generates an ALU-like instruction, that is, an instruction that
// takes three register operands and may have a unified carry operation bit.
func genALUM(withCarry bool, m string) func(uint16) string {
	return func(d uint16) string {
		mnemonic := m

		if withCarry && d&1 == 1 {
			mnemonic += "c"
		}

		mnemonic += " " + getRegsStr(d, 3)

		return mnemonic
	}
}

func genINCDECM(d uint16) string {
	ret := ""
	if d&(1<<6) != 0 {
		ret += "dec "
	} else {
		ret += "inc "
	}
	ret += getRegsStr(d, 1)
	return ret
}

func genROTSHM(d uint16) string {
	ret := ""
	switch (d & (0b111 << 4)) >> 4 {
	case 0:
		ret += "shiftl0 "
	case 1:
		ret += "shiftl1 "
	case 2:
		ret += "shiftr0 "
	case 3:
		ret += "shiftr1 "
	case 4, 5:
		ret += "rotl "
	case 6, 7:
		ret += "rotr "
	}

	ret += getRegsStr(d, 1) + ", "
	ret += fmt.Sprintf("%d", d&0b1111)

	return ret
}

func genMOVM(d uint16) string {
	mnemonic := "mov "

	if d&0b11 == 0b11 {
		mnemonic += "SP, "
		mnemonic += getRegsStr(d, 1)
	} else if d&1 == 1 {
		mnemonic += getRegsStr(d, 1)
		mnemonic += ", SP"
	} else {
		mnemonic += getRegsStr(d, 2)
	}

	return mnemonic
}

// execALU generates a function that executes an ALU-like instruction (that is,
// a binary operation that returns a number), given an actual operation
// function and the information if the instruction uses the carry bit.
func execALU(withCarry bool, operation ALUOpFunc) func(*ICMCProcessor) error {
	return func(pr *ICMCProcessor) error {
		inst := pr.Data[pr.PC]

		// get the operands
		RD := getRegAt(inst, 7)
		RS1 := getRegAt(inst, 4)
		RS2 := getRegAt(inst, 1)

		// execute the operation itself (using uint32s, because the result might
		// overflow a uint16)
		result := operation(uint32(pr.GPRRegs[RS1]), uint32(pr.GPRRegs[RS2]))

		// if the opcode uses the carry bit, the instruction is of the carry
		// variety, and the flag register did in fact have the carry biy set
		if withCarry && inst&1 == 1 && ((pr.fr & carry) != 0) {
			// use the carry bit to increase the value
			result++
		}

		// update the flag register
		pr.updateALUFR(result)

		// and send the result to the destination register
		pr.GPRRegs[RD] = uint16(result)

		return nil
	}
}

// execDIV executes a division in the ICMCProcessor, beeing unique among
// ALU-like functions because it sets the divZero flag state
func execDIV(pr *ICMCProcessor) error {
	RS2 := getRegAt(pr.Data[pr.PC], 1)
	if pr.GPRRegs[RS2] == 0 {
		pr.fr &= divZero
		return nil
	}

	execALU(true, func(a, b uint32) uint32 { return a / b })(pr)
	return nil
}

func execNOT(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]
	RD := getRegAt(inst, 7)
	RS1 := getRegAt(inst, 4)

	pr.GPRRegs[RD] = ^pr.GPRRegs[RS1]

	return nil
}

func execINCDEC(pr *ICMCProcessor) error {
	var result uint32
	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)

	if inst&(1<<6) != 0 {
		// if we are decrementing
		result = uint32(pr.GPRRegs[RD] - 1)
	} else {
		// if we are incrementing
		result = uint32(pr.GPRRegs[RD] + 1)
	}

	pr.updateALUFR(result)

	pr.GPRRegs[RD] = uint16(result)
	return nil
}

func execROTSH(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]
	RD := getRegAt(inst, 7)

	// operand of bits to shift/rotate
	n := inst & 0b1111

  // TODO: this is kind of complicated bit shifting magic, it would be ideal to
  // document all operations, even if they work okay and the result is obvious.
	switch (inst & (0b111 << 4)) >> 4 {
	case 0:
		pr.GPRRegs[RD] = pr.GPRRegs[RD] << n
	case 1:
		pr.GPRRegs[RD] = ^((^pr.GPRRegs[RD]) << n)
	case 2:
		pr.GPRRegs[RD] = pr.GPRRegs[RD] >> n
	case 3:
		pr.GPRRegs[RD] = ^((^pr.GPRRegs[RD]) >> n)
	case 4, 5:
		upper := (pr.GPRRegs[RD] << n)
		lower := pr.GPRRegs[RD] & (((1 << 16) - 1) << (16 - n))
		lower = lower >> (16 - n)
		pr.GPRRegs[RD] = upper + lower
	case 6, 7:
		lower := (pr.GPRRegs[RD] >> n)
		upper := pr.GPRRegs[RD] & ((1 << n) - 1)
		upper = upper << n
		pr.GPRRegs[RD] = upper + lower
	}
	return nil
}

func execMOV(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

  // special cases are needed to handle mov sp, rx or mov rx, sp
	if inst&0b11 == 0b11 {
		RS := getRegAt(inst, 7)
		pr.SP = pr.GPRRegs[RS]
	} else if inst&1 == 1 {
		RD := getRegAt(inst, 7)
		pr.GPRRegs[RD] = pr.SP
	} else {
		RD := getRegAt(inst, 7)
		RS := getRegAt(inst, 4)
		pr.GPRRegs[RD] = pr.GPRRegs[RS]
	}

	return nil
}

// updateALUFR refreshes the flag register after an ALU instruction given an
// ALU output value. It sets the zero, negative and carry/overflow flags.
func (pr *ICMCProcessor) updateALUFR(value uint32) {
	if uint16(value) == 0 {
		pr.fr |= zero
	} else {
		pr.fr &= ^zero
	}

	if uint32(uint16(value)) != value {
		pr.fr |= carry
	} else {
		pr.fr &= ^carry
	}

	if value&(1<<15) != 0 {
		pr.fr |= negative
	} else {
		pr.fr &= ^negative
	}
}
