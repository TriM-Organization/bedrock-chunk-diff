package timeline

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
)

// CompressMethodBytesByID looks up the bytes represents of given compressMethod.
// panic when given compress method ID is unknown.
func CompressMethodBytesByID(compressMethod uint8) []byte {
	switch compressMethod {
	case CompressMethodGzip, CompressMethodZlib:
		return []byte{compressMethod}
	default:
		panic(fmt.Sprintf("CompressMethodBytesByID: Unknown compress method %d was found", compressMethod))
	}
}

// CompressMethodByBytes looks up the compress method id by given compressMethodBytes.
// panic if compressMethodBytes is broken or the got compress method ID is unknown.
func CompressMethodByBytes(compressMethodBytes []byte) uint8 {
	if len(compressMethodBytes) != 1 {
		panic(fmt.Sprintf("CompressMethodByBytes: Broken compress method bytes %#v", compressMethodBytes))
	}
	compressMethod := compressMethodBytes[0]
	switch compressMethod {
	case CompressMethodGzip, CompressMethodZlib:
		return compressMethod
	default:
		panic(fmt.Sprintf("CompressMethodByBytes: Unknown compress method %d was found", compressMethod))
	}
}

// CompresserFuncByID looks up the compresser implements by given compressMethod.
// panic when given compress method ID is unknown.
func CompresserFuncByID(compressMethod uint8) *utils.Compresser {
	switch compressMethod {
	case CompressMethodGzip:
		newReaderFunc := func(r io.Reader) (utils.CompresserReader, error) {
			return gzip.NewReader(r)
		}
		newWriterFunc := func(w io.Writer, level int) (utils.CompresserWriter, error) {
			return gzip.NewWriterLevel(w, level)
		}
		return utils.NewCompresser(newReaderFunc, newWriterFunc)
	case CompressMethodZlib:
		newReaderFunc := func(r io.Reader) (utils.CompresserReader, error) {
			return zlib.NewReader(r)
		}
		newWriterFunc := func(w io.Writer, level int) (utils.CompresserWriter, error) {
			return zlib.NewWriterLevel(w, level)
		}
		return utils.NewCompresser(newReaderFunc, newWriterFunc)
	default:
		panic(fmt.Sprintf("CompresserFuncByID: Unknown compress method %d was found", compressMethod))
	}
}
