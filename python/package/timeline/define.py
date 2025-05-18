from dataclasses import dataclass, field
from bedrockworldoperator import Range, RANGE_OVERWORLD


@dataclass
class ChunkData:
    """
    ChunkData represents the data for a chunk.

    A single chunk could holds the block matrix
    data and its block entities NBT data.

    Args:
        sub_chunks (list[bytes]): The payload (block matrix data) of this chunk.
                                  The length of this list must equal to 24 if this chunk is from Overworld,
                                  or 8 if this chunk is from Nether, or 16 if this chunk is from End.
                                  For example, a Overworld chunk have 24 sub chunks, and sub_chunks list is
                                  holds all the sub chunk data for this chunk, so len(sub_chunks) is 24.
        nbts: (list[bytes]): The block entities NBT data of this chunk. For example, if this chunk has 2018
                             NBT blocks, then len(nbts) will be 2018.
                             Note that each element in this list is little endian TAG_Compound NBT.
        chunk_range: (Range, optional):
            The range of this chunk.
            For a Overworld chunk, this is Range(-64, 319);
            for a Nether chunk, this is Range(0, 127);
            for a End chunk, this is Range(0, 255).
            Defaults to RANGE_OVERWORLD.
    """

    sub_chunks: list[bytes] = field(default_factory=lambda: [])
    nbts: list[bytes] = field(default_factory=lambda: [])
    chunk_range: Range = RANGE_OVERWORLD
