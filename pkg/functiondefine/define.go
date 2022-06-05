package functiondefine

// todo
type Define interface {
	Name() string
	Manifest() *Manifest
}

type Manifest struct {
}
