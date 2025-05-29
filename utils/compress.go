package utils

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

type (
	CompresserReader    io.ReadCloser
	CompresserWriter    io.WriteCloser
	CompresserNewReader func(r io.Reader) (CompresserReader, error)
	CompresserNewWriter func(w io.Writer, level int) (CompresserWriter, error)
)

// Compresser ..
type Compresser struct {
	r CompresserNewReader
	w CompresserNewWriter
}

// NewCompresser ..
func NewCompresser(newReaderFunc CompresserNewReader, newWriterFunc CompresserNewWriter) *Compresser {
	return &Compresser{
		r: newReaderFunc,
		w: newWriterFunc,
	}
}

// Compress ..
func (c Compresser) Compress(in []byte) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	w, err := c.w(buf, flate.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("Compress: %v", err)
	}

	_, err = w.Write(in)
	if err != nil {
		return nil, fmt.Errorf("Compress: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("Compress: %v", err)
	}

	return buf.Bytes(), nil
}

// Decompress ..
func (c Compresser) Decompress(in []byte) (result []byte, err error) {
	r, err := c.r(bytes.NewBuffer(in))
	if err != nil {
		return nil, fmt.Errorf("Decompress: %v", err)
	}
	defer r.Close()

	bytesGet, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Decompress: %v", err)
	}
	return bytesGet, nil
}
