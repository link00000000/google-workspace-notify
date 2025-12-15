package ctl_test

import (
	"testing"

	"github.com/link00000000/gwsn/internal/ctl"
	"github.com/link00000000/gwsn/internal/ctl/transport"
	"golang.org/x/sync/errgroup"
)

func TestTODO(t *testing.T) {
	coordinator := transport.NewInProcessConnCoordinator()
	t1 := transport.NewInProcessTransport(transport.InProcessTransportConfig{Coordinator: coordinator})
	t2 := transport.NewInProcessTransport(transport.InProcessTransportConfig{Coordinator: coordinator})

	c1 := ctl.NewPeer(t1)
	c1.Listen()

	<-c1.RecvPing()
	<-c1.RecvPong()
	c1.SendPing()
	c1.SendPong()

	c2 := ctl.NewPeer(t2)
	c2.Listen()

	<-c2.RecvPing()
	<-c2.RecvPong()
	c2.SendPing()
	c2.SendPong()
}

func TestTODO2(t *testing.T) {
	t1 := transport.NewSocketTransport(SocketTransportConfig{Mode: SocketTransportMode_Host, Socket: "/tmp/gwsn.sock"})
	t2 := transport.NewSocketTransport(SocketTransportConfig{Mode: SocketTransportMode_Client, Socket: "/tmp/gwsn.sock"})

	c1 := ctl.NewPeer(t1)
	c1.Listne()

	<-c1.RecvPing()
	<-c1.RecvPong()
	c1.SendPing()
	c1.SendPong()

	c2 := ctl.NewPeer(t2)
	c2.Listen()

	<-c2.RecvPing()
	<-c2.RecvPong()
	c2.SendPing()
	c2.SendPong()
}
