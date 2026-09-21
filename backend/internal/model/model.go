// Package model contiene los tipos de datos compartidos entre el cliente de
// Warcraft Logs y los analizadores.
package model

// Event es un evento de combate de la API v2 de Warcraft Logs. Solo se
// declaran los campos que usan los analizadores.
type Event struct {
	Timestamp            int64   `json:"timestamp"`
	Type                 string  `json:"type"`
	SourceID             int     `json:"sourceID"`
	TargetID             int     `json:"targetID"`
	AbilityGameID        int     `json:"abilityGameID"`
	Amount               float64 `json:"amount"`
	Absorbed             float64 `json:"absorbed"`
	Overkill             float64 `json:"overkill"`
	HitPoints            int64   `json:"hitPoints"`
	MaxHitPoints         int64   `json:"maxHitPoints"`
	KillerID             int     `json:"killerID"`
	KillingAbilityGameID int     `json:"killingAbilityGameID"`
	Tick                 bool    `json:"tick"`
}

// Fight es un intento de encuentro dentro de un reporte.
type Fight struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	EncounterID     int     `json:"encounterID"`
	Difficulty      int     `json:"difficulty"`
	Kill            bool    `json:"kill"`
	StartTime       int64   `json:"startTime"`
	EndTime         int64   `json:"endTime"`
	FightPercentage float64 `json:"fightPercentage"`
	FriendlyPlayers []int   `json:"friendlyPlayers"`
}

// DurationMs devuelve la duración del intento en milisegundos.
func (f Fight) DurationMs() int64 { return f.EndTime - f.StartTime }

// Player es un jugador de la raid en un intento concreto.
type Player struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"` // p.ej. "DeathKnight", "Priest"
	Spec  string `json:"spec"`  // puede estar vacío si no se conoce
	Role  string `json:"role"`  // "tank", "healer", "dps"
}

// ReportMeta es el resumen de un reporte (lista de peleas).
type ReportMeta struct {
	Code      string  `json:"code"`
	Title     string  `json:"title"`
	Guild     string  `json:"guild"`
	StartTime int64   `json:"startTime"`
	Fights    []Fight `json:"fights"`
}

// Ability es el nombre e icono de una habilidad, tal como los devuelve
// masterData.abilities de la API de Warcraft Logs.
type Ability struct {
	Name string
	Icon string // nombre de fichero (p.ej. "spell_fire_fireball02.jpg"), puede estar vacío
}

// FightData agrupa todo lo necesario para analizar un intento.
type FightData struct {
	Report      ReportMeta
	Fight       Fight
	Players     []Player
	Abilities   map[int]Ability // gameID -> nombre/icono
	Deaths      []Event
	DamageTaken []Event // daño recibido por amigos (se filtra a jugadores al analizar)
	Casts       []Event // casts y begincast de amigos
}
