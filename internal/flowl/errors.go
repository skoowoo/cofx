package flowl

const (
	emptyFile = iota
	invalidCharacter
)

type FlowfileError struct {
	errcode int
}

func (e *FlowfileError) Error() string {
	switch e.errcode {
	case emptyFile:
		return "File is empty"
	}
	return "Unknow error"
}
