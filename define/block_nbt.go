package define

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/TriM-Organization/bedrock-chunk-diff/utils"
	"github.com/cespare/xxhash/v2"
	"github.com/kr/binarydist"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// NBTWithIndex represents a single NBT block entity data who in a chunk.
// Index is an integer that range from 0 to 4095, and could be decode as a
// block relative position to the chunk.
type NBTWithIndex struct {
	Index ChunkBlockIndex
	NBT   map[string]any
}

// NewBlockNBT creates a new NBTWithIndex by given relativePos and nbt.
// relativePos is the relative position of this block NBT to chunk,
// and nbt is the block NBT data of this NBT block.
func NewBlockNBT(relativePos [3]uint16, nbt map[string]any) *NBTWithIndex {
	n := &NBTWithIndex{NBT: nbt}
	n.Index.UpdateIndex(uint8(relativePos[0]), int16(relativePos[1]), uint8(relativePos[2]))
	return n
}

// NBTDeepCopy return the deep copy of src.
func NBTDeepCopy(src []NBTWithIndex) (result []NBTWithIndex, err error) {
	for _, value := range src {
		var m map[string]any
		buf := bytes.NewBuffer(nil)

		err := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian).Encode(value.NBT)
		if err != nil {
			return nil, fmt.Errorf("NBTDeepCopy: %v", err)
		}

		err = nbt.NewDecoderWithEncoding(bytes.NewBuffer(buf.Bytes()), nbt.LittleEndian).Decode(&m)
		if err != nil {
			return nil, fmt.Errorf("NBTDeepCopy: %v", err)
		}

		result = append(result, NBTWithIndex{
			Index: value.Index,
			NBT:   m,
		})
	}
	return
}

// DiffNBTWithIndex represents the difference between the same NBT block but on different time.
// Index is an integer that range from 0 to 4095, and could be decode as a block relative position
// to the chunk.
type DiffNBTWithIndex struct {
	Index   ChunkBlockIndex
	DiffNBT []byte
}

// NewDiffNBT returns the difference between between olderNBT and newerNBT.
// Note that olderNBT and newerNBT must represents the NBt block in the same position.
//
// Time complexity: O(C).
// Note that C is not very small and is little big due to
//   - Use bsdiff to do restore to reduce bytes use.
//   - Use xxhash to ensure when user do restore operation, they can verify the data they get is correct.
func NewDiffNBT(olderNBT *NBTWithIndex, newerNBT *NBTWithIndex) (result *DiffNBTWithIndex, err error) {
	if olderNBT == nil || newerNBT == nil {
		return nil, fmt.Errorf("NewDiffNBT: olderNBT or newerNBT is nil")
	}
	if olderNBT.Index != newerNBT.Index {
		return nil, fmt.Errorf("NewDiffNBT: Can't do difference operation between two blocks in different position")
	}
	n := &DiffNBTWithIndex{Index: olderNBT.Index}

	buf := bytes.NewBuffer(nil)
	utils.MarshalNBT(buf, olderNBT.NBT, "")
	olderNBTBytes := buf.Bytes()

	buf = bytes.NewBuffer(nil)
	utils.MarshalNBT(buf, newerNBT.NBT, "")
	newerNBTBytes := buf.Bytes()

	olderHash := xxhash.Sum64(olderNBTBytes)
	olderHashBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(olderHashBytes, olderHash)

	newerHash := xxhash.Sum64(newerNBTBytes)
	newerHashBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newerHashBytes, newerHash)

	buf = bytes.NewBuffer(nil)
	err = binarydist.Diff(bytes.NewBuffer(olderNBTBytes), bytes.NewBuffer(newerNBTBytes), buf)
	if err != nil {
		return nil, fmt.Errorf("NewDiffNBT: %v", err)
	}

	n.DiffNBT = append(olderHashBytes, newerHashBytes...)
	n.DiffNBT = append(n.DiffNBT, buf.Bytes()...)

	return n, nil
}

// Restore use olderNBT and DiffNBTWithIndex it self to compute the newer block NBT data,
// and return the restor result that with the same chunk block index to olderNBt but newer NBT.
//
// Note that you could do this operation for all difference like []DiffNBTWithIndex,
// then you will get the final block NBT data that represents the latest one.
//
// In this case, the time complexity is O(C×n) where n is the length of these difference array.
// Note that C is not very small and is little big due to we use bsdiff to do restore and use
// xxhash to ensure the data we get is correct.
func (d DiffNBTWithIndex) Restore(olderNBT NBTWithIndex) (result *NBTWithIndex, err error) {
	if d.Index != olderNBT.Index {
		return nil, fmt.Errorf("Restore: Can't do restore operation between two blocks in different position")
	}

	if len(d.DiffNBT) < 16 {
		return nil, fmt.Errorf("Restore: Broken diff")
	}

	buf := bytes.NewBuffer(nil)
	utils.MarshalNBT(buf, olderNBT.NBT, "")
	olderNBTBytes := buf.Bytes()

	if xxhash.Sum64(olderNBTBytes) != binary.LittleEndian.Uint64(d.DiffNBT) {
		return nil, fmt.Errorf("Restore: Given older NBT bytes is not the correct one (hash mismatch)")
	}

	buf = bytes.NewBuffer(nil)
	err = binarydist.Patch(bytes.NewBuffer(olderNBTBytes), buf, bytes.NewBuffer(d.DiffNBT[16:]))
	if err != nil {
		return nil, fmt.Errorf("Restore: %v", err)
	}
	newerNBTBytes := buf.Bytes()

	if xxhash.Sum64(newerNBTBytes) != binary.LittleEndian.Uint64(d.DiffNBT[8:]) {
		return nil, fmt.Errorf("Restore: Data changed")
	}

	result = &NBTWithIndex{Index: olderNBT.Index}
	err = nbt.NewDecoderWithEncoding(bytes.NewBuffer(newerNBTBytes), nbt.LittleEndian).Decode(&result.NBT)
	if err != nil {
		return nil, fmt.Errorf("Restore: %v", err)
	}
	return
}

// MultipleDiffNBT represents the difference between NBT blocks in the same chunk but different times.
// All the NBT blocks should in the same position in this chunk, and MultipleDiffNBT just refer to the
// states (add/remove/modify) of these blocks in different times.
type MultipleDiffNBT struct {
	Removed  []ChunkBlockIndex
	Added    []NBTWithIndex
	Modified []DiffNBTWithIndex
}

// NBTDifference computes the difference between multiple block NBT changes in one single chunk.
// older and newer are represents the different time of these NBT blocks in the same chunk.
//
// Time complexity: O(C×k + (a+b)), a=len(older), b=len(newer).
//
// k is the number that shown the counts of changed (modified) NBT blocks.
// Note that C is not very small and is little big due to we use bsdiff and xxhash for each modified NBT block.
func NBTDifference(older []NBTWithIndex, newer []NBTWithIndex) (result *MultipleDiffNBT, err error) {
	olderSet := make(map[ChunkBlockIndex]*NBTWithIndex)
	newerSet := make(map[ChunkBlockIndex]*NBTWithIndex)
	for _, value := range older {
		olderSet[value.Index] = &value
	}
	for _, value := range newer {
		newerSet[value.Index] = &value
	}

	removed := make([]ChunkBlockIndex, 0)
	removedSet := make(map[ChunkBlockIndex]bool)
	added := make([]NBTWithIndex, 0)
	modified := make([]DiffNBTWithIndex, 0)

	for key := range olderSet {
		if newerSet[key] == nil {
			removed = append(removed, key)
			removedSet[key] = true
		}
	}

	for key := range newerSet {
		if olderSet[key] == nil {
			added = append(added, *newerSet[key])
		}
	}

	for key, value := range olderSet {
		if removedSet[key] {
			continue
		}

		if reflect.DeepEqual(value.NBT, newerSet[key].NBT) {
			continue
		}

		diff, err := NewDiffNBT(value, newerSet[key])
		if err != nil {
			return nil, fmt.Errorf("NBTDifference: %v", err)
		}

		modified = append(modified, *diff)
	}

	return &MultipleDiffNBT{
		Removed:  removed,
		Added:    added,
		Modified: modified,
	}, nil
}

// NBTRestore computes the newer block NBT data of this chunk by given old and diff.
//
// Time complexity: O(a+C×b), a=len(old), b=len(diff.Modified).
// Note that C is not very small and is little big due to we use bsdiff and xxhash for each modified NBT block.
func NBTRestore(old []NBTWithIndex, diff MultipleDiffNBT) (result []NBTWithIndex, err error) {
	// Deep copy
	oldCopy, err := NBTDeepCopy(old)
	if err != nil {
		return nil, fmt.Errorf("NBTRestore: %v", err)
	}

	// Added
	result = append(result, diff.Added...)

	// Modified
	olderSet := make(map[ChunkBlockIndex]NBTWithIndex)
	for _, value := range oldCopy {
		olderSet[value.Index] = value
	}
	for _, value := range diff.Modified {
		newer, err := value.Restore(olderSet[value.Index])
		if err != nil {
			return nil, fmt.Errorf("NBTRestore: %v", err)
		}
		result = append(result, *newer)
	}

	// No change
	changedSet := make(map[ChunkBlockIndex]bool)
	for _, value := range diff.Removed {
		changedSet[value] = true
	}
	for _, value := range diff.Modified {
		changedSet[value.Index] = true
	}
	for _, value := range oldCopy {
		if !changedSet[value.Index] {
			result = append(result, value)
		}
	}

	return
}
