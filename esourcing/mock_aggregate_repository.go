// Code generated by mockery v2.36.0. DO NOT EDIT.

package esourcing

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockAggregateRepository is an autogenerated mock type for the AggregateRepository type
type MockAggregateRepository[T EventSourcedAggregate] struct {
	mock.Mock
}

// Load provides a mock function with given fields: _a0, _a1
func (_m *MockAggregateRepository[T]) Load(_a0 context.Context, _a1 string) (T, error) {
	ret := _m.Called(_a0, _a1)

	var r0 T
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (T, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) T); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(T)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: _a0, _a1
func (_m *MockAggregateRepository[T]) Save(_a0 context.Context, _a1 T) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, T) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockAggregateRepository creates a new instance of MockAggregateRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAggregateRepository[T EventSourcedAggregate](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAggregateRepository[T] {
	mock := &MockAggregateRepository[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
