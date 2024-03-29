// Code generated by MockGen. DO NOT EDIT.
// Source: transformer.go

// Package transformerclient is a generated GoMock package.
package transformerclient

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockTransformer is a mock of Transformer interface.
type MockTransformer struct {
	ctrl     *gomock.Controller
	recorder *MockTransformerMockRecorder
}

// MockTransformerMockRecorder is the mock recorder for MockTransformer.
type MockTransformerMockRecorder struct {
	mock *MockTransformer
}

// NewMockTransformer creates a new mock instance.
func NewMockTransformer(ctrl *gomock.Controller) *MockTransformer {
	mock := &MockTransformer{ctrl: ctrl}
	mock.recorder = &MockTransformerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransformer) EXPECT() *MockTransformerMockRecorder {
	return m.recorder
}

// Expand mocks base method.
func (m *MockTransformer) Expand(ctx context.Context, in *ExpandRequest, opts ...grpc.CallOption) (*ExpandResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Expand", varargs...)
	ret0, _ := ret[0].(*ExpandResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Expand indicates an expected call of Expand.
func (mr *MockTransformerMockRecorder) Expand(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Expand", reflect.TypeOf((*MockTransformer)(nil).Expand), varargs...)
}

// Shorten mocks base method.
func (m *MockTransformer) Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Shorten", varargs...)
	ret0, _ := ret[0].(*ShortenResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Shorten indicates an expected call of Shorten.
func (mr *MockTransformerMockRecorder) Shorten(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shorten", reflect.TypeOf((*MockTransformer)(nil).Shorten), varargs...)
}
