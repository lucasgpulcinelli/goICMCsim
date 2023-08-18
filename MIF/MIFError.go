package MIF

import "fmt"

// MIFError denotes a parsing or lexer error at a certain position of a MIF
// file.
type MIFError struct {
	component string
	line      int
	col       int
	cause     string
}

func (e MIFError) Error() string {
	return fmt.Sprintf("%s failed at line %d, col %d: %s", e.component, e.line,
		e.col, e.cause)
}
