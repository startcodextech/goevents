package registry

import "sync"

type (
	Registrable interface {
		Key() string
	}

	Serializer   func(interface{}) ([]byte, error)
	Deserializer func([]byte, interface{}) error

	Registry interface {
		Serialize(string, interface{}) ([]byte, error)
		MustSerialize(string, interface{}) []byte
		Build(string, ...BuildOption) (interface{}, error)
		MustBuild(string, ...BuildOption) interface{}
		Deserialize(string, []byte, ...BuildOption) (interface{}, error)
		MustDeserialize(string, []byte, ...BuildOption) interface{}
		register(string, func() interface{}, Serializer, Deserializer, []BuildOption) error
	}

	registered struct {
		factory      func() interface{}
		serializer   Serializer
		deserializer Deserializer
		options      []BuildOption
	}

	registry struct {
		registered map[string]registered
		mu         sync.RWMutex
	}
)

var _ Registry = (*registry)(nil)

func New() *registry {
	return &registry{
		registered: make(map[string]registered),
	}
}

func (r *registry) Serialize(key string, v interface{}) ([]byte, error) {
	reg, exists := r.registered[key]
	if !exists {
		return nil, UnregisteredKey(key)
	}
	return reg.serializer(v)
}

func (r *registry) MustSerialize(key string, v interface{}) []byte {
	data, err := r.Serialize(key, v)
	if err != nil {
		panic(err)
	}
	return data
}

func (r *registry) Deserialize(key string, data []byte, options ...BuildOption) (interface{}, error) {
	v, err := r.Build(key, options...)
	if err != nil {
		return nil, err
	}

	err = r.registered[key].deserializer(data, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (r *registry) MustDeserialize(key string, data []byte, options ...BuildOption) interface{} {
	v, err := r.Deserialize(key, data, options...)
	if err != nil {
		panic(err)
	}
	return v
}

func (r *registry) Build(key string, options ...BuildOption) (interface{}, error) {
	reg, exists := r.registered[key]
	if !exists {
		return nil, UnregisteredKey(key)
	}

	v := reg.factory()
	uos := append(r.registered[key].options, options...)

	for _, option := range uos {
		err := option(v)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (r *registry) MustBuild(key string, options ...BuildOption) interface{} {
	v, err := r.Build(key, options...)
	if err != nil {
		panic(err)
	}
	return v
}

func (r *registry) register(key string, fn func() interface{}, s Serializer, d Deserializer, o []BuildOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.registered[key]; exists {
		return AlreadyRegisteredKey(key)
	}

	r.registered[key] = registered{
		factory:      fn,
		serializer:   s,
		deserializer: d,
		options:      o,
	}

	return nil
}
