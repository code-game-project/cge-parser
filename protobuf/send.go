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
	out     io.Writer
	lastMsg schema.MsgType_Type
}

func NewSender(out io.Writer) *ProtobufSender {
	return &ProtobufSender{
		out:     out,
		lastMsg: -1,
	}
}

func (p *ProtobufSender) SendMetadata(cgeVersion string) error {
	p.setMsgType(schema.MsgType_METADATA)

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
	p.setMsgType(schema.MsgType_DIAGNOSTIC)
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

func (p *ProtobufSender) SendToken(tokenType parser.TokenType, lexeme string, line, column int) error {
	p.setMsgType(schema.MsgType_TOKEN)
	data, err := proto.Marshal(&schema.Token{
		Type:   schema.Token_Type(tokenType),
		Lexeme: lexeme,
		Pos: &schema.Pos{
			Line:   int32(line),
			Column: int32(column),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to encode token as protobuf message: %w", err)
	}

	_, err = p.out.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write token message: %w", err)
	}

	return nil
}

func objectToProtobufObj(object parser.Object) *schema.Object {
	var comment *string
	if object.Comment != "" {
		comment = &object.Comment
	}
	properties := make([]*schema.Property, 0, len(object.Properties))
	for _, p := range object.Properties {
		properties = append(properties, propertyToProtobufProp(object.Type == parser.TTEnum, p))
	}
	return &schema.Object{
		Type:       schema.Object_Type(object.Type - parser.TTConfig),
		Name:       object.Name.Lexeme,
		Properties: properties,
		Comment:    comment,
	}
}

func propertyToProtobufProp(isEnum bool, property parser.Property) *schema.Property {
	var comment *string
	if property.Comment != "" {
		comment = &property.Comment
	}
	return &schema.Property{
		Name:    property.Name,
		Type:    propertyTypeToProtobufPropType(isEnum, property.Type),
		Comment: comment,
	}
}

func propertyTypeToProtobufPropType(isEnum bool, propertyType *parser.PropertyType) *schema.Property_Type {
	var pType schema.Property_Type_DataType
	if propertyType.Token.Type == parser.TTIdentifier {
		if isEnum {
			pType = schema.Property_Type_ENUM_VALUE
		} else {
			pType = schema.Property_Type_CUSTOM
		}
	} else {
		pType = schema.Property_Type_DataType(propertyType.Token.Type - parser.TTString)
	}

	var generic *schema.Property_Type
	if propertyType.Generic != nil {
		generic = propertyTypeToProtobufPropType(false, propertyType.Generic)
	}

	return &schema.Property_Type{
		Name:    propertyType.Token.Lexeme,
		Type:    pType,
		Generic: generic,
	}
}

func (p *ProtobufSender) SendObject(object parser.Object) error {
	p.setMsgType(schema.MsgType_OBJECT)
	data, err := proto.Marshal(objectToProtobufObj(object))
	if err != nil {
		return fmt.Errorf("failed to encode object as protobuf message: %w", err)
	}

	_, err = p.out.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write object message: %w", err)
	}
	return nil
}

func (p *ProtobufSender) setMsgType(msgType schema.MsgType_Type) {
	if p.lastMsg == msgType {
		return
	}
	p.lastMsg = msgType
	data, err := proto.Marshal(&schema.MsgType{
		Type: msgType,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode protobuf message type: %s", err)
		return
	}

	_, err = p.out.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write protobuf message type: %s", err)
		return
	}
}
