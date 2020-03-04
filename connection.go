package sshtest

import (
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
		Stat: new(ConnectionStat),
	}
}

type Connection struct {
	net.Conn
	*MockData
	Stat *ConnectionStat
}

type ConnectionStat struct {
	StartTime time.Time
	StopTime  time.Time

	ClientVersion    string
	ClientRemoteAddr string
	Authenticated    bool
	AuthTries        []AuthType

	ServedChannels []*Channel
}

func (c *Connection) handle(serverConfig *ssh.ServerConfig) {
	c.Stat.StartTime = time.Now()
	defer func() {
		_ = c.Close()
		debugf("connection from '%s' closed", c.Conn.RemoteAddr().String())

	}()

	clientConn, channels, reqs, err := ssh.NewServerConn(c, serverConfig)
	if err != nil {
		log.Fatal("failed to handshake: ", err)
	}
	debugf("client '%s' connected from %s", clientConn.ClientVersion(), c.RemoteAddr().String())
	c.Stat.ClientVersion = string(clientConn.ClientVersion())
	c.Stat.ClientRemoteAddr = c.RemoteAddr().String()

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	for newChannel := range channels {
		debugf("channel '%s' accepted", newChannel.ChannelType())
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			return
		}

		ch1 := NewChannel(newChannel)
		ch1.MockData = c.MockData
		c.Stat.ServedChannels = append(c.Stat.ServedChannels, ch1)

		go ch1.handle()
	}
	c.Stat.StopTime = time.Now()

	for _, ch := range c.Stat.ServedChannels {
		for _, r := range ch.Stat.Requests() {
			debugf("accepted request: %v", r)
		}
	}

	debugf("client from '%s' disconnected. Duration: %s", c.Conn.RemoteAddr().String(), c.Stat.StopTime.Sub(c.Stat.StartTime).String())
}
