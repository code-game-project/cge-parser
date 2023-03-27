package parser

import (
	"fmt"
	"io"
)

type Config struct {
	IncludeComments bool
	OnlyMetadata    bool
	SendTokens      bool
	SendAST         bool
	NoObjects       bool
}

type Sender interface {
	SendMetadata(version string) error
}

type parser struct {
	out Sender

	config Config

	scanner *scanner
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
	for {
		t := p.scanner.nextToken()
		fmt.Println(t)
		if t.Type == TTEOF {
			break
		}
	}
	return nil
}
