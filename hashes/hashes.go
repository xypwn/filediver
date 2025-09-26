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

//go:embed dl_type_names.txt
var DLTypeNames string
