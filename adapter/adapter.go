package adapter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/code-game-project/cge-parser/parser"
)

// ParseMetadata reads the header of the CGE file and returns the metadata fields and a new io.Reader (even present if err != nil),
// which still contains the header enabling it to be used for complete parsing.
func ParseMetadata(file io.Reader) (Metadata, io.Reader, []Diagnostic, error) {
	reader := newPeekReader(file, 1024)

	var metadata Metadata
	diagnostics := make([]Diagnostic, 0)

	err := parser.Parse(reader, &callbackSender{
		CBMetadata: func(cgeVersion string) {
			metadata = Metadata{
				CGEVersion: cgeVersion,
			}
		},
		CBDiagnostic: func(diagnosticType parser.DiagnosticType, message string, startLine, startCol, endLine, endCol int) {
			if diagnosticType == parser.DiagnosticError {
				diagnostics = append(diagnostics, Diagnostic{
					Type:        DiagnosticType(diagnosticType),
					Message:     message,
					StartLine:   startLine,
					StartColumn: startCol,
					EndLine:     endLine,
					EndColumn:   endCol,
				})
			}
		},
	}, parser.Config{
		OnlyMetadata:    true,
		IncludeComments: false,
		SendTokens:      false,
		NoObjects:       true,
		DisableWarnings: true,
	})
	if len(diagnostics) > 0 && err == nil {
		err = errors.New("parsing error")
	}
	return metadata, io.MultiReader(reader.buffer, file), diagnostics, err
}

type Config struct {
	IncludeComments bool
	SendTokens      bool
	NoObjects       bool
	DisableWarnings bool
}

func (c Config) toArgs() []string {
	args := make([]string, 0, 2)
	if c.IncludeComments {
		args = append(args, "--comments")
	}
	if c.SendTokens {
		args = append(args, "--tokens")
	}
	if c.NoObjects {
		args = append(args, "--no-objects")
	}
	if c.DisableWarnings {
		args = append(args, "--no-warn")
	}
	return args
}

type ParserResponse struct {
	Metadata    Metadata
	Objects     []Object
	Tokens      []Token
	Diagnostics []Diagnostic
}

func ParseCGE(file io.Reader, cgeParserPath string, config Config) (ParserResponse, []error) {
	cmd := exec.Command(cgeParserPath, config.toArgs()...)
	cmd.Stdin = file
	reader, writer := io.Pipe()
	cmd.Stdout = writer

	errBuf := new(bytes.Buffer)
	cmd.Stderr = errBuf

	err := cmd.Start()
	if err != nil {
		return ParserResponse{}, []error{fmt.Errorf("failed to start cge-parser: %w", err)}
	}

	var response ParserResponse
	receiveDone := make(chan struct{})
	go func() {
		response, err = receiveProtobufs(reader, config)
		receiveDone <- struct{}{}
	}()

	cmdErr := cmd.Wait()
	writer.Close()

	<-receiveDone

	if cmdErr != nil {
		if errBuf.Len() > 0 {
			scanner := bufio.NewScanner(errBuf)
			errs := make([]error, 0, 1)
			for scanner.Scan() {
				errs = append(errs, errors.New(scanner.Text()))
			}
			if len(errs) == 0 {
				errs = append(errs, fmt.Errorf("failed to execute cge-parser"))
			}
			return ParserResponse{}, errs
		} else {
			return ParserResponse{}, []error{fmt.Errorf("failed to execute cge-parser: %w", cmdErr)}
		}
	}

	if err != nil {
		return ParserResponse{}, []error{fmt.Errorf("failed to receive data from cge-parser: %w", err)}
	}

	return response, nil
}
