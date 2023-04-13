package adapter

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protodelim"

	"github.com/code-game-project/cge-parser/protobuf/schema"
)

func receiveProtobufs(input io.Reader, config Config) (ParserResponse, error) {
	in := bufio.NewReader(input)

	objCap := 0
	if !config.NoObjects {
		objCap = 8
	}

	tokenCap := 0
	if config.SendTokens {
		tokenCap = 64
	}

	response := ParserResponse{
		Events:      make([]Object, 0, objCap),
		Commands:    make([]Object, 0, objCap),
		Types:       make([]Object, 0, objCap),
		Enums:       make([]Object, 0, objCap),
		Tokens:      make([]Token, 0, tokenCap),
		Diagnostics: make([]Diagnostic, 0),
	}

	for {
		msgType := new(schema.MsgType)
		err := protodelim.UnmarshalFrom(in, msgType)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return ParserResponse{}, fmt.Errorf("failed to decode protobuf message: %w", err)
		}
		switch msgType.Type {
		case schema.MsgType_METADATA:
			metadata := new(schema.Metadata)
			err = protodelim.UnmarshalFrom(in, metadata)
			if err == nil {
				response.Metadata = metadataFromProtobuf(metadata)
			}
		case schema.MsgType_TOKEN:
			token := new(schema.Token)
			err = protodelim.UnmarshalFrom(in, token)
			if err == nil {
				response.Tokens = append(response.Tokens, tokenFromProtobuf(token))
			}
		case schema.MsgType_OBJECT:
			object := new(schema.Object)
			err = protodelim.UnmarshalFrom(in, object)
			if err == nil {
				switch object.Type {
				case schema.Object_CONFIG:
					response.Config = objectFromProtobuf(object)
				case schema.Object_EVENT:
					response.Events = append(response.Events, objectFromProtobuf(object))
				case schema.Object_COMMAND:
					response.Commands = append(response.Commands, objectFromProtobuf(object))
				case schema.Object_TYPE:
					response.Types = append(response.Types, objectFromProtobuf(object))
				case schema.Object_ENUM:
					response.Enums = append(response.Enums, objectFromProtobuf(object))
				}
			}
		case schema.MsgType_DIAGNOSTIC:
			diagnostic := new(schema.Diagnostic)
			err = protodelim.UnmarshalFrom(in, diagnostic)
			if err == nil {
				response.Diagnostics = append(response.Diagnostics, diagnosticFromProtobuf(diagnostic))
			}
		}
		if err != nil {
			return ParserResponse{}, fmt.Errorf("failed to decode protobuf message: %w", err)
		}
	}

	return response, nil
}
