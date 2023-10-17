// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/txn/client/types.go

// Package mock_frontend is a generated GoMock package.
package mock_frontend

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	lock "github.com/matrixorigin/matrixone/pkg/pb/lock"
	timestamp "github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	txn "github.com/matrixorigin/matrixone/pkg/pb/txn"
	client "github.com/matrixorigin/matrixone/pkg/txn/client"
	rpc "github.com/matrixorigin/matrixone/pkg/txn/rpc"
)

// MockTxnClient is a mock of TxnClient interface.
type MockTxnClient struct {
	ctrl     *gomock.Controller
	recorder *MockTxnClientMockRecorder
}

// MockTxnClientMockRecorder is the mock recorder for MockTxnClient.
type MockTxnClientMockRecorder struct {
	mock *MockTxnClient
}

// NewMockTxnClient creates a new mock instance.
func NewMockTxnClient(ctrl *gomock.Controller) *MockTxnClient {
	mock := &MockTxnClient{ctrl: ctrl}
	mock.recorder = &MockTxnClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxnClient) EXPECT() *MockTxnClientMockRecorder {
	return m.recorder
}

// AbortAllRunningTxn mocks base method.
func (m *MockTxnClient) AbortAllRunningTxn() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AbortAllRunningTxn")
}

// AbortAllRunningTxn indicates an expected call of AbortAllRunningTxn.
func (mr *MockTxnClientMockRecorder) AbortAllRunningTxn() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AbortAllRunningTxn", reflect.TypeOf((*MockTxnClient)(nil).AbortAllRunningTxn))
}

// CNBasedConsistencyEnabled mocks base method.
func (m *MockTxnClient) CNBasedConsistencyEnabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CNBasedConsistencyEnabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// CNBasedConsistencyEnabled indicates an expected call of CNBasedConsistencyEnabled.
func (mr *MockTxnClientMockRecorder) CNBasedConsistencyEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CNBasedConsistencyEnabled", reflect.TypeOf((*MockTxnClient)(nil).CNBasedConsistencyEnabled))
}

// Close mocks base method.
func (m *MockTxnClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockTxnClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTxnClient)(nil).Close))
}

// GetLatestCommitTS mocks base method.
func (m *MockTxnClient) GetLatestCommitTS() timestamp.Timestamp {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestCommitTS")
	ret0, _ := ret[0].(timestamp.Timestamp)
	return ret0
}

// GetLatestCommitTS indicates an expected call of GetLatestCommitTS.
func (mr *MockTxnClientMockRecorder) GetLatestCommitTS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestCommitTS", reflect.TypeOf((*MockTxnClient)(nil).GetLatestCommitTS))
}

// GetSyncLatestCommitTSTimes mocks base method.
func (m *MockTxnClient) GetSyncLatestCommitTSTimes() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSyncLatestCommitTSTimes")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetSyncLatestCommitTSTimes indicates an expected call of GetSyncLatestCommitTSTimes.
func (mr *MockTxnClientMockRecorder) GetSyncLatestCommitTSTimes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSyncLatestCommitTSTimes", reflect.TypeOf((*MockTxnClient)(nil).GetSyncLatestCommitTSTimes))
}

// IterTxns mocks base method.
func (m *MockTxnClient) IterTxns(arg0 func(client.TxnOverview) bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "IterTxns", arg0)
}

// IterTxns indicates an expected call of IterTxns.
func (mr *MockTxnClientMockRecorder) IterTxns(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IterTxns", reflect.TypeOf((*MockTxnClient)(nil).IterTxns), arg0)
}

// MinTimestamp mocks base method.
func (m *MockTxnClient) MinTimestamp() timestamp.Timestamp {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MinTimestamp")
	ret0, _ := ret[0].(timestamp.Timestamp)
	return ret0
}

// MinTimestamp indicates an expected call of MinTimestamp.
func (mr *MockTxnClientMockRecorder) MinTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MinTimestamp", reflect.TypeOf((*MockTxnClient)(nil).MinTimestamp))
}

// New mocks base method.
func (m *MockTxnClient) New(ctx context.Context, commitTS timestamp.Timestamp, options ...client.TxnOption) (client.TxnOperator, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, commitTS}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "New", varargs...)
	ret0, _ := ret[0].(client.TxnOperator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New.
func (mr *MockTxnClientMockRecorder) New(ctx, commitTS interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, commitTS}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockTxnClient)(nil).New), varargs...)
}

// NewWithSnapshot mocks base method.
func (m *MockTxnClient) NewWithSnapshot(snapshot []byte) (client.TxnOperator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewWithSnapshot", snapshot)
	ret0, _ := ret[0].(client.TxnOperator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewWithSnapshot indicates an expected call of NewWithSnapshot.
func (mr *MockTxnClientMockRecorder) NewWithSnapshot(snapshot interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewWithSnapshot", reflect.TypeOf((*MockTxnClient)(nil).NewWithSnapshot), snapshot)
}

// Pause mocks base method.
func (m *MockTxnClient) Pause() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Pause")
}

// Pause indicates an expected call of Pause.
func (mr *MockTxnClientMockRecorder) Pause() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockTxnClient)(nil).Pause))
}

// RefreshExpressionEnabled mocks base method.
func (m *MockTxnClient) RefreshExpressionEnabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshExpressionEnabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// RefreshExpressionEnabled indicates an expected call of RefreshExpressionEnabled.
func (mr *MockTxnClientMockRecorder) RefreshExpressionEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshExpressionEnabled", reflect.TypeOf((*MockTxnClient)(nil).RefreshExpressionEnabled))
}

// Resume mocks base method.
func (m *MockTxnClient) Resume() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Resume")
}

// Resume indicates an expected call of Resume.
func (mr *MockTxnClientMockRecorder) Resume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resume", reflect.TypeOf((*MockTxnClient)(nil).Resume))
}

// SyncLatestCommitTS mocks base method.
func (m *MockTxnClient) SyncLatestCommitTS(arg0 timestamp.Timestamp) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SyncLatestCommitTS", arg0)
}

// SyncLatestCommitTS indicates an expected call of SyncLatestCommitTS.
func (mr *MockTxnClientMockRecorder) SyncLatestCommitTS(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncLatestCommitTS", reflect.TypeOf((*MockTxnClient)(nil).SyncLatestCommitTS), arg0)
}

// WaitLogTailAppliedAt mocks base method.
func (m *MockTxnClient) WaitLogTailAppliedAt(ctx context.Context, ts timestamp.Timestamp) (timestamp.Timestamp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitLogTailAppliedAt", ctx, ts)
	ret0, _ := ret[0].(timestamp.Timestamp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitLogTailAppliedAt indicates an expected call of WaitLogTailAppliedAt.
func (mr *MockTxnClientMockRecorder) WaitLogTailAppliedAt(ctx, ts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitLogTailAppliedAt", reflect.TypeOf((*MockTxnClient)(nil).WaitLogTailAppliedAt), ctx, ts)
}

// MockTxnOperator is a mock of TxnOperator interface.
type MockTxnOperator struct {
	ctrl     *gomock.Controller
	recorder *MockTxnOperatorMockRecorder
}

// MockTxnOperatorMockRecorder is the mock recorder for MockTxnOperator.
type MockTxnOperatorMockRecorder struct {
	mock *MockTxnOperator
}

// NewMockTxnOperator creates a new mock instance.
func NewMockTxnOperator(ctrl *gomock.Controller) *MockTxnOperator {
	mock := &MockTxnOperator{ctrl: ctrl}
	mock.recorder = &MockTxnOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxnOperator) EXPECT() *MockTxnOperatorMockRecorder {
	return m.recorder
}

// AddLockTable mocks base method.
func (m *MockTxnOperator) AddLockTable(locktable lock.LockTable) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddLockTable", locktable)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddLockTable indicates an expected call of AddLockTable.
func (mr *MockTxnOperatorMockRecorder) AddLockTable(locktable interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLockTable", reflect.TypeOf((*MockTxnOperator)(nil).AddLockTable), locktable)
}

// AddWaitLock mocks base method.
func (m *MockTxnOperator) AddWaitLock(tableID uint64, rows [][]byte, opt lock.LockOptions) uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddWaitLock", tableID, rows, opt)
	ret0, _ := ret[0].(uint64)
	return ret0
}

// AddWaitLock indicates an expected call of AddWaitLock.
func (mr *MockTxnOperatorMockRecorder) AddWaitLock(tableID, rows, opt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWaitLock", reflect.TypeOf((*MockTxnOperator)(nil).AddWaitLock), tableID, rows, opt)
}

// AddWorkspace mocks base method.
func (m *MockTxnOperator) AddWorkspace(workspace client.Workspace) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddWorkspace", workspace)
}

// AddWorkspace indicates an expected call of AddWorkspace.
func (mr *MockTxnOperatorMockRecorder) AddWorkspace(workspace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWorkspace", reflect.TypeOf((*MockTxnOperator)(nil).AddWorkspace), workspace)
}

// AppendEventCallback mocks base method.
func (m *MockTxnOperator) AppendEventCallback(event client.EventType, callbacks ...func(txn.TxnMeta)) {
	m.ctrl.T.Helper()
	varargs := []interface{}{event}
	for _, a := range callbacks {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AppendEventCallback", varargs...)
}

// AppendEventCallback indicates an expected call of AppendEventCallback.
func (mr *MockTxnOperatorMockRecorder) AppendEventCallback(event interface{}, callbacks ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{event}, callbacks...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendEventCallback", reflect.TypeOf((*MockTxnOperator)(nil).AppendEventCallback), varargs...)
}

// ApplySnapshot mocks base method.
func (m *MockTxnOperator) ApplySnapshot(data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplySnapshot", data)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplySnapshot indicates an expected call of ApplySnapshot.
func (mr *MockTxnOperatorMockRecorder) ApplySnapshot(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplySnapshot", reflect.TypeOf((*MockTxnOperator)(nil).ApplySnapshot), data)
}

// Commit mocks base method.
func (m *MockTxnOperator) Commit(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit.
func (mr *MockTxnOperatorMockRecorder) Commit(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTxnOperator)(nil).Commit), ctx)
}

// Debug mocks base method.
func (m *MockTxnOperator) Debug(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Debug", ctx, ops)
	ret0, _ := ret[0].(*rpc.SendResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Debug indicates an expected call of Debug.
func (mr *MockTxnOperatorMockRecorder) Debug(ctx, ops interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockTxnOperator)(nil).Debug), ctx, ops)
}

// GetOverview mocks base method.
func (m *MockTxnOperator) GetOverview() client.TxnOverview {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOverview")
	ret0, _ := ret[0].(client.TxnOverview)
	return ret0
}

// GetOverview indicates an expected call of GetOverview.
func (mr *MockTxnOperatorMockRecorder) GetOverview() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOverview", reflect.TypeOf((*MockTxnOperator)(nil).GetOverview))
}

// GetWorkspace mocks base method.
func (m *MockTxnOperator) GetWorkspace() client.Workspace {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkspace")
	ret0, _ := ret[0].(client.Workspace)
	return ret0
}

// GetWorkspace indicates an expected call of GetWorkspace.
func (mr *MockTxnOperatorMockRecorder) GetWorkspace() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkspace", reflect.TypeOf((*MockTxnOperator)(nil).GetWorkspace))
}

// IsRetry mocks base method.
func (m *MockTxnOperator) IsRetry() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRetry")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsRetry indicates an expected call of IsRetry.
func (mr *MockTxnOperatorMockRecorder) IsRetry() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRetry", reflect.TypeOf((*MockTxnOperator)(nil).IsRetry))
}

// Read mocks base method.
func (m *MockTxnOperator) Read(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", ctx, ops)
	ret0, _ := ret[0].(*rpc.SendResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockTxnOperatorMockRecorder) Read(ctx, ops interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockTxnOperator)(nil).Read), ctx, ops)
}

// RemoveWaitLock mocks base method.
func (m *MockTxnOperator) RemoveWaitLock(key uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveWaitLock", key)
}

// RemoveWaitLock indicates an expected call of RemoveWaitLock.
func (mr *MockTxnOperatorMockRecorder) RemoveWaitLock(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveWaitLock", reflect.TypeOf((*MockTxnOperator)(nil).RemoveWaitLock), key)
}

// ResetRetry mocks base method.
func (m *MockTxnOperator) ResetRetry(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ResetRetry", arg0)
}

// ResetRetry indicates an expected call of ResetRetry.
func (mr *MockTxnOperatorMockRecorder) ResetRetry(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetRetry", reflect.TypeOf((*MockTxnOperator)(nil).ResetRetry), arg0)
}

// Rollback mocks base method.
func (m *MockTxnOperator) Rollback(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *MockTxnOperatorMockRecorder) Rollback(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockTxnOperator)(nil).Rollback), ctx)
}

// Snapshot mocks base method.
func (m *MockTxnOperator) Snapshot() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Snapshot")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Snapshot indicates an expected call of Snapshot.
func (mr *MockTxnOperatorMockRecorder) Snapshot() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Snapshot", reflect.TypeOf((*MockTxnOperator)(nil).Snapshot))
}

// SnapshotTS mocks base method.
func (m *MockTxnOperator) SnapshotTS() timestamp.Timestamp {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SnapshotTS")
	ret0, _ := ret[0].(timestamp.Timestamp)
	return ret0
}

// SnapshotTS indicates an expected call of SnapshotTS.
func (mr *MockTxnOperatorMockRecorder) SnapshotTS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SnapshotTS", reflect.TypeOf((*MockTxnOperator)(nil).SnapshotTS))
}

// Status mocks base method.
func (m *MockTxnOperator) Status() txn.TxnStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(txn.TxnStatus)
	return ret0
}

// Status indicates an expected call of Status.
func (mr *MockTxnOperatorMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockTxnOperator)(nil).Status))
}

// Txn mocks base method.
func (m *MockTxnOperator) Txn() txn.TxnMeta {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Txn")
	ret0, _ := ret[0].(txn.TxnMeta)
	return ret0
}

// Txn indicates an expected call of Txn.
func (mr *MockTxnOperatorMockRecorder) Txn() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Txn", reflect.TypeOf((*MockTxnOperator)(nil).Txn))
}

// TxnRef mocks base method.
func (m *MockTxnOperator) TxnRef() *txn.TxnMeta {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxnRef")
	ret0, _ := ret[0].(*txn.TxnMeta)
	return ret0
}

// TxnRef indicates an expected call of TxnRef.
func (mr *MockTxnOperatorMockRecorder) TxnRef() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxnRef", reflect.TypeOf((*MockTxnOperator)(nil).TxnRef))
}

// UpdateSnapshot mocks base method.
func (m *MockTxnOperator) UpdateSnapshot(ctx context.Context, ts timestamp.Timestamp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSnapshot", ctx, ts)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSnapshot indicates an expected call of UpdateSnapshot.
func (mr *MockTxnOperatorMockRecorder) UpdateSnapshot(ctx, ts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSnapshot", reflect.TypeOf((*MockTxnOperator)(nil).UpdateSnapshot), ctx, ts)
}

// Write mocks base method.
func (m *MockTxnOperator) Write(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", ctx, ops)
	ret0, _ := ret[0].(*rpc.SendResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockTxnOperatorMockRecorder) Write(ctx, ops interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockTxnOperator)(nil).Write), ctx, ops)
}

// WriteAndCommit mocks base method.
func (m *MockTxnOperator) WriteAndCommit(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteAndCommit", ctx, ops)
	ret0, _ := ret[0].(*rpc.SendResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WriteAndCommit indicates an expected call of WriteAndCommit.
func (mr *MockTxnOperatorMockRecorder) WriteAndCommit(ctx, ops interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteAndCommit", reflect.TypeOf((*MockTxnOperator)(nil).WriteAndCommit), ctx, ops)
}

// MockTxnIDGenerator is a mock of TxnIDGenerator interface.
type MockTxnIDGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockTxnIDGeneratorMockRecorder
}

// MockTxnIDGeneratorMockRecorder is the mock recorder for MockTxnIDGenerator.
type MockTxnIDGeneratorMockRecorder struct {
	mock *MockTxnIDGenerator
}

// NewMockTxnIDGenerator creates a new mock instance.
func NewMockTxnIDGenerator(ctrl *gomock.Controller) *MockTxnIDGenerator {
	mock := &MockTxnIDGenerator{ctrl: ctrl}
	mock.recorder = &MockTxnIDGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTxnIDGenerator) EXPECT() *MockTxnIDGeneratorMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockTxnIDGenerator) Generate() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Generate indicates an expected call of Generate.
func (mr *MockTxnIDGeneratorMockRecorder) Generate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockTxnIDGenerator)(nil).Generate))
}

// MockTimestampWaiter is a mock of TimestampWaiter interface.
type MockTimestampWaiter struct {
	ctrl     *gomock.Controller
	recorder *MockTimestampWaiterMockRecorder
}

// MockTimestampWaiterMockRecorder is the mock recorder for MockTimestampWaiter.
type MockTimestampWaiterMockRecorder struct {
	mock *MockTimestampWaiter
}

// NewMockTimestampWaiter creates a new mock instance.
func NewMockTimestampWaiter(ctrl *gomock.Controller) *MockTimestampWaiter {
	mock := &MockTimestampWaiter{ctrl: ctrl}
	mock.recorder = &MockTimestampWaiterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTimestampWaiter) EXPECT() *MockTimestampWaiterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockTimestampWaiter) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockTimestampWaiterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTimestampWaiter)(nil).Close))
}

// GetTimestamp mocks base method.
func (m *MockTimestampWaiter) GetTimestamp(arg0 context.Context, arg1 timestamp.Timestamp) (timestamp.Timestamp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTimestamp", arg0, arg1)
	ret0, _ := ret[0].(timestamp.Timestamp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTimestamp indicates an expected call of GetTimestamp.
func (mr *MockTimestampWaiterMockRecorder) GetTimestamp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTimestamp", reflect.TypeOf((*MockTimestampWaiter)(nil).GetTimestamp), arg0, arg1)
}

// NotifyLatestCommitTS mocks base method.
func (m *MockTimestampWaiter) NotifyLatestCommitTS(appliedTS timestamp.Timestamp) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "NotifyLatestCommitTS", appliedTS)
}

// NotifyLatestCommitTS indicates an expected call of NotifyLatestCommitTS.
func (mr *MockTimestampWaiterMockRecorder) NotifyLatestCommitTS(appliedTS interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyLatestCommitTS", reflect.TypeOf((*MockTimestampWaiter)(nil).NotifyLatestCommitTS), appliedTS)
}

// MockWorkspace is a mock of Workspace interface.
type MockWorkspace struct {
	ctrl     *gomock.Controller
	recorder *MockWorkspaceMockRecorder
}

// MockWorkspaceMockRecorder is the mock recorder for MockWorkspace.
type MockWorkspaceMockRecorder struct {
	mock *MockWorkspace
}

// NewMockWorkspace creates a new mock instance.
func NewMockWorkspace(ctrl *gomock.Controller) *MockWorkspace {
	mock := &MockWorkspace{ctrl: ctrl}
	mock.recorder = &MockWorkspaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkspace) EXPECT() *MockWorkspaceMockRecorder {
	return m.recorder
}

// Adjust mocks base method.
func (m *MockWorkspace) Adjust() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Adjust")
	ret0, _ := ret[0].(error)
	return ret0
}

// Adjust indicates an expected call of Adjust.
func (mr *MockWorkspaceMockRecorder) Adjust() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Adjust", reflect.TypeOf((*MockWorkspace)(nil).Adjust))
}

// Commit mocks base method.
func (m *MockWorkspace) Commit(ctx context.Context) ([]txn.TxnRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", ctx)
	ret0, _ := ret[0].([]txn.TxnRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockWorkspaceMockRecorder) Commit(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockWorkspace)(nil).Commit), ctx)
}

// EndStatement mocks base method.
func (m *MockWorkspace) EndStatement() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EndStatement")
}

// EndStatement indicates an expected call of EndStatement.
func (mr *MockWorkspaceMockRecorder) EndStatement() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndStatement", reflect.TypeOf((*MockWorkspace)(nil).EndStatement))
}

// GetSQLCount mocks base method.
func (m *MockWorkspace) GetSQLCount() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSQLCount")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetSQLCount indicates an expected call of GetSQLCount.
func (mr *MockWorkspaceMockRecorder) GetSQLCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSQLCount", reflect.TypeOf((*MockWorkspace)(nil).GetSQLCount))
}

// IncrSQLCount mocks base method.
func (m *MockWorkspace) IncrSQLCount() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "IncrSQLCount")
}

// IncrSQLCount indicates an expected call of IncrSQLCount.
func (mr *MockWorkspaceMockRecorder) IncrSQLCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrSQLCount", reflect.TypeOf((*MockWorkspace)(nil).IncrSQLCount))
}

// IncrStatementID mocks base method.
func (m *MockWorkspace) IncrStatementID(ctx context.Context, commit bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncrStatementID", ctx, commit)
	ret0, _ := ret[0].(error)
	return ret0
}

// IncrStatementID indicates an expected call of IncrStatementID.
func (mr *MockWorkspaceMockRecorder) IncrStatementID(ctx, commit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrStatementID", reflect.TypeOf((*MockWorkspace)(nil).IncrStatementID), ctx, commit)
}

// Rollback mocks base method.
func (m *MockWorkspace) Rollback(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *MockWorkspaceMockRecorder) Rollback(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockWorkspace)(nil).Rollback), ctx)
}

// RollbackLastStatement mocks base method.
func (m *MockWorkspace) RollbackLastStatement(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RollbackLastStatement", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// RollbackLastStatement indicates an expected call of RollbackLastStatement.
func (mr *MockWorkspaceMockRecorder) RollbackLastStatement(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RollbackLastStatement", reflect.TypeOf((*MockWorkspace)(nil).RollbackLastStatement), ctx)
}

// StartStatement mocks base method.
func (m *MockWorkspace) StartStatement() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartStatement")
}

// StartStatement indicates an expected call of StartStatement.
func (mr *MockWorkspaceMockRecorder) StartStatement() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartStatement", reflect.TypeOf((*MockWorkspace)(nil).StartStatement))
}
