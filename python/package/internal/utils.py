import struct
from io import BytesIO


def unpack_bytes_list(payload: bytes) -> list[bytes]:
    result: list[bytes] = []
    ptr = 0

    while ptr < len(payload):
        length: int = struct.unpack("<I", payload[ptr : ptr + 4])[0]
        result.append(payload[ptr + 4 : ptr + 4 + length])
        ptr = ptr + 4 + length

    return result


def pack_bytes_list(sub_chunks: list[bytes]) -> bytes:
    w = BytesIO()
    for i in sub_chunks:
        w.write(struct.pack("<I", len(i)))
        w.write(i)
    return w.getvalue()


def unpack_next_or_last(
    payload: bytes, read_is_last_element: bool
) -> tuple[list[bytes], int, int, list[bytes], int, bool, bool]:
    if len(payload) == 0:
        return [], 0, 0, [], 0, False, False
    r = BytesIO(payload)

    length: int = struct.unpack("<I", r.read(4))[0]
    chunk_payload_bytes = r.read(length)
    sub_chunks = unpack_bytes_list(chunk_payload_bytes)

    range_start: int = struct.unpack("<h", r.read(2))[0]
    range_end: int = struct.unpack("<h", r.read(2))[0]

    length = struct.unpack("<I", r.read(4))[0]
    nbt_payload = r.read(length)
    nbts = unpack_bytes_list(nbt_payload)

    update_unix_time: int = struct.unpack("<q", r.read(8))[0]

    is_last_element = False
    if read_is_last_element:
        is_last_element = bool(int(r.read(1)))

    return (
        sub_chunks,
        range_start,
        range_end,
        nbts,
        update_unix_time,
        is_last_element,
        True,
    )
