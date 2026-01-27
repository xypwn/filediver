package main

import (
	"bytes"
	"io"
	"os"

	"github.com/klauspost/compress/gzip"
)

func compressFile(filename string, buf *bytes.Buffer) {
	buf.Reset()
	w, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filename+".gz", buf.Bytes(), 0666); err != nil {
		panic(err)
	}
}

func main() {
	filenames := []string{
		"dl_library.dl_typelib",
		"generated_arc_settings.dl_bin",
		"generated_customization_armor_sets.dl_bin",
		"generated_customization_passive_bonuses.dl_bin",
		"generated_damage_settings.dl_bin",
		"generated_unit_customization_settings.dl_bin",
		"generated_weapon_customization_settings.dl_bin",
		"generated_entities.dl_bin",
		"generated_entity_deltas.dl_bin",
	}
	var buf bytes.Buffer
	for _, filename := range filenames {
		compressFile(filename, &buf)
	}
}
