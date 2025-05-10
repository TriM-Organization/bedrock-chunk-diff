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
// If not exist, the layer is created, as well as
// all layers between the current highest layer
// and the new highest layer.
func (l *Layers) Layer(layer int) BlockMatrix {
	for layer >= len(*l) {
		temp := *l
		temp = append(temp, BlockMatrix{})
		*l = temp
	}
	return (*l)[layer]
}

// Layer get the the difference block matrix which in layer.
// If not exist, the layer is created, as well as all layers
// between the current highest layer and the new highest layer.
func (d *LayersDiff) Layer(layer int) DiffMatrix {
	for layer >= len(*d) {
		temp := *d
		temp = append(temp, DiffMatrix{})
		*d = temp
	}
	return (*d)[layer]
}
