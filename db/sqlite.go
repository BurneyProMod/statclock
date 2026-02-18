package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB connection
type DB struct {
	conn *sql.DB
}

// OpenDB opens the SQLite database file
func OpenDB(path string) *DB {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	d := &DB{conn: conn}
	d.CreateTable()
	return d
}

// CloseDB closes the database connection
func (d *DB) CloseDB() {
	if err := d.conn.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

// CreateTable creates a table if it doesn't exist
func (d *DB) CreateTable() {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			player_id  TEXT PRIMARY KEY,
			nickname   TEXT NOT NULL,
			steam_id   TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

// Player mirrors the Player struct in main.go for database operations
type Player struct {
	PlayerID string
	Nickname string
	SteamID  string
}

// SavePlayer inserts or replaces a player record
func (d *DB) SavePlayer(p Player) {
	_, err := d.conn.Exec(`
		INSERT INTO players (player_id, nickname, steam_id)
		VALUES (?, ?, ?)
		ON CONFLICT(player_id) DO UPDATE SET
			nickname = excluded.nickname,
			steam_id = excluded.steam_id;
	`, p.PlayerID, p.Nickname, p.SteamID)
	if err != nil {
		log.Fatalf("Error saving player: %v", err)
	}
}

// LoadPlayer retrieves a player by player_id, returns nil if not found
func (d *DB) LoadPlayer(playerID string) *Player {
	row := d.conn.QueryRow(`
		SELECT player_id, nickname, steam_id FROM players WHERE player_id = ?;
	`, playerID)

	var p Player
	if err := row.Scan(&p.PlayerID, &p.Nickname, &p.SteamID); err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		log.Fatalf("Error getting player: %v", err)
	}
	return &p
}
