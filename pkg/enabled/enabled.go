package enabled

var (
	debug bool
)

func OpenDebug(b bool) {
	debug = b
}

func Debug() bool {
	return debug
}
