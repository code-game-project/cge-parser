package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/code-game-project/cge-parser/cge"
)

type Config struct {
	IncludeComments bool
	OnlyMetadata    bool
	SendTokens      bool
	SendAST         bool
	NoObjects       bool
}

type DiagnosticType int32

const (
	DiagnosticInfo DiagnosticType = iota
	DiagnosticWarning
	DiagnosticError
)

type Sender interface {
	SendMetadata(version string) error
	SendDiagnostic(diagnosticType DiagnosticType, message string, startLine, startColumn, endLine, endColumn int) error
	SendToken(tokenType TokenType, lexeme string, line, column int)
}

type parser struct {
	out Sender

	config Config

	scanner *scanner

	previous Token
}

func Parse(input io.Reader, output Sender, config Config) error {
	p := parser{
		out:     output,
		config:  config,
		scanner: newScanner(input),
	}
	return p.parse()
}

func (p *parser) parse() error {
	err := p.metadata()
	if err != nil {
		if _, ok := err.(ParserError); ok {
			err = nil
		}
		return err
	}
	if p.config.OnlyMetadata {
		return nil
	}
	return nil
}

func (p *parser) metadata() error {
	for p.match(TTComment) {
		p.warn(p.previous, "game comments are deprecated")
	}

	if p.match(TTGameName) {
		p.warn(p.previous, "the 'name' metadata field is deprecated")
		if !p.match(TTIdentifier) {
			return p.error(p.peek(0), "expected identifier after 'name' keyword", false)
		}
	}

	var version Token
	for p.match(TTCGEVersion) {
		switch p.previous.Type {
		case TTCGEVersion:
			if version.Lexeme != "" {
				return p.error(p.previous, "duplicate 'cge' metadata field", false)
			}
			if !p.match(TTVersionNumber) {
				return p.error(p.peek(0), "expected version number after 'cge' keyword", false)
			}
			version = p.previous
			if version.Lexeme == "version" {
				p.warn(version, "the 'version' metadata field is deprecated; use 'cge' instead")
			}
		}
	}
	if version.Lexeme == "" {
		return p.error(p.peek(0), "missing required 'cge' metadata field", false)
	}
	if !p.config.OnlyMetadata && !isVersionCompatible(version.Lexeme, cge.CGEVersion) {
		return p.error(version, fmt.Sprintf("incompatible CGE version (file: %s, parser: %s)", version.Lexeme, cge.CGEVersion), false)
	}

	return p.out.SendMetadata(version.Lexeme)
}

func (p *parser) advance() Token {
	token := p.scanner.nextToken()

	if p.config.SendTokens && token.Type != TTError {
		p.out.SendToken(token.Type, token.Lexeme, token.Line, token.Column)
	}

	p.previous = token
	return token
}

func (p *parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.peek(0).Type == t {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) peek(offset int) Token {
	return p.scanner.peekToken(offset)
}

func (p *parser) skipBlock(inBlock bool) {
	if p.peek(0).Type == TTEOF {
		return
	}

	if !inBlock {
		for {
			if p.match(TTEOF) {
				return
			}
			if p.match(TTOpenCurly) {
				break
			}
			p.advance()
		}
	}

	nestingLevel := 1
	for p.peek(0).Type != TTEOF && nestingLevel > 0 {
		if p.peek(0).Type == TTOpenCurly {
			nestingLevel++
		} else if p.peek(0).Type == TTCloseCurly {
			nestingLevel--
		}
		p.advance()
	}
}

func (p *parser) skipProperty() {
	if p.peek(0).Type == TTEOF {
		return
	}

	nestingLevel := 0
	for p.peek(0).Type != TTEOF {
		if p.peek(0).Type == TTOpenCurly {
			nestingLevel++
		} else if p.peek(0).Type == TTCloseCurly {
			nestingLevel--
			if nestingLevel == -1 {
				return
			}
		}
		if nestingLevel == 0 && p.match(TTComma) {
			return
		}
		p.advance()
	}
}

func (p *parser) warn(token Token, message string) {
	err := p.out.SendDiagnostic(DiagnosticWarning, message, token.Line, token.Column, token.Line, token.Column+len(token.Lexeme))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send warning '[%d:%d] %s': %s", token.Line, token.Column, message, err)
	}
}

type ParserError struct {
	Token   Token
	Message string
	inBlock bool
}

func (p ParserError) Error() string {
	return p.Message
}

func (p *parser) error(token Token, message string, inBlock bool) error {
	if token.Type == TTError {
		message = token.Lexeme
		token.Lexeme = " "
	}
	perr := ParserError{
		Token:   token,
		Message: message,
		inBlock: inBlock,
	}
	err := p.out.SendDiagnostic(DiagnosticError, message, token.Line, token.Column, token.Line, token.Column+len(token.Lexeme))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send error '[%d:%d] %s': %s", token.Line, token.Column, message, err)
	}
	return perr
}

func isVersionCompatible(fileVersion, parserVersion string) bool {
	if parserVersion == "dev" {
		return true
	}

	fileParts := strings.Split(fileVersion, ".")
	parserParts := strings.Split(parserVersion, ".")

	if fileParts[0] != parserParts[0] {
		return false
	}

	if parserParts[0] == "0" && fileParts[1] != parserParts[1] {
		return false
	}

	fileMinor, err := strconv.Atoi(fileParts[1])
	if err != nil {
		return false
	}

	parserMinor, err := strconv.Atoi(parserParts[1])
	if err != nil {
		return false
	}

	return parserMinor >= fileMinor
}
