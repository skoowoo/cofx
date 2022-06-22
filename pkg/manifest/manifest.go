package manifest

import "context"

// todo
type Manifester interface {
	Name() string
	Manifest() Manifest
}

type Manifest struct {
	Description    string                                                              `json:"description"`
	Driver         string                                                              `json:"driver"`
	EntryPoint     string                                                              `json:"entrypoint"`
	EntrypointFunc func(context.Context, map[string]string) (map[string]string, error) `json:"-"`
	Args           map[string]string                                                   `json:"args"`
	RetryOnFailure int                                                                 `json:"retry_on_failure"`
}
