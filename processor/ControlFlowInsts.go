package processor

import "fmt"

// the control flow subOpcode to string array.
var cFlowM = []string{
	"", "eq", "ne", "z", "nz", "c", "nc", "gr",
	"le", "eg", "el", "ov", "nov", "n", "dz",
}

func genJMPM(inst uint16) string {
	subOpcode := (inst & (0b1111 << 6)) >> 6
	if subOpcode == 0 {
		return "jmp"
	}

	return "j" + cFlowM[subOpcode]
}

func genCALLM(inst uint16) string {
	subOpcode := (inst & (0b1111 << 6)) >> 6
	if subOpcode == 0 {
		return "call"
	}
	return "c" + cFlowM[subOpcode]
}

func execCMP(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]

  // get register indicies
	RS1 := getRegAt(inst, 7)
	RS2 := getRegAt(inst, 4)

  // get the data at those registers
	RS1data := pr.GPRRegs[RS1]
	RS2data := pr.GPRRegs[RS2]

  // and update FR based on them
	if RS1data > RS2data {
		pr.fr |= greater
		pr.fr &= ^lesser
		pr.fr &= ^equal
	} else if RS1data < RS2data {
		pr.fr |= lesser
		pr.fr &= ^greater
		pr.fr &= ^equal
	} else {
		pr.fr |= equal
		pr.fr &= ^greater
		pr.fr &= ^lesser
	}

	return nil
}

func execRTS(pr *ICMCProcessor) error {
  // if the stack is empty, we cannot return anywhere!
	if pr.SP >= ((1 << 15) - 1) {
		return fmt.Errorf("invalid stack pointer value")
	}

  // get the PC we (hopefully) stored before at a call
	pr.PC = pr.Data[pr.SP+1] - 1

  // and increment the stack pointer to return it to the original position
	pr.SP++
	return nil
}

// shouldExecute returns if a branching instruction should or not actually 
// branch based on the flag register status.
func shouldExecute(fr flagRegisterState, subOpcode uint16) (bool, error) {
	switch subOpcode {
	case 0:
		return true, nil
	case 1:
		return fr&equal == equal, nil
	case 2:
		return fr&equal == 0, nil
	case 3:
		return fr&zero == zero, nil
	case 4:
		return fr&zero == 0, nil
	case 5:
		return fr&carry == carry, nil
	case 6:
		return fr&carry == 0, nil
	case 7:
		return fr&greater == greater, nil
	case 8:
		return fr&lesser == lesser, nil
	case 9:
		return fr&(greater|equal) != 0, nil
	case 10:
		return fr&(lesser|equal) != 0, nil
	case 11:
		return fr&carry == carry, nil // overflow is an alias for carry
	case 12:
		return fr&carry == 0, nil
	case 13:
		return fr&negative == negative, nil
	case 14:
		return fr&divZero == divZero, nil
	}

	return false, fmt.Errorf("invalid branch with subopcode %d", subOpcode)
}

func execJMP(pr *ICMCProcessor) error {
	subOpcode := (pr.Data[pr.PC] >> 6) & 0b1111
	should, err := shouldExecute(pr.fr, subOpcode)
	if err != nil {
		return err
	}
	if !should {
		return nil
	}

	if pr.PC >= (1<<15)-1 {
		return fmt.Errorf("jump at the end of data section")
	}

  // actually jump: set the PC to our immediate argument... -2 because at the 
  // end of RunInstruction we still increment PC.
	pr.PC = pr.Data[pr.PC+1] - 2
	return nil
}

func execCALL(pr *ICMCProcessor) error {
	subOpcode := (pr.Data[pr.PC] & (0b1111 << 6)) >> 6
	should, err := shouldExecute(pr.fr, subOpcode)
	if err != nil {
		return err
	}
	if !should {
		return nil
	}

	if pr.PC >= (1<<15)-1 {
		return fmt.Errorf("call at the end of data section")
	}

	if pr.SP >= (1<<15) || pr.SP == 0 {
		return fmt.Errorf("invalid stack pointer value")
	}

  // the return address is the next instruction in relation to us
	pr.Data[pr.SP] = pr.PC + 2

  // decrement the stack pointer, aka finalize a push PC+2
	pr.SP--

  // actually call: set the PC to our immediate argument... -2 because at the 
  // end of RunInstruction we still increment PC.
	pr.PC = pr.Data[pr.PC+1] - 2
	return nil
}
