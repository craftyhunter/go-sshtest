package sshtest

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/craftyhunter/go-sshtest/protocol"
)

func NewChannel(channel ssh.NewChannel) *Channel {
	return &Channel{
		Type:       channel.ChannelType(),
		newChannel: channel,
		Stat: &ChannelStat{
			mu: sync.Mutex{},
		},
	}
}

type Channel struct {
	Type       string
	newChannel ssh.NewChannel
	mockData *MockData
	ssh.Channel
	Stat *ChannelStat
}

type ChannelStat struct {
	mu       sync.Mutex
	requests []interface{}
}

func (s *ChannelStat) appendRequest(msg interface{}) {
	s.mu.Lock()
	s.requests = append(s.requests, msg)
	s.mu.Unlock()
}

func (s *ChannelStat) Requests() []interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]interface{}, 0, len(s.requests))
	copy(result, s.requests)
	return result
}

func (ch *Channel) handle() {
	channel, requests, err := ch.newChannel.Accept()
	if err != nil {
		log.Fatalf("could not accept channel: %v", err)
	}
	ch.Channel = channel

	go ch.handleRequests(requests)

	ch.handleChannel()

}

func (ch *Channel) handleChannel() {

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
				ch.Stat.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
				return
			}
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypePTYWindowChange:
			msg = new(protocol.MsgRequestPTYWindowChange)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.Stat.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
				return
			}
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		case protocol.MsgTypeEnv:
			msg = new(protocol.MsgRequestSetEnv)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.Stat.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				return
			}

		case protocol.MsgTypeExec:
			msg = new(protocol.MsgRequestExec)
			if err := ssh.Unmarshal(request.Payload, msg); err != nil {
				ch.Stat.appendRequest(protocol.NewUnparsedMsg(request.Type, request.Payload))
				sendReplyFalse(ch.newChannel.ChannelType(), request)
				return
			}

			go func(ch *Channel, msg *protocol.MsgRequestExec) {
				defer func() {
					_ = ch.Close()
				}()
				for in, out := range ch.mockData.mockedExecRequests {
					if in == msg.Command {
						_, _ = ch.Write([]byte(out.result))
						_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(&protocol.MsgExitStatus{ExitStatus: out.exitStatus}))
						return
					}
				}
				_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(protocol.MsgExitStatus{ExitStatus: 0}))
			}(ch, msg.(*protocol.MsgRequestExec))
			sendReplyTrue(ch.newChannel.ChannelType(), request)

		//case "auth-agent-req@openssh.com":
		//	if request.WantReply {
		//		_ = request.Reply(true, nil)
		//	}

		case protocol.MsgTypeShell:
			msg = new(protocol.MsgRequestShell)
			sendReplyTrue(ch.newChannel.ChannelType(), request)
		default:
			msg = protocol.NewUnparsedMsg(request.Type, request.Payload)
			sendReplyFalse(ch.newChannel.ChannelType(), request)
		}
		ch.Stat.appendRequest(msg)
	}
}
