// Command stattrak prints one FACEIT stat for a player, or with -device keeps
// the LED display updated with their Elo.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurneyProMod/stattrak-esp32/api"
	"github.com/BurneyProMod/stattrak-esp32/db"
	"github.com/BurneyProMod/stattrak-esp32/display"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	metric := flag.String("metric", getenvDefault("FACEIT_METRIC", "elo"), "stat to print: elo | matches | age | wl")
	game := flag.String("game", getenvDefault("FACEIT_GAME", "cs2"), "game: cs2 | csgo")
	device := flag.String("device", "", "display URL, e.g. http://faceit-clock.local (keeps it updated with the Elo)")
	interval := flag.Duration("interval", time.Minute, "how often to update the display, with -device")
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

	if *device != "" {
		if *interval < 10*time.Second {
			fatal("-interval must be at least 10s to stay under FACEIT's rate limits")
		}
		watch(ctx, client, nickname, *game, *device, *interval)
		return
	}

	player, err := findPlayer(ctx, client, nickname, game)
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

// watch sends the player's Elo to the display every interval until the program
// is stopped. Errors are logged and tried again on the next update, so a
// Wi-Fi drop or a FACEIT outage doesn't stop it.
func watch(ctx context.Context, client *api.Client, nickname, game, deviceURL string, interval time.Duration) {
	log.Printf("updating %s every %v (Ctrl+C to stop)", deviceURL, interval)
	for {
		player, err := findPlayer(ctx, client, nickname, &game)
		if err == nil {
			elo := player.Elo(game)
			if err = display.SetElo(ctx, deviceURL, elo); err == nil {
				log.Printf("sent Elo %d", elo)
			}
		}
		if err != nil {
			log.Printf("error: %v", err)
		}
		time.Sleep(interval)
	}
}

// findPlayer looks up the player. If they have no cs2 profile it tries csgo,
// since some older accounts are only registered there, and updates *game.
func findPlayer(ctx context.Context, client *api.Client, nickname string, game *string) (*api.Player, error) {
	player, err := client.GetPlayerByNickname(ctx, nickname, *game)
	if errors.Is(err, api.ErrNotFound) && *game == "cs2" {
		fmt.Fprintf(os.Stderr, "note: no cs2 profile for %q, trying csgo\n", nickname)
		*game = "csgo"
		player, err = client.GetPlayerByNickname(ctx, nickname, *game)
	}
	return player, err
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
