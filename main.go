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
	noObjects := pflag.Bool("no-objects", false, "do not return objects")
	noWarn := pflag.Bool("no-warn", false, "disable warnings")
	pflag.Parse()

	return parser.Parse(os.Stdin, protobuf.NewSender(os.Stdout), parser.Config{
		IncludeComments: *comments,
		OnlyMetadata:    *onlyMeta,
		SendTokens:      *tokens,
		NoObjects:       *noObjects,
		DisableWarnings: *noWarn,
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
