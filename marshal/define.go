package marshal

import "github.com/TriM-Organization/bedrock-chunk-diff/define"

const (
	MatrixStateEmpty uint8 = iota
	MatrixStateNotEmpty
)

var (
	emptyBlockMatrix = define.BlockMatrix{}
	emptyDiffMatrix  = define.DiffMatrix{}
)
