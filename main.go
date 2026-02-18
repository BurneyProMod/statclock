package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"statclock/api"
	"statclock/db"

	"github.com/joho/godotenv"
)

// Player stores the player information from the API
type Player struct {
	PlayerID  string
	Nickname  string
	SteamID   string
	FaceitElo int
}

// TranslateToPlayerID converts a FACEIT nickname to a player ID
func TranslateToPlayerID(client *api.Client, ctx context.Context, nickname, game string) string {
	player, err := client.GetPlayerByNickname(ctx, nickname, game)
	if err != nil {
		log.Fatalf("Error translating nickname to player ID: %v", err)
	}
	return player.PlayerID
}

// GetPlayer fetches player details and returns a Player object
func GetPlayer(client *api.Client, ctx context.Context, nickname, game string) *Player {
	playerDetails, err := client.GetPlayerByNickname(ctx, nickname, game)
	if err != nil {
		log.Fatalf("Error fetching player details: %v", err)
	}
	player := &Player{
		PlayerID: playerDetails.PlayerID,
		Nickname: playerDetails.Nickname,
		SteamID:  playerDetails.SteamID64,
	}

	return player
}

func LoadEnv() (apiKey string, game string) {
	// Load .env file if it exists
	_ = godotenv.Load()
	// Get API key from environment
	apiKey = os.Getenv("FACEIT_API_KEY")
	if apiKey == "" {
		log.Fatal("FACEIT_API_KEY environment variable not set")
	}
	// Get game from environment, default to cs2
	game = os.Getenv("FACEIT_GAME")
	if game == "" {
		game = "cs2"
	}
	return apiKey, game
}

func main() {
	apiKey, game := LoadEnv()

	// Prompt for nickname
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter FACEIT nickname: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	nickname := strings.TrimSpace(input)
	if nickname == "" {
		log.Fatal("Nickname cannot be empty")
	}

	// Create Faceit API client
	client := api.NewClient(apiKey)
	ctx := context.Background()

	// Get player object
	player := GetPlayer(client, ctx, nickname, game)
	// Open database
	database := db.OpenDB("statclock.db")
	defer database.Close()

	// Save player to database
	database.SavePlayer(db.Player{
		PlayerID: player.PlayerID,
		Nickname: player.Nickname,
		SteamID:  player.SteamID,
	})

	// Display player information
	fmt.Printf("Player ID: %s\n", player.PlayerID)
	fmt.Printf("Nickname: %s\n", player.Nickname)
	fmt.Printf("Steam ID: %s\n", player.SteamID)
}
