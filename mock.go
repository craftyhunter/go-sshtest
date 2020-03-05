package sshtest

import (
	"sync"
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
	result     string
	exitStatus uint32
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

func (m *MockData) MockExecResult(command, result string, exitStatus uint32) {
	m.mu.Lock()
	m.mockedExecRequests[command] = mockedExecResultStatus{
		result:     result,
		exitStatus: exitStatus,
	}
	m.mu.Unlock()
}
