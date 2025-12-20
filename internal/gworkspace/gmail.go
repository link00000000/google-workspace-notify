package gworkspace

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

type GmailHistoryId struct {
	id      uint64
	isValid bool
}

func NewGmailHistoryId() *GmailHistoryId {
	return &GmailHistoryId{
		id:      0,
		isValid: false,
	}
}

func NewGmailHistoryIdWithValue(id uint64) *GmailHistoryId {
	return &GmailHistoryId{
		id:      id,
		isValid: true,
	}
}

func (h *GmailHistoryId) SetId(id uint64) {
	h.id = id
	h.isValid = true
}

func (h *GmailHistoryId) GetId() uint64 {
	if !h.isValid {
		panic("attempted to get id of invalid GmailHistoryId. check IsValid() before accessing")
	}

	return h.id
}

func (h *GmailHistoryId) IsValid() bool {
	return h.isValid
}

func (h *GmailHistoryId) Clear() {
	h.isValid = false
}

type GmailMessage struct {
	To      string
	From    string
	Subject string
}

type GmailMonitor struct {
	mu  sync.Mutex
	svc *gmail.Service

	isInitialized bool
	historyId     *GmailHistoryId
	updateFreq    time.Duration

	msgsChan chan []*GmailMessage
}

func NewGmailMonitor(svc *gmail.Service, updateFreq time.Duration) *GmailMonitor {
	return &GmailMonitor{
		svc: svc,

		isInitialized: false,
		historyId:     NewGmailHistoryId(),
		updateFreq:    updateFreq,

		msgsChan: make(chan []*GmailMessage, 32),
	}
}

func (g *GmailMonitor) Initialize(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	err := g.refreshHistoryId(ctx)
	if err != nil {
		return fmt.Errorf("error while fetching latest history id: %v", err)
	}

	g.isInitialized = true

	return nil
}

func (g *GmailMonitor) Watch(ctx context.Context) error {
	ticker := time.NewTicker(g.updateFreq)

	tick := func() {
		slog.Debug("GmailMonitor Watch checking for new messages")

		err := g.CheckNow(ctx)
		if err != nil {
			slog.Error("error while checking for new messages", "error", err)
		}

		slog.Debug("GmailMonitor Watch waiting before checking again", "duration", g.updateFreq)
	}

	slog.Debug("starting GmailMonitor ticker")

	// tick once before waiting
	tick()

	for {
		select {
		case <-ticker.C:
			tick()
		case <-ctx.Done():
			return nil
		}
	}
}

func (g *GmailMonitor) CheckNow(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	slog.Debug("checking for new messages")

	msgs, err := g.fetchNewMessages(ctx)

	// 404 when history id is invalid
	if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == http.StatusNotFound {
		slog.Debug("gmail responded 404 when fetching new messages. refreshing history id and trying again")

		err := g.refreshHistoryId(ctx)
		if err != nil {
			return fmt.Errorf("error while refreshing history id: %v", err)
		}

		msgs, err = g.fetchNewMessages(ctx)
		if err != nil {
			return fmt.Errorf("error while fetching new messages: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("error while fetching new messages: %v", err)
	}

	if len(msgs) > 0 {
		slog.Info("received new messages from gmail", "numMessages", len(msgs))
		for _, msg := range msgs {
			slog.Debug("new messages", "to", msg.To, "from", msg.From, "subject", msg.Subject)
		}

		select {
		case g.msgsChan <- msgs:
		case <-ctx.Done():
		}
	}

	return nil
}

func (g *GmailMonitor) Messages() <-chan []*GmailMessage {
	return g.msgsChan
}

func (g *GmailMonitor) fetchNewMessages(ctx context.Context) ([]*GmailMessage, error) {
	if !g.isInitialized || !g.historyId.IsValid() {
		panic("attempted to check for messages, but GmailMonitor was not initialized. call Initialize() first")
	}

	slog.Debug("fetching new messages from gmail")

	msgIds := make([]string, 0)

	forEachPage := func(res *gmail.ListHistoryResponse) error {
		if res.HistoryId > g.historyId.GetId() {
			slog.Debug("updating history id", "old", g.historyId.GetId(), "new", res.HistoryId)
			g.historyId.SetId(res.HistoryId)
		}

		for _, h := range res.History {
			for _, m := range h.MessagesAdded {
				msgIds = append(msgIds, m.Message.Id)
			}
		}

		return nil
	}

	err := g.svc.Users.History.List("me").
		StartHistoryId(g.historyId.GetId()).
		HistoryTypes("messageAdded").
		LabelId("INBOX").
		Pages(ctx, forEachPage)

	if err != nil {
		return []*GmailMessage{}, fmt.Errorf("error while fetching history from gmail (last history id = %d): %v", g.historyId.GetId(), err)
	}

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(16) // TODO: Make configurable

	msgs := make([]*GmailMessage, len(msgIds))

	for i, id := range msgIds {
		group.Go(func() error {
			res, err := g.svc.Users.Messages.Get("me", id).
				Context(ctx).
				Format("metadata").
				MetadataHeaders("To", "From", "Subject").
				Do()

			if err != nil {
				return fmt.Errorf("error while fetching metadata for message (message id = %s): %v", id, err)
			}

			msg := &GmailMessage{}
			for _, h := range res.Payload.Headers {
				switch h.Name {
				case "To":
					msg.To = h.Value
				case "From":
					msg.From = h.Value
				case "Subject":
					msg.Subject = h.Value
				}
			}

			msgs[i] = msg

			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		slog.Error("error whie getting message details", "error", err)
	}

	msgs = slices.DeleteFunc(msgs, func(msg *GmailMessage) bool {
		return msg == nil
	})

	return msgs, nil
}

func (g *GmailMonitor) refreshHistoryId(ctx context.Context) error {
	res, err := g.svc.Users.GetProfile("me").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("error getting profile from Gmail: %v", err)
	}

	g.historyId.SetId(res.HistoryId)

	return nil
}
