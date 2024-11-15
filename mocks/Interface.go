// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	getpocket "github.com/whitekid/getpocket"
)

// Interface is an autogenerated mock type for the Interface type
type Interface struct {
	mock.Mock
}

// Articles provides a mock function with given fields:
func (_m *Interface) Articles() getpocket.ArticleAPI {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Articles")
	}

	var r0 getpocket.ArticleAPI
	if rf, ok := ret.Get(0).(func() getpocket.ArticleAPI); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(getpocket.ArticleAPI)
		}
	}

	return r0
}

// AuthorizedURL provides a mock function with given fields: ctx, redirectURI
func (_m *Interface) AuthorizedURL(ctx context.Context, redirectURI string) (string, string, error) {
	ret := _m.Called(ctx, redirectURI)

	if len(ret) == 0 {
		panic("no return value specified for AuthorizedURL")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, string, error)); ok {
		return rf(ctx, redirectURI)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, redirectURI)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) string); ok {
		r1 = rf(ctx, redirectURI)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, redirectURI)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Modify provides a mock function with given fields:
func (_m *Interface) Modify() getpocket.ModifyAPI {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Modify")
	}

	var r0 getpocket.ModifyAPI
	if rf, ok := ret.Get(0).(func() getpocket.ModifyAPI); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(getpocket.ModifyAPI)
		}
	}

	return r0
}

// NewAccessToken provides a mock function with given fields: ctx, requestToken
func (_m *Interface) NewAccessToken(ctx context.Context, requestToken string) (string, string, error) {
	ret := _m.Called(ctx, requestToken)

	if len(ret) == 0 {
		panic("no return value specified for NewAccessToken")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, string, error)); ok {
		return rf(ctx, requestToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, requestToken)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) string); ok {
		r1 = rf(ctx, requestToken)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, requestToken)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewInterface creates a new instance of Interface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *Interface {
	mock := &Interface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
