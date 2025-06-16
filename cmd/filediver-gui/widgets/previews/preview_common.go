package previews

import "github.com/xypwn/filediver/stingray"

type GetResourceFunc func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error)
