// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	getpocket "github.com/whitekid/getpocket"
)

// Interface is an autogenerated mock type for the Interface type
type Interface struct {
	mock.Mock
}

// Add provides a mock function with given fields: url
func (_m *Interface) Add(url string) getpocket.AddRequester {
	ret := _m.Called(url)

	var r0 getpocket.AddRequester
	if rf, ok := ret.Get(0).(func(string) getpocket.AddRequester); ok {
		r0 = rf(url)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(getpocket.AddRequester)
		}
	}

	return r0
}

// Get provides a mock function with given fields:
func (_m *Interface) Get() *getpocket.GetRequest {
	ret := _m.Called()

	var r0 *getpocket.GetRequest
	if rf, ok := ret.Get(0).(func() *getpocket.GetRequest); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*getpocket.GetRequest)
		}
	}

	return r0
}

// Modify provides a mock function with given fields:
func (_m *Interface) Modify() getpocket.ModifyRequester {
	ret := _m.Called()

	var r0 getpocket.ModifyRequester
	if rf, ok := ret.Get(0).(func() getpocket.ModifyRequester); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(getpocket.ModifyRequester)
		}
	}

	return r0
}

type mockConstructorTestingTNewInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewInterface creates a new instance of Interface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewInterface(t mockConstructorTestingTNewInterface) *Interface {
	mock := &Interface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
