package processor

import (
	"fmt"
)

type ALUOpFunc func(uint32, uint32) uint32

func genALUM(withCarry bool, m string) func(uint16) string {
	return func(d uint16) string {
		mnemonic := m

		if withCarry && d&1 == 1 {
			mnemonic += "c"
		}

		mnemonic += " " + getRegsStr(d, []int{7, 4, 1})

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
	ret += getRegsStr(d, []int{7})
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
	case 4:
	case 5:
		ret += "rotl "
	case 6:
	case 7:
		ret += "rotr "
	}

	ret += getRegsStr(d, []int{7}) + ", "
	ret += fmt.Sprintf("%d", d&0b1111)

	return ret
}

func genMOVM(d uint16) string {
	mnemonic := "mov "

	if d&0b11 == 0b11 {
		mnemonic += "SP, "
		mnemonic += getRegsStr(d, []int{7})
	} else if d&1 == 1 {
		mnemonic += getRegsStr(d, []int{7})
		mnemonic += ", SP"
	} else {
		mnemonic += getRegsStr(d, []int{7, 4})
	}

	return mnemonic
}

func execALU(withCarry bool, operation ALUOpFunc) func(*ICMCProcessor) error {
	return func(pr *ICMCProcessor) error {
		inst := pr.Data[pr.PC]
		RD := getRegAt(inst, 7)
		RS1 := getRegAt(inst, 4)
		RS2 := getRegAt(inst, 1)

		result := operation(uint32(pr.GPRRegs[RS1]), uint32(pr.GPRRegs[RS2]))
		if withCarry && inst&1 == 1 && ((pr.fr & carry) != 0) {
			result++
		}

		pr.updateALUFR(result)
		pr.GPRRegs[RD] = uint16(result)

		return nil
	}
}

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
		result = uint32(pr.GPRRegs[RD] - 1)
		pr.updateALUFR(result)
	} else {
		result = uint32(pr.GPRRegs[RD] + 1)
		pr.updateALUFR(result)
	}

	pr.GPRRegs[RD] = uint16(result)
	return nil
}

func execROTSH(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]
	RD := getRegAt(inst, 7)
	n := inst & 0b1111

	switch (inst & (0b111 << 4)) >> 4 {
	case 0:
		pr.GPRRegs[RD] = pr.GPRRegs[RD] << n
	case 1:
		pr.GPRRegs[RD] = ^((^pr.GPRRegs[RD]) << n)
	case 2:
		pr.GPRRegs[RD] = pr.GPRRegs[RD] >> n
	case 3:
		pr.GPRRegs[RD] = ^((^pr.GPRRegs[RD]) >> n)
	case 4:
	case 5:
		upper := (pr.GPRRegs[RD] << n)
		lower := pr.GPRRegs[RD] & (((1 << 16) - 1) << (16 - n))
		lower = lower >> (16 - n)
		pr.GPRRegs[RD] = upper + lower
	case 6:
	case 7:
		lower := (pr.GPRRegs[RD] >> n)
		upper := pr.GPRRegs[RD] & ((1 << n) - 1)
		upper = upper << n
		pr.GPRRegs[RD] = upper + lower
	}
	return nil
}

func execMOV(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

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
