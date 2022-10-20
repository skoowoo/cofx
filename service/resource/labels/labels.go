package labels

type Labels map[string]string

// Get returns the value of the label.
func (l Labels) Get(key string) string {
	return l[key]
}

// Set sets the value of the label.
func (l Labels) Set(key, value string) {
	l[key] = value
}

// GetFlowID returns the value of the label 'flow_id'.
func (l Labels) GetFlowID() string {
	return l.Get("flow_id")
}

// GetNodeSeq returns the value of the label 'node_seq'.
func (l Labels) GetNodeSeq() string {
	return l.Get("node_seq")
}

// GetNodeName returns the value of the label 'node_name'.
func (l Labels) GetNodeName() string {
	return l.Get("node_name")
}
