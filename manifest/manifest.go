package manifest

type Manifest struct {
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Driver         string            `json:"driver"`
	Entrypoint     string            `json:"entrypoint"`
	Args           map[string]string `json:"args"`
	RetryOnFailure int               `json:"retry_on_failure"`
	IgnoreFailure  bool              `json:"ignore_failure"`
	Usage          Usage             `json:"usage"`
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
