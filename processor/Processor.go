package processor

import (
	"fmt"
)

// flagRegisterState defines the possible flag register conditions
type flagRegisterState int16

// the constants related to the flag register
const (
	equal = 1 << iota
	zero
	carry // overflow is an alias for carry
	greater
	lesser
	negative
	divZero
)

// ICMCProcessor defines a complete simulated processor, with 1 << 15 (32768)
// 16 bit words for data and code.
// The processor is mostly the same as it's VHDL counterpart, except for the
// flag register, that was not documented, therefore there are some assumptions
// here on how it works.
type ICMCProcessor struct {
	Code    [1 << 15]uint16 // this is needed for when reseting, not used during runtime
	Data    [1 << 15]uint16 // the whole code, global variables and stack is here.
	GPRRegs [8]uint16       // the list of register values stored
	SP      uint16          // stack pointer
	PC      uint16          // program counter

	fr flagRegisterState // the flag register, internal because of it's non portability

	IsRunning bool
	inChar    func() (uint8, error)        // inchar environment hook
	outChar   func(char, pos uint16) error // outchar environment hook
}

func NewEmptyProcessor(inChar func() (uint8, error),
	outChar func(char, pos uint16) error) *ICMCProcessor {

	return &ICMCProcessor{
		SP:      (1 << 15) - 1,
		inChar:  inChar,
		outChar: outChar,
	}
}

// fetchInstruction gets, based on an opcode and the AllInstructions list, the
// instruction associated with the opocde. 
// It returns false if the opcode does not exist.
//
// For now, fetchInstruction executes a linear search in the list. This is 
// efficient in most cases because the most used instructions are at the 
// start of the list.
// 
// TODO: maybe a binary search and some kind of instruction index switching 
// mechanism would be much faster.
func fetchInstruction(op Opcode) (Instruction, bool) {
	for _, inst := range AllInstructions {
		if inst.Op == op {
			return inst, true
		}
	}
	return Instruction{}, false
}

// RunInstruction runs a single instruction, incrementing the program counter
// unless the instruction is halt. It returns if the processor should still
// run (true) or should stop immediatly (in the case of a halt or breakp).
func (pr *ICMCProcessor) RunInstruction() (bool, error) {
	if pr.PC >= ((1 << 15) - 1) {
		return false, fmt.Errorf("PC at the end of data section")
	}

	currentOpcode := Opcode(pr.Data[pr.PC] >> 10)

	if currentOpcode == OpHALT {
		return false, nil
	}

	inst, ok := fetchInstruction(currentOpcode)
	if !ok {
		pr.PC++ // skip this instruction in order not to loop on the same error
		return false, fmt.Errorf("instruction does not exist")
	}

	err := inst.Execute(pr)

	pr.PC += uint16(inst.Size)
	return currentOpcode != OpBREAKP, err
}

// RunUntilHalt runs every instruction until a halt is found or an error occurs.
// If an error happens the program counter is still incremented, but if a halt
// is read it will stop right before the increment.
func (pr *ICMCProcessor) RunUntilHalt() (err error) {
	var remain bool
	pr.IsRunning = true

	for pr.IsRunning {
		remain, err = pr.RunInstruction()
		if !remain || err != nil {
			pr.IsRunning = false
		}
	}
	return err
}

// Reset returns all registers to their initial state, and cleans the data
// used, returning it to the initial Code provided.
// Nothing related to screen cleaning is done.
func (pr *ICMCProcessor) Reset() {
	pr.SP = (1 << 15) - 1
	pr.PC = 0
	for i := range pr.GPRRegs {
		pr.GPRRegs[i] = 0
	}

	pr.fr = flagRegisterState(0)

	copy(pr.Data[:], pr.Code[:])
}

// GetMnemonic gets the assembly string that describes the data at a certain
// location. If the data at that location is right before is a 32 bit
// instructon, or if the opcode is invalid, the return value is the decimal
// representation for an immediate; otherwise, the assembly representation is
// returned.
func (pr *ICMCProcessor) GetMnemonic(loc int) string {
	instData := pr.Data[loc]

	if pr.isOperand(loc) {
		return fmt.Sprintf("#%d", instData)
	} else {
		inst, ok := fetchInstruction(Opcode(instData >> 10))
		if !ok {
			return fmt.Sprintf("<invalid opcode %d>", instData>>10)
		}
		return inst.GenMnemonic(instData)
	}
}

// isOperand checks if the data at a certain location is an instruction or an
// operand for one, used in GetMnemonic.
// TODO: this method may be not optimized, if we create a MIF with too many
// 32 bit instructions following each other the code will recurse too much.
func (pr *ICMCProcessor) isOperand(loc int) bool {
	if loc == 0 {
		// loc == 0 cannot have a 32 bit instruction before it
		return false
	}

	instPrevData := pr.Data[loc-1]
	instPrev, ok := fetchInstruction(Opcode(instPrevData >> 10))
	if !ok {
		// if the instruction before does not exist, it is certainly not 32 bits
		return false
	}

	if instPrev.Size == 2 {
		// se if the previous instruction was actually an operand for the
		// instruction before!
		return !pr.isOperand(loc - 1)
	}

	// instructon is 16 bits, we are not it's operand!
	return false
}
