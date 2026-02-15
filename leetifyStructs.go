package statclock

// Since Leetify statistics are either percentages or decimals, the default is float64 rather than int
// Player Stats holds all major statistics (both gameplay and leetify) in one object

type PlayerStats struct {
	steam64ID                     string  `json:"steam64_id"`
	name                          string  `json:"name"`
	mvps                          float64 `json:"mvps"`
	preaim                        float64 `json:"preaim"`
	reactionTime                  float64 `json:"reaction_time"`
	accuracy                      float64 `json:"accuracy"`
	accuracyEnemySpotted          float64 `json:"accuracy_enemy_spotted"`
	accuracyHead                  float64 `json:"accuracy_head"`
	shotsFiredEnemySpotted        float64 `json:"shots_fired_enemy_spotted"`
	shotsFired                    float64 `json:"shots_fired"`
	shotsHitEnemySpotted          float64 `json:"shots_hit_enemy_spotted"`
	shotsHitFriend                float64 `json:"shots_hit_friend"`
	shotsHitFriendHead            float64 `json:"shots_hit_friend_head"`
	shotsHitFoe                   float64 `json:"shots_hit_foe"`
	shotsHitFoeHead               float64 `json:"shots_hit_foe_head"`
	utilityOnDeathAvg             float64 `json:"utility_on_death_avg"`
	heFoesDamageAvg               float64 `json:"he_foes_damage_avg"`
	heFriendsDamageAvg            float64 `json:"he_friends_damage_avg"`
	heThrown                      float64 `json:"he_thrown"`
	molotovThrown                 float64 `json:"molotov_thrown"`
	smokeThrown                   float64 `json:"smoke_thrown"`
	counterStrafingShotsAll       float64 `json:"counter_strafing_shots_all"`
	counterStrafingShotsBad       float64 `json:"counter_strafing_shots_bad"`
	counterStrafingShotsGood      float64 `json:"counter_strafing_shots_good"`
	counterStrafingShotsGoodRatio float64 `json:"counter_strafing_shots_good_ratio"`
	flashbangHitFoe               float64 `json:"flashbang_hit_foe"`
	flashbangLeadingToKill        float64 `json:"flashbang_leading_to_kill"`
	flashbangHitFoeAvgDuration    float64 `json:"flashbang_hit_foe_avg_duration"`
	flashbangHitFriend            float64 `json:"flashbang_hit_friend"`
	flashbangThrown               float64 `json:"flashbang_thrown"`
	flashAssist                   float64 `json:"flash_assist"`
	score                         float64 `json:"score"`
	initialTeamNumber             float64 `json:"initial_team_number"`
	sprayAccuracy                 float64 `json:"spray_accuracy"`
	totalKills                    float64 `json:"total_kills"`
	totalDeaths                   float64 `json:"total_deaths"`
	kdRatio                       float64 `json:"kd_ratio"`
	roundsSurvived                float64 `json:"rounds_survived"`
	roundsSurvivedPercentage      float64 `json:"rounds_survived_percentage"`
	// Damage Per Round
	dpr                               float64 `json:"dpr"`
	totalAssists                      float64 `json:"total_assists"`
	totalDamage                       float64 `json:"total_damage"`
	leetifyRating                     float64 `json:"leetify_rating"`
	ctLeetifyRating                   float64 `json:"ct_leetify_rating"`
	tLeetifyRating                    float64 `json:"t_leetify_rating"`
	multi1K                           float64 `json:"multi1k"`
	multi2K                           float64 `json:"multi2k"`
	multi3K                           float64 `json:"multi3k"`
	multi4K                           float64 `json:"multi4k"`
	multi5K                           float64 `json:"multi5k"`
	roundsCount                       float64 `json:"rounds_count"`
	roundsWon                         float64 `json:"rounds_won"`
	roundsLost                        float64 `json:"rounds_lost"`
	totalHSKills                      float64 `json:"total_hs_kills"`
	tradeKillOpportunities            float64 `json:"trade_kill_opportunities"`
	tradeKillAttempts                 float64 `json:"trade_kill_attempts"`
	tradeKillsSucceed                 float64 `json:"trade_kills_succeed"`
	tradeKillAttemptsPercentage       float64 `json:"trade_kill_attempts_percentage"`
	tradeKillsSuccessPercentage       float64 `json:"trade_kills_success_percentage"`
	tradeKillOpportunitiesPerRound    float64 `json:"trade_kill_opportunities_per_round"`
	tradedDeathOpportunities          float64 `json:"traded_death_opportunities"`
	tradedDeathAttempts               float64 `json:"traded_death_attempts"`
	tradedDeathsSucceed               float64 `json:"traded_deaths_succeed"`
	tradedDeathAttemptsPercentage     float64 `json:"traded_death_attempts_percentage"`
	tradedDeathsSuccessPercentage     float64 `json:"traded_deaths_success_percentage"`
	tradedDeathsOpportunitiesPerRound float64 `json:"traded_deaths_opportunities_per_round"`
}

// Player Profile Retrieval
type PlayerProfile struct {
	// 0 for Public
	PrivacyMode      bool               `json:"privacy_mode"`
	winrate          float64            `json:"winrate"`
	total_matches    float64            `json:"total_matches"`
	name             string             `json:"name"`
	bans             []*Bans            `json:"bans"`
	steam64_id       string             `json:"steam64_id"`
	id               string             `json:"id"`
	ranks            []*Ranks           `json:"ranks"`
	rating           []*Rating          `json:"rating"`
	stats            []*Stats           `json:"stats"`
	recent_matches   []*RecentMatches   `json:"recent_matches"`
	recent_teammates []*RecentTeammates `json:"recent_teammates"`
}

// Stats holds player gameplay statistics
type Stats struct {
	accuracyEnemySpotted           float64 `json:"enemy_spotted_acc"`
	accuracyHeadshot               float64 `json:"accuracy_headshot"`
	counterStrafeGoodShotsRatio    float64 `json:"counter_strafe_good_shots_ratio"`
	ctOpeningAggressionSuccess     float64 `json:"ct_opening_aggression_success_ratio"`
	ctOpeningDuelSuccessPercentage float64 `json:"ct_opening_duel_success_percentage"`
	flashbangHitFoeAvgDuration     float64 `json:"flashbang_hit_foe_avg_duration"`
	flashbangHitPerFlashbang       float64 `json:"flashbang_hit_foe_per_flashbang"`
	flashbangHitFriendPerFlashbang float64 `json:"flashbang_hit_friend_per_flashbang"`
	flashbangLeadingToKill         float64 `json:"flashbang_leading_to_kill"`
	flashbangsThrown               float64 `json:"flashbang_thrown"`
	heDmgAvgFoe                    float64 `json:"he_foes_damage_avg"`
	heDmgAvgFriend                 float64 `json:"he_friends_damage_avg"`
	preAim                         float64 `json:"pre_aim"`
	reactionTime                   float64 `json:"reaction_time"`
	sprayAccuracy                  float64 `json:"spray_accuracy"`
	tOpeningAggressionSuccessRate  float64 `json:"t_opening_aggression_success_rate"`
	tOpeningDuelSuccessPercentage  float64 `json:"t_opening_duel_success_rate"`
	tradedDeathsSuccessPercentage  float64 `json:"traded_deaths_success_percentage"`
	tradeKillOpportunitiesPerRound float64 `json:"trade_kill_opportunities_per_round"`
	tradeKillSuccessPercentage     float64 `json:"trade_kill_success_percentage"`
	utilityOnDeathAvg              float64 `json:"utility_on_death_avg"`
}

// Rating holds 0-100 values for player Leetify Specific metrics
type Rating struct {
	aim         int32 `json:"aim"`
	positioning int32 `json:"positioning"`
	utility     int32 `json:"utility"`
	clutch      int32 `json:"clutch"`
	opening     int32 `json:"opening"`
	ct_leetify  int32 `json:"ct_leetify"`
	t_leetify   int32 `json:"t_leetify"`
}

// Ranks holds player rank information
type Ranks struct {
	leetify     int64          `json:"leetify"`
	premiere    int64          `json:"premiere"`
	faceit      int64          `json:"faceit"`
	faceit_elo  int64          `json:"faceit_elo"`
	wingman     int64          `json:"wingman"`
	renown      int64          `json:"renown"`
	competitive []*Competitive `json:"competitive"`
}
