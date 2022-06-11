package manifest

// todo
type Manifester interface {
	Name() string
	Manifest() Manifest
}

type Manifest struct {
	Description    string                                             `json:"description"`
	Driver         string                                             `json:"driver"`
	EntryPoint     string                                             `json:"entrypoint"`
	EntryPointFunc func(map[string]string) (map[string]string, error) `json:"-"`
	Args           map[string]string                                  `json:"args"`
	RetryOnFailure int                                                `json:"retry_on_failure"`
}
