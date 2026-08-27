<div align="center">
# FileDiver Reverse Engineering Tools
</div>

## Setup
- Download and install [Go](https://go.dev/dl/)
- Download this repository as [Zip](https://github.com/xypwn/filediver/archive/refs/heads/master.zip) (or use git clone)
- In the command line, navigate into the repository folder

# Tools

### Hash Tool
**See [Hash Cracker](#hash-cracker) for advanced hash cracking.**

Calculate and crack murmur64a hashes.

- `go run ./cmd/tools/hash-tool` for a list of options

### Hash Cracker
Fast GPU-based hash cracking.

- `cd ./cmd/tools/hash-cracker`
- `go run . -h` for a list of options
- pattern usage: `go run . pattern "pattern_expression"`

#### Pattern language
- strings without any special characters are parsed as strings
- `<`, `>`: group inner contents together (works like parentheses)
- `<a><b>`: concatenate a and b
- `a|b|c`: a or b or c
- `[a-z0-9~_-]`, `[^a-z]`: character classes mostly like in regex
- `...{m,n}`, `...{m}`: repeat expression m-n times (`,n` can be omitted if `m == n`)
- `...{m,n,s}`, `...{m,s}`: repeat expression m-n times, separated by s (`,n` can be omitted if `m == n` and if `s` doesn't start with a digit)
- `#{var = ...}`: assign value of to variable `var`
- `#{var}`: expand value of variable `var`
- `#{func arg1 arg2 ...}`: call function with arguments `arg1`, `arg2` etc.
##### Builtin variable and functions
- `#{load filename}`: loads pattern from file `filename` and returns it
- `#{import filename}`: import variables from file `filename` (never returns anything)
- `#{known}`: list of all known file and type names
- `#{known-words}`: list of words in all known filenames (separated by `[/_:]`)
##### Examples
- `content/#{known-words}{1,3,[_:]}`: `content/` followed by 1-3 known words, separated by `_` or `:`

### Crossref-checker
Check if selected game files reference any other game files by hash.

- `go run ./cmd/tools/crossref-checker` for a list of options

### Entity Modder
Create patches using modified entity json exports.

1. Export entities to json
2. Find the entity to modify
3. Edit the json file and change some values
4. Run the entity-json-parser tool to write a patch for any number of entity files

#### Entity JSON Parser Options
```
go run ./cmd/tools/entity-json-parser/ --entity-names ... --entities ... [--output <archive-hash>.patch_0] [--levels ...]
```
```
--entity-names 0x1234 0x5678 ...
```
This is a list of entity hashes to be overwritten in the patch

```
--entities ./path/to/modified1.entity.json ./path/to/modified2.entity.json ...
```
This is a list of json files to be parsed and saved as the given `entity-names` in the generated patch

The `entity-names` and `entities` lists must either be the same length, or the `entities` list of files must only include one file. If only a single file is included, every entity in the `entity-names` list will be overwritten with the same file.

```
--output 9ba626afa44a3aa3.patch_0
```
This is the location of the output patch file, which defaults to the boot archive

```
--levels ./path/to/modified1.level.json ./path/to/modified2.level.json
```
This is a list of level json files (which must have embedded entity files) that will have their embedded entity file updated in the patch. This is an advanced, optional feature.
