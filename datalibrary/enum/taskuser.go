package enum

type TaskUser uint32

const (
	TaskUser_None        TaskUser = 0x0
	TaskUser_Value_1     TaskUser = 0x1
	TaskUser_Value_2     TaskUser = 0x2
	TaskUser_Value_4     TaskUser = 0x4
	TaskUser_Value_8     TaskUser = 0x8
	TaskUser_Value_10    TaskUser = 0x10
	TaskUser_Value_20    TaskUser = 0x20
	TaskUser_Value_40    TaskUser = 0x40
	TaskUser_Value_80    TaskUser = 0x80
	TaskUser_Value_100   TaskUser = 0x100
	TaskUser_Value_200   TaskUser = 0x200
	TaskUser_Value_800   TaskUser = 0x800
	TaskUser_Value_1000  TaskUser = 0x1000
	TaskUser_Value_2000  TaskUser = 0x2000
	TaskUser_Value_4000  TaskUser = 0x4000
	TaskUser_Value_8000  TaskUser = 0x8000
	TaskUser_Value_10000 TaskUser = 0x10000
	TaskUser_Value_20000 TaskUser = 0x20000
	TaskUser_Value_40000 TaskUser = 0x40000
)

func (p TaskUser) MarshalText() ([]byte, error) {
	if p == TaskUser_None {
		return []byte(p.String()), nil
	}
	toReturn := ""
	i := TaskUser_Value_1
	for i <= TaskUser_Value_40000 {
		if i&p != TaskUser_None {
			if len(toReturn) > 0 {
				toReturn += "|"
			}
			toReturn += i.String()
		}
		i <<= 1
	}
	return []byte(toReturn), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=TaskUser
