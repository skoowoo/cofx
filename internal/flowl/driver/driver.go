package driver

type FunctionDriver interface {
	Name() string
	Load() error
	Run() error
}
