package parser

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

type scanner struct {
	input *bufio.Scanner

	tokenBuffer *tokenBuffer

	line       int
	column     int
	tokenRunes []rune

	nextRune rune
}

func newScanner(input io.Reader) *scanner {
	inputScanner := bufio.NewScanner(input)
	inputScanner.Split(bufio.ScanRunes)
	s := &scanner{
		input:       inputScanner,
		tokenBuffer: newTokenBuffer(32),
		tokenRunes:  make([]rune, 0, 32),
	}
	s.nextChar()
	s.column = 0
	s.tokenRunes = s.tokenRunes[:0]
	return s
}

func (s *scanner) nextToken() Token {
	if s.tokenBuffer.length == 0 {
		s.scanToken()
	}
	return s.tokenBuffer.pop()
}

// peekToken(0) returns the next token to be returned by nextToken().
func (s *scanner) peekToken(offset int) Token {
	for s.tokenBuffer.length <= offset {
		s.scanToken()
	}
	return s.tokenBuffer.peek(offset)
}

func (s *scanner) scanToken() {
	c := s.nextChar()

	for {
		if c == '\000' {
			s.addToken(TTEOF)
			return
		}

		if unicode.IsSpace(c) {
			s.tokenRunes = s.tokenRunes[:0]
			if c == '\n' {
				s.newLine()
			}
			c = s.nextChar()
			continue
		}

		switch c {
		case '{':
			s.addToken(TTOpenCurly)
		case '}':
			s.addToken(TTCloseCurly)
		case '/':
			if s.match('/') {
				s.comment()
			} else if s.match('*') {
				s.blockComment()
			}
		case ':':
			s.addToken(TTColon)
		case ',':
			s.addToken(TTComma)
		case '<':
			s.addToken(TTLess)
		case '>':
			s.addToken(TTGreater)
		default:
			if isLowerAlpha(c) {
				s.identifier()
			} else if isDigit(c) {
				err := s.versionNumber()
				if err != nil {
					s.newErrorAtNext(err.Error())
				}
			} else {
				s.newErrorAtPrev(fmt.Sprintf("unexpected character '%c'", c))
			}
		}
		break
	}
}

func (s *scanner) identifier() {
	for isLowerAlphaNum(s.peekChar()) {
		s.nextChar()
	}

	name := string(s.tokenRunes)
	switch name {
	case "name":
		s.addToken(TTGameName)
	case "version", "cge":
		s.addToken(TTCGEVersion)
	case "config":
		s.addToken(TTConfig)
	case "event":
		s.addToken(TTEvent)
	case "command":
		s.addToken(TTCommand)
	case "type":
		s.addToken(TTType)
	case "enum":
		s.addToken(TTEnum)
	case "string":
		s.addToken(TTString)
	case "bool":
		s.addToken(TTBool)
	case "int", "int32":
		s.addToken(TTInt32)
	case "int64":
		s.addToken(TTInt64)
	case "float32":
		s.addToken(TTFloat32)
	case "float", "float64":
		s.addToken(TTFloat64)
	case "list":
		s.addToken(TTList)
	case "map":
		s.addToken(TTMap)
	default:
		s.addToken(TTIdentifier)
	}
}

func (s *scanner) versionNumber() error {
	for isDigit(s.peekChar()) {
		s.nextChar()
	}

	if !s.match('.') {
		return fmt.Errorf("expected '.' after major version")
	}

	if !isDigit(s.peekChar()) {
		return fmt.Errorf("expected digit after '.'")
	}
	for isDigit(s.peekChar()) {
		s.nextChar()
	}

	s.addToken(TTVersionNumber)
	return nil
}

func (s *scanner) comment() {
	for s.peekChar() != '\n' && s.peekChar() != '\000' {
		s.nextChar()
	}
	s.addToken(TTComment)
}

func (s *scanner) blockComment() {
	line := s.line
	column := s.column - 2
	for s.peekChar() != '\000' {
		c := s.nextChar()
		if c == '*' && s.match('/') {
			break
		}
		if c == '\n' {
			s.line++
			s.column = 0
		}
	}
	s.addTokenWithPos(TTComment, line, column)
}

func (s *scanner) match(char rune) bool {
	if s.peekChar() != char {
		return false
	}
	s.nextChar()
	return true
}

func (s *scanner) nextChar() rune {
	current := s.nextRune
	for {
		if !s.input.Scan() {
			if err := s.input.Err(); err != nil {
				panic(fmt.Errorf("failed to read input data: %w", err))
			}
			s.nextRune = '\000'
			return current
		}
		if s.input.Bytes()[0] != '\r' {
			break
		}
	}
	r, _ := utf8.DecodeRune(s.input.Bytes())
	s.nextRune = r
	s.column++
	s.tokenRunes = append(s.tokenRunes, current)
	return current
}

func (s *scanner) peekChar() rune {
	return s.nextRune
}

func (s *scanner) addToken(tokenType TokenType) {
	lexeme := string(s.tokenRunes)
	s.addTokenWithPos(tokenType, s.line, s.column-len(lexeme))
}

func (s *scanner) addTokenWithPos(tokenType TokenType, line, column int) {
	lexeme := string(s.tokenRunes)
	s.tokenBuffer.push(Token{
		Line:   line,
		Column: column,
		Type:   tokenType,
		Lexeme: lexeme,
	})
	s.tokenRunes = s.tokenRunes[:0]
}

func (s *scanner) newLine() {
	s.line++
	s.column = 0
	s.tokenRunes = s.tokenRunes[:0]
}

func (s *scanner) newErrorAtNext(message string) {
	s.tokenBuffer.push(Token{
		Line:   s.line,
		Column: s.column,
		Type:   TTError,
		Lexeme: message,
	})
	s.tokenRunes = s.tokenRunes[:0]
}

func (s *scanner) newErrorAtPrev(message string) {
	s.tokenBuffer.push(Token{
		Line:   s.line,
		Column: s.column - 1,
		Type:   TTError,
		Lexeme: message,
	})
	s.tokenRunes = s.tokenRunes[:0]
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isLowerAlpha(char rune) bool {
	return char >= 'a' && char <= 'z' || char == '_'
}

func isLowerAlphaNum(char rune) bool {
	return isDigit(char) || isLowerAlpha(char)
}

type tokenBuffer struct {
	tokens  []Token
	length  int
	current int
}

func newTokenBuffer(size int) *tokenBuffer {
	return &tokenBuffer{
		tokens: make([]Token, size),
	}
}

func (t *tokenBuffer) push(token Token) {
	if t.length == len(t.tokens) {
		t.grow()
	}
	t.tokens[(t.current+t.length)%len(t.tokens)] = token
	t.length++
}

func (t *tokenBuffer) pop() Token {
	if t.length == 0 {
		panic("cannot pop from empty buffer")
	}
	ret := t.tokens[t.current]
	t.current = (t.current + 1) % len(t.tokens)
	t.length--
	return ret
}

func (t *tokenBuffer) peek(offset int) Token {
	if t.length <= offset {
		panic("cannot peek behind end of buffer")
	}
	return t.tokens[(t.current+offset)%len(t.tokens)]
}

func (t *tokenBuffer) grow() {
	newBuf := make([]Token, len(t.tokens)*2)
	if t.current+t.length > len(t.tokens) {
		firstHalf := t.tokens[t.current:]
		copy(newBuf, firstHalf)
		copy(newBuf[len(firstHalf):], t.tokens[:t.length-len(firstHalf)])
	} else {
		copy(newBuf, t.tokens[t.current:t.current+t.length])
	}
	t.current = 0
	t.tokens = newBuf
}
