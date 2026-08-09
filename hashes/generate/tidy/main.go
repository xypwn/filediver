package main

import (
	"github.com/xypwn/filediver/hashes"
)

func main() {
	if _, err := hashes.TidyHashesFile("cracked.txt"); err != nil {
		panic(err)
	}
}
