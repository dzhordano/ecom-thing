// code generated by MockGen. DO NOT EDIT.
// Source: internal/application/interfaces/product.go

// Package mock_interfaces is a generated GoMock package.
package mock_interfaces

import (
	context "context"
	reflect "reflect"

	domain "github.com/dzhordano/ecom-thing/services/product/internal/domain"
	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
)

// MockProductService is a mock of ProductService interface.
type MockProductService struct {
	ctrl     *gomock.Controller
	recorder *MockProductServiceMockRecorder
}

// MockProductServiceMockRecorder is the mock recorder for MockProductService.
type MockProductServiceMockRecorder struct {
	mock *MockProductService
}

// NewMockProductService creates a new mock instance.
func NewMockProductService(ctrl *gomock.Controller) *MockProductService {
	mock := &MockProductService{ctrl: ctrl}
	mock.recorder = &MockProductServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProductService) EXPECT() *MockProductServiceMockRecorder {
	return m.recorder
}

// CreateProduct mocks base method.
func (m *MockProductService) CreateProduct(ctx context.Context, name, description, category string, price float64) (*domain.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProduct", ctx, name, description, category, price)
	ret0, _ := ret[0].(*domain.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateProduct indicates an expected call of CreateProduct.
func (mr *MockProductServiceMockRecorder) CreateProduct(ctx, name, description, category, price interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProduct", reflect.TypeOf((*MockProductService)(nil).CreateProduct), ctx, name, description, category, price)
}

// DeactivateProduct mocks base method.
func (m *MockProductService) DeactivateProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeactivateProduct", ctx, id)
	ret0, _ := ret[0].(*domain.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeactivateProduct indicates an expected call of DeactivateProduct.
func (mr *MockProductServiceMockRecorder) DeactivateProduct(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeactivateProduct", reflect.TypeOf((*MockProductService)(nil).DeactivateProduct), ctx, id)
}

// GetById mocks base method.
func (m *MockProductService) GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", ctx, id)
	ret0, _ := ret[0].(*domain.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockProductServiceMockRecorder) GetById(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockProductService)(nil).GetById), ctx, id)
}

// SearchProducts mocks base method.
func (m *MockProductService) SearchProducts(ctx context.Context, filters map[string]any) ([]*domain.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchProducts", ctx, filters)
	ret0, _ := ret[0].([]*domain.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchProducts indicates an expected call of SearchProducts.
func (mr *MockProductServiceMockRecorder) SearchProducts(ctx, filters interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchProducts", reflect.TypeOf((*MockProductService)(nil).SearchProducts), ctx, filters)
}

// UpdateProduct mocks base method.
func (m *MockProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, isActive bool, price float64) (*domain.Product, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProduct", ctx, id, name, description, category, isActive, price)
	ret0, _ := ret[0].(*domain.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateProduct indicates an expected call of UpdateProduct.
func (mr *MockProductServiceMockRecorder) UpdateProduct(ctx, id, name, description, category, isActive, price interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProduct", reflect.TypeOf((*MockProductService)(nil).UpdateProduct), ctx, id, name, description, category, isActive, price)
}
