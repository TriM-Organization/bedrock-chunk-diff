package define

type (
	// Layers represents the block matrix of a sub chunk.
	// A single sub chunk could have multiple layers,
	// and each layer holds 4096 blocks.
	Layers []BlockMatrix
	// LayersDiff represents the difference for all layers
	// in the target sub chunk between two different times.
	LayersDiff []DiffMatrix
)

// Layer get the the block matrix which in layer.
// If not exist, then create empty layer, as well as
// all layers between the current highest layer
// and the new highest layer.
func (l *Layers) Layer(layer int) BlockMatrix {
	for layer >= len(*l) {
		temp := *l
		temp = append(temp, NewMatrix[BlockMatrix](true))
		*l = temp
	}
	return (*l)[layer]
}

// Layer get the the difference block matrix which in layer.
// If not exist, then create empty layer, as well as all layers
// between the current highest layer and the new highest layer.
func (d *LayersDiff) Layer(layer int) DiffMatrix {
	for layer >= len(*d) {
		temp := *d
		temp = append(temp, NewMatrix[DiffMatrix](true))
		*d = temp
	}
	return (*d)[layer]
}

// LayerDifference computes the difference between older and newer.
// Time complexity: O(4096×n), n = max(len(older), len(newer)).
func LayerDifference(older Layers, newer Layers) LayersDiff {
	var result LayersDiff

	for i := range older {
		_ = result.Layer(i)
	}
	for i := range newer {
		_ = result.Layer(i)
	}

	for i := range result {
		result[i] = BlockDifference(older.Layer(i), newer.Layer(i))
	}

	return result
}

// LayerRestore use old and diff to compute the newer layers
// Time complexity: O(4096×n), n = len(diff).
//
// Note that you could do this operation for all difference array,
// then you will get the final layers that represents the latest one.
//
// In this case, the time complexity is O(n×L×4096) where n is the
// length of these difference array and n is the average count of
// each len(diff).
func LayerRestore(old Layers, diff LayersDiff) Layers {
	var result Layers

	for i := range diff {
		_ = result.Layer(i)
	}

	for i := range result {
		result[i] = BlockRestore(old.Layer(i), diff.Layer(i))
	}

	return result
}
