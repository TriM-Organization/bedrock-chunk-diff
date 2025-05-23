from .define import Dimension, ChunkPos
from .constant import DIMENSION_OVERWORLD
from dataclasses import dataclass
from .chunk_timeline import ChunkTimeline
from ..internal.symbol_export_timeline_db import (
    new_timeline_db,
    release_timeline_db,
    tldb_close_timeline_db,
    tldb_delete_chunk_timeline,
    tldb_load_latest_time_point_unix_time,
    tldb_new_chunk_timeline,
    tldb_save_latest_time_point_unix_time,
)


@dataclass
class TimelineDatabase:
    """
    TimelineDatabase wrapper and implements all features from Timeline,
    and as a provider to provide timeline of chunk related functions.
    """

    _database_id: int = -1

    def __del__(self):
        if self._database_id >= 0 and release_timeline_db is not None:
            release_timeline_db(self._database_id)

    def is_valid(self) -> bool:
        """
        is_valid check current timeline database is valid or not.

        If not valid, it means the timeline database actually not exist,
        not only Python but also in Go.

        Try to use an invalid timeline database is not allowed,
        and any operation will be terminated.

        Returns:
            bool: Whether the timeline database is valid or not.
        """
        return self._database_id >= 0

    def close_timeline_db(self):
        """
        close_timeline_db closes the timeline database.

        It will wait until all the timelines in use are
        released before closing the database.

        Raises:
            Exception: When failed to close the timeline database.
        """
        err = tldb_close_timeline_db(self._database_id)
        if len(err) > 0:
            raise Exception(err)

    def new_chunk_timeline(
        self,
        pos: ChunkPos,
        read_only: bool = False,
        dm: Dimension = DIMENSION_OVERWORLD,
    ) -> ChunkTimeline:
        """
        new_chunk_timeline gets the timeline of a chunk who is at pos.
        If current timeline database is not exist or meet error, the you get a invalid ChunkTimeline.
        Therefore, you need use ChunkTimeline.is_valid() to check whether the timeline you get is valid or not.

        Note that if timeline of current chunk is not exist, then we will not create a timeline
        but return an empty one so you can modify it. The time to create the timeline is only when you
        save a timeline that not empty to the database.

        If read_only is true, then returned a timeline but only can read.
        For a read only timeline, you also need use ChunkTimeline.save to release it.

        Important:
            - Once any modifications have been made to the returned timeline, you must save them
              at the end; otherwise, the timeline will not be able to maintain data consistency
              (only need to save at the last modification).

            - Timeline of one chunk can't be using by multiple threads. Therefore, you will
              get blocking when a thread calling new_chunk_timeline but there is still some
              threads are using target chunk.

            - Calling ChunkTimeline.save to release the timeline.

            - Returned ChunkTimeline can't shared with multiple threads, and it's your responsibility
              to ensure this thing.

        Args:
            pos (ChunkPos): The chunk position of the target chunk.
            read_only (bool, optional): You want to the target timeline is read only or not.
                                        Defaults to False.
            dm (Dimension, optional): The dimension of the target chunk.
                                      Defaults to DIMENSION_OVERWORLD.
        """
        return ChunkTimeline(
            tldb_new_chunk_timeline(self._database_id, int(dm), pos.x, pos.z, read_only)
        )

    def delete_chunk_timeline(self, pos: ChunkPos, dm: Dimension = DIMENSION_OVERWORLD):
        """
        delete_chunk_timeline deletes the timeline of chunk who at pos.
        If timeline is not exist, then do no operation.

        Time complexity: O(n).
        n is the time point that this chunk have.

        Args:
            pos (ChunkPos): The chunk position of the target chunk.
            dm (Dimension, optional): The dimension of the target chunk.
                                      Defaults to DIMENSION_OVERWORLD.

        Raises:
            Exception: When failed to delete target timeline.
        """
        err = tldb_delete_chunk_timeline(self._database_id, int(dm), pos.x, pos.z)
        if len(err) > 0:
            raise Exception(err)

    def load_latest_time_point_unix_time(
        self, pos: ChunkPos, dm: Dimension = DIMENSION_OVERWORLD
    ):
        """
        load_latest_time_point_unix_time loads
        the time when latest time point update.

        If not exist, then return 0.

        Args:
            pos (ChunkPos): The chunk position of the target chunk.
            dm (Dimension, optional): The dimension of the target chunk.
                                      Defaults to DIMENSION_OVERWORLD.

        Returns:
            int: The unix time of the latest time point.
                 Return 0 for not exist or current timeline database is not exist.
        """
        result = tldb_load_latest_time_point_unix_time(
            self._database_id, int(dm), pos.x, pos.z
        )
        if result == -1:
            return 0
        return result

    def save_latest_time_point_unix_time(
        self, pos: ChunkPos, time_stamp: int, dm: Dimension = DIMENSION_OVERWORLD
    ):
        """
        save_latest_time_point_unix_timesaves the time
        when the latest time point is generated.

        If timeStamp is 0, then delete the time from the database.

        Args:
            pos (ChunkPos): The chunk position of the target chunk.
            time_stamp (int): The unix time to update.
            dm (Dimension, optional): The dimension of the target chunk.
                                      Defaults to DIMENSION_OVERWORLD.

        Raises:
            Exception: When failed to update the unix time.
        """
        err = tldb_save_latest_time_point_unix_time(
            self._database_id, int(dm), pos.x, pos.z, time_stamp
        )
        if len(err) > 0:
            raise Exception(err)


def new_timeline_database(
    path: str, no_grow_sync: bool = False, no_sync: bool = False
) -> TimelineDatabase:
    """
    new_timeline_database open a level database that used for
    chunk delta update whose at path.

    If not exist, then create a new database.

    Note that you could use TimelineDatabase.is_valid() to check
    whether the timeline database is valid or not.

    Args:
        path (str): The path of the timeline database want to open or create.
        no_grow_sync (bool, optional): When no_grow_sync is true, skips the truncate call when growing the database.
                             Setting this to true is only safe on non-ext3/ext4 systems.
                             Skipping truncation avoids preallocation of hard drive space and
                             bypasses a truncate() and fsync() syscall on remapping.
                                - See also: https://github.com/boltdb/bolt/issues/284
                            Defaults to False.
        no_sync (bool, optional): Setting the no_sync flag will cause the database to skip fsync()
                        calls after each commit. This can be useful when bulk loading data
                        into a database and you can restart the bulk load in the event of
                        a system failure or database corruption. Do not set this flag for
                        normal use.
                        THIS IS UNSAFE. PLEASE USE WITH CAUTION.
                        Defaults to False.

    Returns:
        TimelineDatabase: The opened timeline database.
    """
    return TimelineDatabase(new_timeline_db(path, no_grow_sync, no_sync))
