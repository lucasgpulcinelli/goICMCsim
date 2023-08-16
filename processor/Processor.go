package processor

import (
	"fmt"
)

type flagRegisterState int16

const (
	equal = 1 << iota
	zero
	carry // overflow is an alias for carry
	greater
	lesser
	negative
	divZero
)

type ICMCProcessor struct {
	Code    [1 << 15]uint16
	Data    [1 << 15]uint16
	GPRRegs [8]uint16
	SP      uint16
	PC      uint16

	fr flagRegisterState

	IsRunning bool
	inChar    func() (uint8, error)
	outChar   func(char, pos uint16) error
}

func NewEmptyProcessor(inChar func() (uint8, error),
	outChar func(char, pos uint16) error) *ICMCProcessor {

	return &ICMCProcessor{
		SP:      (1 << 15) - 1,
		inChar:  inChar,
		outChar: outChar,
	}
}

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

func (pr *ICMCProcessor) RunInstruction() error {
	if pr.PC >= ((1 << 15) - 1) {
		return fmt.Errorf("PC at the end of data section")
	}

	currentOpcode := Opcode(pr.Data[pr.PC] >> 10)

	inst, ok := AllInstructions[currentOpcode]
	if !ok {
		return fmt.Errorf("instruction does not exist")
	}

	if err := inst.Execute(pr); err != nil {
		return err
	}

	pr.PC += uint16(inst.Size)
	return nil
}

func (pr *ICMCProcessor) RunUntilHalt() error {
	pr.IsRunning = true
	for {
		op := Opcode(pr.Data[pr.PC] >> 10)
		if op == OpHALT {
			pr.IsRunning = false
			return nil
		}
		if err := pr.RunInstruction(); err != nil {
			pr.IsRunning = false
			return err
		}
		if op == OpBREAKP {
			pr.IsRunning = false
			return nil
		}
	}
}

func (pr *ICMCProcessor) Reset() {
	pr.SP = (1 << 15) - 1
	pr.PC = 0
	for i := range pr.GPRRegs {
		pr.GPRRegs[i] = 0
	}

	pr.fr = flagRegisterState(0)

	copy(pr.Data[:], pr.Code[:])
}

func (pr *ICMCProcessor) GetMnemonic(loc int) string {
	instData := pr.Data[loc]

	inst, ok := AllInstructions[Opcode(instData>>10)]
	if !ok {
		return fmt.Sprintf("#%d", instData)
	}

	if loc < 1 {
		return inst.GenMnemonic(instData)
	}

	instPrevData := pr.Data[loc-1]
	instPrev, ok := AllInstructions[Opcode(instPrevData>>10)]

	if ok && instPrev.Size == 2 {
		if loc < 2 {
			return inst.GenMnemonic(instData)
		}

		instPPrevData := pr.Data[loc-2]
		instPPrev, ok := AllInstructions[Opcode(instPPrevData>>10)]

		if !ok || instPPrev.Size != 2 {
			return fmt.Sprintf("#%d", instData)
		}
	}

	return inst.GenMnemonic(instData)
}
