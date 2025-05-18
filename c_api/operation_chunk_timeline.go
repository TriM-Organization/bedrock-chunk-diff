package main

import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	"github.com/TriM-Organization/bedrock-world-operator/define"
)

var savedChunkTimeline = NewSimpleManager[*timeline.ChunkTimeline]()

// append ..
func append(
	id C.longlong,
	chunkPayload *C.char, nbtPayload *C.char,
	rangeStart C.int, rangeEnd C.int,
	e chunk.Encoding,
) *C.char {
	subChunks := unpackChunks(asGoBytes(chunkPayload))
	nbts, err := unpackNBTs(asGoBytes(nbtPayload))
	if err != nil {
		return C.CString(fmt.Sprintf("append: %v", err))
	}

	c, err := utils.FromChunkPayload(subChunks, define.Range{int(rangeStart), int(rangeEnd)}, e)
	if err != nil {
		return C.CString(fmt.Sprintf("append: %v", err))
	}

	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return C.CString("append: Chunk timeline not found")
	}

	(*ctl).Append(c, nbts)
	return C.CString("")
}

//export AppendDiskChunk
func AppendDiskChunk(id C.longlong, chunkPayload *C.char, nbtPayload *C.char, rangeStart C.int, rangeEnd C.int) *C.char {
	return append(id, chunkPayload, nbtPayload, rangeStart, rangeEnd, chunk.DiskEncoding)
}

//export AppendNetworkChunk
func AppendNetworkChunk(id C.longlong, chunkPayload *C.char, nbtPayload *C.char, rangeStart C.int, rangeEnd C.int) *C.char {
	return append(id, chunkPayload, nbtPayload, rangeStart, rangeEnd, chunk.NetworkEncoding)
}

//export Empty
func Empty(id C.longlong) C.int {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return -1
	}
	return asCbool((*ctl).Empty())
}

//export ReadOnly
func ReadOnly(id C.longlong) C.int {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return -1
	}
	return asCbool((*ctl).ReadOnly())
}

//export AllTimePoint
func AllTimePoint(id C.longlong) *C.char {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return asCbytes(nil)
	}

	allTimePoint := (*ctl).AllTimePoint()
	buf := bytes.NewBuffer(nil)

	for _, value := range allTimePoint {
		temp := make([]byte, 8)
		binary.LittleEndian.PutUint64(temp, uint64(value))
		buf.Write(temp)
	}

	return asCbytes(buf.Bytes())
}

//export AllTimePointLen
func AllTimePointLen(id C.longlong) C.int {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return -1
	}
	return C.int((*ctl).AllTimePointLen())
}

//export SetMaxLimit
func SetMaxLimit(id C.longlong, maxLimit C.int) *C.char {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return C.CString("SetMaxLimit: Chunk timeline not found")
	}

	err := (*ctl).SetMaxLimit(uint(maxLimit))
	if err != nil {
		return C.CString(fmt.Sprintf("SetMaxLimit: %v", err))
	}

	return C.CString("")
}

//export Compact
func Compact(id C.longlong) *C.char {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return C.CString("Compact: Chunk timeline not found")
	}

	err := (*ctl).Compact()
	if err != nil {
		return C.CString(fmt.Sprintf("Compact: %v", err))
	}

	return C.CString("")
}

// next ..
func next(id C.longlong, e chunk.Encoding) (complexReturn *C.char) {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return asCbytes(nil)
	}

	c, nbts, updateUnixTime, isLastElement, err := (*ctl).Next()
	if err != nil {
		return asCbytes(nil)
	}

	return packNextOrLast(c, e, nbts, updateUnixTime, &isLastElement)
}

//export NextDiskChunk
func NextDiskChunk(id C.longlong) (complexReturn *C.char) {
	return next(id, chunk.DiskEncoding)
}

//export NextNetworkChunk
func NextNetworkChunk(id C.longlong) (complexReturn *C.char) {
	return next(id, chunk.NetworkEncoding)
}

// last ..
func last(id C.longlong, e chunk.Encoding) (complexReturn *C.char) {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return asCbytes(nil)
	}

	c, nbts, updateUnixTime, err := (*ctl).Last()
	if err != nil {
		return asCbytes(nil)
	}

	return packNextOrLast(c, e, nbts, updateUnixTime, nil)
}

//export LastDiskChunk
func LastDiskChunk(id C.longlong) (complexReturn *C.char) {
	return last(id, chunk.DiskEncoding)
}

//export LastNetworkChunk
func LastNetworkChunk(id C.longlong) (complexReturn *C.char) {
	return last(id, chunk.NetworkEncoding)
}

//export Pop
func Pop(id C.longlong) *C.char {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return C.CString("Pop: Chunk timeline not found")
	}

	err := (*ctl).Pop()
	if err != nil {
		return C.CString(fmt.Sprintf("Pop: %v", err))
	}

	return C.CString("")
}

//export Save
func Save(id C.longlong) *C.char {
	ctl := savedChunkTimeline.LoadObject(int(id))
	if ctl == nil {
		return C.CString("Save: Chunk timeline not found")
	}

	err := (*ctl).Save()
	if err != nil {
		return C.CString(fmt.Sprintf("Save: %v", err))
	}

	return C.CString("")
}
