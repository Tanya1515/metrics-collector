// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Tanya1515/metrics-collector.git/cmd/storage (interfaces: RepositoryInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRepositoryInterface is a mock of RepositoryInterface interface.
type MockRepositoryInterface struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryInterfaceMockRecorder
}

// MockRepositoryInterfaceMockRecorder is the mock recorder for MockRepositoryInterface.
type MockRepositoryInterfaceMockRecorder struct {
	mock *MockRepositoryInterface
}

// NewMockRepositoryInterface creates a new mock instance.
func NewMockRepositoryInterface(ctrl *gomock.Controller) *MockRepositoryInterface {
	mock := &MockRepositoryInterface{ctrl: ctrl}
	mock.recorder = &MockRepositoryInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoryInterface) EXPECT() *MockRepositoryInterfaceMockRecorder {
	return m.recorder
}

// CheckConnection mocks base method.
func (m *MockRepositoryInterface) CheckConnection(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckConnection", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckConnection indicates an expected call of CheckConnection.
func (mr *MockRepositoryInterfaceMockRecorder) CheckConnection(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckConnection", reflect.TypeOf((*MockRepositoryInterface)(nil).CheckConnection), arg0)
}

// GetAllCounterMetrics mocks base method.
func (m *MockRepositoryInterface) GetAllCounterMetrics() (map[string]int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllCounterMetrics")
	ret0, _ := ret[0].(map[string]int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllCounterMetrics indicates an expected call of GetAllCounterMetrics.
func (mr *MockRepositoryInterfaceMockRecorder) GetAllCounterMetrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllCounterMetrics", reflect.TypeOf((*MockRepositoryInterface)(nil).GetAllCounterMetrics))
}

// GetAllGaugeMetrics mocks base method.
func (m *MockRepositoryInterface) GetAllGaugeMetrics() (map[string]float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllGaugeMetrics")
	ret0, _ := ret[0].(map[string]float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllGaugeMetrics indicates an expected call of GetAllGaugeMetrics.
func (mr *MockRepositoryInterfaceMockRecorder) GetAllGaugeMetrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllGaugeMetrics", reflect.TypeOf((*MockRepositoryInterface)(nil).GetAllGaugeMetrics))
}

// GetCounterValueByName mocks base method.
func (m *MockRepositoryInterface) GetCounterValueByName(arg0 string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounterValueByName", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCounterValueByName indicates an expected call of GetCounterValueByName.
func (mr *MockRepositoryInterfaceMockRecorder) GetCounterValueByName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounterValueByName", reflect.TypeOf((*MockRepositoryInterface)(nil).GetCounterValueByName), arg0)
}

// GetGaugeValueByName mocks base method.
func (m *MockRepositoryInterface) GetGaugeValueByName(arg0 string) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGaugeValueByName", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGaugeValueByName indicates an expected call of GetGaugeValueByName.
func (mr *MockRepositoryInterfaceMockRecorder) GetGaugeValueByName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGaugeValueByName", reflect.TypeOf((*MockRepositoryInterface)(nil).GetGaugeValueByName), arg0)
}

// Init mocks base method.
func (m *MockRepositoryInterface) Init() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init")
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockRepositoryInterfaceMockRecorder) Init() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockRepositoryInterface)(nil).Init))
}

// RepositoryAddCounterValue mocks base method.
func (m *MockRepositoryInterface) RepositoryAddCounterValue(arg0 string, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RepositoryAddCounterValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RepositoryAddCounterValue indicates an expected call of RepositoryAddCounterValue.
func (mr *MockRepositoryInterfaceMockRecorder) RepositoryAddCounterValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryAddCounterValue", reflect.TypeOf((*MockRepositoryInterface)(nil).RepositoryAddCounterValue), arg0, arg1)
}

// RepositoryAddGaugeValue mocks base method.
func (m *MockRepositoryInterface) RepositoryAddGaugeValue(arg0 string, arg1 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RepositoryAddGaugeValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RepositoryAddGaugeValue indicates an expected call of RepositoryAddGaugeValue.
func (mr *MockRepositoryInterfaceMockRecorder) RepositoryAddGaugeValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryAddGaugeValue", reflect.TypeOf((*MockRepositoryInterface)(nil).RepositoryAddGaugeValue), arg0, arg1)
}

// RepositoryAddValue mocks base method.
func (m *MockRepositoryInterface) RepositoryAddValue(arg0 string, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RepositoryAddValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RepositoryAddValue indicates an expected call of RepositoryAddValue.
func (mr *MockRepositoryInterfaceMockRecorder) RepositoryAddValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryAddValue", reflect.TypeOf((*MockRepositoryInterface)(nil).RepositoryAddValue), arg0, arg1)
}
