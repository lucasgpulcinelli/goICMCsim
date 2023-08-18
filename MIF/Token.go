package MIF

import "fmt"

// Token identifies a single syntatical unit in a MIF program.
type Token byte

const (
	TokNone Token = iota // for errors
	TokEOF
	TokIdent
	TokEq
	TokStmtEnd
	TokContent
	TokBegin
	TokOpen
	TokClose
	TokNumber
	TokRange
	TokColon
	TokEnd
)

var TokMap = map[Token]string{
	TokNone:    "<TokNone>",
	TokEOF:     "EOF",
	TokIdent:   "identifier",
	TokEq:      "'='",
	TokStmtEnd: "';'",
	TokContent: "CONTENT",
	TokBegin:   "BEGIN",
	TokOpen:    "'['",
	TokClose:   "']'",
	TokNumber:  "number",
	TokRange:   "'..'",
	TokColon:   "':'",
	TokEnd:     "END",
}

func (tok Token) String() string {
	s, ok := TokMap[tok]
	if !ok {
		return fmt.Sprintf("(unknown token %d)", int(tok))
	}
	return s
}
