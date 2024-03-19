<div align="center">

# FileDiver

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/xypwn/filediver/.github%2Fworkflows%2Fbuild-release.yml)](https://github.com/xypwn/filediver/releases/latest/)
[![CodeFactor Grade](https://img.shields.io/codefactor/grade/github/xypwn/filediver)](https://www.codefactor.io/repository/github/xypwn/filediver)
[![GitHub License](https://img.shields.io/github/license/xypwn/filediver)](https://opensource.org/license/bsd-3-clause)

[![GitHub Release](https://img.shields.io/github/v/release/xypwn/filediver)](https://github.com/xypwn/filediver/releases/latest/)
[![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/xypwn/filediver/total)](https://github.com/xypwn/filediver/releases/latest/)

An unofficial Helldivers 2 game asset extractor.
</div>

## Download
- [Windows (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-windows-amd64.zip)
- [Linux (64-bit)](https://github.com/xypwn/filediver/releases/latest/download/filediver-linux-amd64.tar.gz)

Extract the achive into a new folder.

## Usage
While you can simply double-click the executable to run it, using the terminal will grant you some more options.

In your terminal, navigate to the folder containing the executable. From there, run `filediver -h` to get a list of options.

Here are some example commands:

`$` `filediver` - simply running the app should automatically detect your installation directory and dump all files into the "extracted" directory in your current folder.

`$` `filediver -h` - print a detailed description of all command line options.

`$` `filediver -o "custom_dir"` - extract the files into a directory called "custom_dir".

`$` `filediver -c "enable:video"` - extract only video files.

`$` `filediver -c "audio:format=ogg"` - extract audio as Ogg (more storage-efficient).

`$` `filediver -c "audio:format=mp3" -i "content/audio/291227525.wwise_stream"` - extract the Super Earth anthem as mp3.

## Features
### File Types/Formats
- **Audio**: Audiokinetic wwise bnk/wem; automatically converted to WAV
- **Video**: Bink; automatically converted to MP4 via FFmpeg (shipped with Windows binary)
- **Textures**: Direct Draw Surface (.dds)
- **Models (WIP)**: Stingray Unit; automatically converted to GLB (glTF)

Planned: bones, animations

## Credits/Links
This app builds on a lot of work from other people. This includes:
- [Hellextractor by Xaymar](https://github.com/Xaymar/Hellextractor)
	- Basic binary file structure
	- Unhashed resource names/types (.txt files)
- [vgmstream](https://github.com/vgmstream/vgmstream), [ww2ogg by hcs](https://github.com/hcs64/ww2ogg) and [bnkextr by eXpl0it3r](https://github.com/eXpl0it3r/bnkextr)
	- Wwise audio formats

Some useful discussion on the topic of HD2 resource extraction: https://reshax.com/topic/507-helldivers-2-model-extraction-help/

## License
Copyright (c) Darwin Schuppan

FileDiver is licensed under the 3-Clause BSD License (https://opensource.org/license/bsd-3-clause).
