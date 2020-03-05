package sshtest

import (
	"sync"
	"time"
)

func NewMockData() *MockData {
	return &MockData{
		mu: sync.Mutex{},

		mockedExecRequests: make(map[string]mockedExecResultStatus),
	}
}

type MockData struct {
	mu sync.Mutex

	mockedExecRequests map[string]mockedExecResultStatus
}

type mockedExecResultStatus struct {
	exitStatus uint32
	result     string
	timeout    time.Duration
}

func (m *MockData) getMocksExecResult() map[string]mockedExecResultStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]mockedExecResultStatus)
	for k, v := range m.mockedExecRequests {
		result[k] = v
	}
	return result
}

func (m *MockData) MockExecResult(command, result string, timeout time.Duration, exitStatus uint32) {
	m.mu.Lock()
	m.mockedExecRequests[command] = mockedExecResultStatus{
		exitStatus: exitStatus,
		result:     result,
		timeout:    timeout,
	}
	m.mu.Unlock()
}
