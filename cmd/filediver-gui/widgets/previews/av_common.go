package previews

import (
	"fmt"
	"strings"
)

func formatPlayerTimeF(timeSeconds float32, showMS bool) string {
	return formatPlayerTimeMS(int(timeSeconds*1000), showMS)
}

func formatPlayerTimeMS(timeMilliseconds int, showMS bool) string {
	t := timeMilliseconds
	var b strings.Builder
	if t < 0 {
		b.WriteString("-")
		t = -t
	}

	var hrs, mins, secs, msecs int
	msecs = t % 1000
	t /= 1000
	secs = t % 60
	t /= 60
	mins = t % 60
	t /= 60
	hrs = t

	if hrs > 0 {
		fmt.Fprintf(&b, "%d:", hrs)
	}
	fmt.Fprintf(&b, "%02d:%02d", mins, secs)
	if showMS {
		fmt.Fprintf(&b, ":%03d", msecs)
	}

	return b.String()
}
