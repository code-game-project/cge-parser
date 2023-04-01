package adapter

import (
	"github.com/code-game-project/cge-parser/parser"
	"github.com/code-game-project/cge-parser/protobuf/schema"
)

type Metadata struct {
	CGEVersion string
}

func metadataFromProtobuf(metadata *schema.Metadata) Metadata {
	return Metadata{
		CGEVersion: metadata.CgeVersion,
	}
}

type Object struct {
	Name       string
	Type       ObjectType
	Comment    string
	Properties []Property
}

type ObjectType int

const (
	OTConfig  = ObjectType(schema.Object_CONFIG)
	OTCommand = ObjectType(schema.Object_COMMAND)
	OTEvent   = ObjectType(schema.Object_EVENT)
	OTType    = ObjectType(schema.Object_TYPE)
	OTEnum    = ObjectType(schema.Object_ENUM)
)

func objectFromProtobuf(object *schema.Object) Object {
	if object.Comment == nil {
		var c string
		object.Comment = &c
	}

	properties := make([]Property, 0, len(object.Properties))
	for _, p := range object.Properties {
		properties = append(properties, propertyFromProtobuf(p))
	}

	return Object{
		Name:       object.Name,
		Type:       ObjectType(object.Type),
		Comment:    *object.Comment,
		Properties: properties,
	}
}

type Property struct {
	Name    string
	Type    *PropertyType
	Comment string
}

type PropertyType struct {
	Name    string
	Type    DataType
	Generic *PropertyType
}

type DataType int

const (
	DTString    = DataType(schema.Property_Type_STRING)
	DTBool      = DataType(schema.Property_Type_BOOL)
	DTInt32     = DataType(schema.Property_Type_INT32)
	DTInt64     = DataType(schema.Property_Type_INT64)
	DTFloat32   = DataType(schema.Property_Type_FLOAT32)
	DTFloat64   = DataType(schema.Property_Type_FLOAT64)
	DTMap       = DataType(schema.Property_Type_MAP)
	DTList      = DataType(schema.Property_Type_LIST)
	DTEnumValue = DataType(schema.Property_Type_ENUM_VALUE)
	DTCustom    = DataType(schema.Property_Type_CUSTOM)
)

func propertyFromProtobuf(property *schema.Property) Property {
	if property.Comment == nil {
		var c string
		property.Comment = &c
	}

	return Property{
		Name:    property.Name,
		Type:    propertyTypeFromProtobuf(property.Type),
		Comment: *property.Comment,
	}
}

func propertyTypeFromProtobuf(propertyType *schema.Property_Type) *PropertyType {
	var generic *PropertyType
	if propertyType.Generic != nil {
		generic = propertyTypeFromProtobuf(propertyType.Generic)
	}
	return &PropertyType{
		Name:    propertyType.Name,
		Type:    DataType(propertyType.Type),
		Generic: generic,
	}
}

type DiagnosticType int

const (
	DiagInfo    = DiagnosticType(schema.Diagnostic_INFO)
	DiagWarning = DiagnosticType(schema.Diagnostic_WARNING)
	DiagError   = DiagnosticType(schema.Diagnostic_ERROR)
)

type Diagnostic struct {
	Type        DiagnosticType
	Message     string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

func diagnosticFromProtobuf(diagnostic *schema.Diagnostic) Diagnostic {
	return Diagnostic{
		Type:        DiagnosticType(diagnostic.Type),
		Message:     diagnostic.Msg,
		StartLine:   int(diagnostic.Start.Line),
		StartColumn: int(diagnostic.Start.Column),
		EndLine:     int(diagnostic.End.Line),
		EndColumn:   int(diagnostic.End.Column),
	}
}

type Token struct {
	Type   parser.TokenType
	Lexeme string
	Line   int
	Column int
}

func tokenFromProtobuf(token *schema.Token) Token {
	return Token{
		Type:   parser.TokenType(token.Type),
		Lexeme: token.Lexeme,
		Line:   int(token.Pos.Line),
		Column: int(token.Pos.Column),
	}
}
