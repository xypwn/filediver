package hashes

import (
	_ "embed"
)

//go:embed hashes.txt
var Hashes string

//go:embed material_textures.txt
var MaterialTextures string

//go:embed thinhashes.txt
var ThinHashes string
