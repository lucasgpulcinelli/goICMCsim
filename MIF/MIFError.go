package MIF

import "fmt"

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
