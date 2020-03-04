package sshtest

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
		Stat: &ConnectionStat{
			mu: sync.Mutex{},
		},
	}
}

type Connection struct {
	net.Conn
	ClientConn *ssh.ServerConn
	mockData   *MockData
	Stat       *ConnectionStat
}

type ConnectionStat struct {
	mu        sync.Mutex
	StartTime time.Time
	StopTime  time.Time

	servedChannels []*Channel
}

func (s *ConnectionStat) appendChannel(ch *Channel) {
	s.mu.Lock()
	s.servedChannels = append(s.servedChannels, ch)
	s.mu.Unlock()
}

func (s *ConnectionStat) ServedChannels() []*Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*Channel{}, s.servedChannels...)
}

func (c *Connection) handle(serverConfig *ssh.ServerConfig) {
	c.Stat.StartTime = time.Now()
	defer func() {
		_ = c.Close()
		debugf("connection from '%s' closed", c.Conn.RemoteAddr().String())

	}()

	clientConn, channels, reqs, err := ssh.NewServerConn(c, serverConfig)
	if err != nil {
		if err != io.EOF {
			log.Fatal("failed to handshake: ", err)
		}
		return
	}
	debugf("client '%s' connected from %s", clientConn.ClientVersion(), c.RemoteAddr().String())
	c.ClientConn = clientConn

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	for newChannel := range channels {
		debugf("channel '%s' accepted", newChannel.ChannelType())
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			return
		}

		ch1 := NewChannel(newChannel)
		ch1.mockData = c.mockData
		c.Stat.appendChannel(ch1)

		go ch1.handle()
	}
	c.Stat.StopTime = time.Now()

	for _, ch := range c.Stat.ServedChannels() {
		for _, r := range ch.Stat.Requests() {
			debugf("accepted request: %[1]T %[1]v", r)
		}
	}

	debugf("client from '%s' disconnected. Duration: %s", c.Conn.RemoteAddr().String(), c.Stat.StopTime.Sub(c.Stat.StartTime).String())
}
