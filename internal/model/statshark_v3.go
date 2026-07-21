package model

type SSPlayerDetail struct {
	Nickname    string `json:"nickname"`
	UID         int    `json:"uid"`
	Level       int    `json:"level"`
	LevelProg   float64 `json:"level_progress"`
	Title       string `json:"title"`
	TitleDisc   string `json:"title_disc"`
	Squadron    string `json:"squadron"`
	SquadronID  int    `json:"squadron_id"`
	Avatar      string `json:"avatar"`
	BanStatus   string `json:"ban_status"`
	TotalViews  int    `json:"total_views"`
	RecentViews int    `json:"recent_views"`
	LastUpdate  string `json:"last_update"`
	RegisterDay int64  `json:"register_day"`
	LastOnline  int64  `json:"last_online"`

	Arcade    *SSDetailMode `json:"arcade"`
	Realistic *SSDetailMode `json:"realistic"`
	Simulator *SSDetailMode `json:"simulator"`

	SpadedTotal int                    `json:"spaded_total"`
	SpadedByCountry map[string]int     `json:"spaded_by_country"`

	Leaderboard   *SSLeaderboardData   `json:"leaderboard"`
	Ratings       *SSRatings           `json:"ratings"`
	RatingsNeo    *SSRatingsNeo        `json:"ratings_neo"`
	NameHistory   []SSNameHistory      `json:"name_history"`
	SquadronHistory []SSSquadronHistory `json:"squadron_history"`
	Titles        []SSTitle            `json:"titles"`
	UserInfo      *SSUserInfo          `json:"user_info"`
}

type SSDetailMode struct {
	PvP      *SSDetailPvP      `json:"pvp"`
	Skirmish *SSDetailSkirmish `json:"skirmish"`
}

type SSDetailPvP struct {
	Games       int     `json:"games"`
	Wins        int     `json:"wins"`
	WinRate     float64 `json:"win_rate"`
	AirKills    int     `json:"air_kills"`
	GroundKills int     `json:"ground_kills"`
	NavalKills  int     `json:"naval_kills"`
	Kills       int     `json:"kills"`
	AIBotKills  int     `json:"ai_bot_kills"`
	Respawns    int     `json:"respawns"`
	KPB         float64 `json:"kills_per_battle"`
	KD          float64 `json:"kd"`
	TimePlayed  int     `json:"time_played"`
	FighterTime int     `json:"fighter_time"`
	BomberTime  int     `json:"bomber_time"`
	TankTime    int     `json:"tank_time"`
	HeavyTankTime int   `json:"heavy_tank_time"`
	TDTime      int     `json:"td_time"`
	SPAATime    int     `json:"spaa_time"`
	ShipTime    int     `json:"ship_time"`
	HeliTime    int     `json:"heli_time"`
	AssaultTime int     `json:"assault_time"`
	TorpedoBoatTime int `json:"torpedo_boat_time"`
	GunBoatTime     int `json:"gun_boat_time"`
	DestroyerTime   int `json:"destroyer_time"`
	CruiserTime     int `json:"cruiser_time"`
	HumanTime       int `json:"human_time"`
}

type SSDetailSkirmish struct {
	Games       int `json:"games"`
	Wins        int `json:"wins"`
	AirKills    int `json:"air_kills"`
	GroundKills int `json:"ground_kills"`
	NavalKills  int `json:"naval_kills"`
	Respawns    int `json:"respawns"`
	TimePlayed  int `json:"time_played"`
}

type SSLeaderboardEntry struct {
	Value float64 `json:"value"`
	Rank  int     `json:"rank"`
}

type SSLeaderboardData struct {
	AirArcade      map[string]*SSLeaderboardEntry `json:"air_arcade"`
	AirRealistic   map[string]*SSLeaderboardEntry `json:"air_realistic"`
	AirSimulation  map[string]*SSLeaderboardEntry `json:"air_simulation"`
	Arcade         map[string]*SSLeaderboardEntry `json:"arcade"`
	HelicopterArcade map[string]*SSLeaderboardEntry `json:"helicopter_arcade"`
	Historical     map[string]*SSLeaderboardEntry `json:"historical"`
	Simulation     map[string]*SSLeaderboardEntry `json:"simulation"`
	TankArcade     map[string]*SSLeaderboardEntry `json:"tank_arcade"`
	TankRealistic  map[string]*SSLeaderboardEntry `json:"tank_realistic"`
	TankSimulation map[string]*SSLeaderboardEntry `json:"tank_simulation"`
}

type SSRatingsEntry struct {
	Rating float64 `json:"rating"`
}

type SSRatings struct {
	Total map[string]*SSRatingsEntry `json:"total"`
}

type SSRatingsNeoEntry struct {
	Monthly struct {
		Wt8          float64 `json:"wt8"`
		OriginalWt8  float64 `json:"original_wt8"`
		IsSuspicious bool    `json:"isSuspicious"`
		Contributions struct {
			Kills     float64 `json:"kills"`
			Position  float64 `json:"position"`
			Win       float64 `json:"win"`
			Score     float64 `json:"score"`
		} `json:"contributions"`
	} `json:"monthly"`
}

type SSRatingsNeo struct {
	TankRealistic *SSRatingsNeoEntry `json:"tank_realistic"`
	TankArcade    *SSRatingsNeoEntry `json:"tank_arcade"`
	AirRealistic  *SSRatingsNeoEntry `json:"air_realistic"`
	AirArcade     *SSRatingsNeoEntry `json:"air_arcade"`
}

type SSNameHistory struct {
	IGN  string `json:"ign"`
	Date string `json:"date"`
}

type SSSquadronHistory struct {
	ClanID   int    `json:"clan_id"`
	ClanTag  string `json:"clan_tag"`
	Date     string `json:"date"`
}

type SSTitle struct {
	Name string `json:"name"`
	Lang string `json:"lang"`
	Disc string `json:"disc"`
}

type SSUserInfo struct {
	Likes       int `json:"likes"`
	Dislikes    int `json:"dislikes"`
	CommentCount int `json:"comment_count"`
}
