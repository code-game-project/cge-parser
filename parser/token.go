package parser

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Column int
}

type TokenType int

const (
	TTGameName TokenType = iota
	TTCGEVersion

	TTConfig
	TTCommand
	TTEvent
	TTType
	TTEnum

	TTString
	TTBool
	TTInt32
	TTInt64
	TTFloat32
	TTFloat64

	TTMap
	TTList

	TTIdentifier
	TTVersionNumber

	TTOpenCurly
	TTCloseCurly
	TTColon
	TTComma
	TTGreater
	TTLess

	TTComment

	TTError
	TTEOF
)
