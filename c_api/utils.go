package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func asCbool(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func asGoBool(b C.int) bool {
	return (int(b) != 0)
}

func asCbytes(b []byte) *C.char {
	result := make([]byte, 4)
	binary.LittleEndian.PutUint32(result, uint32(len(b)))
	result = append(result, b...)
	return (*C.char)(C.CBytes(result))
}

func asGoBytes(p *C.char) []byte {
	l := binary.LittleEndian.Uint32(C.GoBytes(unsafe.Pointer(p), 4))
	return C.GoBytes(unsafe.Pointer(p), C.int(4+l))[4:]
}

func packChunks(subChunks [][]byte) []byte {
	buf := bytes.NewBuffer(nil)
	for _, value := range subChunks {
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(value)))
		buf.Write(length)
		buf.Write(value)
	}
	return buf.Bytes()
}

func unpackChunks(payload []byte) (subChunks [][]byte) {
	for len(payload) > 0 {
		length := binary.LittleEndian.Uint32(payload)
		subChunks = append(subChunks, payload[4:length+4])
		payload = payload[length+4:]
	}
	return
}

func packNBTs(nbts []map[string]any) (payload []byte, err error) {
	buf := bytes.NewBuffer(nil)

	for _, value := range nbts {
		w := bytes.NewBuffer(nil)
		if err = nbt.NewEncoderWithEncoding(w, nbt.LittleEndian).Encode(value); err != nil {
			return nil, fmt.Errorf("packNBTs: %v", err)
		}

		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(w.Len()))
		buf.Write(length)

		buf.Write(w.Bytes())
	}

	return buf.Bytes(), nil
}

func unpackNBTs(payload []byte) (nbts []map[string]any, err error) {
	for len(payload) > 0 {
		var m map[string]any
		length := binary.LittleEndian.Uint32(payload)

		err := nbt.NewDecoderWithEncoding(bytes.NewBuffer(payload[4:length+4]), nbt.LittleEndian).Decode(&m)
		if err != nil {
			return nil, fmt.Errorf("AppendDiskChunk: %v", err)
		}
		nbts = append(nbts, m)

		payload = payload[length+4:]
	}
	return
}

func packNextOrLast(
	c *chunk.Chunk, e chunk.Encoding, nbts []map[string]any,
	updateUnixTime int64,
	isLastElement *bool,
) *C.char {
	result := bytes.NewBuffer(nil)

	// c
	{
		chunkPayload, r := utils.ChunkPayload(c, e)
		chunkPayloadBytes := packChunks(chunkPayload)

		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(chunkPayloadBytes)))
		result.Write(length)
		result.Write(chunkPayloadBytes)

		rangeStart := make([]byte, 2)
		rangeEnd := make([]byte, 2)

		binary.LittleEndian.PutUint16(rangeStart, uint16(r[0]))
		binary.LittleEndian.PutUint16(rangeEnd, uint16(r[1]))

		result.Write(rangeStart)
		result.Write(rangeEnd)
	}

	// nbts
	{
		nbtPayload, err := packNBTs(nbts)
		if err != nil {
			return asCbytes(nil)
		}

		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(nbtPayload)))
		result.Write(length)
		result.Write(nbtPayload)
	}

	// updateUnixTime
	unixTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(unixTimeBytes, uint64(updateUnixTime))
	result.Write(unixTimeBytes)

	// isLastElement
	if isLastElement != nil {
		if *isLastElement {
			result.WriteByte(1)
		} else {
			result.WriteByte(0)
		}
	}

	return asCbytes(result.Bytes())
}
