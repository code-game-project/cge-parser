package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/code-game-project/cge-parser/parser"
	"github.com/code-game-project/cge-parser/protobuf"
)

func run() error {
	comments := pflag.Bool("comments", false, "include doc comments in output")
	onlyMeta := pflag.Bool("only-meta", false, "stop parsing after sending the metadata message")
	tokens := pflag.Bool("tokens", false, "return all parsed tokens")
	ast := pflag.Bool("ast", false, "return the complete AST")
	noObjects := pflag.Bool("no-objects", false, "do not return objects")
	pflag.Parse()

	return parser.Parse(os.Stdin, protobuf.NewSender(os.Stdout), parser.Config{
		IncludeComments: *comments,
		OnlyMetadata:    *onlyMeta,
		SendTokens:      *tokens,
		SendAST:         *ast,
		NoObjects:       *noObjects,
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
