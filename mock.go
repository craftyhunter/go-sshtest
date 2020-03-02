package sshtest

func NewMockData() *MockData {
	return &MockData{
		mockedExecRequests: make(map[string]mockedExecResultStatus),
	}
}

type MockData struct {
	mockedExecRequests map[string]mockedExecResultStatus
}

type mockedExecResultStatus struct {
	result     string
	exitStatus uint32
}

func (m *MockData) MockExecResult(command, result string, exitStatus uint32) {
	m.mockedExecRequests[command] = mockedExecResultStatus{
		result:     result,
		exitStatus: exitStatus,
	}
}
