package wc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a JWT-authenticated HTTP client for the WC2026 REST API.
type Client struct {
	base  string
	token string
	http  *http.Client
}

// New creates a new Client. base is the API root URL, token is the JWT bearer token.
func New(base, token string) *Client {
	return &Client{
		base:  base,
		token: token,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d for %s", resp.StatusCode, path)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// FetchMatches fetches matches filtered by status ("live", "upcoming", "finished").
func (c *Client) FetchMatches(ctx context.Context, status string) ([]Match, error) {
	var result []Match
	err := c.get(ctx, "/matches?status="+status, &result)
	return result, err
}

// FetchStandings fetches current group standings.
func (c *Client) FetchStandings(ctx context.Context) (map[string][]GroupRow, error) {
	result := make(map[string][]GroupRow)
	err := c.get(ctx, "/standings", &result)
	return result, err
}
