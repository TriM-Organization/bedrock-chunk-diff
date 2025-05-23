from dataclasses import dataclass, field


@dataclass(frozen=True)
class Range:
    """
    Range represents the height range of a Dimension in blocks.

    The first value of the Range holds the minimum Y value,
    the second value holds the maximum Y value.

    Note that Range is a hashable and cannot be further modified object.
    """

    start_range: int = 0
    end_range: int = 0


@dataclass(frozen=True)
class Dimension:
    """
    Dimension is a dimension of a World.

    It influences a variety of properties of a World such as the building range,
    the sky color and the behavior of liquid blocks.

    Note that Dimension is a hashable and cannot be further modified object.

    Args:
        dm (int): The id of this dimension.
    """

    dm: int = 0

    def __int__(self) -> int:
        return self.dm

    def range(self) -> Range:
        """range returns the range that player could build block in this dimension.

        Returns:
            Range: The range that player could build block in this dimension.
                   If this dimension is not standard dimension, then redirect
                   to Overworld range.
        """
        match self.dm:
            case 0:
                return Range(-64, 319)
            case 1:
                return Range(0, 127)
            case 2:
                return Range(0, 255)
            case _:
                return Range(-64, 319)

    def height(self) -> int:
        """
        height returns the height of this dimension.

        For example, the height of Overworld is 384
        due to "384 = 319 - (-64) + 1", and 319 is
        the max Y that Overworld could build, and -64
        is the min Y that Overworld could build.

        Returns:
            int: The height of this dimension.
                 If this dimension is not standard dimension, then redirect
                 to Overworld height.
        """
        match self.dm:
            case 0:
                return 384
            case 1:
                return 128
            case 2:
                return 256
            case _:
                return 384

    def __str__(self) -> str:
        match self.dm:
            case 0:
                return "Overworld"
            case 1:
                return "Nether"
            case 2:
                return "End"
            case _:
                return f"Custom (id={self.dm})"


@dataclass(frozen=True)
class ChunkPos:
    """
    ChunkPos holds the position of a chunk. The type is provided as a utility struct for keeping track of a
    chunk's position.

    Chunks do not themselves keep track of that. Chunk positions are different from block positions in the
    way that increasing the X/Z by one means increasing the absolute value on the X/Z axis in terms of blocks
    by 16.

    Note that ChunkPos is a hashable and cannot be further modified object.
    """

    x: int = 0
    z: int = 0


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
            Defaults to Range(-64, 319).
    """

    sub_chunks: list[bytes] = field(default_factory=lambda: [])
    nbts: list[bytes] = field(default_factory=lambda: [])
    chunk_range: Range = Range(-64, 319)
