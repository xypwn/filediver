package ah_bin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

var NotFound error = errors.New("did not find any ah_bin files")

type BuildInfo struct {
	Commit   string `json:"commit"`
	Hash     string `json:"hash"`
	Version  string `json:"version"`
	Combined string `json:"combined"`
	Year     uint32 `json:"year"`
	Month    uint32 `json:"month"`
	Day      uint32 `json:"day"`
	Hour     uint32 `json:"hour"`
	Minute   uint32 `json:"minute"`
	Second   uint32 `json:"second"`
	BuildId  uint32 `json:"build_id"`
}

func LoadBuildInfo(r io.Reader) (*BuildInfo, error) {
	commit, err := util.ReadCString(r)
	if err != nil {
		return nil, err
	}

	hash, err := util.ReadCStringWithSkip(r)
	if err != nil {
		return nil, err
	}

	version, err := util.ReadCStringWithSkip(r)
	if err != nil {
		return nil, err
	}

	combined, err := util.ReadCStringWithSkip(r)
	if err != nil {
		return nil, err
	}

	var data []byte = make([]byte, 1)
	for {
		read, err := r.Read(data)
		if read == 0 {
			return nil, fmt.Errorf("string read past the end of r")
		}

		if err != nil {
			return nil, err
		}

		// Break reading string on null terminator
		if data[0] != 0 {
			break
		}
	}

	data = append(data, 0, 0, 0)
	_, err = r.Read(data[1:])
	if err != nil {
		return nil, err
	}
	var year, month, day, hour, minute, second, buildId uint32
	if _, err := binary.Decode(data, binary.LittleEndian, &year); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &month); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &day); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hour); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &minute); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &second); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &buildId); err != nil {
		return nil, err
	}

	return &BuildInfo{
		Commit:   *commit,
		Hash:     *hash,
		Version:  *version,
		Combined: *combined,
		Year:     year,
		Month:    month,
		Day:      day,
		Hour:     hour,
		Minute:   minute,
		Second:   second,
		BuildId:  buildId,
	}, nil
}

func LoadFromDataDir(dataDir *stingray.DataDir) (*BuildInfo, error) {
	ahBinType := stingray.Sum("ah_bin")
	for fileId := range dataDir.Files {
		if fileId.Type != ahBinType {
			continue
		}
		ahBinData, err := dataDir.Read(fileId, stingray.DataMain)
		if err != nil {
			return nil, fmt.Errorf("failed to read %v.ah_bin", fileId.Name.String())
		}
		return LoadBuildInfo(bytes.NewReader(ahBinData))
	}
	return nil, NotFound
}
