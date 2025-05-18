import numpy
from dataclasses import dataclass
from bedrockworldoperator import Range
from .define import ChunkData
from ..internal.symbol_export_timeline_db import release_chunk_timeline
from ..internal.symbol_export_chunk_timeline import (
    ctl_all_time_point,
    ctl_all_time_point_len,
    ctl_append_disk_chunk,
    ctl_append_network_chunk,
    ctl_compact,
    ctl_empty,
    ctl_last_disk_chunk,
    ctl_last_network_chunk,
    ctl_next_disk_chunk,
    ctl_next_network_chunk,
    ctl_pop,
    ctl_read_only,
    ctl_save,
    ctl_set_max_limit,
)


@dataclass
class ChunkTimeline:
    """
    ChunkTimeline records the timeline of a chunk,
    and it contains the change logs about this chunk
    on this timeline.

    In other words, the ChunkTimeline holds the history
    of this chunk.

    Note that it's unsafe for multiple thread to access this
    struct due to we don't use mutex to ensure the operation
    is atomic.

    So, it's your responsibility to make ensure there is only
    one thread is using this object.

    Additionally, before you use this object, please ensure
    you use ChunkTimeline.is_valid() to check whether the timeline
    is valid or not.
    Note that check only need do just once, and after you do the
    check, you don't need to do more checks when you use the same object.
    """

    _chunk_timeline_id: int = -1

    def __del__(self):
        if self._chunk_timeline_id >= 0 and release_chunk_timeline is not None:
            release_chunk_timeline(self._chunk_timeline_id)

    def is_valid(self) -> bool:
        """
        is_valid check current chunk timeline is valid or not.

        If not valid, it means the chunk timeline actually not exist,
        not only Python but also in Go.

        Try to use an invalid chunk timeline is not allowed,
        and any operation will be terminated.

        Returns:
            bool: Whether the chunk timeline is valid or not.
        """
        return self._chunk_timeline_id >= 0

    def append_disk_chunk(self, chunk_data: ChunkData):
        """
        append_disk_chunk tries append a new chunk whose is disk
        encoding to the timeline of current chunk. Additionally,
        we will also append the block entities of this chunk.

        If the size of timeline will overflow max limit, then we will
        firstly pop some time point from the underlying timeline.

        Note the poped time points must be the most earliest one.

        If current timeline is read only, then calling append_disk_chunk
        will do no operation.

        Args:
            chunk_data (ChunkData): The chunk you want to append to the timeline.

        Raises:
            Exception: When failed to append the chunk.
        """
        err = ctl_append_disk_chunk(
            self._chunk_timeline_id,
            chunk_data.sub_chunks,
            chunk_data.nbts,
            chunk_data.chunk_range.start_range,
            chunk_data.chunk_range.end_range,
        )
        if len(err) > 0:
            raise Exception(err)

    def append_network_chunk(self, chunk_data: ChunkData):
        """
        append_network_chunk tries append a new chunk whose is network
        encoding to the timeline of current chunk.
        Additionally, we will also append the block entities of this chunk.

        If the size of timeline will overflow max limit, then we will
        firstly pop some time point from the underlying timeline.

        Note the poped time points must be the most earliest one.

        If current timeline is read only, then calling append_network_chunk
        will do no operation.

        Args:
            chunk_data (ChunkData): The chunk you want to append to the timeline.

        Raises:
            Exception: When failed to append the chunk.
        """
        err = ctl_append_network_chunk(
            self._chunk_timeline_id,
            chunk_data.sub_chunks,
            chunk_data.nbts,
            chunk_data.chunk_range.start_range,
            chunk_data.chunk_range.end_range,
        )
        if len(err) > 0:
            raise Exception(err)

    def empty(self) -> bool:
        """
        empty returns whether this timeline is empty or not.
        If is empty, then calling Save will result in no operation.

        Returns:
            bool: Return True for exist.
                  Return False for not exist or current timeline is not exist.
        """
        result = ctl_empty(self._chunk_timeline_id)
        return result == 1

    def read_only(self) -> bool:
        """
        read_only returns whether this timeline is read only or not.
        If is read only, then calling any function that will modify
        underlying timeline will result in no operation.

        Returns:
            bool: Return True for this timeline is read only.
                  Return False for this timeline could be modified,
                  or this timeline is not exist.
        """
        result = ctl_read_only(self._chunk_timeline_id)
        return result == 1

    def all_time_point(self) -> numpy.ndarray:
        """
        all_time_point returns a list that holds
        the unix time of all time points this timeline
        have.

        Granted the returned array is non-decreasing.

        Note that the retuened list is read only.
        If need to modify, please copy a new one.

        Returns:
            numpy.ndarray: The list that holds the unix time
                           for all time points in this timeline.
                           If is empty, then this timeline have
                           no time point or this timeline is not
                           exist.
        """
        return ctl_all_time_point(self._chunk_timeline_id)

    def all_time_point_len(self) -> int:
        """
        all_time_point_len returns the length of
        the time point that this timeline have.

        Returns:
            int: The length of this timeline.
        """
        return ctl_all_time_point_len(self._chunk_timeline_id)

    def set_max_limit(self, max_limit: int):
        """
        set_max_limit sets the timeline could record how many time point.
        max_limit must bigger than 0. If less, then set the limit to 1.

        After calling set_max_limit if overflow immediately, then we will
        pop some time point from the underlying timeline.
        Poped time points must be the most earliest one.

        Note that calling set_max_limit will not change the empty states
        of this timeline.

        If current timeline is read only, then calling set_max_limit will
        do no operation.

        Args:
            max_limit (int): The max limit of this timeline.

        Raises:
            Exception: When failed to update the max limit.
        """
        err = ctl_set_max_limit(self._chunk_timeline_id, max_limit)
        if len(err) > 0:
            raise Exception(err)

    def compact(self):
        """
        compact compacts the underlying block palette as much as possible, try to delete all
        unused blocks from it.

        If current timeline is empty or read only, then calling compact will do no operation.

        compact is very expensive due to its time complexity is O(C×k×4096×N×L).
            - k is the count of sub chunks that this chunk have.
            - N is the count of time point that this timeline have.
            - L is the average count of layers for each sub chunks in this timeline.
            - C is a little big (bigger than 2) due to there are multiple difference/prefix-sum
              operations need to do.

        Raises:
            Exception: When failed to compact the underlying block palette.
        """
        err = ctl_compact(self._chunk_timeline_id)
        if len(err) > 0:
            raise Exception(err)

    def next_disk_chunk(self) -> tuple[ChunkData, int, bool]:
        """
        next_disk_chunk gets the next time point of current chunk and the NBT blocks in it.
        Note that the returned ChunkData is in disk encoding.

        With the call to next_disk_chunk, we granted that the returned time keeps increasing
        until the entire time series is traversed.

        When it is already at the end of the timeline, calling next_disk_chunk again will back
        to the earliest time point.
        In other words, next_disk_chunk is self-loop and can be called continuously.

        Returns:
            tuple[ChunkData, int, bool]:
                The chunk data in the next time point of this current chunk timeline.
                Returned int is the update unix time of this time point.
                The returned bool can inform whether the element obtained after the
                current call to next_disk_chunk is at the end of the time series.
        """
        sub_chunks, range_start, range_end, nbts, update_unix_time, is_last_element = (
            ctl_next_disk_chunk(self._chunk_timeline_id)
        )
        return (
            ChunkData(sub_chunks, nbts, Range(range_start, range_end)),
            update_unix_time,
            is_last_element,
        )

    def next_network_chunk(self) -> tuple[ChunkData, int, bool]:
        """
        next_network_chunk gets the next time point of current chunk and the NBT blocks in it.
        Note that the returned ChunkData is in network encoding.

        With the call to next_network_chunk, we granted that the returned time keeps increasing
        until the entire time series is traversed.

        When it is already at the end of the timeline, calling next_network_chunk again will back
        to the earliest time point.
        In other words, next_network_chunk is self-loop and can be called continuously.

        Returns:
            tuple[ChunkData, int, bool]:
                The chunk data in the next time point of this current chunk timeline.
                Returned int is the update unix time of this time point.
                The returned bool can inform whether the element obtained after the
                current call to next_network_chunk is at the end of the time series.
        """
        sub_chunks, range_start, range_end, nbts, update_unix_time, is_last_element = (
            ctl_next_network_chunk(self._chunk_timeline_id)
        )
        return (
            ChunkData(sub_chunks, nbts, Range(range_start, range_end)),
            update_unix_time,
            is_last_element,
        )

    def last_disk_chunk(self) -> tuple[ChunkData, int]:
        """
        last_disk_chunk gets the latest time point
        of current chunk and the NBT blocks in it.

        Time complexity: O(1).

        Returns:
            tuple[ChunkData, int, bool]:
                The chunk data who is encoded in disk encoding.
                Returned int is the update unix time of the time point.
        """
        sub_chunks, range_start, range_end, nbts, update_unix_time = (
            ctl_last_disk_chunk(self._chunk_timeline_id)
        )
        return ChunkData(
            sub_chunks, nbts, Range(range_start, range_end)
        ), update_unix_time

    def last_network_chunk(self) -> tuple[ChunkData, int]:
        """
        last_network_chunk gets the latest time point
        of current chunk and the NBT blocks in it.

        Time complexity: O(1).

        Returns:
            tuple[ChunkData, int, bool]:
                The chunk data who is encoded in network encoding.
                Returned int is the update unix time of the time point.
        """
        sub_chunks, range_start, range_end, nbts, update_unix_time = (
            ctl_last_network_chunk(self._chunk_timeline_id)
        )
        return ChunkData(
            sub_chunks, nbts, Range(range_start, range_end)
        ), update_unix_time

    def pop(self):
        """
        pop tries to delete the first time point from this timeline.

        If current timeline is empty, or it is read only, or there is
        only one time point, then we will do no operation.

        Raises:
            Exception: When failed to pop.
        """
        err = ctl_pop(self._chunk_timeline_id)
        if len(err) > 0:
            raise Exception(err)

    def save(self):
        """
        save saves current timeline into the underlying database, and also release current timeline.

        Read only timeline should also calling save to release the resource.
        But read only timeline calling this function will only release but don't do further operation.
        Additionally, empty non read only timeline is also follow the same behavior.

        Note that you could use s.Empty() and s.ReadOnly() to check.

        If you calling save with no exception, then this timeline is released and can't be used again.
        Also, you can't call save multiple times.
        But, if exception happened, then this object will not released.

        Note that we will not check whether it has been released, nor will we check whether you have called
        save multiple times.

        save must calling at the last modification of the timeline; otherwise, the timeline will not be able
        to maintain data consistency.

        Raises:
            Exception: When failed to save this timeline.
        """
        err = ctl_save(self._chunk_timeline_id)
        if len(err) > 0:
            raise Exception(err)
