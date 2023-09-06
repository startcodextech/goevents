package ddd

// Metadata is a map of any values.
type Metadata map[string]any

// NewMetadata creates a new Metadata.
func (m Metadata) Set(key string, value any) {
	m[key] = value
}

// Get returns a value by key.
func (m Metadata) Get(key string) any {
	return m[key]
}

// Del deletes a value by key.
func (m Metadata) Del(key string) {
	delete(m, key)
}

// Keys returns all keys.
func (m Metadata) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func (m Metadata) configureEvent(e *event) {
	for key, value := range m {
		e.metadata[key] = value
	}
}

func (m Metadata) configureCommand(c *command) {
	for key, value := range m {
		c.metadata[key] = value
	}
}

func (m Metadata) configureReply(r *reply) {
	for key, value := range m {
		r.metadata[key] = value
	}
}
