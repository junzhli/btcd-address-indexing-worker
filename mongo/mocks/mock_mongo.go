// Code generated by MockGen. DO NOT EDIT.
// Source: mongo.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	mongo "github.com/junzhli/btcd-address-indexing-worker/mongo"
	reflect "reflect"
)

// MockMongo is a mock of Mongo interface
type MockMongo struct {
	ctrl     *gomock.Controller
	recorder *MockMongoMockRecorder
}

// MockMongoMockRecorder is the mock recorder for MockMongo
type MockMongoMockRecorder struct {
	mock *MockMongo
}

// NewMockMongo creates a new mock instance
func NewMockMongo(ctrl *gomock.Controller) *MockMongo {
	mock := &MockMongo{ctrl: ctrl}
	mock.recorder = &MockMongoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMongo) EXPECT() *MockMongoMockRecorder {
	return m.recorder
}

// PutUserHistory mocks base method
func (m *MockMongo) PutUserHistory(doc *mongo.UserHistory) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutUserHistory", doc)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutUserHistory indicates an expected call of PutUserHistory
func (mr *MockMongoMockRecorder) PutUserHistory(doc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutUserHistory", reflect.TypeOf((*MockMongo)(nil).PutUserHistory), doc)
}

// GetUserHistory mocks base method
func (m *MockMongo) GetUserHistory(addr string) (*mongo.UserHistory, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserHistory", addr)
	ret0, _ := ret[0].(*mongo.UserHistory)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserHistory indicates an expected call of GetUserHistory
func (mr *MockMongoMockRecorder) GetUserHistory(addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserHistory", reflect.TypeOf((*MockMongo)(nil).GetUserHistory), addr)
}
