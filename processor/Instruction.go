package processor

type Opcode byte

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

type Instruction struct {
	GenMnemonic func(uint16) string
	Size        byte
	Execute     func(*ICMCProcessor) error
}

var AllInstructions = map[Opcode]Instruction{
	OpNOP: {func(uint16) string { return "nop" }, 1,
		func(*ICMCProcessor) error { return nil }},
	OpADD: {genALUM(true, "add"), 1,
		execALU(true, func(a, b uint32) uint32 { return a + b })},
	OpSUB: {genALUM(true, "sub"), 1,
		execALU(true, func(a, b uint32) uint32 { return a - b })},
	OpMULT: {genALUM(true, "mult"), 1,
		execALU(true, func(a, b uint32) uint32 { return a * b })},
	OpDIV: {genALUM(true, "div"), 1, execDIV},
	OpMOD: {genALUM(false, "mod"), 1,
		execALU(false, func(a, b uint32) uint32 { return a % b })},
	OpAND: {genALUM(false, "and"), 1,
		execALU(false, func(a, b uint32) uint32 { return a & b })},
	OpOR: {genALUM(false, "or"), 1,
		execALU(false, func(a, b uint32) uint32 { return a | b })},
	OpXOR: {genALUM(false, "xor"), 1,
		execALU(false, func(a, b uint32) uint32 { return a ^ b })},
	OpNOT:     {gen2RegM("not"), 1, execNOT},
	OpINCDEC:  {genINCDECM, 1, execINCDEC},
	OpCMP:     {gen2RegM("cmp"), 1, execCMP},
	OpROTSH:   {genROTSHM, 1, execROTSH},
	OpMOV:     {genMOVM, 1, execMOV},
	OpPUSH:    {gen1RegM("push"), 1, execPUSH},
	OpPOP:     {gen1RegM("pop"), 1, execPOP},
	OpLOADN:   {gen1RegM("loadn"), 2, execLOADN},
	OpLOAD:    {gen1RegM("load"), 2, execLOAD},
	OpSTORE:   {gen1RegM("store"), 2, execSTORE},
	OpLOADI:   {gen2RegM("loadi"), 1, execLOADI},
	OpSTOREI:  {gen2RegM("storei"), 1, execSTOREI},
	OpRTS:     {func(uint16) string { return "rts" }, 1, execRTS},
	OpJMP:     {genJMPM, 2, execJMP},
	OpCALL:    {genCALLM, 2, execCALL},
	OpINCHAR:  {gen1RegM("inchar"), 1, execINCHAR},
	OpOUTCHAR: {gen2RegM("outchar"), 1, execOUTCHAR},
	OpHALT: {func(uint16) string { return "halt" }, 1,
		func(*ICMCProcessor) error { return nil }},
	OpBREAKP: {func(uint16) string { return "breakp" }, 1,
		func(*ICMCProcessor) error { return nil }},
	OpCSCARRY: {genCSCARRYM, 1, execCSCARRY},
}

func toRegStr(value uint16) string {
	return "R" + string(byte(value)+'0')
}

func getRegsStr(inst uint16, startBits []int) string {
	ret := ""
	for _, sb := range startBits {
		ret += toRegStr(getRegAt(inst, sb)) + ", "
	}

	ret = ret[:len(ret)-2]
	return ret
}

func gen1RegM(m string) func(uint16) string {
	return func(d uint16) string {
		mnemonic := m

		mnemonic += " " + getRegsStr(d, []int{7})

		return mnemonic
	}
}

func gen2RegM(m string) func(uint16) string {
	return func(d uint16) string {
		mnemonic := m

		mnemonic += " " + getRegsStr(d, []int{7, 4})

		return mnemonic
	}
}

func getRegAt(inst uint16, bit int) uint16 {
	return (inst & (0b111 << bit)) >> bit
}
