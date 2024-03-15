// The following is manually converted from hcs's ww2ogg (https://github.com/hcs64/ww2ogg)
// and vgmstream (https://github.com/vgmstream/vgmstream)
package vorbis

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/bitio"
)

func ww2oggConvertPacket(d *Decoder, writer io.Writer, wp wPacket, reader io.Reader) error {
	bw := bitio.NewWriter(writer)
	br := bitio.NewReader(reader)
	copyBits := func(nBits uint8) (uint64, error) {
		return copyBits(bw, br, nBits)
	}

	packetType := uint64(0) // audio packet
	if _, err := bw.WriteBits(packetType, 1); err != nil {
		return err
	}

	modeNumber, err := copyBits(d.modeBits)
	if err != nil {
		return err
	}

	remainder, _, err := br.ReadBits(8 - d.modeBits)
	if err != nil {
		return err
	}

	if d.modeBlockFlag[modeNumber] {
		var nextBlockFlag bool
		if wp.HasNext {
			// Get next first byte to read nextModeNumber
			nextModeNumber := wp.NextB & ((1 << d.modeBits) - 1)
			nextBlockFlag = d.modeBlockFlag[nextModeNumber]
		} else {
			// EOF => probably doesn't matter
			nextBlockFlag = false
		}

		prevWindowType := d.prevBlockFlag
		if err := bw.WriteBit(prevWindowType); err != nil {
			return err
		}
		nextWindowType := nextBlockFlag
		if err := bw.WriteBit(nextWindowType); err != nil {
			return err
		}
	}

	d.prevBlockFlag = d.modeBlockFlag[modeNumber]

	if _, err := bw.WriteBits(remainder, 8-d.modeBits); err != nil {
		return err
	}

	// Copy rest of packet bit by bit (since not byte aligned)
	bitsToCopy := (wp.PacketSize - 1) * 8
	for i := 0; i < int(bitsToCopy); i++ {
		b, err := br.ReadBit()
		if err != nil {
			return err
		}
		if err := bw.WriteBit(b); err != nil {
			return err
		}
	}

	if err := bw.FlushByte(); err != nil {
		return err
	}

	return nil
}

func ww2oggConvertSetup(d *Decoder, writer io.Writer, reader io.Reader) error {
	bw := bitio.NewWriter(writer)
	br := bitio.NewReader(reader)
	copyBits := func(nBits uint8) (uint64, error) {
		return copyBits(bw, br, nBits)
	}

	pktHdr := struct {
		Type uint8
		ID   [6]byte
	}{
		Type: 0x05,                                  // packet type: setup
		ID:   [6]byte{'v', 'o', 'r', 'b', 'i', 's'}, // id: always "vorbis"
	}
	if err := binary.Write(bw, binary.LittleEndian, pktHdr); err != nil {
		return err
	}

	codebookCountLess1, err := copyBits(8)
	if err != nil {
		return err
	}
	codebookCount := codebookCountLess1 + 1

	// Rebuild external wwise codebooks
	for i := uint64(0); i < codebookCount; i++ {
		codebookID, _, err := br.ReadBits(10)
		if err != nil {
			return err
		}

		if err := codebooksAOTUV603.RebuildByID(bw, int(codebookID)); err != nil {
			return err
		}
	}

	// Time domain transforms
	timeCountLess1 := uint64(0)
	if _, err := bw.WriteBits(timeCountLess1, 6); err != nil {
		return err
	}
	dummyTimeValue := uint64(0)
	if _, err := bw.WriteBits(dummyTimeValue, 16); err != nil {
		return err
	}

	// Floors
	floorCountLess1, err := copyBits(6)
	if err != nil {
		return err
	}
	floorCount := floorCountLess1 + 1
	for i := uint64(0); i < floorCount; i++ {
		floorType := uint64(1) // floor type always 1
		if _, err := bw.WriteBits(floorType, 16); err != nil {
			return err
		}

		floor1Partitions, err := copyBits(5)
		if err != nil {
			return err
		}

		var maximumClass uint64
		var floor1PartitionClassList [32]uint64 // max 5 bits
		for j := uint64(0); j < floor1Partitions; j++ {
			floor1PartitionClass, err := copyBits(4)
			if err != nil {
				return err
			}

			floor1PartitionClassList[j] = floor1PartitionClass

			if floor1PartitionClass > maximumClass {
				maximumClass = floor1PartitionClass
			}
		}

		var floor1ClassDimensionsList [16 + 1]uint64 // max 4 bits + 1
		for j := uint64(0); j <= maximumClass; j++ {
			classDimensionsLess1, err := copyBits(3)
			if err != nil {
				return err
			}

			floor1ClassDimensionsList[j] = classDimensionsLess1 + 1

			classSubclasses, err := copyBits(2)
			if err != nil {
				return err
			}

			if classSubclasses != 0 {
				masterbook, err := copyBits(8)
				if err != nil {
					return err
				}

				if masterbook > codebookCount {
					return errors.New("invalid floor1 masterbook")
				}
			}

			for k := uint64(0); k < uint64(1)<<classSubclasses; k++ {
				subclassBookPlus1, err := copyBits(8)
				if err != nil {
					return err
				}

				subclassBook := int64(subclassBookPlus1) - 1
				if subclassBook >= 0 && subclassBook >= int64(codebookCount) {
					return errors.New("invalid floor1 subclass book")
				}
			}
		}

		floor1MultiplierLess1, err := copyBits(2)
		if err != nil {
			return err
		}
		_ = floor1MultiplierLess1

		rangebits, err := copyBits(4)
		if err != nil {
			return err
		}

		for j := uint64(0); j < floor1Partitions; j++ {
			currentClassNumber := floor1PartitionClassList[j]
			for k := uint64(0); k < floor1ClassDimensionsList[currentClassNumber]; k++ {
				_, err := copyBits(uint8(rangebits))
				if err != nil {
					return err
				}
			}
		}
	}

	// Residues
	residueCountLess1, err := copyBits(6)
	if err != nil {
		return err
	}
	residueCount := residueCountLess1 + 1
	for i := uint64(0); i < residueCount; i++ {
		residueType, _, err := br.ReadBits(2)
		if err != nil {
			return err
		}
		if _, err := bw.WriteBits(residueType, 16); err != nil {
			return err
		}

		if residueType > 2 {
			return errors.New("invalid residue type")
		}

		residueBegin, err := copyBits(24)
		if err != nil {
			return err
		}
		_ = residueBegin
		residueEnd, err := copyBits(24)
		if err != nil {
			return err
		}
		_ = residueEnd
		residuePartitionSizeLess1, err := copyBits(24)
		if err != nil {
			return err
		}
		_ = residuePartitionSizeLess1
		residueClassificationsLess1, err := copyBits(6)
		if err != nil {
			return err
		}
		residueClassifications := residueClassificationsLess1 + 1
		residueClassbook, err := copyBits(8)
		if err != nil {
			return err
		}

		if residueClassbook >= codebookCount {
			return errors.New("invalid residue classbook")
		}

		var residueCascade [64 + 1]uint64 // max 6 bytes + 1
		for j := uint64(0); j < residueClassifications; j++ {
			loblPut, err := copyBits(3)
			if err != nil {
				return err
			}
			bitflag, err := copyBits(1)
			if err != nil {
				return err
			}
			highBits := uint64(0)
			if bitflag != 0 {
				highBits, err = copyBits(5)
				if err != nil {
					return err
				}
			}
			residueCascade[j] = highBits*8 + loblPut
		}

		for j := uint64(0); j < residueClassifications; j++ {
			for k := uint64(0); k < 8; k++ {
				if residueCascade[j]&(1<<k) != 0 {
					residueBook, err := copyBits(8)
					if err != nil {
						return err
					}
					if residueBook >= codebookCount {
						return errors.New("invalid residue book")
					}
				}
			}
		}
	}

	// Mappings
	mappingCountLess1, err := copyBits(6)
	if err != nil {
		return err
	}
	mappingCount := mappingCountLess1 + 1
	for i := uint64(0); i < mappingCount; i++ {
		mappingType := uint64(0) // always 0, only mapping type
		if _, err := bw.WriteBits(mappingType, 16); err != nil {
			return err
		}

		submapsFlag, err := copyBits(1)
		if err != nil {
			return err
		}
		submaps := uint64(1)
		if submapsFlag != 0 {
			submapsLess1, err := copyBits(4)
			if err != nil {
				return err
			}
			submaps = submapsLess1 + 1
		}

		squarePolarFlag, err := copyBits(1)
		if err != nil {
			return err
		}
		if squarePolarFlag != 0 {
			couplingStepsLess1, err := copyBits(8)
			if err != nil {
				return err
			}
			couplingSteps := couplingStepsLess1 + 1

			for j := uint64(0); j < couplingSteps; j++ {
				magnitudeBits := ilog(uint64(d.cfg.Channels) - 1)
				angleBits := ilog(uint64(d.cfg.Channels) - 1)

				magnitude, err := copyBits(magnitudeBits)
				if err != nil {
					return err
				}
				angle, err := copyBits(angleBits)
				if err != nil {
					return err
				}

				if angle == magnitude || magnitude >= uint64(d.cfg.Channels) || angle >= uint64(d.cfg.Channels) {
					return fmt.Errorf("invalid coupling: angle=%v, mag=%v, ch=%v", angle, magnitude, d.cfg.Channels)
				}
			}
		}

		mappingReserved, err := copyBits(2)
		if mappingReserved != 0 {
			return errors.New("mapping reserved field nonzero")
		}

		if submaps > 1 {
			for j := 0; j < int(d.cfg.Channels); j++ {
				mappingMux, err := copyBits(4)
				if err != nil {
					return err
				}
				if mappingMux >= submaps {
					return errors.New("mappingMux >= submaps")
				}
			}
		}

		for j := uint64(0); j < submaps; j++ {
			// Unused time domain transform config packet
			timeConfig, err := copyBits(8)
			if err != nil {
				return err
			}
			_ = timeConfig

			floorNumber, err := copyBits(8)
			if err != nil {
				return err
			}
			if floorNumber >= floorCount {
				return errors.New("invalid floor mapping")
			}

			residueNumber, err := copyBits(8)
			if err != nil {
				return err
			}
			if residueNumber >= residueCount {
				return errors.New("invalid residue mapping")
			}
		}
	}

	// Modes
	modeCountLess1, err := copyBits(6)
	if err != nil {
		return err
	}
	modeCount := modeCountLess1 + 1
	d.modeBits = ilog(modeCountLess1)
	for i := uint64(0); i < modeCount; i++ {
		blockFlag, err := copyBits(1)
		if err != nil {
			return err
		}
		d.modeBlockFlag[i] = blockFlag != 0

		windowType := uint64(0)
		if _, err := bw.WriteBits(windowType, 16); err != nil {
			return err
		}
		transformType := uint64(0)
		if _, err := bw.WriteBits(transformType, 16); err != nil {
			return err
		}

		mapping, err := copyBits(8)
		if err != nil {
			return err
		}
		if mapping >= mappingCount {
			return errors.New("invalid mode mapping")
		}
	}

	// End flag
	{
		framing := uint64(1)
		if _, err := bw.WriteBits(framing, 1); err != nil {
			return err
		}
	}

	// Flush / align to bytes
	if err := bw.FlushByte(); err != nil {
		return err
	}

	return nil
}
