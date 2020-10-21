package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Now provides the current time, defaults to time.Now. Change if you want to control the time
var Now = time.Now

type source struct {
	ClientID     string
	ClientSecret string
	URL          string

	refreshTTL time.Duration
	httpClient http.Client
	token      *oauth2.Token
	Config     clientcredentials.Config
}

func NewTokenSource(id, secret, url string, refreshTTL time.Duration, httpClient http.Client) oauth2.TokenSource {
	return source{
		ClientID:     id,
		ClientSecret: secret,
		URL:          url,

		httpClient: httpClient,
		refreshTTL: refreshTTL,
		Config: clientcredentials.Config{
			ClientID:     id,
			ClientSecret: secret,
			TokenURL:     url,
			AuthStyle:    oauth2.AuthStyleInParams,
		},
	}
}

func (t source) Token() (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &t.httpClient)

	if t.token == nil || t.token.Expiry.Sub(Now()) < t.refreshTTL {
		token, err := t.Config.Token(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve a new token: %w", err)
		}

		t.token = token
	}

	return t.token, nil
}
