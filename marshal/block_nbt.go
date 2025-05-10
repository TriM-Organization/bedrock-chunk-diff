package marshal

import (
	"bytes"
	"fmt"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockNBTBytes return the bytes represents of blockNBT.
// blockNBT must contains all NBT blocks from the same sub chunk
// and in the same time.
func BlockNBTBytes(blockNBT []define.NBTWithIndex) []byte {
	if len(blockNBT) == 0 {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	w := protocol.NewWriter(buf, 0)

	for _, value := range blockNBT {
		index := uint32(value.Index)
		w.Varuint32(&index)
		utils.MarshalNBT(buf, value.NBT, "")
	}

	return utils.Gzip(buf.Bytes())
}

// BytesToBlockNBT decode multiple NBTWithIndex from bytes.
// Ensure all element in returned slice all represents the NBT blocks
// in the same sub chunk and in the same time.
func BytesToBlockNBT(in []byte) (result []define.NBTWithIndex, err error) {
	if len(in) == 0 {
		return
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToBlockNBT: %v", err)
		return
	}
	result = make([]define.NBTWithIndex, 0)

	buf := bytes.NewBuffer(originBytes)
	r := protocol.NewReader(buf, 0, false)

	for buf.Len() > 0 {
		var index uint32
		var m map[string]any
		r.Varuint32(&index)
		r.NBT(&m, nbt.LittleEndian)
		result = append(result, define.NBTWithIndex{
			Index: define.SubChunkBlockIndex(index),
			NBT:   m,
		})
	}

	return result, nil
}

// MultipleDiffNBTBytes return the bytes represents of diff.
func MultipleDiffNBTBytes(diff define.MultipleDiffNBT) []byte {
	if len(diff.Added) == 0 && len(diff.Modified) == 0 {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	w := protocol.NewWriter(buf, 0)

	length := uint32(len(diff.Added))
	w.Varuint32(&length)
	for _, value := range diff.Added {
		index := uint32(value.Index)
		w.Varuint32(&index)
		utils.MarshalNBT(buf, value.NBT, "")
	}

	for _, value := range diff.Modified {
		index := uint32(value.Index)
		w.Varuint32(&index)
		w.ByteSlice(&value.DiffNBT)
	}

	return utils.Gzip(buf.Bytes())
}

// BytesToMultipleDiffNBT decode MultipleDiffNBT from bytes.
func BytesToMultipleDiffNBT(in []byte) (result define.MultipleDiffNBT, err error) {
	if len(in) == 0 {
		return
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToMultipleDiffNBT: %v", err)
		return
	}
	result.Modified = make([]define.DiffNBTWithIndex, 0)

	buf := bytes.NewBuffer(originBytes)
	r := protocol.NewReader(buf, 0, false)

	var length uint32
	r.Varuint32(&length)
	result.Added = make([]define.NBTWithIndex, length)

	for i := range length {
		var object define.NBTWithIndex
		var index uint32
		r.Varuint32(&index)
		r.NBT(&object.NBT, nbt.LittleEndian)
		object.Index = define.SubChunkBlockIndex(index)
		result.Added[i] = object
	}

	for buf.Len() > 0 {
		var object define.DiffNBTWithIndex
		var index uint32
		r.Varuint32(&index)
		r.ByteSlice(&object.DiffNBT)
		object.Index = define.SubChunkBlockIndex(index)
		result.Modified = append(result.Modified, object)
	}

	return
}
