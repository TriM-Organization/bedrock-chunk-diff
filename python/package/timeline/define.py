from dataclasses import dataclass, field
from bedrockworldoperator import Range, RANGE_OVERWORLD


@dataclass
class ChunkData:
    """
    ChunkData represents the data for a chunk.
    A single chunk could holds the block matrix data and its block entities NBT data.

    Note that for nbts field in this class, we recommend the length of nbts is equal
    to the block entities in this chunk.
    That means, each element in this list is represents the NBT data of one block entity.
    For example, if this chunk has T NBT blocks, then len(nbts) will be T.

    Note that the length of nbts for you is not strict, you can combine multiple block
    entities payload to just one element, this will not case problems.
    However, we granted that the length of the nbts list you get by calling the function
    (next_disk_chunk, next_network_chunk, last_disk_chunk and last_network_chunk) must be
    the number of block entities within this chunk, so that each element in this list is
    just one little endian TAG_Compound NBT.

    Args:
        sub_chunks (list[bytes]): The payload (block matrix data) of this chunk.
                                  The length of this list must equal to 24 if this chunk is from Overworld,
                                  or 8 if this chunk is from Nether, or 16 if this chunk is from End.
                                  For example, a Overworld chunk have 24 sub chunks, and sub_chunks list is
                                  holds all the sub chunk data for this chunk, so len(sub_chunks) is 24.
        nbts: (list[bytes]): The block entities NBT data of this chunk.
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
