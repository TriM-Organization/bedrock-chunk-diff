from .types import LIB
from .types import as_c_string, as_python_string
from .types import CInt, CLongLong, CString


LIB.NewTimelineDB.argtypes = [CString, CInt, CInt]
LIB.ReleaseTimelineDB.argtypes = [CLongLong]
LIB.CloseTimelineDB.argtypes = [CLongLong]
LIB.NewChunkTimeline.argtypes = [CLongLong, CInt, CInt, CInt, CInt]
LIB.ReleaseChunkTimeline.argtypes = [CLongLong]
LIB.DeleteChunkTimeline.argtypes = [CLongLong, CInt, CInt, CInt]
LIB.LoadLatestTimePointUnixTime.argtypes = [CLongLong, CInt, CInt, CInt]
LIB.SaveLatestTimePointUnixTime.argtypes = [CLongLong, CInt, CInt, CInt, CLongLong]

LIB.NewTimelineDB.restype = CLongLong
LIB.ReleaseTimelineDB.restype = None
LIB.CloseTimelineDB.restype = CString
LIB.NewChunkTimeline.restype = CLongLong
LIB.ReleaseChunkTimeline.restype = None
LIB.DeleteChunkTimeline.restype = CString
LIB.LoadLatestTimePointUnixTime.restype = CLongLong
LIB.SaveLatestTimePointUnixTime.restype = CString


def new_timeline_db(path: str, no_grow_sync: bool, no_sync: bool) -> int:
    return int(LIB.NewTimelineDB(as_c_string(path), CInt(no_grow_sync), CInt(no_sync)))


def release_timeline_db(id: int) -> None:
    LIB.ReleaseTimelineDB(CLongLong(id))


def tldb_close_timeline_db(id: int) -> str:
    return as_python_string(LIB.CloseTimelineDB(CLongLong(id)))


def tldb_new_chunk_timeline(
    id: int, dm: int, posx: int, posz: int, read_only: bool
) -> int:
    return int(
        LIB.NewChunkTimeline(
            CLongLong(id), CInt(dm), CInt(posx), CInt(posz), CInt(read_only)
        )
    )


def release_chunk_timeline(id: int) -> None:
    LIB.ReleaseChunkTimeline(CLongLong(id))


def tldb_delete_chunk_timeline(id: int, dm: int, posx: int, posz: int) -> str:
    return as_python_string(
        LIB.DeleteChunkTimeline(CLongLong(id), CInt(dm), CInt(posx), CInt(posz))
    )


def tldb_load_latest_time_point_unix_time(
    id: int, dm: int, posx: int, posz: int
) -> int:
    return int(
        LIB.LoadLatestTimePointUnixTime(CLongLong(id), CInt(dm), CInt(posx), CInt(posz))
    )


def tldb_save_latest_time_point_unix_time(
    id: int, dm: int, posx: int, posz: int, time_stamp: int
) -> str:
    return as_python_string(
        LIB.SaveLatestTimePointUnixTime(
            CLongLong(id), CInt(dm), CInt(posx), CInt(posz), CLongLong(time_stamp)
        )
    )
