package labels

type Labels map[string]string

func (l Labels) Get(key string) string {
	return l[key]
}

func (l Labels) Set(key, value string) {
	l[key] = value
}
