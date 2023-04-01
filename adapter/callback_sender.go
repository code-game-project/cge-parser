package adapter

import (
	"github.com/code-game-project/cge-parser/parser"
)

type callbackSender struct {
	CBMetadata   func(cgeVersion string)
	CBDiagnostic func(diagnosticType parser.DiagnosticType, message string, startLine, startColumn, endLine, endColumn int)
	CBToken      func(tokenType parser.TokenType, lexeme string, line, column int)
	CBObject     func(object parser.Object)
}

func (c *callbackSender) SendMetadata(cgeVersion string) error {
	if c.CBMetadata != nil {
		c.CBMetadata(cgeVersion)
	}
	return nil
}

func (c *callbackSender) SendDiagnostic(diagnosticType parser.DiagnosticType, message string, startLine, startColumn, endLine, endColumn int) error {
	if c.CBDiagnostic != nil {
		c.CBDiagnostic(diagnosticType, message, startLine, startColumn, endLine, endColumn)
	}
	return nil
}

func (c *callbackSender) SendToken(tokenType parser.TokenType, lexeme string, line, column int) error {
	if c.CBToken != nil {
		c.CBToken(tokenType, lexeme, line, column)
	}
	return nil
}

func (c *callbackSender) SendObject(object parser.Object) error {
	if c.CBObject != nil {
		c.CBObject(object)
	}
	return nil
}
