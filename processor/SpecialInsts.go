package processor

func genCSCARRYM(inst uint16) string {
  if inst & (1<<9) != 0 {
    return "clearc"
  } else {
    return "setc"
  }
}

func execINCHAR(pr *ICMCProcessor) error {
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
