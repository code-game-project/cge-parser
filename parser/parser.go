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
	NoObjects       bool
	DisableWarnings bool
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
	SendToken(tokenType TokenType, lexeme string, line, column int) error
	SendObject(object Object) error
}

type ObjectType string

type Object struct {
	Comment    string
	Type       TokenType
	Name       Token
	Properties []Property
}

func (o Object) String() string {
	var text string
	if o.Comment != "" {
		text = "// " + o.Comment + "\n"
	}
	text = fmt.Sprintf("%s%s {", text, o.Name.Lexeme)

	for _, p := range o.Properties {
		text = fmt.Sprintf("%s\n%s,", text, p)
	}

	return text + "\n}"
}

type Property struct {
	Comment string
	Name    string
	Type    *PropertyType
}

type PropertyType struct {
	Token   Token
	Generic *PropertyType
}

func (p Property) String() string {
	if p.Type == nil {
		return p.Name
	}
	return fmt.Sprintf("%s: %s", p.Name, p.Type.Token.Lexeme)
}

type parser struct {
	out Sender

	config Config

	scanner *scanner

	previous Token

	objects                 []Object
	commands                map[string]struct{}
	events                  map[string]struct{}
	types                   map[string]struct{}
	configObj               bool
	accessedTypeIdentifiers []Token

	hadError bool
}

func Parse(input io.Reader, output Sender, config Config) error {
	p := parser{
		out:                     output,
		config:                  config,
		scanner:                 newScanner(input),
		objects:                 make([]Object, 0, 32),
		commands:                make(map[string]struct{}),
		events:                  make(map[string]struct{}),
		types:                   make(map[string]struct{}),
		accessedTypeIdentifiers: make([]Token, 0),
	}
	return p.parse()
}

func (p *parser) parse() (err error) {
	defer func() {
		r := recover()
		if e, ok := r.(error); ok {
			err = e
		} else if r != nil {
			panic(r)
		}
	}()

	err = p.metadata()
	if err != nil {
		if _, ok := err.(ParserError); ok {
			err = nil
		}
		return err
	}
	if p.config.OnlyMetadata {
		return nil
	}

	for p.peek(0).Type != TTEOF {
		decl, err := p.declaration()
		if err != nil {
			if e, ok := err.(ParserError); ok {
				p.skipBlock(e.inBlock)
			}
			continue
		}
		p.objects = append(p.objects, decl)
	}

	for _, id := range p.accessedTypeIdentifiers {
		if _, ok := p.types[id.Lexeme]; !ok {
			p.error(id, fmt.Sprintf("undefined type '%s'.", id.Lexeme), true)
		}
	}

	if !p.configObj {
		p.objects = append(p.objects, Object{
			Type: TTConfig,
		})
	}

	p.detectDeclarationCycles()

	if !p.config.NoObjects && !p.hadError {
		for _, o := range p.objects {
			err := p.out.SendObject(o)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to send object: %s", err)
				break
			}
		}
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

	err := p.out.SendMetadata(version.Lexeme)
	if err != nil {
		return err
	}

	if !p.config.OnlyMetadata && !isVersionCompatible(version.Lexeme, cge.CGEVersion) {
		return p.error(version, fmt.Sprintf("incompatible CGE version (file: %s, parser: %s)", version.Lexeme, cge.CGEVersion), false)
	}

	return nil
}

func (p *parser) declaration() (Object, error) {
	comment := p.comment()

	if !p.match(TTConfig, TTCommand, TTEvent, TTType, TTEnum) {
		if p.peek(0).Type == TTComment {
			return Object{}, p.error(p.peek(0), "comment does not belong to an object", false)
		}
		return Object{}, p.error(p.peek(0), "expected type declaration", false)
	}

	objectKeyword := p.previous

	if objectKeyword.Type == TTConfig {
		if p.configObj {
			return Object{}, p.error(p.previous, "duplicate config object", false)
		}
		p.configObj = true
	} else {
		if !p.match(TTIdentifier) {
			return Object{}, p.error(p.peek(0), fmt.Sprintf("expected identifier after '%s' keyword.", p.previous.Lexeme), false)
		}
	}
	name := p.previous

	switch objectKeyword.Type {
	case TTCommand:
		if _, ok := p.commands[name.Lexeme]; ok {
			return Object{}, p.error(name, fmt.Sprintf("command '%s' already defined", name.Lexeme), false)
		}
		p.commands[name.Lexeme] = struct{}{}
	case TTEvent:
		if _, ok := p.events[name.Lexeme]; ok {
			return Object{}, p.error(name, fmt.Sprintf("event '%s' already defined", name.Lexeme), false)
		}
		p.events[name.Lexeme] = struct{}{}
	case TTType, TTEnum:
		if _, ok := p.types[name.Lexeme]; ok {
			return Object{}, p.error(name, fmt.Sprintf("type '%s' already defined", name.Lexeme), false)
		}
		p.types[name.Lexeme] = struct{}{}
	}

	if !p.match(TTOpenCurly) {
		if objectKeyword.Type == TTConfig {
			return Object{}, p.error(p.peek(0), fmt.Sprintf("expected block after '%s' keyword", objectKeyword.Lexeme), true)
		} else {
			return Object{}, p.error(p.peek(0), fmt.Sprintf("expected block after %s name", objectKeyword.Lexeme), true)
		}
	}

	var properties []Property
	var err error
	if objectKeyword.Type == TTEnum {
		properties, err = p.enumBlock()
	} else {
		properties, err = p.block()
	}
	if err != nil {
		return Object{}, err
	}

	return Object{
		Comment:    comment,
		Type:       objectKeyword.Type,
		Name:       name,
		Properties: properties,
	}, nil
}

func (p *parser) block() ([]Property, error) {
	properties := make([]Property, 0)

	for p.peek(0).Type != TTEOF && p.peek(0).Type != TTCloseCurly {
		property, err := p.property()
		if err != nil {
			p.skipProperty()
			continue
		}
		properties = append(properties, property)
		if !p.match(TTComma) {
			break
		}
	}

	if !p.match(TTCloseCurly) {
		return nil, p.error(p.peek(0), "expected '}' after block", true)
	}

	return properties, nil
}

func (p *parser) enumBlock() ([]Property, error) {
	properties := make([]Property, 0)

	for p.peek(0).Type != TTEOF && p.peek(0).Type != TTCloseCurly {
		property, err := p.enumValue()
		if err != nil {
			p.skipProperty()
			continue
		}
		properties = append(properties, property)
		if !p.match(TTComma) {
			break
		}
	}

	if !p.match(TTCloseCurly) {
		return nil, p.error(p.peek(0), "expected '}' after block", true)
	}

	return properties, nil
}

func (p *parser) property() (Property, error) {
	comment := p.comment()

	if !p.match(TTIdentifier) {
		return Property{}, p.error(p.peek(0), "expected property name", true)
	}
	name := p.previous

	if !p.match(TTColon) {
		return Property{}, p.error(p.peek(0), "expected ':' after property name", true)
	}

	propertyType, err := p.propertyType()
	if err != nil {
		return Property{}, err
	}

	return Property{
		Comment: comment,
		Name:    name.Lexeme,
		Type:    propertyType,
	}, nil
}

func (p *parser) enumValue() (Property, error) {
	comment := p.comment()

	if !p.match(TTIdentifier) {
		return Property{}, p.error(p.peek(0), "expected property name", true)
	}
	name := p.previous

	return Property{
		Comment: comment,
		Name:    name.Lexeme,
	}, nil
}

func (p *parser) propertyType() (*PropertyType, error) {
	if !p.match(TTString, TTBool, TTInt32, TTInt64, TTFloat32, TTFloat64, TTMap, TTList, TTIdentifier, TTType, TTEnum) {
		return &PropertyType{}, p.error(p.peek(0), "expected type after property name", true)
	}

	propertyType := p.previous
	var generic *PropertyType

	switch propertyType.Type {
	case TTIdentifier:
		p.accessedTypeIdentifiers = append(p.accessedTypeIdentifiers, propertyType)
	case TTType, TTEnum:
		if !p.match(TTIdentifier) {
			return &PropertyType{}, p.error(p.peek(0), "expected identifier after 'type' keyword", true)
		}

		identifier := p.previous
		if _, ok := p.types[identifier.Lexeme]; ok {
			return &PropertyType{}, p.error(identifier, fmt.Sprintf("type '%s' is already defined", identifier.Lexeme), true)
		}
		p.types[identifier.Lexeme] = struct{}{}

		if !p.match(TTOpenCurly) {
			return &PropertyType{}, p.error(p.peek(0), "expected block after type name", true)
		}

		var properties []Property
		var err error
		if propertyType.Type == TTType {
			properties, err = p.block()
		} else {
			properties, err = p.enumBlock()
		}
		if err != nil {
			return &PropertyType{}, err
		}

		p.objects = append(p.objects, Object{
			Type:       propertyType.Type,
			Name:       identifier,
			Properties: properties,
		})

		propertyType = identifier
	case TTMap, TTList:
		if !p.match(TTLess) {
			return &PropertyType{}, p.error(p.peek(0), "expected generic", true)
		}

		var err error
		generic, err = p.propertyType()
		if err != nil {
			return &PropertyType{}, err
		}

		if !p.match(TTGreater) {
			return &PropertyType{}, p.error(p.peek(0), "expected '>' after generic value", true)
		}
	}

	return &PropertyType{
		Token:   propertyType,
		Generic: generic,
	}, nil
}

func (p *parser) comment() string {
	var comments []string
	for p.match(TTComment) {
		if p.config.IncludeComments {
			comments = append(comments, p.previous.Lexeme)
		}
	}
	return strings.Join(comments, "\n")
}

func (p *parser) advance() Token {
	token := p.scanner.nextToken()

	if p.config.SendTokens && token.Type != TTError {
		err := p.out.SendToken(token.Type, token.Lexeme, token.Line, token.Column)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to send token: %s", err)
		}
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
	if p.config.DisableWarnings {
		return
	}
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
	p.hadError = true
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
