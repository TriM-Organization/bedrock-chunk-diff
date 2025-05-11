package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Gzip ..
func Gzip(in []byte) (result []byte, err error) {
	buf := bytes.NewBuffer(nil)

	w, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("Gzip: %v", err)
	}

	_, err = w.Write(in)
	if err != nil {
		return nil, fmt.Errorf("Gzip: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("Gzip: %v", err)
	}

	return buf.Bytes(), nil
}

// Ungzip ..
func Ungzip(in []byte) (result []byte, err error) {
	r, err := gzip.NewReader(bytes.NewBuffer(in))
	if err != nil {
		return nil, fmt.Errorf("Ungzip: %v", err)
	}
	defer r.Close()

	bytesGet, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Ungzip: %v", err)
	}
	return bytesGet, nil
}
