package stingray

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
)

type Hash struct{ Value uint64 }

func HashFromString(s string) (Hash, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Hash{}, err
	}
	return Hash{Value: binary.BigEndian.Uint64(b)}, nil
}

// Murmur64a hash
func Sum64(b []byte) Hash {
	var seed uint64 = 0
	var mix uint64 = 0xC6A4A7935BD1E995
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

	for i := len(b) - 1; i >= 0; i-- {
		hash ^= uint64(b[i]) << uint64(8*i)
	}

	hash *= mix
	hash ^= hash >> shifts

	hash *= mix
	hash ^= hash >> shifts

	return Hash{Value: hash}
}

// Hellextractor uses Big Endian hashes, but really,
// that's not correct, since the Stingray engine
// basically only runs on Little Endian devices
/*func (h Hash) ToBigEndian() uint64 {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], h.Value)
	return binary.LittleEndian.Uint64(b[:])
}*/

func (h Hash) String() string {
	s := strconv.FormatUint(h.Value /*h.ToBigEndian()*/, 16)
	return strings.Repeat("0", 16-len(s)) + s
}

type ThinHash struct{ Value uint32 }
