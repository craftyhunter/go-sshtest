package sshtest

import (
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewConnection(conn net.Conn, mockData *MockData) *Connection {
	return &Connection{
		Conn:     conn,
		mockData: mockData,
		mu:       sync.Mutex{},
	}
}

type Connection struct {
	net.Conn
	ClientConn *ssh.ServerConn
	mockData   *MockData

	mu             sync.Mutex
	startTime      time.Time
	stopTime       time.Time
	servedChannels []*Channel
}

type ConnectionStat struct {
}

func (s *Connection) appendChannel(ch *Channel) {
	s.mu.Lock()
	s.servedChannels = append(s.servedChannels, ch)
	s.mu.Unlock()
}

func (s *Connection) ServedChannels() []*Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*Channel{}, s.servedChannels...)
}

func (c *Connection) handle(serverConfig *ssh.ServerConfig) {
	c.startTime = time.Now()
	defer func() {
		_ = c.Close()
		debugf("connection from '%s' closed", c.Conn.RemoteAddr().String())

	}()

	clientConn, channels, reqs, err := ssh.NewServerConn(c, serverConfig)
	if err != nil {
		if err != io.EOF {
			debugf("failed to handshake: %s", err)
			return
		}
		return
	}
	debugf("client '%s' connected from %s", clientConn.ClientVersion(), c.RemoteAddr().String())
	c.ClientConn = clientConn

	var wg sync.WaitGroup
	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	for newChannel := range channels {
		debugf("channel '%s' accepted", newChannel.ChannelType())
		switch newChannel.ChannelType() {
		case "session":
			ch1 := NewChannel(newChannel, c.mockData)
			c.appendChannel(ch1)
			wg.Add(1)
			go func() {
				ch1.handle()
				wg.Done()
			}()
		case "auth-agent@openssh.com":
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		default:
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		}
	}
	wg.Wait()

	c.stopTime = time.Now()
	debugf("client from '%s' disconnected. Duration: %s", c.Conn.RemoteAddr().String(), c.stopTime.Sub(c.startTime).String())

	for _, ch := range c.ServedChannels() {
		for _, r := range ch.Requests() {
			debugf("accepted request: %[1]T %[1]v", r)
		}
	}
}
