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
// blockNBT must contains all NBT blocks from the same chunk
// and in the same time.
func BlockNBTBytes(blockNBT []define.NBTWithIndex) (result []byte, err error) {
	if len(blockNBT) == 0 {
		return nil, nil
	}

	buf := bytes.NewBuffer(nil)

	for _, value := range blockNBT {
		value.Index.Marshal(buf)
		utils.MarshalNBT(buf, value.NBT, "")
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("BlockNBTBytes: %v", err)
	}
	return
}

// BytesToBlockNBT decode multiple NBTWithIndex from bytes.
// Ensure all element in returned slice all represents the NBT blocks
// in the same chunk and in the same time.
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
		var index define.ChunkBlockIndex
		var m map[string]any

		index.Unmarshal(buf)
		r.NBT(&m, nbt.LittleEndian)

		result = append(result, define.NBTWithIndex{
			Index: index,
			NBT:   m,
		})
	}

	return result, nil
}

// MultipleDiffNBTBytes return the bytes represents of diff.
func MultipleDiffNBTBytes(diff define.MultipleDiffNBT) (result []byte, err error) {
	if len(diff.Removed) == 0 && len(diff.Added) == 0 && len(diff.Modified) == 0 {
		return nil, nil
	}

	buf := bytes.NewBuffer(nil)
	w := protocol.NewWriter(buf, 0)

	length := uint32(len(diff.Removed))
	w.Varuint32(&length)
	for _, value := range diff.Removed {
		value.Marshal(buf)
	}

	length = uint32(len(diff.Added))
	w.Varuint32(&length)
	for _, value := range diff.Added {
		value.Index.Marshal(buf)
		utils.MarshalNBT(buf, value.NBT, "")
	}

	for _, value := range diff.Modified {
		value.Index.Marshal(buf)
		w.ByteSlice(&value.DiffNBT)
	}

	result, err = utils.Gzip(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("MultipleDiffNBTBytes: %v", err)
	}
	return
}

// BytesToMultipleDiffNBT decode MultipleDiffNBT from bytes.
func BytesToMultipleDiffNBT(in []byte) (result define.MultipleDiffNBT, err error) {
	var length uint32

	if len(in) == 0 {
		return
	}

	originBytes, err := utils.Ungzip(in)
	if err != nil {
		err = fmt.Errorf("BytesToMultipleDiffNBT: %v", err)
		return
	}

	buf := bytes.NewBuffer(originBytes)
	r := protocol.NewReader(buf, 0, false)

	r.Varuint32(&length)
	result.Removed = make([]define.ChunkBlockIndex, length)
	for i := range length {
		var index define.ChunkBlockIndex
		index.Unmarshal(buf)
		result.Removed[i] = index
	}

	r.Varuint32(&length)
	result.Added = make([]define.NBTWithIndex, length)
	for i := range length {
		var object define.NBTWithIndex
		var index define.ChunkBlockIndex
		index.Unmarshal(buf)
		r.NBT(&object.NBT, nbt.LittleEndian)
		object.Index = index
		result.Added[i] = object
	}

	for buf.Len() > 0 {
		var object define.DiffNBTWithIndex
		var index define.ChunkBlockIndex
		index.Unmarshal(buf)
		r.ByteSlice(&object.DiffNBT)
		object.Index = define.ChunkBlockIndex(index)
		result.Modified = append(result.Modified, object)
	}

	return
}
