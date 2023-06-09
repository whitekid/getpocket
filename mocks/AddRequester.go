// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	getpocket "github.com/whitekid/getpocket"
)

// AddRequester is an autogenerated mock type for the AddRequester type
type AddRequester struct {
	mock.Mock
}

// Do provides a mock function with given fields: ctx
func (_m *AddRequester) Do(ctx context.Context) (*getpocket.AddResponse, error) {
	ret := _m.Called(ctx)

	var r0 *getpocket.AddResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*getpocket.AddResponse, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *getpocket.AddResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*getpocket.AddResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAddRequester interface {
	mock.TestingT
	Cleanup(func())
}

// NewAddRequester creates a new instance of AddRequester. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAddRequester(t mockConstructorTestingTNewAddRequester) *AddRequester {
	mock := &AddRequester{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
