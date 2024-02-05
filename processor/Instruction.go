package processor

import "fmt"

type Opcode byte

// all opcodes in the ICMC architecture
const (
	OpNOP     Opcode = 0b000000
	OpADD            = 0b100000
	OpSUB            = 0b100001
	OpMULT           = 0b100010
	OpDIV            = 0b100011
	OpMOD            = 0b100101
	OpAND            = 0b010010
	OpOR             = 0b010011
	OpXOR            = 0b010100
	OpNOT            = 0b010101
	OpINCDEC         = 0b100100
	OpCMP            = 0b010110
	OpROTSH          = 0b010000
	OpMOV            = 0b110011
	OpPUSH           = 0b000101
	OpPOP            = 0b000110
	OpLOADN          = 0b111000
	OpLOAD           = 0b110000
	OpSTORE          = 0b110001
	OpLOADI          = 0b111100
	OpSTOREI         = 0b111101
	OpRTS            = 0b000100
	OpJMP            = 0b000010
	OpCALL           = 0b000011
	OpINCHAR         = 0b110101
	OpOUTCHAR        = 0b110010
	OpHALT           = 0b001111
	OpBREAKP         = 0b001110
	OpCSCARRY        = 0b001000
)

// Instruction describes all data a single instruction needs to be fully
// described for execution and display.
type Instruction struct {
	Op          Opcode
	GenMnemonic func(uint16) string
	Size        byte
	Execute     func(*ICMCProcessor) error
}

// The vector of all instructions in the ICMC architecture.
// To add a new instruction, just add an entry here (and in the Opcode consts)
// and create the two function hooks.
// The ALU instructions are complicated because they use functional programming
// and reuse some ALU properties, such as flag register settings.
// The order here is important! The most executed instructions must appear
// first for faster instruction fetching
var AllInstructions = []Instruction{
	{OpJMP, genJMPM, 2, execJMP},
	{OpINCDEC, genINCDECM, 1, execINCDEC},
	{OpINCHAR, genRegM("inchar", 1), 1, execINCHAR},
	{OpCMP, genRegM("cmp", 2), 1, execCMP},
	{OpADD, genALUM(true, "add"), 1,
		execALU(true, func(a, b uint32) uint32 { return a + b }),
	},
	{OpSUB, genALUM(true, "sub"), 1,
		execALU(true, func(a, b uint32) uint32 { return a - b }),
	},
	{OpMULT, genALUM(true, "mult"), 1,
		execALU(true, func(a, b uint32) uint32 { return a * b }),
	},
	{OpMOD, genALUM(false, "mod"), 1,
		execALU(false, func(a, b uint32) uint32 { return a % b }),
	},
	{OpCALL, genCALLM, 2, execCALL},
	{OpOUTCHAR, genRegM("outchar", 2), 1, execOUTCHAR},
	{OpAND, genALUM(false, "and"), 1,
		execALU(false, func(a, b uint32) uint32 { return a & b }),
	},
	{OpOR, genALUM(false, "or"), 1,
		execALU(false, func(a, b uint32) uint32 { return a | b }),
	},
	{OpXOR, genALUM(false, "xor"), 1,
		execALU(false, func(a, b uint32) uint32 { return a ^ b }),
	},
	{OpDIV, genALUM(true, "div"), 1, execDIV},
	{OpNOT, genRegM("not", 2), 1, execNOT},
	{OpLOADI, genRegM("loadi", 2), 1, execLOADI},
	{OpSTOREI, genRegM("storei", 2), 1, execSTOREI},
	{OpPUSH, genRegM("push", 1), 1, execPUSH},
	{OpPOP, genRegM("pop", 1), 1, execPOP},
	{OpLOADN, genRegM("loadn", 1), 2, execLOADN},
	{OpLOAD, genRegM("load", 1), 2, execLOAD},
	{OpSTORE, genRegM("store", 1), 2, execSTORE},
	{OpRTS, genRegM("rts", 0), 1, execRTS},
	{OpNOP, genRegM("nop", 0), 1, execNOP},
	{OpHALT, genRegM("halt", 0), 0, execSTOP},
	{OpBREAKP, genRegM("breakp", 0), 1, execSTOP},
	{OpCSCARRY, genCSCARRYM, 1, execCSCARRY},
	{OpROTSH, genROTSHM, 1, execROTSH},
	{OpMOV, genMOVM, 1, execMOV},
}

// toRegStr gets the name of a register based on it's index.
func toRegStr(value uint16) string {
	return "R" + string(byte(value)+'0')
}

// getRegsStr gets, based on a 16 bit instruction and the number of register
// operands, an asm-like string with the register names used.
func getRegsStr(inst uint16, numRegs int) string {
	if numRegs == 0 {
		return ""
	} else if numRegs > 3 {
		panic(fmt.Sprint("invalid register number in getRegsStr: ", numRegs))
	}

	ret := ""

	bit := 7
	for i := 0; i < numRegs; i++ {
		ret += toRegStr(getRegAt(inst, bit)) + ", "
		bit -= 3
	}

	ret = ret[:len(ret)-2]
	return ret
}

// genRegM generates a function that, given a 16 bit instruction, returns an
// asm-like dissasembly of that instruction. This must be done in runtime
// because there are instructions that change mnemonic based on bits different
// than their opcode
func genRegM(m string, numRegs int) func(uint16) string {
	return func(d uint16) string {
		return m + " " + getRegsStr(d, numRegs)
	}
}

// getRegAt gets the register index in an instruction, considering that the
// register describing part of the instruction starts at bit.
func getRegAt(inst uint16, bit int) uint16 {
	return (inst & (0b111 << bit)) >> bit
}
