package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

var socketPath = filepath.Join(os.TempDir() + "gwsn.sock")

/*
var (
	_ CtlTransport = (*SocketHost)(nil)
	_ CtlTransport = (*SocketClient)(nil)
)
*/

type SocketHostCfg struct {
	SocketPath string
}

type SocketHost struct {
	cfg SocketHostCfg

	cSend    chan []byte
	cReceive chan []byte
}

func NewSocketHost(cfg SocketHostCfg) *SocketHost {
	return &SocketHost{
		cfg: cfg,
	}
}

func (h *SocketHost) Listen(ctx context.Context) error {
	err := os.Remove(h.cfg.SocketPath)
	if err != nil {
		slog.Warn("error while removing file, but still continuing", "file", h.cfg.SocketPath, "error", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to open socket %s: %v", socketPath, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	_ = g

	for {
		conn, err := listener.Accept()

		if err != nil {
			slog.Warn("error while accepting connectiong, but still continuing", "error", err)
		} else {
			slog.Debug("accepted new connection", "local_address", conn.LocalAddr(), "remote_address", conn.RemoteAddr())

			/*
				// Receive
				g.Go(func() error {
					// TODO: Set deadline to periodically check ctx status
					conn.Read()
					// TODO: Handle message
					//h.cReceive <- msg
				})

				// Send
				g.Go(func() error {
					for {
						select {
						case msg := <-h.cSend:
							s, err := json.Marshal(msg)
							if err != nil {
								slog.Warn("failed to marshal message, discarding and continuing", "message_type", msg.Type, "error", err)
							}

							n, err := conn.Write(s)
							if err != nil {
								cerr := conn.Close()
								return fmt.Errorf("failed to write message to connection, closing connection: %v", errors.Join(err, cerr))
							}

							slog.Debug("wrote message to connection", "message_type", msg.Type, "num_bytes_written", n)
						case <-ctx.Done():
							return nil
						}
					}
				})
			*/
		}
	}

	return nil
}

/*
type SocketClient struct {
}

func NewSocketClient() *SocketClient {
	return &SocketClient{}
}

// TODO: Socket client listen

*/
