package MIF

import "fmt"

type Token byte

const (
	TokNone    Token = iota // for errors
	TokEOF                  // EOF
	TokIdent                // identifiers
	TokEq                   // =
	TokStmtEnd              // ;
	TokContent              // CONTENT
	TokBegin                // BEGIN
	TokOpen                 // [
	TokClose                // ]
	TokNumber               // numbers
	TokRange                // ..
	TokColon                // :
	TokEnd                  // END
)

var TokMap = map[Token]string{
	TokNone:    "TokNone",
	TokEOF:     "TokEOF",
	TokIdent:   "TokIdent",
	TokEq:      "TokEq",
	TokStmtEnd: "TokStmtEnd",
	TokContent: "TokContent",
	TokBegin:   "TokBegin",
	TokOpen:    "TokOpen",
	TokClose:   "TokClose",
	TokNumber:  "TokNumber",
	TokRange:   "TokRange",
	TokColon:   "TokColon",
	TokEnd:     "TokEnd",
}

func (tok Token) String() string {
	s, ok := TokMap[tok]
	if !ok {
		return fmt.Sprintf("(unknown token %d)", int(tok))
	}
	return s
}
