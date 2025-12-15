package transport_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/link00000000/gwsn/internal/ctl/transport"
)

func TestListenToSocket(t *testing.T) {
	cfg := transport.SocketHostCfg{
		SocketPath: filepath.Join(t.TempDir(), "gwsn.sock"),
	}
	h := transport.NewSocketHost(cfg)

	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*200)
	defer cancel()

	err := h.Listen(ctx)
	if err != nil {
		t.Errorf("Listen return error: %v", err)
	}
}
