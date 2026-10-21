// Command stattrak prints one FACEIT stat for a player.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/BurneyProMod/stattrak-esp32/api"
	"github.com/BurneyProMod/stattrak-esp32/db"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	metric := flag.String("metric", getenvDefault("FACEIT_METRIC", "elo"), "stat to print: elo | matches | age | wl")
	game := flag.String("game", getenvDefault("FACEIT_GAME", "cs2"), "game: cs2 | csgo")
	flag.Parse()

	apiKey := os.Getenv("FACEIT_API_KEY")
	if apiKey == "" {
		fatal("FACEIT_API_KEY is required")
	}
	nickname := os.Getenv("FACEIT_NICKNAME")
	if nickname == "" {
		fatal("FACEIT_NICKNAME is required")
	}

	ctx := context.Background()
	client := api.NewClient(apiKey)

	player, err := client.GetPlayerByNickname(ctx, nickname, *game)
	if errors.Is(err, api.ErrNotFound) && *game == "cs2" {
		// Some older accounts are only registered under csgo.
		fmt.Fprintf(os.Stderr, "note: no cs2 profile for %q, trying csgo\n", nickname)
		*game = "csgo"
		player, err = client.GetPlayerByNickname(ctx, nickname, *game)
	}
	if err != nil {
		fatal(err.Error())
	}

	database := db.OpenDB("stattrak.db")
	defer database.Close()
	database.SavePlayer(db.Player{
		PlayerID: player.PlayerID,
		Nickname: player.Nickname,
		SteamID:  player.SteamID64,
	})

	switch *metric {
	case "elo":
		fmt.Printf("Elo: %d\n", player.Elo(*game))

	case "age":
		days := int(time.Since(player.ActivatedAt).Hours() / 24)
		fmt.Printf("Account Age (Days): %d\n", days)

	case "matches":
		stats := getStats(ctx, client, player.PlayerID, *game)
		fmt.Printf("Total Matches: %d\n", lifetimeInt(stats, "Matches", "Total Matches"))

	case "wl":
		stats := getStats(ctx, client, player.PlayerID, *game)
		matches := lifetimeInt(stats, "Matches", "Total Matches")
		wins := lifetimeInt(stats, "Wins")
		fmt.Printf("Wins/Loss: %d/%d\n", wins, max(matches-wins, 0))

	default:
		fatal("unknown -metric (use: elo | matches | age | wl)")
	}
}

func getStats(ctx context.Context, client *api.Client, playerID, game string) *api.PlayerStats {
	stats, err := client.GetPlayerStats(ctx, playerID, game)
	if err != nil {
		fatal(err.Error())
	}
	return stats
}

func lifetimeInt(stats *api.PlayerStats, keys ...string) int {
	n, err := stats.LifetimeInt(keys...)
	if err != nil {
		fatal(err.Error())
	}
	return n
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}
