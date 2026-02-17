// Look up by faceit_id first then get the nickname by the faceit id
// the faceit id does not change.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL = "https://open.faceit.com/data/v4"
)

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// sendRequest sets auth/JSON headers, executes the request, and decodes the response directly into v.
// Errors from the API are returned with their message when possible.
func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(c.APIKey)))
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: time.Minute}
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}
	// Decode directly into the provided interface
	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

// GetPlayerByNickname demonstrates a basic API call (GET /players?nickname=&game=)
func (c *Client) GetPlayerByNickname(ctx context.Context, nickname, game string) (*PlayerDetails, error) {
	if strings.TrimSpace(nickname) == "" {
		return nil, fmt.Errorf("nickname is required")
	}

	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u, err := url.Parse(base + "/players")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("nickname", nickname)
	if strings.TrimSpace(game) != "" {
		params.Set("game", strings.TrimSpace(game))
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var res PlayerDetails
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	if strings.TrimSpace(res.PlayerID) == "" {
		return nil, fmt.Errorf("player_id missing in response for nickname=%q", nickname)
	}

	return &res, nil
}

// GetPlayer retrieves player details by player ID (GET /players/{player_id}).
func (c *Client) GetPlayer(ctx context.Context, playerID string) (*PlayerDetails, error) {
	if strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("playerID is required")
	}

	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s", base, url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerDetails
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerBans fetches bans for a player (GET /players/{player_id}/bans).
func (c *Client) GetPlayerBans(ctx context.Context, playerID string) (*PlayerBans, error) {
	if strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("playerID is required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/bans", base, url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerBans
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerStatsInRange fetches stats for a player & game (GET /players/{player_id}/games/{game_id}/stats).
func (c *Client) GetPlayerStatsInRange(ctx context.Context, playerID, gameID string) (*PlayerStatsInRange, error) {
	if strings.TrimSpace(playerID) == "" || strings.TrimSpace(gameID) == "" {
		return nil, fmt.Errorf("playerID and gameID are required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/games/%s/stats", base, url.PathEscape(playerID), url.PathEscape(gameID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerStatsInRange
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerMatches fetches match history (GET /players/{player_id}/history).
func (c *Client) GetPlayerMatches(ctx context.Context, playerID string) (*PlayerMatches, error) {
	if strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("playerID is required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/history", base, url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerMatches
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerStats fetches game stats for a player (GET /players/{player_id}/games/{game_id}/stats).
func (c *Client) GetPlayerStats(ctx context.Context, playerID, gameID string) (*PlayerStats, error) {
	if strings.TrimSpace(playerID) == "" || strings.TrimSpace(gameID) == "" {
		return nil, fmt.Errorf("playerID and gameID are required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/games/%s/stats", base, url.PathEscape(playerID), url.PathEscape(gameID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerStats
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerTeams fetches teams for a player (GET /players/{player_id}/teams).
func (c *Client) GetPlayerTeams(ctx context.Context, playerID string) (*PlayerTeams, error) {
	if strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("playerID is required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/teams", base, url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerTeams
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerTournaments fetches tournaments for a player (GET /players/{player_id}/tournaments).
func (c *Client) GetPlayerTournaments(ctx context.Context, playerID string) (*PlayerTournaments, error) {
	if strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("playerID is required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/players/%s/tournaments", base, url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerTournaments
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetPlayerGlobalRanking fetches ranking info (GET /rankings/games/{game_id}/regions/{region}/players/{player_id}).
func (c *Client) GetPlayerGlobalRanking(ctx context.Context, gameID, region, playerID string) (*PlayerGlobalRanking, error) {
	if strings.TrimSpace(gameID) == "" || strings.TrimSpace(region) == "" || strings.TrimSpace(playerID) == "" {
		return nil, fmt.Errorf("gameID, region, and playerID are required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = strings.TrimRight(baseURL, "/")
	}

	u := fmt.Sprintf("%s/rankings/games/%s/regions/%s/players/%s", base, url.PathEscape(gameID), url.PathEscape(region), url.PathEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res PlayerGlobalRanking
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Error and Success responses follow the same structure, we can make them into structs
type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type successResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}
type PlayerDetails struct {
	ActivatedAt        time.Time `json:"activated_at"`
	Avatar             string    `json:"avatar"`
	Country            string    `json:"country"`
	CoverFeaturedImage string    `json:"cover_featured_image"`
	CoverImage         string    `json:"cover_image"`
	FaceitURL          string    `json:"faceit_url"`
	FriendsIds         []string  `json:"friends_ids"`
	Games              struct {
		Property1 struct {
			FaceitElo       int    `json:"faceit_elo"`
			GamePlayerID    string `json:"game_player_id"`
			GamePlayerName  string `json:"game_player_name"`
			GameProfileID   string `json:"game_profile_id"`
			Region          string `json:"region"`
			Regions         any    `json:"regions"`
			SkillLevel      int    `json:"skill_level"`
			SkillLevelLabel string `json:"skill_level_label"`
		} `json:"property1"`
		Property2 struct {
			FaceitElo       int    `json:"faceit_elo"`
			GamePlayerID    string `json:"game_player_id"`
			GamePlayerName  string `json:"game_player_name"`
			GameProfileID   string `json:"game_profile_id"`
			Region          string `json:"region"`
			Regions         any    `json:"regions"`
			SkillLevel      int    `json:"skill_level"`
			SkillLevelLabel string `json:"skill_level_label"`
		} `json:"property2"`
	} `json:"games"`
	Infractions    any      `json:"infractions"`
	MembershipType string   `json:"membership_type"`
	Memberships    []string `json:"memberships"`
	NewSteamID     string   `json:"new_steam_id"`
	Nickname       string   `json:"nickname"`
	Platforms      struct {
		Property1 string `json:"property1"`
		Property2 string `json:"property2"`
	} `json:"platforms"`
	PlayerID string `json:"player_id"`
	Settings struct {
		Language string `json:"language"`
	} `json:"settings"`
	SteamID64     string `json:"steam_id_64"`
	SteamNickname string `json:"steam_nickname"`
	Verified      bool   `json:"verified"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/bans
type PlayerBans struct {
	End   int `json:"end"`
	Items []struct {
		EndsAt   time.Time `json:"ends_at"`
		Game     string    `json:"game"`
		Nickname string    `json:"nickname"`
		Reason   string    `json:"reason"`
		StartsAt time.Time `json:"starts_at"`
		Type     string    `json:"type"`
		UserID   string    `json:"user_id"`
	} `json:"items"`
	Start int `json:"start"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/games/{game_id}/stats
type PlayerStatsInRange struct {
	End   int `json:"end"`
	Items []struct {
		Stats struct {
			Property1 any `json:"property1"`
			Property2 any `json:"property2"`
		} `json:"stats"`
	} `json:"items"`
	Start int `json:"start"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/history
type PlayerMatches struct {
	End   int `json:"end"`
	From  int `json:"from"`
	Items []struct {
		CompetitionID   string   `json:"competition_id"`
		CompetitionName string   `json:"competition_name"`
		CompetitionType string   `json:"competition_type"`
		FaceitURL       string   `json:"faceit_url"`
		FinishedAt      int      `json:"finished_at"`
		GameID          string   `json:"game_id"`
		GameMode        string   `json:"game_mode"`
		MatchID         string   `json:"match_id"`
		MatchType       string   `json:"match_type"`
		MaxPlayers      int      `json:"max_players"`
		OrganizerID     string   `json:"organizer_id"`
		PlayingPlayers  []string `json:"playing_players"`
		Region          string   `json:"region"`
		Results         struct {
			Score struct {
				Property1 int `json:"property1"`
				Property2 int `json:"property2"`
			} `json:"score"`
			Winner string `json:"winner"`
		} `json:"results"`
		StartedAt int    `json:"started_at"`
		Status    string `json:"status"`
		Teams     struct {
			Property1 struct {
				Avatar   string `json:"avatar"`
				Nickname string `json:"nickname"`
				Players  []struct {
					Avatar         string `json:"avatar"`
					FaceitURL      string `json:"faceit_url"`
					GamePlayerID   string `json:"game_player_id"`
					GamePlayerName string `json:"game_player_name"`
					Nickname       string `json:"nickname"`
					PlayerID       string `json:"player_id"`
					SkillLevel     int    `json:"skill_level"`
				} `json:"players"`
				TeamID string `json:"team_id"`
				Type   string `json:"type"`
			} `json:"property1"`
			Property2 struct {
				Avatar   string `json:"avatar"`
				Nickname string `json:"nickname"`
				Players  []struct {
					Avatar         string `json:"avatar"`
					FaceitURL      string `json:"faceit_url"`
					GamePlayerID   string `json:"game_player_id"`
					GamePlayerName string `json:"game_player_name"`
					Nickname       string `json:"nickname"`
					PlayerID       string `json:"player_id"`
					SkillLevel     int    `json:"skill_level"`
				} `json:"players"`
				TeamID string `json:"team_id"`
				Type   string `json:"type"`
			} `json:"property2"`
		} `json:"teams"`
		TeamsSize int `json:"teams_size"`
	} `json:"items"`
	Start int `json:"start"`
	To    int `json:"to"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/games/{game_id}/stats
type PlayerStats struct {
	End   int `json:"end"`
	Items []struct {
		Avatar          string `json:"avatar"`
		BackgroundImage string `json:"background_image"`
		ChatRoomID      string `json:"chat_room_id"`
		CoverImage      string `json:"cover_image"`
		Description     string `json:"description"`
		FaceitURL       string `json:"faceit_url"`
		GameData        struct {
			Assets struct {
				Cover        string `json:"cover"`
				FeaturedImgL string `json:"featured_img_l"`
				FeaturedImgM string `json:"featured_img_m"`
				FeaturedImgS string `json:"featured_img_s"`
				FlagImgIcon  string `json:"flag_img_icon"`
				FlagImgL     string `json:"flag_img_l"`
				FlagImgM     string `json:"flag_img_m"`
				FlagImgS     string `json:"flag_img_s"`
				LandingPage  string `json:"landing_page"`
			} `json:"assets"`
			GameID       string   `json:"game_id"`
			LongLabel    string   `json:"long_label"`
			Order        int      `json:"order"`
			ParentGameID string   `json:"parent_game_id"`
			Platforms    []string `json:"platforms"`
			Regions      []string `json:"regions"`
			ShortLabel   string   `json:"short_label"`
		} `json:"game_data"`
		GameID         string `json:"game_id"`
		HubID          string `json:"hub_id"`
		JoinPermission string `json:"join_permission"`
		MaxSkillLevel  int    `json:"max_skill_level"`
		MinSkillLevel  int    `json:"min_skill_level"`
		Name           string `json:"name"`
		OrganizerData  struct {
			Avatar         string `json:"avatar"`
			Cover          string `json:"cover"`
			Description    string `json:"description"`
			Facebook       string `json:"facebook"`
			FaceitURL      string `json:"faceit_url"`
			FollowersCount int    `json:"followers_count"`
			Name           string `json:"name"`
			OrganizerID    string `json:"organizer_id"`
			Twitch         string `json:"twitch"`
			Twitter        string `json:"twitter"`
			Type           string `json:"type"`
			Vk             string `json:"vk"`
			Website        string `json:"website"`
			Youtube        string `json:"youtube"`
		} `json:"organizer_data"`
		OrganizerID   string `json:"organizer_id"`
		PlayersJoined int    `json:"players_joined"`
		Region        string `json:"region"`
		RuleID        string `json:"rule_id"`
	} `json:"items"`
	Start int `json:"start"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/teams
type PlayerTeams struct {
	End   int `json:"end"`
	Items []struct {
		Avatar      string `json:"avatar"`
		ChatRoomID  string `json:"chat_room_id"`
		CoverImage  string `json:"cover_image"`
		Description string `json:"description"`
		Facebook    string `json:"facebook"`
		FaceitURL   string `json:"faceit_url"`
		Game        string `json:"game"`
		Leader      string `json:"leader"`
		Members     []struct {
			Avatar         string   `json:"avatar"`
			Country        string   `json:"country"`
			FaceitURL      string   `json:"faceit_url"`
			MembershipType string   `json:"membership_type"`
			Memberships    []string `json:"memberships"`
			Nickname       string   `json:"nickname"`
			SkillLevel     int      `json:"skill_level"`
			UserID         string   `json:"user_id"`
		} `json:"members"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		TeamID   string `json:"team_id"`
		TeamType string `json:"team_type"`
		Twitter  string `json:"twitter"`
		Website  string `json:"website"`
		Youtube  string `json:"youtube"`
	} `json:"items"`
	Start int `json:"start"`
}

// GET https://open.faceit.com/data/v4/players/{player_id}/tournaments
type PlayerTournaments struct {
	End   int `json:"end"`
	Items []struct {
		AnticheatRequired           bool     `json:"anticheat_required"`
		Custom                      bool     `json:"custom"`
		FaceitURL                   string   `json:"faceit_url"`
		FeaturedImage               string   `json:"featured_image"`
		GameID                      string   `json:"game_id"`
		InviteType                  string   `json:"invite_type"`
		MatchType                   string   `json:"match_type"`
		MaxSkill                    int      `json:"max_skill"`
		MembershipType              string   `json:"membership_type"`
		MinSkill                    int      `json:"min_skill"`
		Name                        string   `json:"name"`
		NumberOfPlayers             int      `json:"number_of_players"`
		NumberOfPlayersCheckedin    int      `json:"number_of_players_checkedin"`
		NumberOfPlayersJoined       int      `json:"number_of_players_joined"`
		NumberOfPlayersParticipants int      `json:"number_of_players_participants"`
		OrganizerID                 string   `json:"organizer_id"`
		PrizeType                   string   `json:"prize_type"`
		Region                      string   `json:"region"`
		StartedAt                   int      `json:"started_at"`
		Status                      string   `json:"status"`
		SubscriptionsCount          int      `json:"subscriptions_count"`
		TeamSize                    int      `json:"team_size"`
		TotalPrize                  any      `json:"total_prize"`
		TournamentID                string   `json:"tournament_id"`
		WhitelistCountries          []string `json:"whitelist_countries"`
	} `json:"items"`
	Start int `json:"start"`
}

// GET https://open.faceit.com/data/v4/rankings/games/{game_id}/regions/{region}/players/{player_id}
type PlayerGlobalRanking struct {
	End   int `json:"end"`
	Items []struct {
		Country        string `json:"country"`
		FaceitElo      int    `json:"faceit_elo"`
		GameSkillLevel int    `json:"game_skill_level"`
		Nickname       string `json:"nickname"`
		PlayerID       string `json:"player_id"`
		Position       int    `json:"position"`
	} `json:"items"`
	Position int `json:"position"`
	Start    int `json:"start"`
}
