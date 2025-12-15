package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/link00000000/gwsn/internal/ctl/transport"
)

type Processor struct {
	t transport.Transport
}

func NewProcessor(t transport.Transport) *Processor {
	return &Processor{t: t}
}

func (p *Processor) Ping(ctx context.Context) error {
	msg, err := makeTransportMsg(CmdType_Ping, PingPayload{})
	if err != nil {
		return fmt.Errorf("failed to make transport message: %v", err)
	}

	err = p.t.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message via transport: %v", err)
	}

	return nil
}

func (p *Processor) RecvPing(ctx context.Context) (*PingPayload, error) {

}

func makeTransportMsg[T any](cmdType CmdType, payload T) (*transport.Msg, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	buf, err = json.Marshal(Msg{CmdType: cmdType, Payload: buf})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := transport.Msg(buf)
	return &msg, nil
}
