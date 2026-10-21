// Package api is a small client for the FACEIT Data API v4.
// Docs: https://docs.faceit.com/api/data/
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://open.faceit.com/data/v4"
	maxAttempts    = 3
)

// Client calls the FACEIT Data API. Create one with NewClient and reuse it.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	RetryDelay time.Duration // wait before the first retry; doubles each time
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		RetryDelay: time.Second,
	}
}

// Player holds the fields we use from GET /players.
type Player struct {
	PlayerID    string              `json:"player_id"`
	Nickname    string              `json:"nickname"`
	SteamID64   string              `json:"steam_id_64"`
	ActivatedAt time.Time           `json:"activated_at"`
	Games       map[string]GameInfo `json:"games"`
}

// GameInfo is a player's rating in one game (e.g. "cs2").
type GameInfo struct {
	FaceitElo  int    `json:"faceit_elo"`
	SkillLevel int    `json:"skill_level"`
	Region     string `json:"region"`
}

// PlayerStats is GET /players/{player_id}/stats/{game_id}.
// FACEIT sends most lifetime numbers as strings, so Lifetime is kept loosely
// typed and read with LifetimeInt.
type PlayerStats struct {
	PlayerID string         `json:"player_id"`
	GameID   string         `json:"game_id"`
	Lifetime map[string]any `json:"lifetime"`
}

// GetPlayerByNickname looks up a player (GET /players?nickname=&game=).
func (c *Client) GetPlayerByNickname(ctx context.Context, nickname, game string) (*Player, error) {
	query := url.Values{"nickname": {nickname}, "game": {game}}

	var p Player
	if err := c.get(ctx, "/players?"+query.Encode(), &p); err != nil {
		return nil, fmt.Errorf("look up player %q: %w", nickname, err)
	}
	if p.PlayerID == "" {
		return nil, fmt.Errorf("look up player %q: response has no player_id", nickname)
	}
	return &p, nil
}

// GetPlayerStats fetches a player's lifetime stats for a game.
func (c *Client) GetPlayerStats(ctx context.Context, playerID, game string) (*PlayerStats, error) {
	path := "/players/" + url.PathEscape(playerID) + "/stats/" + url.PathEscape(game)

	var s PlayerStats
	if err := c.get(ctx, path, &s); err != nil {
		return nil, fmt.Errorf("get stats for %s: %w", playerID, err)
	}
	return &s, nil
}

// Elo returns the player's FACEIT Elo for game, or 0 if they haven't played it.
func (p *Player) Elo(game string) int {
	return p.Games[game].FaceitElo
}

// LifetimeInt returns the first lifetime stat matching one of keys
// (case-insensitive) as an int.
func (s *PlayerStats) LifetimeInt(keys ...string) (int, error) {
	for _, want := range keys {
		for key, value := range s.Lifetime {
			if !strings.EqualFold(key, want) {
				continue
			}
			n, ok := toInt(value)
			if !ok {
				return 0, fmt.Errorf("lifetime stat %q is not a number: %v", key, value)
			}
			return n, nil
		}
	}
	return 0, fmt.Errorf("lifetime stats have none of %v", keys)
}

// get sends an authenticated GET request and decodes the JSON response into v.
// Rate-limit (429) and server (5xx) errors are retried with a growing delay.
func (c *Client) get(ctx context.Context, path string, v any) error {
	delay := c.RetryDelay
	for attempt := 1; ; attempt++ {
		err := c.getOnce(ctx, path, v)

		var apiErr *APIError
		retryable := errors.As(err, &apiErr) &&
			(apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500)
		if !retryable || attempt == maxAttempts {
			return err
		}

		wait := delay
		if apiErr.RetryAfter > 0 {
			wait = apiErr.RetryAfter
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay *= 2
	}
}

// getOnce makes a single request with no retries.
func (c *Client) getOnce(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		retryAfter, _ := strconv.Atoi(res.Header.Get("Retry-After"))
		return &APIError{
			StatusCode: res.StatusCode,
			Body:       strings.TrimSpace(string(body)),
			RetryAfter: time.Duration(retryAfter) * time.Second,
		}
	}
	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// toInt converts a decoded JSON number or numeric string to an int.
func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		return n, err == nil
	}
	return 0, false
}
