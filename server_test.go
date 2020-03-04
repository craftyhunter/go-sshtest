package sshtest

import (
	"fmt"
	"github.com/craftyhunter/go-sshtest/protocol"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	server := NewMockedServer()
	server.AllowClientNoAuth()
	server.MockExecResult("echo OK", "OK\n", 0)
	host, port, err := server.Start()
	require.NoError(t, err)
	require.True(t, serverIsAlive(host, port))

	client := NewTestClient()
	client.Command = "echo OK"
	client.EnvVars = map[string]string{
		"VAR1": "VALUE1",
	}

	err = client.Connect(host, port)
	require.NoError(t, err)

	//check connection
	require.Len(t, server.ServedConnections(), 2)
	servedConn := server.ServedConnections()[1]
	require.Equal(t, client.ClientVersion, string(servedConn.ClientConn.ClientVersion()))
	require.Equal(t, client.ClientAddress, servedConn.RemoteAddr().String())
	require.Equal(t, client.User, servedConn.ClientConn.User())

	// check channel
	require.Len(t, servedConn.Stat.ServedChannels(), 1)
	servedChan := servedConn.Stat.ServedChannels()[0]
	require.Equal(t, "session", servedChan.Type)

	// check requests
	require.Len(t, servedChan.Stat.Requests(), 2)
	envReqRaw := servedChan.Stat.Requests()[0]
	require.IsType(t, &protocol.MsgRequestSetEnv{}, envReqRaw)
	envReq := envReqRaw.(*protocol.MsgRequestSetEnv)
	require.Equal(t, "VAR1", envReq.Name)
	require.Equal(t, "VALUE1", envReq.Value)

	execReqRaw := servedChan.Stat.Requests()[1]
	require.IsType(t, &protocol.MsgRequestExec{}, execReqRaw)
	execReq := execReqRaw.(*protocol.MsgRequestExec)
	require.Equal(t, "echo OK", execReq.Command)

	server.Stop()
	require.False(t, serverIsAlive(host, port))

	server.Wait()
}

func TestServer_AddAuthorizedKey(t *testing.T) {
	privateKey, _ := generateRSAKey(2048)
	signer, err := ssh.NewSignerFromKey(privateKey)
	require.NoError(t, err)
	publicRSAKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	server := NewMockedServer()
	server.AddAuthorizedKey(publicRSAKey)
	host, port, err := server.Start()
	require.NoError(t, err)

	client := NewTestClient()
	client.ClientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	client.Command = "echo OK"

	err = client.Connect(host, port)
	require.NoError(t, err)
	server.Stop()
	server.Wait()
}

func serverIsAlive(host string, port uint16) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Millisecond*300)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func NewTestClient() *TestClient {
	return &TestClient{
		ClientConfig: &ssh.ClientConfig{
			Timeout:         time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			ClientVersion:   "SSH-2.0-testClient",
			User:            "user1",
		},
	}
}

type TestClient struct {
	*ssh.ClientConfig
	Command       string
	EnvVars       map[string]string
	ClientAddress string
}

func (client *TestClient) Connect(host string, port uint16) (err error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	clientConn, err := ssh.Dial("tcp", addr, client.ClientConfig)
	if err != nil {
		return errors.Wrapf(err, "couldn't connect to %s", addr)
	}
	client.ClientAddress = clientConn.LocalAddr().String()
	defer func() {
		_ = clientConn.Close()
	}()

	session, err := clientConn.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	for k, v := range client.EnvVars {
		if err = session.Setenv(k, v); err != nil {
			return
		}
	}

	if err = session.Start(client.Command); err != nil {
		return
	}

	return session.Wait()
}
