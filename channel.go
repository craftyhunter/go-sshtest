package sshtest

import (
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/craftyhunter/go-sshtest/protocol"
)

func NewChannel(channel ssh.NewChannel, mockData *MockData) *Channel {
	return &Channel{
		Type:       channel.ChannelType(),
		newChannel: channel,
		mockData:   mockData,
		mu:         sync.Mutex{},
	}
}

type Channel struct {
	ssh.Channel
	Type       string
	newChannel ssh.NewChannel
	mockData   *MockData

	mu       sync.Mutex
	requests []interface{}
}

type ChannelStat struct {
}

func (s *Channel) appendRequest(msg interface{}) {
	s.mu.Lock()
	s.requests = append(s.requests, msg)
	s.mu.Unlock()
}

func (s *Channel) Requests() []interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]interface{}{}, s.requests...)
}

func (ch *Channel) handle() {
	channel, requests, err := ch.newChannel.Accept()
	if err != nil {
		debugf("could not accept channel: %v", err)
		return
	}
	ch.Channel = channel

	go ch.handleRequests(requests)
}

func sendReplyTrue(chType string, request *ssh.Request) {
	if request.WantReply {
		_ = request.Reply(true, nil)
		debugf("channel '%s' msg '%s' replied '%v' payload: '%v'", chType, request.Type, true, request.Payload)
	}
}

func sendReplyFalse(chType string, request *ssh.Request) {
	if request.WantReply {
		_ = request.Reply(false, nil)
		debugf("channel '%s' msg '%s' replied '%v' payload: '%v'", chType, request.Type, false, request.Payload)
	}
}

func (ch *Channel) handleRequests(in <-chan *ssh.Request) {
	for request := range in {
		debugf("channel '%s' msg '%s' wantReply '%v' with payload: '%v'", ch.newChannel.ChannelType(), request.Type, request.WantReply, request.Payload)
		var msg interface{}
		switch request.Type {
		case protocol.MsgTypePTYReq:
			msg = new(protocol.MsgRequestPTY)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
			}
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypePTYWindowChange:
			msg = new(protocol.MsgRequestPTYWindowChange)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
			}
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypeEnv:
			msg = new(protocol.MsgRequestSetEnv)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
			}
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypeExec:
			msg = new(protocol.MsgRequestExec)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
			}

			sendReplyTrue(ch.newChannel.ChannelType(), request)
			go func(ch *Channel, msg *protocol.MsgRequestExec) {
				defer func() {
					_ = ch.Close()
				}()
				for in, out := range ch.mockData.getMocksExecResult() {
					if in == msg.Command {
						time.Sleep(out.timeout)
						_, _ = ch.Write([]byte(out.result))
						_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(&protocol.MsgExitStatus{ExitStatus: out.exitStatus}))
						return
					}
				}
				_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(protocol.MsgExitStatus{ExitStatus: 0}))
			}(ch, msg.(*protocol.MsgRequestExec))

		case protocol.MsgTypeAuthAgent:
			msg = new(protocol.MsgRequestAuthAgent)
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypeShell:
			msg = new(protocol.MsgRequestShell)
			sendReplyTrue(ch.newChannel.ChannelType(), request)
		default:
			msg = protocol.NewUnparsedMsg(request.Type, request.Payload)
			sendReplyFalse(ch.newChannel.ChannelType(), request)
		}
		ch.appendRequest(msg)
	}
}
