// Code generated by MockGen. DO NOT EDIT.
// Source: F:\GeekGoProjects\src\webookpro\internal\repository\cache\code.go

// Package cachemocks is a generated GoMock package.
package cachemocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockCodeCache is a mock of CodeCache interface.
type MockCodeCache struct {
	ctrl     *gomock.Controller
	recorder *MockCodeCacheMockRecorder
}

// MockCodeCacheMockRecorder is the mock recorder for MockCodeCache.
type MockCodeCacheMockRecorder struct {
	mock *MockCodeCache
}

// NewMockCodeCache creates a new mock instance.
func NewMockCodeCache(ctrl *gomock.Controller) *MockCodeCache {
	mock := &MockCodeCache{ctrl: ctrl}
	mock.recorder = &MockCodeCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeCache) EXPECT() *MockCodeCacheMockRecorder {
	return m.recorder
}

// Key mocks base method.
func (m *MockCodeCache) Key(ctx context.Context, biz, phone string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Key", ctx, biz, phone)
	ret0, _ := ret[0].(string)
	return ret0
}

// Key indicates an expected call of Key.
func (mr *MockCodeCacheMockRecorder) Key(ctx, biz, phone interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Key", reflect.TypeOf((*MockCodeCache)(nil).Key), ctx, biz, phone)
}

// Set mocks base method.
func (m *MockCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, biz, phone, code)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCodeCacheMockRecorder) Set(ctx, biz, phone, code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCodeCache)(nil).Set), ctx, biz, phone, code)
}

// Verify mocks base method.
func (m *MockCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", ctx, biz, phone, inputCode)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockCodeCacheMockRecorder) Verify(ctx, biz, phone, inputCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockCodeCache)(nil).Verify), ctx, biz, phone, inputCode)
}
