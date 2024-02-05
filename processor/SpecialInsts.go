package processor

import "fmt"

func genCSCARRYM(inst uint16) string {
	if inst&(1<<9) != 0 {
		return "clearc"
	} else {
		return "setc"
	}
}

func execINCHAR(pr *ICMCProcessor) error {
	// because inchar is dependent on the environment (for instance, the visual
	// toolkit used), the ICMCProcessor just calls a hook that must be defined.

	inst := pr.Data[pr.PC]

	RD := getRegAt(inst, 7)
	v, err := pr.inChar()
	if err != nil {
		return err
	}

	pr.GPRRegs[RD] = uint16(v)
	return nil
}

func execOUTCHAR(pr *ICMCProcessor) error {
	// because outchar is dependent on the environment (for instance, the visual
	// toolkit used), the ICMCProcessor just calls a hook that must be defined.

	inst := pr.Data[pr.PC]

	RS1 := getRegAt(inst, 7)
	RS2 := getRegAt(inst, 4)

	return pr.outChar(pr.GPRRegs[RS1], pr.GPRRegs[RS2])
}

func execCSCARRY(pr *ICMCProcessor) error {
	if pr.Data[pr.PC]&(1<<9) != 0 {
		pr.fr &= carry
	} else {
		pr.fr &= ^carry
	}
	return nil
}

// execNOP is a placeholder for instructions that don't do anything. This
// includes the actual NOP instruction, but also instructions that have effects
// in RunInstruction directly just via their opcode, such as breakp and halt.
func execNOP(pr *ICMCProcessor) error {
	return nil
}

func execSTOP(pr *ICMCProcessor) error {
	return fmt.Errorf("stop")
}
