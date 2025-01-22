<div align="center">
# FileDiver Reverse Engineering Tools
</div>

## Setup
- Download and install [Go](https://go.dev/dl/)
- Download this repository as [Zip](https://github.com/xypwn/filediver/archive/refs/heads/master.zip) (or use git clone)
- In the command line, navigate into the repository folder

# Tools

### Hash Tool
Calculate and crack murmur64a hashes.

- `go run ./cmd/tools/hash_tool` for a list of options

### Crossref-checker
Check if selected game files reference any other game files by hash.

- `go run ./cmd/tools/crossref-checker` for a list of options