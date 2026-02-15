package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type playerLookupResponse struct {
	Nickname    string                    `json:"nickname"`
	PlayerID    string                    `json:"player_id"`
	ActivatedAt string                    `json:"activated_at"`
	Games       map[string]playerGameInfo `json:"games"`
}

type playerGameInfo struct {
	FaceitElo int `json:"faceit_elo"`
}

type playerStatsResponse struct {
	GameID   string                 `json:"game_id"`
	PlayerID string                 `json:"player_id"`
	Lifetime map[string]interface{} `json:"lifetime"`
}

type metric string

const (
	MetricElo          metric = "elo"
	MetricTotalMatches metric = "matches"
	MetricAccountAge   metric = "age"
	MetricWinLoss      metric = "wl"
)

func main() {
	_ = godotenv.Load()

	var (
		m      = flag.String("metric", getenvDefault("FACEIT_METRIC", string(MetricElo)), "Metric: elo | matches | age | wl")
		game   = flag.String("game", getenvDefault("FACEIT_GAME", "cs2"), "Game: cs2 | csgo | ...")
		to     = flag.Duration("timeout", 15*time.Second, "HTTP timeout")
		apiURL = flag.String("api", "https://open.faceit.com/data/v4", "FACEIT Data API base URL")
	)
	flag.Parse()

	apiKey := os.Getenv("FACEIT_API_KEY")
	if apiKey == "" {
		fatal("FACEIT_API_KEY is required")
	}

	nickname := os.Getenv("FACEIT_NAME")
	if nickname == "" {
		fatal("FACEIT_NAME is required")
	}

	ctx := context.Background()

	player, err := getPlayerByNickname(ctx, *apiURL, apiKey, nickname, *game, *to)
	if err != nil {
		// Optional fallback for some accounts
		if *game == "cs2" {
			if p2, err2 := getPlayerByNickname(ctx, *apiURL, apiKey, nickname, "csgo", *to); err2 == nil {
				player = p2
				*game = "csgo"
			} else {
				fatalErr(err)
			}
		} else {
			fatalErr(err)
		}
	}

	switch metric(strings.ToLower(strings.TrimSpace(*m))) {
	case MetricElo:
		cur := currentElo(player, *game)
		fmt.Printf("Elo: %d\n", cur)

	case MetricTotalMatches:
		stats, err := getPlayerStats(ctx, *apiURL, apiKey, player.PlayerID, *game, *to)
		if err != nil {
			fatalErr(err)
		}
		matches, err := lifetimeInt(stats.Lifetime, "Matches", "matches", "Total Matches", "total matches")
		if err != nil {
			fatalErr(err)
		}
		fmt.Printf("Total Matches: %d\n", matches)

	case MetricAccountAge:
		days, err := accountAgeDays(player.ActivatedAt, time.Now().UTC())
		if err != nil {
			fatalErr(err)
		}
		fmt.Printf("Account Age (Days): %d\n", days)

	case MetricWinLoss:
		stats, err := getPlayerStats(ctx, *apiURL, apiKey, player.PlayerID, *game, *to)
		if err != nil {
			fatalErr(err)
		}
		matches, err := lifetimeInt(stats.Lifetime, "Matches", "matches", "Total Matches", "total matches")
		if err != nil {
			fatalErr(err)
		}
		wins, err := lifetimeInt(stats.Lifetime, "Wins", "wins")
		if err != nil {
			fatalErr(err)
		}
		losses := matches - wins
		if losses < 0 {
			losses = 0
		}
		fmt.Printf("Wins/Loss: %d/%d\n", wins, losses)

	default:
		fatal("unknown -metric (use: elo | matches | age | wl)")
	}
}

// GET /players?nickname=...&game=...
func getPlayerByNickname(ctx context.Context, apiBase, apiKey, nickname, game string, timeout time.Duration) (*playerLookupResponse, error) {
	u, err := url.Parse(strings.TrimRight(apiBase, "/") + "/players")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("nickname", nickname)
	q.Set("game", game)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("player lookup http %d: %s", resp.StatusCode, string(b))
	}

	var pr playerLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	if pr.PlayerID == "" {
		return nil, fmt.Errorf("player_id missing in response for nickname=%q", nickname)
	}
	return &pr, nil
}

// GET /players/{player_id}/stats/{game_id}
func getPlayerStats(ctx context.Context, apiBase, apiKey, playerID, game string, timeout time.Duration) (*playerStatsResponse, error) {
	u := strings.TrimRight(apiBase, "/") + "/players/" + url.PathEscape(playerID) + "/stats/" + url.PathEscape(game)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("player stats http %d: %s", resp.StatusCode, string(b))
	}

	var sr playerStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

func currentElo(player *playerLookupResponse, game string) int {
	if player == nil || player.Games == nil {
		return 0
	}
	if g, ok := player.Games[game]; ok {
		return g.FaceitElo
	}
	for _, g := range player.Games {
		if g.FaceitElo > 0 {
			return g.FaceitElo
		}
	}
	return 0
}

func accountAgeDays(activatedAt string, now time.Time) (int, error) {
	s := strings.TrimSpace(activatedAt)
	if s == "" {
		return 0, fmt.Errorf("activated_at missing/empty in /players response")
	}

	t, err := parseISOTime(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse activated_at=%q: %w", s, err)
	}

	if now.Before(t) {
		return 0, fmt.Errorf("activated_at is in the future (%s)", t.UTC().Format(time.RFC3339))
	}

	d := now.Sub(t)
	return int(d.Hours() / 24), nil
}

func parseISOTime(s string) (time.Time, error) {
	// Try common FACEIT/ISO variants
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}
	var lastErr error
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}

func toInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case float32:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case json.Number:
		i, err := t.Int64()
		return int(i), err == nil
	case string:
		s := strings.TrimSpace(t)
		i, err := strconv.Atoi(s)
		return i, err == nil
	default:
		return 0, false
	}
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(2)
}

func fatalErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
