package define

// MatrixSize is the size of block matrix.
// A single sub chunk only holds 4096 blocks,
// so we here use 4096 as the size.
const MatrixSize = 4096

type (
	// BlockMatrix represents a block matrix at a specific point in time.
	BlockMatrix *[MatrixSize]int32
	// DiffMatrix is a matrix that holds the difference of
	// BlockMatrix between time i-1 and time i.
	// Note that i must bigger than 0.
	DiffMatrix *[MatrixSize]int32
)

// NewMatrix creates and returns new T.
// Note that the returned T is not nil.
func NewMatrix[T BlockMatrix | DiffMatrix]() T {
	return &[MatrixSize]int32{}
}

// MatrixIsEmpty checks the martix T is empty or not.
func MatrixIsEmpty[T BlockMatrix | DiffMatrix](matrix T) bool {
	return (matrix == nil)
}

// BlockDifference computes the difference between older and newer.
// Time complexity: O(4096).
func BlockDifference(older BlockMatrix, newer BlockMatrix) DiffMatrix {
	if MatrixIsEmpty(older) {
		if MatrixIsEmpty(newer) {
			return nil
		}
		return DiffMatrix(newer)
	}

	result := NewMatrix[DiffMatrix]()

	if MatrixIsEmpty(newer) {
		for i := range MatrixSize {
			result[i] = -older[i]
		}
		return result
	}

	for i := range MatrixSize {
		result[i] = newer[i] - older[i]
	}

	isAllAir := true
	for _, value := range result {
		if value != 0 {
			isAllAir = false
		}
	}

	if isAllAir {
		return nil
	}
	return result
}

// BlockRestore use old and diff to compute the newer block matrix.
// Time complexity: O(4096).
//
// Note that you could do this operation for all difference array,
// then you will get the final block matrix that represents the latest one.
//
// In this case, the time complexity is O(n×4096) where n is the length of
// these difference array.
func BlockRestore(old BlockMatrix, diff DiffMatrix) BlockMatrix {
	if MatrixIsEmpty(old) {
		if MatrixIsEmpty(diff) {
			return nil
		}
		return BlockMatrix(diff)
	}

	if MatrixIsEmpty(diff) {
		return old
	}

	result := NewMatrix[BlockMatrix]()
	for index, value := range diff {
		result[index] = old[index] + value
	}
	return result
}
