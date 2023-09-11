// Code generated by mockery v2.27.1. DO NOT EDIT.

package mocks

import (
	http "net/http"
	time "time"

	monitor "github.com/stellar/stellar-disbursement-platform-backend/internal/monitor"
	mock "github.com/stretchr/testify/mock"
)

// MockMonitorClient is an autogenerated mock type for the MonitorClient type
type MockMonitorClient struct {
	mock.Mock
}

// GetMetricHttpHandler provides a mock function with given fields:
func (_m *MockMonitorClient) GetMetricHttpHandler() http.Handler {
	ret := _m.Called()

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func() http.Handler); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// GetMetricType provides a mock function with given fields:
func (_m *MockMonitorClient) GetMetricType() monitor.MetricType {
	ret := _m.Called()

	var r0 monitor.MetricType
	if rf, ok := ret.Get(0).(func() monitor.MetricType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(monitor.MetricType)
	}

	return r0
}

// MonitorCounters provides a mock function with given fields: tag, labels
func (_m *MockMonitorClient) MonitorCounters(tag monitor.MetricTag, labels map[string]string) {
	_m.Called(tag, labels)
}

// MonitorDBQueryDuration provides a mock function with given fields: duration, tag, labels
func (_m *MockMonitorClient) MonitorDBQueryDuration(duration time.Duration, tag monitor.MetricTag, labels monitor.DBQueryLabels) {
	_m.Called(duration, tag, labels)
}

// MonitorDuration provides a mock function with given fields: duration, tag, labels
func (_m *MockMonitorClient) MonitorDuration(duration time.Duration, tag monitor.MetricTag, labels map[string]string) {
	_m.Called(duration, tag, labels)
}

// MonitorHistogram provides a mock function with given fields: value, tag, labels
func (_m *MockMonitorClient) MonitorHistogram(value float64, tag monitor.MetricTag, labels map[string]string) {
	_m.Called(value, tag, labels)
}

// MonitorHttpRequestDuration provides a mock function with given fields: duration, labels
func (_m *MockMonitorClient) MonitorHttpRequestDuration(duration time.Duration, labels monitor.HttpRequestLabels) {
	_m.Called(duration, labels)
}

type mockConstructorTestingTNewMockMonitorClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockMonitorClient creates a new instance of MockMonitorClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockMonitorClient(t mockConstructorTestingTNewMockMonitorClient) *MockMonitorClient {
	mock := &MockMonitorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}