package manifest

import "context"

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
	Usage          Usage                                                               `json:"usage"`
}

type Usage struct {
	Args         []UsageDesc `json:"args"`
	ReturnValues []UsageDesc `json:"return_values"`
}

type UsageDesc struct {
	Name           string   `json:"name"`
	OptionalValues []string `json:"optional_values"`
	Desc           string   `json:"desc"`
}
