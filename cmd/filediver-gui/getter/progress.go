package getter

type State int

const (
	Unknown State = iota
	Fetching
	Downloading
	Extracting
	Done
)

type Progress struct {
	State               State
	ContentCurrentBytes int
	ContentTotalBytes   int
}
