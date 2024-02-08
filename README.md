![GO ICMC Simulator](https://github.com/lucasgpulcinelli/goICMCsim/assets/11618151/da81d732-5cb4-4f41-9128-37ae864ceac9)

<p align="center">
<img src="https://img.shields.io/github/go-mod/go-version/lucasgpulcinelli/goICMCsim?logo=go"/>
<a href="https://github.com/lucasgpulcinelli/goICMCsim/issues?q=is%3Aopen+is%3Aissue+label%3Afeature-request+sort%3Areactions-%2B1-desc"><img src="https://img.shields.io/github/issues/lucasgpulcinelli/goICMCsim/feature-request.svg"></a>
<a href="https://github.com/lucasgpulcinelli/goICMCsim/issues?utf8=✓&q=is%3Aissue+is%3Aopen+label%3Abug"><img src="https://img.shields.io/github/issues/lucasgpulcinelli/goICMCsim/bug.svg"></a>
<a href="https://github.com/lucasgpulcinelli/goICMCsim/releases"><img src="https://img.shields.io/github/v/release/lucasgpulcinelli/goICMCsim"/></a>
<img src="https://img.shields.io/github/license/lucasgpulcinelli/goICMCsim"/>
</p>

## Resume
This program is a simulator for the ICMC architecture (defined [here](https://github.com/simoesusp/Processador-ICMC/)), it has many upgraded functionalities in comparision with the c++ simulator, namely:
- A resizable window and fullscreen cabability;
- An instruction scroll to view all instructions and data being modified in real time;
- Ability to edit the stack pointer and program counter directly;
- Better error handling: the simulator will stop and point the error to the programmer instead of carrying on;
- Better parsing of MIF files, relying only in it's syntax definition and showing line, collumn pairs and cause if errors are found;
- Capability to change character mapping MIF during runtime (without reseting);
- Shortcuts that do not rely on keys that may not be present in a laptop keyboard (for instance insert, home and end keys);
- Support for windows, macOS and linux;

## Installation
If you don't want to compile anything, go to the [releases page](https://github.com/lucasgpulcinelli/goICMCsim/releases) and download a precompiled binary for your system.

## Usage
The first thing you will want to do is add a program to run and test it. This can be done by either specifing MIF files in the ICMC architecture format in the command line or by using the file -\> open code/char MIF menu. Remember to always specify a char MIF, otherwise the code's outchars will always output blank characters!

## How to Compile from the Source Code
First, install a recent version of go (at least 1.13), either from your package manager or from [here](https://go.dev/doc/install). After that, you will also need git and a C compiler (on windows, you need to use MinGW).

On debian/ubuntu based systems, you will need to install `libgl1-mesa-dev xorg-dev`;
On fedora and red hat based systems, you will need to install `libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel`;

Then, Just use `go build .` to compile and `./goICMCsim` to start an empty processor. You can see the command line options with `--help`.

If you don't want to clone the repository and just want to compile and install it directly into $GOPATH/bin, just use `go install github.com/lucasgpulcinelli/goICMCsim@latest` (you will still need the tools listed before).

## How to add/modify instructions in the simulator
First, you will need to choose an opcode for your instruction, then add it in the constants list at [processor/Instruction.go](processor/Instruction.go).
After that, right below in the same file, you will need to add your instruction data to the AllInstructions list, for it to be actually found and executed by the simulator.

You will need to add four informations:
- the opcode you just created,
- an instruction to generate it's mnemonic string (if your instruction receives a list of registers, just use genRegM as most instructions, or create a function yourself and add it there),
- the instruction size in 16 bit words,
- and a function to execute it.

To make the execution function, you will need a function that takes the processor context and returns an error (usually nil, to indicate everything went right).

### An instruction example
Let's create a new instruction, called `incmod`, that takes a register and adds 1 to it, but if it becomes greater than or equal to another register, it wraps around in the same way a mod would. Basically, `incmod rx, ry` is equal to `rx = (rx+1) % ry`.
This instruction is going to have opcode 0b111111, plus three bits for the first register, and three more for the second, resulting in 0b111111xxxyyydddd (where x is the first bits for the first register, y is for the second, and d is a don't care value, meaning an unused 0 or 1).

The entry in AllInstructions will be `{OpINCMOD, genRegM("incmod", 2), 1, execINCMOD},`, where OpINCMOD is defined above in the constant block as `OpINCMOD = 0b111111`. `genRegM` is used with those arguments to define the mnemonic and to say that it has two register opcodes in the usual position. 1 is to say that a single word is necessary to encode the instruction. execINCMOD is defined as follows:

```golang
func execINCMOD(pr *ICMCProcessor) error {
    // get the instruction binary data
    inst := pr.Data[pr.PC]

    // remember, those are indices from 0-7, not the actual values
    rx_index := getRegAt(inst, 7) // 7 is the first bit from right to left encoding the register index for rx
    ry_index := getRegAt(inst, 4) // same for ry

    // set the value in rx to the result of our operation
    pr.GPRRegs[rx_index] = (pr.GPRRegs[rx_index] + 1) % pr.GPRRegs[ry_index]

    // no errors happend
    return nil
}
```

## What you can do to help the GO ICMC Simulator's development
**An open source project is never complete.**

There are many ways in which you can participate in this project, for example:

* [Submit bugs and feature requests](https://github.com/lucasgpulcinelli/goICMCsim/issues), and help us verify as they are checked in
* Review [source code changes](https://github.com/lucasgpulcinelli/goICMCsim/pulls)
* [Test on macOS](https://github.com/lucasgpulcinelli/goICMCsim/labels/macOS%20test), and help us maintain compatibility on all platforms

If you're interested in solving problems and contributing directly to the code base, take a look at the [issue page](https://github.com/lucasgpulcinelli/goICMCsim/issues). And if you're making your first contribution, look for the tag [good first issue](https://github.com/lucasgpulcinelli/goICMCsim/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## Contributors
The main contributors to the project are listes here:
- Lucas Eduardo Gulka Pulcinelli ([github](https://github.com/lucasgpulcinelli))
- Isaac Santos Soares ([github](https://github.com/iss2718))

Special thanks to:
- Artur Brenner Weber ([github](https://github.com/ArturWeber)), for providing macOS/arm64 builds and helping with documentation
- Daniel Contente Romanzini ([github](https://github.com/Dauboau)), for providing macOS/amd64 builds
