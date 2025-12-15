package transport_test

import (
	"context"
	"sync"
	"testing"

	"github.com/link00000000/gwsn/internal/ctl/transport"
)

func TestInProcessConnectSendReceive(t *testing.T) {
	// configure transport
	t1 := transport.NewInProcessTransport()
	t2 := transport.NewInProcessTransport()
	t1.SetPeer(t2)
	t2.SetPeer(t1)

	// start listening
	ctx, cancel := context.WithCancel(t.Context())
	var listenWg sync.WaitGroup

	defer listenWg.Wait()
	defer cancel()

	listenWg.Go(func() {
		err := t1.Listen(ctx)
		if err != nil {
			t.Errorf("peer 1 Listen returned an error: %v", err)
		}
	})

	listenWg.Go(func() {
		err := t2.Listen(ctx)
		if err != nil {
			t.Errorf("peer 2 Listen returned an error: %v", err)
		}
	})

	// send / recv message
	msgSend := &transport.Msg{0xa0, 0xb0, 0xc0, 0xd0}
	cMsgRecv := make(chan *transport.Msg, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		err := t1.Send(ctx, msgSend)
		if err != nil {
			t.Errorf("peer 1 failed to send message: %v", err)
		}
	})

	wg.Go(func() {
		msgRecv, err := t2.Recv(ctx)
		if err != nil {
			t.Errorf("peer 2 failed to receive message: %v", err)
		}

		cMsgRecv <- msgRecv
	})

	msgRecv := <-cMsgRecv
	wg.Wait()

	// check message
	if len(*msgSend) != len(*msgRecv) {
		t.Errorf("message received by peer 2 is not the same length as the message send by peer 1. expected %d bytes, received %d bytes", len(*msgSend), len(*msgRecv))
	}

	for n, b := range *msgRecv {
		c := (*msgSend)[n]
		if b != c {
			t.Errorf("message received by peer 2 is not identical to the message send by peer 1. expected byte %d to be 0x%02x, received 0x%02x", n, b, c)
		}
	}
}
