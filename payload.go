package goevents

import "fmt"

func MapType[T interface{}](payload interface{}) (*T, error) {
	if payload == nil {
		return nil, ErrPayloadNil
	}

	v, ok := payload.(T)
	if !ok {
		return nil, fmt.Errorf(ErrPayloadTypeAssertion.Error(), v)
	}

	return &v, nil
}
