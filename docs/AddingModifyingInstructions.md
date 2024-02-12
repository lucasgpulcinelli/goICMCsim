# 📝 Adding/Modifying Instructions

## 🛠️ How to Add/Modify Instructions in the Simulator

To add or modify instructions in the simulator, follow these steps:

1. Choose an opcode for your instruction.
2. Add it to the constants list at [processor/Instruction.go](processor/Instruction.go).
3. Below the constants list in the same file, add your instruction data to the `AllInstructions` list. This allows the simulator to find and execute your instruction.
4. You need to provide four pieces of information:
    - The opcode you just created.
    - An instruction to generate its mnemonic string. If your instruction receives a list of registers, use `genRegM` (as most instructions do), or create a custom function and add it there.
    - The instruction size in 16-bit words.
    - A function to execute it.
5. To create the execution function, you need a function that takes the processor context and returns an error (usually `nil` to indicate success).

## 📋 An Instruction Example

Let's create a new instruction called `incmod`. This instruction takes a register and adds 1 to it. If it becomes greater than or equal to another register, it wraps around like a modulo operation. Essentially, `incmod rx, ry` is equivalent to `rx = (rx + 1) % ry`.

### 🔹 Instruction Encoding
This instruction will have opcode `0b111111`, plus three bits for the first register and three more for the second, resulting in `0b111111xxxyyydddd` (where `x` represents the first bits for the first register, `y` is for the second, and `d` is a don't care value, meaning an unused `0` or `1`).

### 🔹 Instruction Entry
The entry in `AllInstructions` will be:
```golang
{OpINCMOD, genRegM("incmod", 2), 1, execINCMOD},
```
Where `OpINCMOD` is defined above in the constant block as `OpINCMOD = 0b111111`. `genRegM` is used with those arguments to define the mnemonic and to specify that it has two register opcodes in the usual position. `1` indicates that a single word is necessary to encode the instruction.

### 🔹 Execution Function
The `execINCMOD` function is defined as follows:

```golang
func execINCMOD(pr *ICMCProcessor) error {
    // Get the instruction binary data
    inst := pr.Data[pr.PC]

    // Retrieve register indices (0-7)
    rx_index := getRegAt(inst, 7) // 7 is the first bit from right to left encoding the register index for rx
    ry_index := getRegAt(inst, 4) // Same for ry

    // Set the value in rx to the result of our operation
    pr.GPRRegs[rx_index] = (pr.GPRRegs[rx_index] + 1) % pr.GPRRegs[ry_index]

    // No errors occurred
    return nil
}
```

This function performs the operation described by the `incmod` instruction and returns `nil` to indicate success.