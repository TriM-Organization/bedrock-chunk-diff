package define

// MatrixSize is the size of block matrix.
// A single sub chunk only holds 4096 blocks,
// so we here use 4096 as the size.
const MatrixSize = 4096

type (
	// BlockMatrix represents a block matrix at a specific point in time.
	BlockMatrix [MatrixSize]uint16
	// DiffMatrix is a matrix that holds the difference of
	// BlockMatrix between time i-1 and time i.
	// Note that i must bigger than 0.
	DiffMatrix [MatrixSize]int
)

// Difference computes the difference between older and newer.
// Time complexity: O(4096).
func Difference(older BlockMatrix, newer BlockMatrix) DiffMatrix {
	var result DiffMatrix
	for i := range MatrixSize {
		result[i] = int(newer[i]) - int(older[i])
	}
	return result
}

// Restore use old and diff to compute the newer block matrix.
// The modification is carried out directly on old.
//
// Time complexity: O(4096).
//
// Note that you could do this operation for all difference array,
// then you will get the final block matrix that represents the latest one.
//
// In this case, the time complexity is O(n×4096) where n is the length of
// these difference array.
func Restore(old BlockMatrix, diff DiffMatrix) BlockMatrix {
	for index, value := range diff {
		old[index] = uint16(int(old[index]) + value)
	}
	return old
}
