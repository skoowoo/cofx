package manifest

// todo
type Manifester interface {
	Name() string
	Manifest() Manifest
}

type Manifest struct {
	Description    string                 `json:"description"`
	Driver         string                 `json:"driver"`
	EntryPoint     string                 `json:"entrypoint"`
	Args           map[string]interface{} `json:"args"`
	RetryOnFailure int                    `json:"retry_on_failure"`
}
