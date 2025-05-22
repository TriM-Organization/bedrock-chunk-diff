package define

// MatrixSize is the size of block matrix.
// A single sub chunk only holds 4096 blocks,
// so we here use 4096 as the size.
const MatrixSize = 4096

type (
	// BlockMatrix represents a block matrix at a specific point in time.
	BlockMatrix *[MatrixSize]uint32
	// SingleBlockDiff represents a single block change who in a sub chunk.
	SingleBlockDiff struct {
		IndexDelta   uint32
		NewPaletteID uint32
	}
	// DiffMatrix is a matrix that holds the difference of
	// BlockMatrix between time i-1 and time i.
	// Note that i must bigger than 0.
	DiffMatrix []SingleBlockDiff
)

// NewBlockMatrix creates a new BlockMatrix that full of air and is not nil.
func NewBlockMatrix() BlockMatrix {
	return &[MatrixSize]uint32{}
}

// BlockMatrixIsEmpty checks the given block martix is empty or not.
func BlockMatrixIsEmpty(matrix BlockMatrix) bool {
	return (matrix == nil)
}

// BlockDifference computes the difference between older and newer.
// Time complexity: O(4096).
func BlockDifference(older BlockMatrix, newer BlockMatrix) DiffMatrix {
	var result DiffMatrix

	if BlockMatrixIsEmpty(older) && BlockMatrixIsEmpty(newer) {
		return nil
	}

	if BlockMatrixIsEmpty(older) {
		older = NewBlockMatrix()
	}
	if BlockMatrixIsEmpty(newer) {
		newer = NewBlockMatrix()
	}

	lastIndex := 0
	for index := range MatrixSize {
		if newID := newer[index]; newID != older[index] {
			result = append(result, SingleBlockDiff{
				IndexDelta:   uint32(index - lastIndex),
				NewPaletteID: newID,
			})
			lastIndex = index
		}
	}

	return result
}

// BlockRestore use old and diff to compute the newer block matrix.
// Note that the returned block martix is the same object of old when
// old is not empty.
// Time complexity: O(l), l=len(diff).
//
// Note that you could do this operation for all difference array,
// then you will get the final block matrix that represents the latest one.
//
// In this case, the time complexity is O(n×L) where n is the length of
// these difference array, and L is the average length of all the diff slice.
func BlockRestore(old BlockMatrix, diff DiffMatrix) BlockMatrix {
	if len(diff) == 0 {
		return old
	}

	if BlockMatrixIsEmpty(old) {
		old = NewBlockMatrix()
	}

	index := uint32(0)
	for _, value := range diff {
		index += value.IndexDelta
		old[index] = value.NewPaletteID
	}

	return old
}

// BlockNoChange reports diff is empty or not.
func BlockNoChange(diff DiffMatrix) bool {
	return (len(diff) == 0)
}
