package sshtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDebugOn(t *testing.T) {
	debugEnabled = false
	DebugOn()
	require.Equal(t, true, debugEnabled)
}

func TestDebugOff(t *testing.T) {
	debugEnabled = true
	DebugOff()
	require.Equal(t, false, debugEnabled)
}
