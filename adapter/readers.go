package adapter

import (
	"bytes"
	"io"
)

type peekReader struct {
	reader io.Reader
	buffer *bytes.Buffer
}

func newPeekReader(reader io.Reader, initialBufferSize int) *peekReader {
	return &peekReader{
		reader: reader,
		buffer: bytes.NewBuffer(make([]byte, 0, initialBufferSize)),
	}
}

func (r *peekReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	n2, err2 := r.buffer.Write(p[:n])
	if n2 != n || err2 != nil {
		panic("failed to write data to peekReader")
	}
	return n, err
}
