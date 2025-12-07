package main

import (
	"context"
	"net/http"

	"github.com/link00000000/google-workspace-notify/src/ui"
)

func RunHttpServer(ctx context.Context) error {
	s := &http.Server{Addr: ":8080", Handler: ui.NewHandler()}

	go s.ListenAndServe()
	<-ctx.Done()

	return s.Shutdown(context.TODO())
}
