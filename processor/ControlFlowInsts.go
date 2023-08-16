package processor

import "fmt"

func genCFlowM(subOpcode uint16) string {
	switch subOpcode {
  case 0:
    return ""
	case 1:
		return "eq"
	case 2:
		return "ne"
	case 3:
		return "z"
	case 4:
		return "nz"
	case 5:
		return "c"
	case 6:
		return "nc"
	case 7:
		return "gr"
	case 8:
		return "le"
	case 9:
		return "eg"
	case 10:
		return "el"
	case 11:
		return "ov"
	case 12:
		return "nov"
	case 13:
		return "n"
	case 14:
		return "dz"
	}
	return "INVALID"
}

func genJMPM(inst uint16) string {
	subOpcode := (inst & (0b1111 << 6)) >> 6
	if subOpcode == 0 {
		return "jmp"
	}
	return "j" + genCFlowM(subOpcode)
}

func genCALLM(inst uint16) string {
	subOpcode := (inst & (0b1111 << 6)) >> 6
	if subOpcode == 0 {
		return "call"
	}
	return "c" + genCFlowM(subOpcode)
}

func execCMP(pr *ICMCProcessor) error {
	inst := pr.Data[pr.PC]
	RS1 := getRegAt(inst, 7)
	RS2 := getRegAt(inst, 4)

	RS1data := pr.GPRRegs[RS1]
	RS2data := pr.GPRRegs[RS2]

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
	if pr.SP >= ((1 << 15)-1) {
		return fmt.Errorf("invalid stack pointer value")
	}
	pr.PC = pr.Data[pr.SP+1] - 1
	pr.SP++
	return nil
}

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

	return false, fmt.Errorf("invalid jump with subopcode %d", subOpcode)
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

	pr.Data[pr.SP] = pr.PC+2
	pr.SP--

	pr.PC = pr.Data[pr.PC+1] - 2
	return nil
}
