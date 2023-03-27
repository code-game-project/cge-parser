package parser

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Column int
}

type TokenType string

const (
	TTGameName   TokenType = "GAME_NAME"
	TTCGEVersion TokenType = "CGE_VERSION"

	TTConfig  TokenType = "CONFIG"
	TTCommand TokenType = "COMMAND"
	TTEvent   TokenType = "EVENT"
	TTType    TokenType = "TYPE"
	TTEnum    TokenType = "ENUM"

	TTString  TokenType = "STRING"
	TTBool    TokenType = "BOOL"
	TTInt32   TokenType = "INT32"
	TTInt64   TokenType = "INT64"
	TTFloat32 TokenType = "FLOAT32"
	TTFloat64 TokenType = "FLOAT64"

	TTMap  TokenType = "MAP"
	TTList TokenType = "LIST"

	TTIdentifier    TokenType = "IDENTIFIER"
	TTVersionNumber TokenType = "VERSION_NUMBER"

	TTOpenCurly  TokenType = "OPEN_CURLY"
	TTCloseCurly TokenType = "CLOSE_CURLY"
	TTColon      TokenType = "COLON"
	TTComma      TokenType = "COMMA"
	TTGreater    TokenType = "GREATER"
	TTLess       TokenType = "LESS"

	TTComment TokenType = "COMMENT"

	TTError TokenType = "ERROR"
	TTEOF   TokenType = "EOF"
)
