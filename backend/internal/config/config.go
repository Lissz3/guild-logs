// Package config carga la configuración editable: defensivos por clase/spec,
// consumibles, habilidades evitables por encuentro y umbrales.
package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed default.json
var defaultJSON []byte

type Thresholds struct {
	WindowSec          float64 `json:"windowSec"`          // ventana previa a la muerte para sumar daño
	MechanicPct        float64 `json:"mechanicPct"`        // % de HP máx en la ventana => muerte por mecánica
	MechanicHighPct    float64 `json:"mechanicHighPct"`    // igual, severidad alta
	UsedLookbackSec    float64 `json:"usedLookbackSec"`    // cuánto atrás se busca un defensivo usado
	SpikePct           float64 `json:"spikePct"`           // % HP máx en la ventana => pico de daño
	GapSec             float64 `json:"gapSec"`             // hueco mínimo para contar como downtime
	OutlierRatio       float64 `json:"outlierRatio"`       // veces la mediana para marcar daño "posiblemente evitable"
	OutlierMinSharePct float64 `json:"outlierMinSharePct"` // mínimo % del daño total del jugador
	WipeIgnoreAfterPct float64 `json:"wipeIgnoreAfterPct"` // en wipes, se ignoran las muertes tras este % de la raid
}

type Defensive struct {
	Name     string   `json:"name"`
	ID       int      `json:"id"`
	Class    string   `json:"class"`
	Specs    []string `json:"specs"`
	Cooldown float64  `json:"cooldown"` // segundos
	Charges  int      `json:"charges"`  // 0 o 1 => una carga
	Kind     string   `json:"kind"`     // personal | external | raid
	Talent   bool     `json:"talent"`   // requiere talento (no se puede saber si lo tiene)
}

// MaxCharges devuelve el número de cargas (mínimo 1).
func (d Defensive) MaxCharges() int {
	if d.Charges < 1 {
		return 1
	}
	return d.Charges
}

type Consumable struct {
	Name                string `json:"name"`
	Pattern             string `json:"pattern"`
	RequiresClassInRaid string `json:"requiresClassInRaid"`
	// ID e Icon son opcionales: no hay un solo spellID de "Healthstone" o
	// "poción" válido para todas las expansiones/niveles de objeto, así que
	// esto es solo para mostrar un icono/enlace representativo.
	ID   int    `json:"id"`
	Icon string `json:"icon"`
}

type Avoidable struct {
	ByEncounter map[string][]int `json:"byEncounter"`
	Global      []int            `json:"global"`
}

type Activity struct {
	IgnorePatterns []string           `json:"ignorePatterns"`
	GcdByClass     map[string]float64 `json:"gcdByClass"`
	DefaultGcd     float64            `json:"defaultGcd"`
}

type Config struct {
	Thresholds  Thresholds   `json:"thresholds"`
	Defensives  []Defensive  `json:"defensives"`
	Consumables []Consumable `json:"consumables"`
	Avoidable   Avoidable    `json:"avoidable"`
	IgnoreDmg   []int        `json:"ignoreDamage"`
	Activity    Activity     `json:"activity"`
}

// Load carga la configuración por defecto o, si path no está vacío, el archivo
// indicado (que sustituye a la configuración embebida por completo).
func Load(path string) (*Config, error) {
	data := defaultJSON
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		data = b
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &c, nil
}

// AvoidableFor devuelve el conjunto de habilidades evitables conocidas para un encuentro.
func (c *Config) AvoidableFor(encounterID int) map[int]bool {
	set := map[int]bool{}
	for _, id := range c.Avoidable.Global {
		set[id] = true
	}
	for _, id := range c.Avoidable.ByEncounter[fmt.Sprint(encounterID)] {
		set[id] = true
	}
	return set
}
