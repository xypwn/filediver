// The following is manually converted from ww2ogg by hcs (https://github.com/hcs64/ww2ogg)
package vorbis

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/xypwn/filediver/bitio"
)

const errPfx = "codebook: "

//go:embed packed_codebooks_aoTuV_603.bin
var codebooksPackedAOTUV603 []byte

var codebooksAOTUV603 = newCodebookLibrary(codebooksPackedAOTUV603)

func bookMaptype1Quantvals(entries uint64, dimensions uint64) uint64 {
	bits := uint64(ilog(entries))
	vals := entries >> ((bits - 1) * (dimensions - 1) / dimensions)

	for {
		var acc uint64 = 1
		var acc1 uint64 = 1
		for i := uint64(0); i < dimensions; i++ {
			acc *= vals
			acc1 *= vals + 1
		}
		if acc <= entries && acc1 > entries {
			return vals
		} else {
			if acc > entries {
				vals--
			} else {
				vals++
			}
		}
	}
}

type codebookLibrary struct {
	Data    []byte
	Offsets []int
	Count   int
}

func newCodebookLibrary(data []byte) *codebookLibrary {
	le := binary.LittleEndian

	offsetOffset := int(le.Uint32(data[len(data)-4:]))

	res := &codebookLibrary{
		Count: (len(data) - offsetOffset) / 4,
		Data:  make([]byte, offsetOffset),
	}
	res.Offsets = make([]int, res.Count)

	copy(res.Data, data[:offsetOffset])

	for i := 0; i < res.Count; i++ {
		res.Offsets[i] = int(le.Uint32(data[offsetOffset+4*i:]))
	}

	return res
}

func (cl *codebookLibrary) Codebook(id int) ([]byte, error) {
	if id < 0 || id+1 >= len(cl.Offsets) {
		return nil, fmt.Errorf("%vID out of bounds", errPfx)
	}
	return cl.Data[cl.Offsets[id]:cl.Offsets[id+1]], nil
}

func (cl *codebookLibrary) RebuildByID(w *bitio.Writer, id int) error {
	cb, err := cl.Codebook(id)
	if err != nil {
		return err
	}

	r := bitio.NewReader(bytes.NewReader(cb))

	dimensions, _, err := r.ReadBits(4)
	if err != nil {
		return err
	}
	entries, _, err := r.ReadBits(14)
	if err != nil {
		return err
	}

	if _, err := w.WriteBitsMany(
		0x564342, 24, // identifier
		dimensions, 16,
		entries, 24,
	); err != nil {
		return err
	}

	ordered, err := r.ReadBit()
	if err != nil {
		return err
	}

	if err := w.WriteBit(ordered); err != nil {
		return err
	}

	if ordered {
		initialLength, _, err := r.ReadBits(5)
		if err != nil {
			return err
		}

		if _, err := w.WriteBits(initialLength, 5); err != nil {
			return err
		}

		currentEntry := uint64(0)
		for currentEntry < entries {
			bw := ilog(entries - currentEntry)
			number, _, err := r.ReadBits(bw)
			if err != nil {
				return err
			}
			if _, err := w.WriteBits(number, bw); err != nil {
				return err
			}
			currentEntry += number
		}
		if currentEntry > entries {
			return errors.New("codebook: currentEntry out of range")
		}
	} else {
		codewordLenLen, _, err := r.ReadBits(3)
		if err != nil {
			return err
		}
		sparse, err := r.ReadBit()
		if err != nil {
			return err
		}

		if codewordLenLen == 0 || codewordLenLen > 5 {
			return errors.New("codebook: bad codeword length length")
		}

		if err := w.WriteBit(sparse); err != nil {
			return err
		}

		for i := uint64(0); i < entries; i++ {
			present := true

			if sparse {
				present, err = r.ReadBit()
				if err != nil {
					return err
				}

				if err := w.WriteBit(present); err != nil {
					return err
				}
			}

			if present {
				codewordLen, _, err := r.ReadBits(uint8(codewordLenLen))
				if err != nil {
					return err
				}

				if _, err := w.WriteBits(codewordLen, 5); err != nil {
					return err
				}
			}
		}
	}

	lookupType, _, err := r.ReadBits(1)
	if err != nil {
		return err
	}
	if _, err := w.WriteBits(lookupType, 4); err != nil {
		return err
	}

	if lookupType == 1 {
		minLen, _, err := r.ReadBits(32)
		if err != nil {
			return err
		}
		maxLen, _, err := r.ReadBits(32)
		if err != nil {
			return err
		}
		valLen, _, err := r.ReadBits(4)
		if err != nil {
			return err
		}
		sequenceFlag, err := r.ReadBit()
		if err != nil {
			return err
		}

		if _, err := w.WriteBits(minLen, 32); err != nil {
			return err
		}
		if _, err := w.WriteBits(maxLen, 32); err != nil {
			return err
		}
		if _, err := w.WriteBits(valLen, 4); err != nil {
			return err
		}
		if err := w.WriteBit(sequenceFlag); err != nil {
			return err
		}

		quantvals := bookMaptype1Quantvals(entries, dimensions)
		for i := uint64(0); i < quantvals; i++ {
			val, _, err := r.ReadBits(uint8(valLen + 1))
			if err != nil {
				return err
			}
			if _, err := w.WriteBits(val, uint8(valLen+1)); err != nil {
				return err
			}
		}
	}

	if len(cb) != 0 && len(cb) != r.BitsRead()/8+1 {
		return errors.New("codebook: size mismatch")
	}

	return nil
}
