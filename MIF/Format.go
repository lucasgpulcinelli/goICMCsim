package MIF

import (
	"fmt"
	"strconv"
)

type Format int

const (
	FormatNone = iota
	FormatBin
	FormatUnsiged
)

func (p *Parser) toFormat(v string) (Format, error) {
	switch v {
	case "UNS":
		return FormatUnsiged, nil
	case "BIN":
		return FormatBin, nil
	}

	return FormatNone, p.newError(fmt.Sprintf("expected Format type, not %s", v))
}

func (p *Parser) readWithFormat(v string, f Format) (ret int64, err error) {
	switch f {
	case FormatNone:
		return 0, p.newError("invalid format provided")
	case FormatBin:
		ret, err = strconv.ParseInt(v, 2, 64)
		if err != nil {
			return
		}
  case FormatUnsiged:
		ret, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return
		}
	}
	return
}
