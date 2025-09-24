package stingray

import (
	"cmp"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Hash struct{ Value uint64 }

func murmur64aSum(b []byte) Hash {
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

// Murmur64a hash
func Sum[T ~[]byte | string](x T) Hash {
	return murmur64aSum([]byte(x))
}

// 64-bit hash to 32-bit hash
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

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h Hash) Cmp(other Hash) int {
	return cmp.Compare(h.Value, other.Value)
}

// ParseHash parses a big endian murmur64 hash.
// Ignores 0x prefix if present.
func ParseHash(s string) (Hash, error) {
	s = strings.TrimPrefix(s, "0x")
	x, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return Hash{}, fmt.Errorf("parsing hash: %w", err)
	}
	return Hash{Value: x}, nil
}

// ParseHash parses a big endian murmur32 hash.
// Ignores 0x prefix if present.
func ParseThinHash(s string) (ThinHash, error) {
	s = strings.TrimPrefix(s, "0x")
	x, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ThinHash{}, fmt.Errorf("parsing thin hash: %w", err)
	}
	return ThinHash{Value: uint32(x)}, nil
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

func (h ThinHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h ThinHash) Cmp(other ThinHash) int {
	return cmp.Compare(h.Value, other.Value)
}
