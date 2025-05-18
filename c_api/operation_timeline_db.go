package main

import "C"
import (
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/timeline"
	operator_define "github.com/TriM-Organization/bedrock-world-operator/define"
)

var savedTimelineDB = NewSimpleManager[timeline.TimelineDatabase]()

//export NewTimelineDB
func NewTimelineDB(path *C.char, noGrowSync C.int, noSync C.int) C.longlong {
	tldb, err := timeline.Open(C.GoString(path), asGoBool(noGrowSync), asGoBool(noGrowSync))
	if err != nil {
		return -1
	}
	return C.longlong(savedTimelineDB.AddObject(tldb))
}

//export CloseTimelineDB
func CloseTimelineDB(id C.longlong) *C.char {
	tldb := savedTimelineDB.LoadObject(int(id))
	if tldb == nil {
		return C.CString("CloseTimelineDB: Timeline database not found")
	}

	err := (*tldb).CloseTimelineDB()
	if err != nil {
		return C.CString(fmt.Sprintf("CloseTimelineDB: %v", err))
	}

	return C.CString("")
}

//export NewChunkTimeline
func NewChunkTimeline(id C.longlong, dm C.int, chunkPosX C.int, chunkPosZ C.int, readOnly C.int) C.longlong {
	tldb := savedTimelineDB.LoadObject(int(id))
	if tldb == nil {
		return -1
	}

	result, err := (*tldb).NewChunkTimeline(
		define.DimChunk{
			Dimension: operator_define.Dimension(dm),
			ChunkPos:  operator_define.ChunkPos{int32(chunkPosX), int32(chunkPosZ)},
		},
		asGoBool(readOnly),
	)
	if err != nil {
		return -1
	}

	return C.longlong(savedChunkTimeline.AddObject(result))
}

//export DeleteChunkTimeline
func DeleteChunkTimeline(id C.longlong, dm C.int, chunkPosX C.int, chunkPosZ C.int) *C.char {
	tldb := savedTimelineDB.LoadObject(int(id))
	if tldb == nil {
		return C.CString("DeleteChunkTimeline: Timeline database not found")
	}

	err := (*tldb).DeleteChunkTimeline(
		define.DimChunk{
			Dimension: operator_define.Dimension(dm),
			ChunkPos:  operator_define.ChunkPos{int32(chunkPosX), int32(chunkPosZ)},
		},
	)
	if err != nil {
		return C.CString(fmt.Sprintf("DeleteChunkTimeline: %v", err))
	}

	return C.CString("")
}

//export LoadLatestTimePointUnixTime
func LoadLatestTimePointUnixTime(id C.longlong, dm C.int, chunkPosX C.int, chunkPosZ C.int) C.longlong {
	tldb := savedTimelineDB.LoadObject(int(id))
	if tldb == nil {
		return -1
	}

	result := (*tldb).LoadLatestTimePointUnixTime(
		define.DimChunk{
			Dimension: operator_define.Dimension(dm),
			ChunkPos:  operator_define.ChunkPos{int32(chunkPosX), int32(chunkPosZ)},
		},
	)

	return C.longlong(result)
}

//export SaveLatestTimePointUnixTime
func SaveLatestTimePointUnixTime(id C.longlong, dm C.int, chunkPosX C.int, chunkPosZ C.int, timeStamp C.longlong) *C.char {
	tldb := savedTimelineDB.LoadObject(int(id))
	if tldb == nil {
		return C.CString("SaveLatestTimePointUnixTime: Timeline database not found")
	}

	err := (*tldb).SaveLatestTimePointUnixTime(
		define.DimChunk{
			Dimension: operator_define.Dimension(dm),
			ChunkPos:  operator_define.ChunkPos{int32(chunkPosX), int32(chunkPosZ)},
		},
		int64(timeStamp),
	)
	if err != nil {
		return C.CString(fmt.Sprintf("SaveLatestTimePointUnixTime: %v", err))
	}

	return C.CString("")
}
