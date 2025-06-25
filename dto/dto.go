package dto

type Register struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ServerParams struct {
	Creator                 uint   `json:"-"`
	Name                    string `json:"name"`
	Port                    int    `json:"port"`
	GameMode                string `json:"game_mode"`
	Difficult               string `json:"difficult"`
	AllowCheat              bool   `json:"allow_cheats"`
	ViewDistance            int    `json:"view_distance"`
	SeedWorld               string `json:"seed"`
	MaxPlayer               int    `json:"max_player"`
	DefaultPermissionPlayer string `json:"permission_player"`
}

type StartServerReq struct {
	Name    string `json:"name"`
	WorldId uint   `json:"world_id"`
	Port    int    `json:"port"`
}

type PermissionPlayer struct {
	Xuid       string `json:"xuid"`
	Permission string `json:"permission"`
}

type Allowlist struct {
	Xuid     string `json:"xuid"`
	Name     string `json:"name"`
	Priority bool   `json:"ignoresPlayerLimit"`
}

type GetWorlds struct {
	ID      uint   `json:"id"`
	Creator string `json:"creator"`
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Players int    `json:"players"`
}

type GetWorldAndPlayers struct {
	ID                      uint     `json:"id"`
	Creator                 string   `json:"creator"`
	Name                    string   `json:"name"`
	Port                    int      `json:"port"`
	GameMode                string   `json:"game_mode"`
	Difficult               string   `json:"difficult"`
	AllowCheat              bool     `json:"allow_cheats"`
	ViewDistance            int      `json:"view_distance"`
	SeedWorld               string   `json:"seed"`
	MaxPlayer               int      `json:"max_player"`
	DefaultPermissionPlayer string   `json:"permission_player"`
	Players                 []Player `json:"players"`
}

type Player struct {
	Xuid string `json:"xuid"`
}
