package gworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	credentialsFilePath = "credentials.json"
	tokenFilePath       = "token.json"
)

type HttpClient struct {
	*http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		Client: &http.Client{},
	}
}

func (c *HttpClient) Configure(ctx context.Context, scopes ...string) error {
	b, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("error while reading credentials files (%s): %v", credentialsFilePath, err)
	}

	cfg, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return fmt.Errorf("error while configuring oauth: %v", err)
	}

	tok, err := getToken(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get oauth token: %v", err)
	}

	c.Client = cfg.Client(ctx, tok)

	return nil
}

func getToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	tok, err := getCachedToken()
	if err != nil {
		slog.Warn("failed to get cached token", "file", tokenFilePath, "error", err)

		tok, err = getTokenFromWeb(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get token new token: %v", err)
		}

		err = setCachedToken(tok)
		if err != nil {
			// This error is okay because we will just get a new token next time
			slog.Error("failed to set cached token", "error", err)
		}
	}
	return tok, nil
}

func getCachedToken() (*oauth2.Token, error) {
	slog.Debug("getting cached token from file", "file", tokenFilePath)

	f, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file (%s): %v", tokenFilePath, err)
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token file (%s): %v", tokenFilePath, err)
	}

	return tok, nil
}

func setCachedToken(token *oauth2.Token) error {
	slog.Debug("caching token in file", "file", tokenFilePath)

	f, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open token file (%s): %v", tokenFilePath, err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("failed to write token file (%s): %v", tokenFilePath, err)
	}

	return nil
}

func getTokenFromWeb(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	slog.Info("getting new token from the web")

	stateToken := "state-token" // TODO: Generate proper state token
	authUrl := cfg.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)

	// TODO: Automatically open browser or web view
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authUrl)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("failed to read auth code: %v", err)
	}

	tok, err := cfg.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange for auth token: %v", err)
	}

	return tok, nil
}
