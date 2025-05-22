package define

type (
	// ChunkMatrix represents the chunk matrix that holds all
	// block of this chunk. A chunk can have multiple sub chunks,
	// and each sub chunks can have multiple layers.
	// A single layer in one sub chunk only have 4096 blocks.
	ChunkMatrix []Layers
	// ChunkDiffMatrix represents the difference for all sub chunks
	// in the target chunk between two different times.
	ChunkDiffMatrix []LayersDiff
)

// ChunkDifference computes the difference between older and newer.
// We assume len(older) = len(newer).
//
// Time complexity: O(n×L), n=len(older).
// L is the average changes for each sub chunk.
func ChunkDifference(older ChunkMatrix, newer ChunkMatrix) ChunkDiffMatrix {
	result := make(ChunkDiffMatrix, len(older))
	for i := range result {
		result[i] = LayerDifference(older[i], newer[i])
	}
	return result
}

// BlockRestore use old and diff to compute the newer chunk matrix.
// We assume len(old) = len(diff).
//
// Time complexity: O(n×L), n=len(old).
// L is the average count of changes that each sub chunk have.
//
// To reduce the time causes, the block martix in layers of each sub chunk
// in the returned chunk martix is the same one in the corresponding layer
// that come from old.
//
// Note that you could do this operation for all difference array,
// then you will get the final block matrix that represents the
// latest one.
//
// In this case, the time complexity is O(k×n×L) where k is the
// length of these difference array.
func ChunkRestore(old ChunkMatrix, diff ChunkDiffMatrix) ChunkMatrix {
	result := make(ChunkMatrix, len(old))
	for i := range result {
		result[i] = LayerRestore(old[i], diff[i])
	}
	return result
}

// ChunkNoChange reports diff is empty or not.
func ChunkNoChange(diff ChunkDiffMatrix) bool {
	for _, value := range diff {
		if !LayerNoChange(value) {
			return false
		}
	}
	return true
}
