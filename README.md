<div align="center">

# Filediver

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/xypwn/filediver/.github%2Fworkflows%2Fbuild-release.yml)](https://github.com/xypwn/filediver/actions)
[![GitHub License](https://img.shields.io/github/license/xypwn/filediver)](https://opensource.org/license/bsd-3-clause)

[![GitHub Release](https://img.shields.io/github/v/release/xypwn/filediver)](https://github.com/xypwn/filediver/releases/latest/)
[![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/xypwn/filediver/total)](https://github.com/xypwn/filediver/releases/latest/)

An unofficial Helldivers 2 game asset extractor.
</div>

## Download
### Filediver GUI (graphical interface with preview)
![GUI Screenshot](screenshots/gui.png)
- [Download GUI Windows (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-gui-windows.exe)
- [Download GUI Linux (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-gui-linux)

**Simply download and run. It's highly recommended to install all extensions when prompted.**

### Filediver CLI (command-line interface)
- [Download CLI Windows (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-cli-windows.zip)
- [Download CLI Linux (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-cli-linux.zip)

**Extract the archive. This will create a folder named `filediver`, where everything relevant is located.**

The program is called "filediver.exe" (or just "filediver" on Linux). See [usage](#usage).

### Helper scripts (scripts_dist)
- [Windows (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/scripts-dist-windows.zip)
- [Linux (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/scripts-dist-linux.tar.xz)

**CLI Only: Extract the helper scripts achive into the `filediver` folder (the same folder containing the executable, e.g. `filediver.exe`).**

## Export features
- Video to bink/mp4
- Audio to ogg/aac/wav
- Images/textures to png
- 3D models to gltf/blender (with bones, textures and animations [needs flag])
- Prefabs
- Text tables to JSON
- ...and more

## Usage
**See the [Wiki](https://github.com/xypwn/filediver/wiki)**
  - [CLI Basics](https://github.com/xypwn/filediver/wiki/10-CLI-Basics)

## Links
- [HD 2 Archive Labelling](https://docs.google.com/spreadsheets/d/1oQys_OI5DWou4GeRE3mW56j7BIi4M7KftBIPAl1ULFw) (IDs can be used with -t option)
- [Helldivers Archive Discord server](https://discord.gg/helldiversarchive)

## Credits
This app builds on a lot of work from other people. This includes:
- [Hellextractor by Xaymar](https://github.com/Xaymar/Hellextractor)
	- Basic binary file structure
	- Unhashed resource names/types (.txt files)
- [vgmstream](https://github.com/vgmstream/vgmstream), [ww2ogg by hcs](https://github.com/hcs64/ww2ogg) and [bnkextr by eXpl0it3r](https://github.com/eXpl0it3r/bnkextr)
	- Wwise audio formats
- [ImageMagick](https://imagemagick.org)
	- DDS texture decoding
- [Accurate HD2 Shader by Thejudsub](https://discord.com/channels/1210541115829260328/1222290154409033889) on [the Helldivers Archive Discord server](https://discord.gg/helldiversarchive)
	- The most accurate Blender material replicating the game's procedural shaders

## Hacking
### Running
- `go run ./cmd/filediver-cli` / `go run ./cmd/filediver-gui`
### Setup blender importer for development
- `uv venv --python 3.11`
- `uv pip install -r scripts/requirements.txt`
- set env `FILEDIVER_BLENDER_IMPORTER_COMMAND="uv run scripts/hd2_accurate_blender_importer.py"`

## License
Copyright (c) filediver contributors

FileDiver is licensed under the 3-Clause BSD License (https://opensource.org/license/bsd-3-clause).
