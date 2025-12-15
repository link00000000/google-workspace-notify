package ctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type CtlMsgType string

const (
	CtlMsgType_Ping = "ping"
	CtlMsgType_Pong = "pong"
)

var ErrUnsupportedMsg = errors.New("unsupported message")

type Msg struct {
	Type  string
	Bytes []byte
}

type Transport interface {
	Listen(ctx context.Context) error
	Close(ctx context.Context) error

	Send(ctx context.Context, msg *Msg) error
	Recv(ctx context.Context) (*Msg, error)
}

type CtlTransport interface {
	Listen() error

	SendChan() chan<- Msg
	ReceiveChan() <-chan Msg
}

type PingMsg struct{}
type PongMsg struct{}

type CtlPeer struct {
	t    CtlTransport
	wg   sync.WaitGroup
	done chan struct{}

	cStatusRequest  chan *PingMsg
	cStatusResponse chan *PongMsg
}

func NewCtlPeer(transport CtlTransport) *CtlPeer {
	return &CtlPeer{
		t:    transport,
		wg:   sync.WaitGroup{},
		done: make(chan struct{}),

		cStatusRequest: make(chan *PingMsg),
	}
}

func (p *CtlPeer) Listen() error {
	err := p.t.Listen()
	if err != nil {
		return fmt.Errorf("failed to start ctl transport: %v", err)
	}

	p.wg.Go(func() {
		for {
			select {
			case msg := <-p.t.ReceiveChan():
				err := p.handleMsg(&msg)
				if err != nil {
					slog.Error("failed to handle message", "message_type", msg.Type)
				}
			case <-p.done:
				return
			}
		}
	})

	return nil
}

func (p *CtlPeer) Close() error {
	close(p.done)
	p.wg.Wait()

	err := p.t.Close()
	if err != nil {
		return fmt.Errorf("error while stopping ctl transport: %v", err)
	}

	return nil
}

func (p *CtlPeer) handleMsg(msg *Msg) error {
	switch msg.Type {
	case CtlMsgType_Ping:
		m := &PingMsg{}
		err := json.Unmarshal(msg.Bytes, m)
		if err != nil {
			return fmt.Errorf("failed to parse message: %v", err)
		} else {
			p.cStatusRequest <- m
		}
	case CtlMsgType_Pong:
		m := &PongMsg{}
		err := json.Unmarshal(msg.Bytes, m)
		if err != nil {
			return fmt.Errorf("failed to parse message: %v", err)
		} else {
			p.cStatusResponse <- m
		}
	default:
		return ErrUnsupportedMsg
	}

	return nil
}
