// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import (
	context "context"

	data "github.com/stellar/stellar-disbursement-platform-backend/internal/data"
	mock "github.com/stretchr/testify/mock"

	schema "github.com/stellar/stellar-disbursement-platform-backend/pkg/schema"
)

// MockDistributionAccountService is an autogenerated mock type for the DistributionAccountServiceInterface type
type MockDistributionAccountService struct {
	mock.Mock
}

// GetBalance provides a mock function with given fields: _a0, account, asset
func (_m *MockDistributionAccountService) GetBalance(_a0 context.Context, account *schema.TransactionAccount, asset data.Asset) (float64, error) {
	ret := _m.Called(_a0, account, asset)

	if len(ret) == 0 {
		panic("no return value specified for GetBalance")
	}

	var r0 float64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *schema.TransactionAccount, data.Asset) (float64, error)); ok {
		return rf(_a0, account, asset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *schema.TransactionAccount, data.Asset) float64); ok {
		r0 = rf(_a0, account, asset)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *schema.TransactionAccount, data.Asset) error); ok {
		r1 = rf(_a0, account, asset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBalances provides a mock function with given fields: _a0, account
func (_m *MockDistributionAccountService) GetBalances(_a0 context.Context, account *schema.TransactionAccount) (map[data.Asset]float64, error) {
	ret := _m.Called(_a0, account)

	if len(ret) == 0 {
		panic("no return value specified for GetBalances")
	}

	var r0 map[data.Asset]float64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *schema.TransactionAccount) (map[data.Asset]float64, error)); ok {
		return rf(_a0, account)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *schema.TransactionAccount) map[data.Asset]float64); ok {
		r0 = rf(_a0, account)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[data.Asset]float64)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *schema.TransactionAccount) error); ok {
		r1 = rf(_a0, account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockDistributionAccountService creates a new instance of MockDistributionAccountService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDistributionAccountService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDistributionAccountService {
	mock := &MockDistributionAccountService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}