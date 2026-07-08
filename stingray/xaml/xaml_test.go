package xaml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/stingray/xaml"
)

func TestPathData(t *testing.T) {
	require := require.New(t)

	testStrs := []string{
		"M4.12,0 L9.67,5.47 4.12,10.94 0,10.88 5.56,5.47 0,0.06",
		"F1 M0,0 L5,5 10,0 z",
		"M10,80 Q95,10,180,80",
	}

	for _, testStr := range testStrs {
		var p xaml.PathData
		err := p.UnmarshalText([]byte(testStr))
		require.NoError(err)
		b, err := p.MarshalText()
		require.NoError(err)
		require.Equal(testStr, string(b))
	}
}

func TestParseColor(t *testing.T) {
	require := require.New(t)

	testCases := []struct {
		S string
		C xaml.Color
	}{
		{"#000", xaml.Color{0, 0, 0, 0xff}},
		{"#111", xaml.Color{0x11, 0x11, 0x11, 0xff}},
		{"#123", xaml.Color{0x11, 0x22, 0x33, 0xff}},
		{"#a123", xaml.Color{0x11, 0x22, 0x33, 0xaa}},
		{"#ab123456", xaml.Color{0x12, 0x34, 0x56, 0xab}},
		{"DarkCyan", xaml.Color{0x00, 0x8b, 0x8b, 0xff}},
	}

	for _, tc := range testCases {
		var c xaml.Color
		err := c.UnmarshalText([]byte(tc.S))
		require.NoError(err)
		require.Equal(tc.C, c)
	}
}
