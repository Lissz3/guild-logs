// Package analysis contiene los analizadores: muertes evitables/por mecánica,
// daño evitable, uptime de actividad y uso de defensivos.
//
// Todos los tiempos de salida son milisegundos desde el inicio del intento.
package analysis

import "guildlogs/internal/config"

// Options son los parámetros de un análisis.
type Options struct {
	Thresholds    config.Thresholds
	AssumeTalents bool // contar defensivos con talento aunque el jugador no los haya usado
}

// AbilityRef nombra una habilidad con su icono, para que el frontend pueda
// mostrar el icono y enlazar su tooltip (Wowhead). AbilityKind es "item"
// cuando AbilityID es un itemID de Wowhead (wowhead.com/item=); vacío (o
// cualquier otro valor) significa spellID (wowhead.com/spell=), el caso normal.
type AbilityRef struct {
	Name        string `json:"name"`
	AbilityID   int    `json:"abilityId"`
	AbilityIcon string `json:"abilityIcon"`
	AbilityKind string `json:"abilityKind,omitempty"`
}

type AbilityDamage struct {
	Ability     string  `json:"ability"`
	AbilityID   int     `json:"abilityId"`
	AbilityIcon string  `json:"abilityIcon"`
	Amount      float64 `json:"amount"`
	PctMaxHP    float64 `json:"pctMaxHp"`
	Hits        int     `json:"hits"`
	Avoidable   bool    `json:"avoidable"` // marcada como evitable en la config
}

type ExternalOption struct {
	Ability     string `json:"ability"`
	AbilityID   int    `json:"abilityId"`
	AbilityIcon string `json:"abilityIcon"`
	Caster      string `json:"caster"`
	Kind        string `json:"kind"`
}

// DeathReport describe una muerte y por qué se marca (o no).
type DeathReport struct {
	PlayerID int    `json:"playerId"`
	Player   string `json:"player"`
	Class    string `json:"class"`
	Spec     string `json:"spec"`
	Role     string `json:"role"`
	TimeMs   int64  `json:"timeMs"`

	KillingAbility     string          `json:"killingAbility"`
	KillingAbilityIcon string          `json:"killingAbilityIcon"`
	KillingAbilityID   int             `json:"killingAbilityId"`
	WindowPct          float64         `json:"windowPct"` // % de HP máx recibido en la ventana previa
	WindowSec          float64         `json:"windowSec"`
	TopSources         []AbilityDamage `json:"topSources"`
	AvoidableShare     float64         `json:"avoidableShare"` // % del daño de la ventana de habilidades evitables conocidas

	DefensivesAvailable  []AbilityRef     `json:"defensivesAvailable"`
	DefensivesUsed       []AbilityRef     `json:"defensivesUsed"`
	ConsumablesAvailable []AbilityRef     `json:"consumablesAvailable"`
	ExternalsAvailable   []ExternalOption `json:"externalsAvailable"`

	// Flags: "avoidable", "mechanic", "mechanic_high", "avoidable_damage".
	Flags   []string `json:"flags"`
	Verdict string   `json:"verdict"`
	// Ignored es true para muertes posteriores al "punto de wipe" (no cuentan).
	Ignored bool `json:"ignored"`
}

type AvoidableRow struct {
	Ability     string  `json:"ability"`
	AbilityID   int     `json:"abilityId"`
	AbilityIcon string  `json:"abilityIcon"`
	Amount      float64 `json:"amount"`
	Hits        int     `json:"hits"`
	RaidMedian  float64 `json:"raidMedian"`
	Ratio       float64 `json:"ratio"`    // veces la mediana
	SharePct    float64 `json:"sharePct"` // % del daño total recibido por el jugador
	Known       bool    `json:"known"`    // en la lista de evitables de la config
}

type PlayerAvoidable struct {
	PlayerID      int            `json:"playerId"`
	Player        string         `json:"player"`
	Class         string         `json:"class"`
	Spec          string         `json:"spec"`
	Role          string         `json:"role"`
	TotalTaken    float64        `json:"totalTaken"`
	KnownTaken    float64        `json:"knownTaken"`
	PossibleTaken float64        `json:"possibleTaken"`
	AvoidablePct  float64        `json:"avoidablePct"` // (known+possible)/total
	Rows          []AvoidableRow `json:"rows"`
}

type AbilityAvoidable struct {
	Ability     string  `json:"ability"`
	AbilityID   int     `json:"abilityId"`
	AbilityIcon string  `json:"abilityIcon"`
	Known       bool    `json:"known"`
	Total       float64 `json:"total"`
	Hitters     int     `json:"hitters"`
	Median      float64 `json:"median"`
	Worst       string  `json:"worst"`
	WorstAmt    float64 `json:"worstAmount"`
}

type Gap struct {
	StartMs int64 `json:"startMs"`
	EndMs   int64 `json:"endMs"`
}

type PlayerActivity struct {
	PlayerID   int     `json:"playerId"`
	Player     string  `json:"player"`
	Class      string  `json:"class"`
	Spec       string  `json:"spec"`
	Role       string  `json:"role"`
	UptimePct  float64 `json:"uptimePct"`
	AliveMs    int64   `json:"aliveMs"`
	ActiveMs   int64   `json:"activeMs"`
	ExcusedMs  int64   `json:"excusedMs"` // inactividad de toda la raid (fases, movimiento...), no penaliza
	DowntimeMs int64   `json:"downtimeMs"`
	Casts      int     `json:"casts"`
	Gaps       []Gap   `json:"gaps"`       // los huecos más largos
	DeadFromMs int64   `json:"deadFromMs"` // -1 si sigue vivo
}

type CooldownUse struct {
	Name        string  `json:"name"`
	AbilityID   int     `json:"abilityId"`
	AbilityIcon string  `json:"abilityIcon"`
	Kind        string  `json:"kind"`
	Cooldown    float64 `json:"cooldown"`
	Possible    int     `json:"possible"`
	Used        int     `json:"used"`
	UsedAtMs    []int64 `json:"usedAtMs"`
	Missed      int     `json:"missedOpportunities"` // picos de daño con el CD disponible y sin usar
}

type Spike struct {
	StartMs   int64        `json:"startMs"`
	EndMs     int64        `json:"endMs"`
	Pct       float64      `json:"pct"` // % de HP máx recibido en la ventana
	Top       string       `json:"top"`
	TopIcon   string       `json:"topIcon"`
	TopID     int          `json:"topAbilityId"`
	Available []AbilityRef `json:"available"` // defensivos personales disponibles y sin usar
	Used      []AbilityRef `json:"used"`
	Died      bool         `json:"died"`
}

type PlayerCooldowns struct {
	PlayerID  int           `json:"playerId"`
	Player    string        `json:"player"`
	Class     string        `json:"class"`
	Spec      string        `json:"spec"`
	Role      string        `json:"role"`
	Cooldowns []CooldownUse `json:"cooldowns"`
	Spikes    []Spike       `json:"spikes"`
	UsePct    float64       `json:"usePct"` // usados / posibles (defensivos personales)
}

type PlayerSummary struct {
	PlayerID        int     `json:"playerId"`
	Player          string  `json:"player"`
	Class           string  `json:"class"`
	Spec            string  `json:"spec"`
	Role            string  `json:"role"`
	Deaths          int     `json:"deaths"`
	DeathTimestamps []int64 `json:"deathTimestamps"` // timeMs de cada muerte contada (mismo run que Deaths)
	AvoidableDeaths int     `json:"avoidableDeaths"`
	MechanicDeaths  int     `json:"mechanicDeaths"`
	UptimePct       float64 `json:"uptimePct"`
	AvoidableDamage float64 `json:"avoidableDamage"` // known+possible, en daño absoluto (ver AvoidablePct para el %)
	AvoidablePct    float64 `json:"avoidablePct"`
	CooldownUsePct  float64 `json:"cooldownUsePct"`
	MissedSpikes    int     `json:"missedSpikes"`
}

type Summary struct {
	Deaths               int             `json:"deaths"`
	CountedDeaths        int             `json:"countedDeaths"`
	AvoidableDeaths      int             `json:"avoidableDeaths"`
	MechanicDeaths       int             `json:"mechanicDeaths"`
	AvgUptimePct         float64         `json:"avgUptimePct"`
	TotalAvoidableDamage float64         `json:"totalAvoidableDamage"` // suma de daño evitable/posible de toda la raid en el intento (no tanks)
	Players              []PlayerSummary `json:"players"`
}

type FightInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Kill       bool   `json:"kill"`
	Difficulty int    `json:"difficulty"`
	DurationMs int64  `json:"durationMs"`
}

// Result es la respuesta completa de un análisis.
type Result struct {
	ReportCode      string             `json:"reportCode"`
	ReportTitle     string             `json:"reportTitle"`
	Fight           FightInfo          `json:"fight"`
	Options         Options            `json:"options"`
	Summary         Summary            `json:"summary"`
	Deaths          []DeathReport      `json:"deaths"`
	AvoidableDamage []PlayerAvoidable  `json:"avoidableDamage"`
	AvoidableByAbil []AbilityAvoidable `json:"avoidableByAbility"`
	Activity        []PlayerActivity   `json:"activity"`
	Cooldowns       []PlayerCooldowns  `json:"cooldowns"`
	Warnings        []string           `json:"warnings"`
}
