package MIF

import (
	"fmt"
	"strconv"
)

// Format defines the possible formats the MIF data can have
type Format int

const (
	FormatNone    = 0
	FormatBin     = 2
	FormatUnsiged = 10
)

// formatMap is the mapping of format identifiers to their respective Formats.
var formatMap = map[string]Format{
	"BIN": FormatBin,
	"UNS": FormatUnsiged,
}

// readWithFormat reads a string containing a number in a specified MIF format
// if the format has not been provided (aka f == FormatNone), the value
// returned is implied from the prefix (0x, 0b, 0 or decimal).
func (p *Parser) readWithFormat(v string, f Format) (int64, error) {
	ret, err := strconv.ParseInt(v, int(f), 64)
	if err != nil {
		err = p.newError(fmt.Sprintf("invalid number: %s", err.Error()))
	}
	return ret, err
}
