package protobuf

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/code-game-project/cge-parser/parser"
	"github.com/code-game-project/cge-parser/protobuf/schema"
)

type ProtobufSender struct {
	out io.Writer
}

func NewSender(out io.Writer) *ProtobufSender {
	return &ProtobufSender{
		out: out,
	}
}

func (p *ProtobufSender) SendMetadata(cgeVersion string) error {
	data, err := proto.Marshal(&schema.Metadata{
		CgeVersion: cgeVersion,
	})
	if err != nil {
		return fmt.Errorf("failed to encode metadata as protobuf message: %w", err)
	}

	_, err = p.out.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write metadata message: %w", err)
	}

	return nil
}

func (p *ProtobufSender) SendDiagnostic(diagnosticType parser.DiagnosticType, message string, startLine, startColumn, endLine, endColumn int) error {
	data, err := proto.Marshal(&schema.Diagnostic{
		Type: schema.Diagnostic_Type(diagnosticType),
		Msg:  message,
		Start: &schema.Pos{
			Line:   int32(startLine),
			Column: int32(startColumn),
		},
		End: &schema.Pos{
			Line:   int32(endLine),
			Column: int32(endColumn),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to encode diagnostic as protobuf message: %w", err)
	}

	_, err = p.out.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write diagnostic message: %w", err)
	}

	return nil
}

func (p *ProtobufSender) SendToken(tokenType parser.TokenType, lexeme string, line, column int) {
	data, err := proto.Marshal(&schema.Token{
		Type:   string(tokenType),
		Lexeme: lexeme,
		Pos: &schema.Pos{
			Line:   int32(line),
			Column: int32(column),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode token as protobuf message: %s", err)
		return
	}

	_, err = p.out.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write token message: %s", err)
		return
	}
}
