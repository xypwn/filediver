package datalib

type DLInstanceHeader struct {
	_       DLHash // May be a DLHash or uint32
	Magic   [4]byte
	Version uint32
	Type    DLHash
	Size    uint32
	Is64Bit uint8
	_       [7]uint8
}

type DLInstance struct {
	DLInstanceHeader
	Data []byte
}

type DLSubdataHeader struct {
	Magic   [4]byte
	Version uint32
	Type    DLHash
	Size    uint32
	Is64Bit uint8
	_       [7]uint8
}

type DLSubdata struct {
	DLSubdataHeader
	Data []byte
}
