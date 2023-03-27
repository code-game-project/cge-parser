package protobuf

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

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
