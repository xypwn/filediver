<div align="center">

# FileDiver

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/xypwn/filediver/.github%2Fworkflows%2Fbuild-release.yml)](https://github.com/xypwn/filediver/actions)
[![Scrutinizer quality (GitHub/Bitbucket)](https://img.shields.io/scrutinizer/quality/g/xypwn/filediver)](https://scrutinizer-ci.com/g/xypwn/filediver)
[![GitHub License](https://img.shields.io/github/license/xypwn/filediver)](https://opensource.org/license/bsd-3-clause)

[![GitHub Release](https://img.shields.io/github/v/release/xypwn/filediver)](https://github.com/xypwn/filediver/releases/latest/)
[![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/xypwn/filediver/total)](https://github.com/xypwn/filediver/releases/latest/)

An unofficial Helldivers 2 game asset extractor.
</div>

## Download
- [Windows (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-windows-amd64.zip)
- [Linux (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-linux-amd64.tar.gz)

**Extract the achive into a new folder.**

The program is called "filediver.exe" (or just "filediver" on Linux). See [usage](#usage).

<details>
<summary>What is "ffmpeg.exe"?</summary>

"ffmpeg.exe" ([FFmpeg](https://ffmpeg.org/)) is used for converting video and audio files. It is downloaded from an official source by the [GitHub workflow](https://github.com/xypwn/filediver/blob/master/.github/workflows/build-release.yml) that generates the .zip archive you can download.

You only need to keep it in the folder if you don't have it installed on your computer already.
</details>

## Usage
### Windows
Navigate to the folder where you unpacked the program into. `SHIFT`+`Right-Click` **in** the folder and select "Open in PowerShell".

In PowerShell/Terminal, run `./filediver -h` to get a list of options.

### Here are some example commands:

Simply running the app should automatically detect your installation directory and dump all files into the "extracted" directory in your current folder:
```sh
./filediver
```

Print a detailed description of all command line options:
```sh
./filediver -h
```

Extract the files into a directory called "custom_dir":
```sh
./filediver -o "custom_dir"
```

Extract only video files:
```sh
./filediver -c "enable:video"
```

Extract the Super Earth anthem as mp3:
```sh
./filediver -c "audio:format=mp3" -i "content/audio/291227525.wwise_stream"
```

## Features
### File Types/Formats
- **Audio**: Audiokinetic wwise bnk/wem; automatically converted to WAV; other formats require FFmpeg
- **Video**: Bink; automatically converted to MP4 via FFmpeg (shipped with Windows binary)
- **Textures**: Direct Draw Surface (.dds); automatically converted to PNG
- **Models (WIP)**: Stingray Unit; automatically converted to GLB (=glTF); can be imported into [Blender](https://www.blender.org/); for importing bones, see [Importing Bones](#importing-bones)

Planned: animations

### Importing Bones
When importing the .glb into blender, you need to change the "Bone Dir" option from "Blender" to "Temperance", or you will see huge spheres for bones.

### Thejudsub's Accurate Shader
.glb models exported from filediver can be imported into Blender with the accurate shader pre-applied, saving lots of manual work finding and applying textures:

(Prerequisites: `Python 3.11.*` must be installed on your system - Blender 4.1-4.3+ only supports specifically Python 3.11)
1. Ensure you've setup your environment by running `./scripts/setup_environment.ps1` (on Windows) or `bash ./scripts/setup_environment.sh` (on Linux). These scripts will verify you have the correct Python version and will install the virtual environment
2. Export a model that uses procedural materials (most armor pieces and weapons do)
3. Activate the virtual environment with `&./scripts/.venv/Scripts/Activate.ps1` (Windows - powershell), `./scripts/.venv/Scripts/activate.bat` (Windows - cmd), or `source ./scripts/.venv/bin/activate` (Linux - most shells)
4. Run `python ./scripts/hd2_accurate_blender_importer.py path/to/filediver/exported.glb path/to/output.blend`
5. `path/to/output.blend` will be a _**new, completely fresh/overwritten**_ blend file containing the exported models with the shader applied.
6. Run `deactivate` to leave the python virtual environment.

## Credits/Links
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

Some useful discussion on the topic of HD2 resource extraction: https://reshax.com/topic/507-helldivers-2-model-extraction-help/

## Hacking
- Install [Go](https://go.dev/dl/)
- `go run ./cmd/filediver-cli`

## License
Copyright (c) Darwin Schuppan and contributors

FileDiver is licensed under the 3-Clause BSD License (https://opensource.org/license/bsd-3-clause).
