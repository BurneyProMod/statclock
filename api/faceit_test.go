package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testPlayerID = "11111111-2222-3333-4444-555555555555"

// newTestClient returns a Client pointed at a fake FACEIT server that runs handler.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := NewClient("test-key")
	c.BaseURL = srv.URL
	c.RetryDelay = time.Millisecond // keep retry tests fast
	return c
}

// serveFile writes a file from testdata/ as the response body.
func serveFile(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	w.Write(body)
}

func TestGetPlayerByNickname(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/players" {
			t.Errorf("path = %q, want /players", r.URL.Path)
		}
		if got := r.URL.Query().Get("nickname"); got != "test-player" {
			t.Errorf("nickname = %q, want test-player", got)
		}
		if got := r.URL.Query().Get("game"); got != "cs2" {
			t.Errorf("game = %q, want cs2", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}
		serveFile(t, w, "player.json")
	})

	p, err := c.GetPlayerByNickname(context.Background(), "test-player", "cs2")
	if err != nil {
		t.Fatal(err)
	}

	if p.PlayerID != testPlayerID {
		t.Errorf("PlayerID = %q, want %q", p.PlayerID, testPlayerID)
	}
	if p.SteamID64 != "76561190000000000" {
		t.Errorf("SteamID64 = %q", p.SteamID64)
	}
	wantActivated := time.Date(2018, 8, 25, 2, 2, 58, 379_000_000, time.UTC)
	if !p.ActivatedAt.Equal(wantActivated) {
		t.Errorf("ActivatedAt = %v, want %v", p.ActivatedAt, wantActivated)
	}
	for game, want := range map[string]int{"cs2": 2009, "csgo": 1850, "tf2": 0} {
		if got := p.Elo(game); got != want {
			t.Errorf("Elo(%q) = %d, want %d", game, got, want)
		}
	}
}

func TestGetPlayerStats(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		want := "/players/" + testPlayerID + "/stats/cs2"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		serveFile(t, w, "stats.json")
	})

	s, err := c.GetPlayerStats(context.Background(), testPlayerID, "cs2")
	if err != nil {
		t.Fatal(err)
	}

	matches, err := s.LifetimeInt("Matches")
	if err != nil || matches != 1687 {
		t.Errorf("Matches = %d, %v; want 1687", matches, err)
	}
	wins, err := s.LifetimeInt("Wins")
	if err != nil || wins != 849 {
		t.Errorf("Wins = %d, %v; want 849", wins, err)
	}
}

func TestLifetimeInt(t *testing.T) {
	stats := &PlayerStats{Lifetime: map[string]any{
		"Matches":        "1687",
		"Wins":           849.0, // a JSON number rather than a string
		"K/D Ratio":      "1.05",
		"Recent Results": []any{"1", "0"},
	}}

	tests := []struct {
		name    string
		keys    []string
		want    int
		wantErr bool
	}{
		{"numeric string", []string{"Matches"}, 1687, false},
		{"case-insensitive", []string{"matches"}, 1687, false},
		{"falls back to later key", []string{"Total Matches", "Matches"}, 1687, false},
		{"JSON number", []string{"Wins"}, 849, false},
		{"missing key", []string{"Losses"}, 0, true},
		{"not an integer", []string{"K/D Ratio"}, 0, true},
		{"not a number", []string{"Recent Results"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stats.LifetimeInt(tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		status    int
		wantErr   error // named error it should match, or nil
		wantCalls int32 // 1 = no retry, maxAttempts = retried
	}{
		{http.StatusBadRequest, nil, 1},
		{http.StatusUnauthorized, ErrUnauthorized, 1},
		{http.StatusForbidden, ErrUnauthorized, 1},
		{http.StatusNotFound, ErrNotFound, 1},
		{http.StatusTooManyRequests, ErrRateLimited, maxAttempts},
		{http.StatusInternalServerError, nil, maxAttempts},
		{http.StatusServiceUnavailable, nil, maxAttempts},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			var calls atomic.Int32
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tt.status)
			})

			_, err := c.GetPlayerByNickname(context.Background(), "test-player", "cs2")

			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tt.status {
				t.Fatalf("err = %v, want APIError with status %d", err, tt.status)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("errors.Is(err, %v) = false", tt.wantErr)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Errorf("calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestRetryThenSuccess(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			w.WriteHeader(http.StatusTooManyRequests)
		case 2:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			serveFile(t, w, "player.json")
		}
	})

	p, err := c.GetPlayerByNickname(context.Background(), "test-player", "cs2")
	if err != nil {
		t.Fatal(err)
	}
	if p.Elo("cs2") != 2009 {
		t.Errorf("Elo = %d, want 2009", p.Elo("cs2"))
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryAfterHeader(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	// getOnce makes a single request, so this doesn't wait 7 seconds.
	err := c.getOnce(context.Background(), "/players", &Player{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.RetryAfter != 7*time.Second {
		t.Fatalf("err = %v, want APIError with RetryAfter 7s", err)
	}
}

func TestCancelStopsRetrying(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	c.RetryDelay = time.Hour // only the context can end the wait

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.GetPlayerByNickname(ctx, "test-player", "cs2")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestBadJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{not json`))
	})

	_, err := c.GetPlayerByNickname(context.Background(), "test-player", "cs2")
	if err == nil || !strings.Contains(err.Error(), "decode response") {
		t.Errorf("err = %v, want a decode response error", err)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	tests := []struct {
		err  *APIError
		want string
	}{
		{&APIError{StatusCode: 401}, "unauthorized, check FACEIT_API_KEY (http 401)"},
		{&APIError{StatusCode: 404, Body: `{"errors":[]}`}, `not found (http 404): {"errors":[]}`},
		{&APIError{StatusCode: 500, Body: "oops"}, "http 500: oops"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}
