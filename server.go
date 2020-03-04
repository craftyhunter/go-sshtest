package sshtest

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

type AuthType int

const (
	ServerVersion = "SSH-2.0-ServerMock 1.0"
)

type Server struct {
	// ssh server config
	*ssh.ServerConfig

	// server listen address
	listenAddr string

	// opened listener
	listener net.Listener

	// server privateKey
	privateKey *rsa.PrivateKey

	quit chan struct{}
	wg   sync.WaitGroup

	// keys for authorize clients
	authorizedKeys    []ssh.PublicKey
	authorizedKeysMap map[string]struct{}

	*MockData

	servedConnections []*Connection
}

func NewMockedServer() (server *Server) {
	privateKey, _ := generateRSAKey(2048)
	signer, _ := ssh.NewSignerFromKey(privateKey)
	server = NewServer("localhost:0", signer)
	server.privateKey = privateKey
	return
}

func NewServer(listenAddr string, serverKey ssh.Signer) (server *Server) {
	server = &Server{
		ServerConfig:      &ssh.ServerConfig{ServerVersion: ServerVersion},
		authorizedKeysMap: make(map[string]struct{}),
		listenAddr:        listenAddr,
		MockData:          NewMockData(),
		quit:              make(chan struct{}),
	}

	server.ServerConfig.PublicKeyCallback = func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
		if _, ok := server.authorizedKeysMap[string(pubKey.Marshal())]; ok {
			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(pubKey),
				},
			}, nil
		}
		return nil, fmt.Errorf("unknown public key for %q", c.User())
	}

	server.ServerConfig.AddHostKey(serverKey)
	return
}

func (s *Server) ServedConnections() []*Connection {
	return s.servedConnections
}

func (s *Server) AllowClientNoAuth() {
	s.ServerConfig.NoClientAuth = true
}

func (s *Server) AddAuthorizedKey(key ssh.PublicKey) {
	debugf("added client authorized key: '%s'", key.Type())
	s.authorizedKeys = append(s.authorizedKeys, key)
	s.authorizedKeysMap[string(key.Marshal())] = struct{}{}
}

func (s *Server) parseAssressPort(addressString string) (host string, port uint16, err error) {
	parts := strings.SplitN(addressString, ":", 2)
	if len(parts) < 2 {
		return host, port, fmt.Errorf("wrong address string: no one ':' found in %s", addressString)
	}
	host = parts[0]
	portI, err := strconv.Atoi(parts[1])
	if err != nil {
		return host, port, fmt.Errorf("wrong address string: %s", err)
	}
	port = uint16(portI)
	return
}

func (s *Server) Start() (address string, port uint16, err error) {
	s.listener, err = net.Listen("tcp", s.listenAddr)
	if err != nil {
		return
	}

	debugf("server '%s' started on: '%s'", s.ServerVersion, s.listener.Addr().String())

	s.wg.Add(1)
	go s.serve()
	return s.parseAssressPort(s.listener.Addr().String())
}

func (s *Server) serve() {
	defer s.wg.Done()
	for {
		netConn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Fatalf("failed to accept incoming connection: %T", err)
			}
		}

		debugf("accepted new connection from %s", netConn.RemoteAddr().String())
		conn := NewConnection(netConn)
		s.servedConnections = append(s.servedConnections, conn)
		conn.mockData = s.MockData

		s.wg.Add(1)
		go func() {
			conn.handle(s.ServerConfig)
			s.wg.Done()
		}()
	}
}

func (s *Server) Stop() {
	debug("Stopping server ...")
	close(s.quit)
	_ = s.listener.Close()
	s.wg.Wait()
	debug("Server stopped ...")
}

func (s *Server) Wait() {
	<-s.quit
}
