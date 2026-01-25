package util

import (
	"fmt"
	"io"
)

func ReadCString(r io.Reader) (*string, error) {
	var data []byte = make([]byte, 1)
	var toReturn string
	for {
		read, err := r.Read(data)
		if read == 0 {
			return nil, fmt.Errorf("string read past the end of r")
		}

		if err != nil {
			return nil, err
		}

		// Break reading string on null terminator
		if data[0] == 0 {
			break
		}

		toReturn = toReturn + string(data)
	}
	return &toReturn, nil
}

func ReadCStringWithSkip(r io.Reader) (*string, error) {
	var data []byte = make([]byte, 1)
	var toReturn string
	for {
		read, err := r.Read(data)
		if read == 0 {
			return nil, fmt.Errorf("string read past the end of r")
		}

		if err != nil {
			return nil, err
		}

		// Break reading string on null terminator
		if data[0] == 0 {
			if len(toReturn) == 0 {
				continue
			}
			break
		}

		toReturn = toReturn + string(data)
	}
	return &toReturn, nil
}
