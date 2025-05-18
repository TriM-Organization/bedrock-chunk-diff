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
- [Compatibility](#compatibility)
- [Get started quickly](#get-started-quickly)
- [Note](#note)
- [🐍 Pypi Package](#-pypi-package)
- [Others](#others)





# Summary
## Aims
**Bedrock Chunk Diff** is build basd on **Go** language that provide a high speed implements for **Python** that can do delta update operation for Minecraft game saves very fast.



## Details
The finest granularity of delta update is the Chunk.
That means, the user is easily (and also very fast) to record the time point for the Minecraft game saves when the server is running.

So, for a chunk that not loaded, they will nerver to get update, then their is no newer time point to be created.
Therefore, we just need to track the chunks that player loaded, so this package provided a very useful delta update implements.

Additionally, we finally used a single file as the [database](https://github.com/etcd-io/bbolt), so it's very easy for you to backup the timeline database, just copy one is ok...

See [research document](./doc/Sub%20Chunk%20Delta%20Update%20Implements%20Disscussion.pdf) see our research study essay.



## Upstream
This package is based on [bedrock-world-operator](https://github.com/YingLunTown-DreamLand/bedrock-world-operator) that nowadays only support Minecraft `v1.20.51` that align with Minecraft Chinese Version. Therefore, this package can only be used on current Chinese version of Minecraft.

For higher version, you maybe need to modifiy **bedrock-world-operator** and fork this repository to make delta update could running correctly in your server.

Additionally, **bedrock-world-operator** only support the standard Minecraft blocks. For custom blocks, you also need to start a modification.



## Supported OS and architecture
Due to **Bedrock Chunk Diff** is based on **Go**, so we pre-built some dynamic library.

However, maybe some of the systems or architectures are not supportted.
Therefore, if needed, welcome to open new **ISSUE** or **Pull Request**,

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





# Compatibility
`0.0.x` version is still on testing, and we can't ensure all the things are compatibility.





# Get started quickly
```python
from bedrockworldoperator import Range, Dimension, ChunkPos
from bedrockworldoperator import RANGE_OVERWORLD, RANGE_NETHER, RANGE_END
from bedrockworldoperator import DIMENSION_OVERWORLD, DIMENSION_NETHER, DIMENSION_END

from .timeline.timeline_database import new_timeline_database
```

We export those things above by default.<br/>
Therefore, by using `new_timeline_database`, you can create a new timeline database.

There are multiple functions in each class you get by `new_timeline_database`, and you can do more operation based on them.
We ensure there are enough annotations, so we will not provide extra documents for this project.





# Note
You can't used any thing that come from package `internal`, because they are our internal implement details.

If you want to start a contribution on this project, then you maybe need to do some research on this package.
But we most suggest you study on `c_api` and `timeline` folder first, because they are our **Go** implements.





# 🐍 Pypi Package
This package **bedrock-world-operator** is been uploaded to **Pypi** ，and you can use `pip install bedrock-chunk-diff` to install.

See [📦 bedrock-chunk-diff on Pypi](https://pypi.org/project/bedrock-world-operator) to learn more.

We used **CD/CI Github Actions**, so if you are the collaborator of this project, you can trigger the workflows by change **version** or use your hand. Then, the robot will compile this project and upload it to **Pypi**.





# Others
This project is licensed under [MIT LICENSE](./LICENSE).
