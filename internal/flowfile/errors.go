package flowfile

const (
	emptyFile = iota
	invalidCharacter
)

func newEmptyfileErr() *FlowfileError {
	return &FlowfileError{
		emptyFile,
	}
}

func newInvalidCharacterErr() *FlowfileError {
	return &FlowfileError{
		invalidCharacter,
	}
}

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
