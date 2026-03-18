package components

import "github.com/xypwn/filediver/stingray"

type HashLookup interface {
	LookupHash(stingray.Hash) string
	LookupString(uint32) string
	LookupThinHash(stingray.ThinHash) string
}

type BasicLookup struct {
	ThinHash func(stingray.ThinHash) string
	Hash     func(stingray.Hash) string
	Str      func(uint32) string
}

func (b *BasicLookup) LookupHash(hash stingray.Hash) string {
	return b.Hash(hash)
}
func (b *BasicLookup) LookupThinHash(hash stingray.ThinHash) string {
	return b.ThinHash(hash)
}
func (b *BasicLookup) LookupString(val uint32) string {
	return b.Str(val)
}
