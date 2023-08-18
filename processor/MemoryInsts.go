package processor

import "fmt"

func execPUSH(pr *ICMCProcessor) error {
	var value uint16

	inst := pr.Data[pr.PC]

  // see if we are pushing the flag register
	if inst&(1<<6) != 0 {
		// because a sequence of push fr, pop rX is considered seriously bad
		// practice, the meaning in the sequence of bits in the flag register do
		// not match the original implementation.
		value = uint16(pr.fr)
	} else {
		RS := getRegAt(inst, 7)
		value = pr.GPRRegs[RS]
	}

	if pr.SP >= (1<<15) || pr.SP == 0 {
		return fmt.Errorf("invalid stack pointer value")
	}

	pr.Data[pr.SP] = value
	pr.SP--
	return nil
}

func execPOP(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	if pr.SP >= (1 << 15) {
		return fmt.Errorf("invalid stack pointer value")
	}

	pr.SP++

  // see if we are popping the flag register
	if inst&(1<<6) != 0 {
		pr.fr = flagRegisterState(pr.Data[pr.SP])
	} else {
		RD := getRegAt(inst, 7)
		pr.GPRRegs[RD] = pr.Data[pr.SP]
	}
	return nil
}

func execLOADN(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)

	if pr.PC == (1<<15)-1 {
		return fmt.Errorf("loadn at the end of data section")
	}

	pr.GPRRegs[RD] = pr.Data[pr.PC+1]
	return nil
}

func execLOAD(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)

	if pr.PC >= (1<<15)-1 {
		return fmt.Errorf("load at the end of data section")
	}

	loc := pr.Data[pr.PC+1]
	if loc >= (1<<15)-1 {
		return fmt.Errorf("load has invalid memory as operand")
	}

	pr.GPRRegs[RD] = pr.Data[loc]
	return nil
}

func execSTORE(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	RS := getRegAt(inst, 7)

	if pr.PC == (1<<15)-1 {
		return fmt.Errorf("store at the end of data section")
	}

	loc := pr.Data[pr.PC+1]
	if loc >= (1<<15)-1 {
		return fmt.Errorf("store has invalid memory as operand")
	}

	pr.Data[loc] = pr.GPRRegs[RS]
	return nil
}

func execLOADI(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)
	RS := getRegAt(inst, 4)

	loc := pr.GPRRegs[RS]
	if loc >= (1<<15)-1 {
		return fmt.Errorf("loadi has invalid memory as operand")
	}

	pr.GPRRegs[RD] = pr.Data[loc]
	return nil
}

func execSTOREI(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)
	RS := getRegAt(inst, 4)

	loc := pr.GPRRegs[RD]
	if loc >= (1<<15)-1 {
		return fmt.Errorf("storei has invalid memory as operand")
	}

	pr.Data[loc] = pr.GPRRegs[RS]
	return nil
}
