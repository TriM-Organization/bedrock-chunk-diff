import numpy
from .types import LIB
from .types import CInt, CLongLong, CString, CSlice
from .types import as_c_bytes, as_python_bytes, as_python_string
from .utils import pack_bytes_list, unpack_next_or_last


LIB.AppendDiskChunk.argtypes = [CLongLong, CSlice, CSlice, CInt, CInt, CInt]
LIB.AppendNetworkChunk.argtypes = [CLongLong, CSlice, CSlice, CInt, CInt, CInt]
LIB.Empty.argtypes = [CLongLong]
LIB.ReadOnly.argtypes = [CLongLong]
LIB.Pointer.argtypes = [CLongLong]
LIB.ResetPointer.argtypes = [CLongLong]
LIB.AllTimePoint.argtypes = [CLongLong]
LIB.AllTimePointLen.argtypes = [CLongLong]
LIB.SetMaxLimit.argtypes = [CLongLong, CInt]
LIB.Compact.argtypes = [CLongLong]
LIB.NextDiskChunk.argtypes = [CLongLong]
LIB.NextNetworkChunk.argtypes = [CLongLong]
LIB.JumpToDiskChunk.argtypes = [CLongLong, CInt]
LIB.JumpToNetworkChunk.argtypes = [CLongLong, CInt]
LIB.LastDiskChunk.argtypes = [CLongLong]
LIB.LastNetworkChunk.argtypes = [CLongLong]
LIB.Pop.argtypes = [CLongLong]
LIB.Save.argtypes = [CLongLong]

LIB.AppendDiskChunk.restype = CString
LIB.AppendNetworkChunk.restype = CString
LIB.Empty.restype = CInt
LIB.ReadOnly.restype = CInt
LIB.Pointer.restype = CInt
LIB.ResetPointer.restype = CString
LIB.AllTimePoint.restype = CSlice
LIB.AllTimePointLen.restype = CInt
LIB.SetMaxLimit.restype = CString
LIB.Compact.restype = CString
LIB.NextDiskChunk.restype = CSlice
LIB.NextNetworkChunk.restype = CSlice
LIB.JumpToDiskChunk.restype = CSlice
LIB.JumpToNetworkChunk.restype = CSlice
LIB.LastDiskChunk.restype = CSlice
LIB.LastNetworkChunk.restype = CSlice
LIB.Pop.restype = CString
LIB.Save.restype = CString


def ctl_append_disk_chunk(
    id: int,
    chunk_payload: list[bytes],
    nbt_payload: list[bytes],
    range_start: int,
    range_end: int,
    nop_when_no_change: bool,
) -> str:
    return as_python_string(
        LIB.AppendDiskChunk(
            CLongLong(id),
            as_c_bytes(pack_bytes_list(chunk_payload)),
            as_c_bytes(b"".join(nbt_payload)),
            CInt(range_start),
            CInt(range_end),
            CInt(nop_when_no_change),
        )
    )


def ctl_append_network_chunk(
    id: int,
    chunk_payload: list[bytes],
    nbt_payload: list[bytes],
    range_start: int,
    range_end: int,
    nop_when_no_change: bool,
) -> str:
    return as_python_string(
        LIB.AppendNetworkChunk(
            CLongLong(id),
            as_c_bytes(pack_bytes_list(chunk_payload)),
            as_c_bytes(b"".join(nbt_payload)),
            CInt(range_start),
            CInt(range_end),
            CInt(nop_when_no_change),
        )
    )


def ctl_empty(id: int) -> int:
    return int(LIB.Empty(CLongLong(id)))


def ctl_read_only(id: int) -> int:
    return int(LIB.ReadOnly(CLongLong(id)))


def ctl_pointer(id: int) -> int:
    return int(LIB.Pointer(CLongLong(id)))


def ctl_reset_pointer(id: int) -> str:
    return as_python_string(LIB.ResetPointer(CLongLong(id)))


def ctl_all_time_point(id: int) -> numpy.ndarray:
    return numpy.frombuffer(
        as_python_bytes(LIB.AllTimePoint(CLongLong(id))), dtype="<i8"
    )


def ctl_all_time_point_len(id: int) -> int:
    return int(LIB.AllTimePointLen(CLongLong(id)))


def ctl_set_max_limit(id: int, max_limit: int) -> str:
    return as_python_string(LIB.SetMaxLimit(CLongLong(id), CInt(max_limit)))


def ctl_compact(id: int) -> str:
    return as_python_string(LIB.Compact(CLongLong(id)))


def ctl_next_disk_chunk(
    id: int,
) -> tuple[list[bytes], int, int, list[bytes], int, bool, bool]:
    return unpack_next_or_last(as_python_bytes(LIB.NextDiskChunk(CLongLong(id))), True)


def ctl_next_network_chunk(
    id: int,
) -> tuple[list[bytes], int, int, list[bytes], int, bool, bool]:
    return unpack_next_or_last(
        as_python_bytes(LIB.NextNetworkChunk(CLongLong(id))), True
    )


def ctl_jump_to_disk_chunk(
    id: int, index: int
) -> tuple[list[bytes], int, int, list[bytes], int, bool, bool]:
    return unpack_next_or_last(
        as_python_bytes(LIB.JumpToDiskChunk(CLongLong(id), CInt(index))), False
    )


def ctl_jump_to_network_chunk(
    id: int, index: int
) -> tuple[list[bytes], int, int, list[bytes], int, bool, bool]:
    return unpack_next_or_last(
        as_python_bytes(LIB.JumpToNetworkChunk(CLongLong(id), CInt(index))), False
    )


def ctl_last_disk_chunk(
    id: int,
) -> tuple[list[bytes], int, int, list[bytes], int, bool]:
    sub_chunks, range_start, range_end, nbts, update_unix_time, _, success = (
        unpack_next_or_last(as_python_bytes(LIB.LastDiskChunk(CLongLong(id))), False)
    )
    return sub_chunks, range_start, range_end, nbts, update_unix_time, success


def ctl_last_network_chunk(
    id: int,
) -> tuple[list[bytes], int, int, list[bytes], int, bool]:
    sub_chunks, range_start, range_end, nbts, update_unix_time, _, success = (
        unpack_next_or_last(as_python_bytes(LIB.LastNetworkChunk(CLongLong(id))), False)
    )
    return sub_chunks, range_start, range_end, nbts, update_unix_time, success


def ctl_pop(id: int) -> str:
    return as_python_string(LIB.Pop(CLongLong(id)))


def ctl_save(id: int) -> str:
    return as_python_string(LIB.Save(CLongLong(id)))
