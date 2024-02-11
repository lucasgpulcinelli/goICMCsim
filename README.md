# 🖥️ GO ICMC Simulator

![GO ICMC Simulator Cover](https://github.com/lucasgpulcinelli/goICMCsim/assets/11618151/da81d732-5cb4-4f41-9128-37ae864ceac9)

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/lucasgpulcinelli/goICMCsim?logo=go"/>
  <a href="https://github.com/lucasgpulcinelli/goICMCsim/issues?q=is%3Aopen+is%3Aissue+label%3Afeature-request+sort%3Areactions-%2B1-desc">
    <img src="https://img.shields.io/github/issues/lucasgpulcinelli/goICMCsim/feature-request.svg">
  </a>
  <a href="https://github.com/lucasgpulcinelli/goICMCsim/issues?utf8=✓&q=is%3Aissue+is%3Aopen+label%3Abug">
    <img src="https://img.shields.io/github/issues/lucasgpulcinelli/goICMCsim/bug.svg">
  </a>
  <a href="https://github.com/lucasgpulcinelli/goICMCsim/releases">
    <img src="https://img.shields.io/github/v/release/lucasgpulcinelli/goICMCsim"/>
  </a>
  <img src="https://img.shields.io/github/license/lucasgpulcinelli/goICMCsim"/>
</p>

## 📝 Overview
This program is a simulator for the [ICMC architecture](https://github.com/simoesusp/Processador-ICMC/). It features several upgraded functionalities compared to the C++ simulator, including:

- A resizable window and fullscreen capability.
- An instruction scroll to view all instructions and data being modified in real-time.
- Ability to edit the stack pointer and program counter directly.
- Enhanced error handling: the simulator will halt and indicate errors to the programmer.
- Improved parsing of MIF files, adhering strictly to syntax definition and providing detailed error messages.
- Capability to change character mapping MIF during runtime (without resetting).
- Shortcuts that do not rely on keys that may not be present on laptop keyboards (e.g., insert, home, and end keys).
- Support for Windows, macOS, and Linux.

## 💻 Installation
If you prefer not to compile anything, you can download a precompiled binary for your system from the [releases page](https://github.com/lucasgpulcinelli/goICMCsim/releases).

## 🚀 Usage
To get started, add a program to run and test it. You can specify MIF files in the ICMC architecture format in the command line or use the file -> open code/char MIF menu. Always specify a char MIF to ensure proper output of characters.

## 🛠️ How to Compile from Source Code
1. Install a recent version of Go (at least 1.13) from [here](https://go.dev/doc/install).
2. Install Git and a C compiler (on Windows, use MinGW).
3. On Debian/Ubuntu-based systems, install `libgl1-mesa-dev xorg-dev`; on Fedora and Red Hat-based systems, install `libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel`.
4. Clone the repository and navigate to the project directory.
5. Run `go build .` to compile and `./goICMCsim` to start an empty processor. Use `--help` to see command line options.
6. Optionally, you can install directly into `$GOPATH/bin` with `go install github.com/lucasgpulcinelli/goICMCsim@latest`.

## ⚙️ How to Add/Modify Instructions in the Simulator
To add or modify instructions:
1. Choose an opcode for your instruction.
2. Add it to the constants list in `processor/Instruction.go`.
3. Add your instruction data to the `AllInstructions` list in the same file, including the opcode, mnemonic string, instruction size, and execution function.
4. Implement the execution function. See the example in the [documentation](docs/README.md/#go-icmc-simulator-documentation) for details.

## 🤝 Contributing
An open-source project is never complete. You can contribute by:

- [Submitting bugs and feature requests](https://github.com/lucasgpulcinelli/goICMCsim/issues).
- Reviewing [source code changes](https://github.com/lucasgpulcinelli/goICMCsim/pulls).
- [Testing on macOS](https://github.com/lucasgpulcinelli/goICMCsim/labels/macOS%20test) and helping maintain compatibility on all platforms.

If you're interested in solving problems and contributing directly to the codebase, check out the [issue page](https://github.com/lucasgpulcinelli/goICMCsim/issues) and look for [good first issues](https://github.com/lucasgpulcinelli/goICMCsim/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## ✨ Contributors

### 🏆 Main contributors to the project:

- [Lucas Eduardo Gulka Pulcinelli](https://github.com/lucasgpulcinelli)
- [Isaac Santos Soares](https://github.com/iss2718)

### ♥️ Special thanks to:

- [Artur Brenner Weber](https://github.com/ArturWeber) for providing macOS/arm64 builds and assisting with documentation.
- [Daniel Contente Romanzini](https://github.com/Dauboau) for providing macOS/amd64 builds.
