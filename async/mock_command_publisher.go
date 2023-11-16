// Code generated by mockery v2.36.0. DO NOT EDIT.

package async

import (
	context "context"

	ddd "github.com/startcodextech/goevents/ddd"
	mock "github.com/stretchr/testify/mock"
)

// MockCommandPublisher is an autogenerated mock type for the CommandPublisher type
type MockCommandPublisher struct {
	mock.Mock
}

// Publish provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockCommandPublisher) Publish(_a0 context.Context, _a1 string, _a2 ddd.Command) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ddd.Command) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockCommandPublisher creates a new instance of MockCommandPublisher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCommandPublisher(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCommandPublisher {
	mock := &MockCommandPublisher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}