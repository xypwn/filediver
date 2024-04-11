package stingray

import (
	"encoding/binary"
	"encoding/hex"
)

type Hash struct{ Value uint64 }

// Murmur64a hash
func Sum64(b []byte) Hash {
	var seed uint64 = 0
	var mix uint64 = 0xc6a4a7935bd1e995
	const shifts = 47

	var hash uint64 = seed ^ (uint64(len(b)) * mix)

	for len(b) >= 8 {
		key := binary.LittleEndian.Uint64(b)
		b = b[8:]

		key *= mix
		key ^= key >> shifts
		key *= mix

		hash ^= key
		hash *= mix
	}

	if len(b) > 0 {
		for i := len(b) - 1; i >= 0; i-- {
			hash ^= uint64(b[i]) << uint64(8*i)
		}
		hash *= mix
	}

	hash ^= hash >> shifts

	hash *= mix
	hash ^= hash >> shifts

	return Hash{Value: hash}
}

func (h Hash) Thin() ThinHash {
	return ThinHash{Value: uint32(h.Value >> 32)}
}

func (h Hash) StringEndian(endian binary.ByteOrder) string {
	var b [8]byte
	endian.PutUint64(b[:], h.Value)
	return hex.EncodeToString(b[:])
}

func (h Hash) String() string {
	return "0x" + h.StringEndian(binary.BigEndian)
}

type ThinHash struct{ Value uint32 }

func (h ThinHash) StringEndian(endian binary.ByteOrder) string {
	var b [4]byte
	endian.PutUint32(b[:], h.Value)
	return hex.EncodeToString(b[:])
}

func (h ThinHash) String() string {
	return "0x" + h.StringEndian(binary.BigEndian)
}
