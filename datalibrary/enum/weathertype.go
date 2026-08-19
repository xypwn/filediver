package enum

type WeatherType uint32

const (
	WeatherType_NoWeather WeatherType = iota
	WeatherType_Clear
	WeatherType_Cloudy
	WeatherType_Value_3_Len_17
	WeatherType_Raining
	WeatherType_Value_5_Len_17
	WeatherType_Value_6_Len_21
	WeatherType_Value_7_Len_21
	WeatherType_Value_8_Len_21
	WeatherType_Value_9_Len_21
)

func (p WeatherType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeatherType
