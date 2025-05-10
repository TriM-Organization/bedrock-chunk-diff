package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Gzip ..
func Gzip(in []byte) []byte {
	buf := bytes.NewBuffer(in)
	w, _ := gzip.NewWriterLevel(gzip.NewWriter(buf), gzip.BestCompression)
	w.Write(in)
	return buf.Bytes()
}

// Ungzip ..
func Ungzip(in []byte) (result []byte, err error) {
	buf := bytes.NewBuffer(in)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("Ungzip: %v", err)
	}
	bytesGet, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Ungzip: %v", err)
	}
	return bytesGet, nil
}
