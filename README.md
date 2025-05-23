<h1 align="center">Bedrock Chunk Diff</h1>
<h3 align="center">A Minecraft chunk delta update implements based on Go</h3>
<br/>
<p align="center">
<img src="https://forthebadge.com/images/badges/built-with-love.svg">
<p>



[python]: https://img.shields.io/badge/python-3.11-AB70FF?style=for-the-badge
[license]: https://img.shields.io/badge/LICENSE-MIT-228B22?style=for-the-badge



[![][python]](https://www.python.org/)<br/>
[![][license]](LICENSE)<br/>







# Catalogue
- [Catalogue](#catalogue)
- [Summary](#summary)
  - [Aims](#aims)
  - [Details](#details)
  - [Upstream](#upstream)
  - [Supported OS and architecture](#supported-os-and-architecture)
  - [Features](#features)
  - [Recover](#recover)
- [Compatibility](#compatibility)
- [Get started quickly](#get-started-quickly)
- [Note](#note)
  - [Internal Details](#internal-details)
  - [Important](#important)
- [🐍 Pypi Package](#-pypi-package)
- [Others](#others)





# Summary
## Aims
**Bedrock Chunk Diff** is build basd on **Go** language that provide a high speed implements for **Python** that can do delta update operation for Minecraft game saves very fast.



## Details
The finest granularity of delta update is the Chunk.
That means, the user is easily (and also very fast) to record the time point for the Minecraft game saves when the server is running.

So, for a chunk that not loaded, they will nerver get update, then their is no newer time point to be created.
Therefore, we just need to track the chunks that player loaded, so this package provided a very useful delta update implements.

Additionally, we finally used [single file database](https://github.com/etcd-io/bbolt) to record everything, so it's very easy for you to backup the timeline database, just copy one file is OK.

Different to [CoreProtect](https://github.com/PlayPro/CoreProtect), this package is not used for track the single block changes. That means, each time you append a new time point of a chunk to the timeline of this chunk, we are actually creating a snapshot of this chunk. Create snapshot is very helpful for backup the Minecraft game saves, bot not helpful to track the player actions. So, this package is satisfied with large block changes in a single chunk.





## Upstream
This package is based on [bedrock-world-operator](https://github.com/YingLunTown-DreamLand/bedrock-world-operator) that nowadays only support Minecraft `v1.20.51` that align with Minecraft Chinese Version. Therefore, this package can only be used on current Chinese version of Minecraft.

For higher version, you maybe need to modifiy **bedrock-world-operator** and fork this repository to make delta update could running correctly in your server.

Additionally, **bedrock-world-operator** only support the standard Minecraft blocks. For custom blocks, you also need to start a modification.



## Supported OS and architecture
Due to **Bedrock Chunk Diff** is based on **Go**, so we pre-built some dynamic library.

However, maybe some of the systems or architectures are not supportted.
Therefore, if needed, welcome to open new **ISSUE** or **Pull Request**.

- Windows
    * x86_64/amd64
    * x86/i686
- Darwin (MacOS)
    * x86_64/amd64
    * arm64/aarch64
- Linux
    * x86_64/amd64
    * arm64
    * aarch64 (Termux on Android)

By reference [this file](./python/package/internal/load_dynamic_library.py) to know how we load dynamic library in **Python** side.



## Features
- [x] Delta update for blocks in chunk
- [x] Delta update for NBT data in chunk
- [ ] Delta update for map pixel data (Not planned to support, but welcome to open **Pull Request**)
- [ ] Delta update for lodestone data (Not planned to support, but welcome to open **Pull Request**)
- [ ] Delta update for player data (Not planned to support, but welcome to open **Pull Request**)
- [ ] Delta update for mob data in game saves (Not planned to support, but welcome to open **Pull Request**)



## Recover
To use the database to back to a specific time point for each chunk and generated a available Minecraft game save, use [this](./cmd/recover) tools to help you.

Note that we pre-built this tool for some of operating system, and please see [latest release](https://github.com/TriM-Organization/bedrock-chunk-diff/releases/latest) for more information.

Additionally, you can reference these codes to write your own ones, and welcome to contribute more tool for this project.





# Compatibility
`0.0.x` version is still on testing, and we can't ensure all the things are compatibility.





# Get started quickly
```python
from .timeline.define import Range, Dimension, ChunkPos
from .timeline.constant import RANGE_OVERWORLD, RANGE_NETHER, RANGE_END
from .timeline.constant import DIMENSION_OVERWORLD, DIMENSION_NETHER, DIMENSION_END

from .timeline.define import ChunkData
from .timeline.timeline_database import new_timeline_database
```

We export those things above by default.<br/>
Therefore, by using `new_timeline_database`, you can create a new timeline database.

There are multiple functions in each class you get by `new_timeline_database`, and you can do more operation based on them.
We ensure there are enough annotations, so we will not provide extra documents for this project.





# Note
## Internal Details
You can't used any thing that come from package `internal`, because they are our internal implement details.

If you want to start a contribution on this project, then you maybe need to do some research on this package.
But we most suggest you study on `c_api` and `timeline` folder first, because they are our **Go** implements.

## Important
It's unsafe to use `bedrock-chunk-diff` and `bedrock-world-operator` in the same program (See https://github.com/golang/go/issues/65050#issuecomment-1885233457 for more information).

Nowadays we find that **Termux** will have this problem, but for **Windows** and most of the **Linux** devices, everything is work as expected.

Therefore, for **Termux**, you need start multiple programs and use some ways to pass messages between different program when you need use multiple **Python** packages that based on **Go** (e.g. `bedrock-chunk-diff` and `bedrock-world-operator`).





# 🐍 Pypi Package
This package **bedrock-world-operator** is been uploaded to **Pypi** ，and you can use `pip install bedrock-chunk-diff` to install.

See [📦 bedrock-chunk-diff on Pypi](https://pypi.org/project/bedrock-chunk-diff) to learn more.

We used **CD/CI Github Actions**, so if you are the collaborator of this project, you can trigger the workflows by change **version** or use your hand. Then, the robot will compile this project and upload it to **Pypi**.





# Others
This project is licensed under [MIT LICENSE](./LICENSE).
