// Package demo genera un intento sintético y determinista. Sirve para probar la
// herramienta sin credenciales de Warcraft Logs (código de reporte "demo") y
// como fixture de los tests.
package demo

import (
	"fmt"

	"guildlogs/internal/model"
)

const (
	FightStart int64 = 1_000_000
	FightMs    int64 = 300_000

	AbRaidBlast   = 100
	AbVoidZone    = 101 // evitable
	AbTankBuster  = 102
	AbVolley      = 103
	AbPoison      = 104
	AbFiller      = 1001
	AbHealthstone = 6262
	AbPotion      = 431416
)

// IDs de jugadores con un papel concreto en el escenario (los usan los tests).
const (
	PIDMageDeath     = 12 // muere por mecánica con Ice Block + HS/poción disponibles
	PIDWarriorDeath  = 13 // muere por mecánica habiendo usado todo
	PIDPriestDeath   = 14 // muere lentamente con Dispersion disponible
	PIDDKSpike       = 15 // pico de daño sin usar defensivos
	PIDCareless      = 11 // recibe siempre el Void Zone
	PIDRogueDowntime = 9  // 15s sin castear
)

type spec struct {
	id          int
	name, class string
	spec, role  string
}

var roster = []spec{
	{1, "Tanky", "Warrior", "Protection", "tank"},
	{2, "Bulwark", "Paladin", "Protection", "tank"},
	{3, "Holyman", "Priest", "Holy", "healer"},
	{4, "Discy", "Priest", "Discipline", "healer"},
	{5, "Treeman", "Druid", "Restoration", "healer"},
	{6, "Waver", "Shaman", "Restoration", "healer"},
	{7, "Lockie", "Warlock", "Affliction", "dps"},
	{8, "Frosty", "Mage", "Frost", "dps"},
	{9, "Stabby", "Rogue", "Outlaw", "dps"},
	{10, "Hunterino", "Hunter", "Marksmanship", "dps"},
	{11, "Careless", "Demon Hunter", "Havoc", "dps"},
	{12, "Pyro", "Mage", "Fire", "dps"},
	{13, "Berserk", "Warrior", "Fury", "dps"},
	{14, "Shadowy", "Priest", "Shadow", "dps"},
	{15, "Bonez", "DeathKnight", "Unholy", "dps"},
	{16, "Zappy", "Shaman", "Elemental", "dps"},
	{17, "Kitty", "Druid", "Feral", "dps"},
	{18, "Drako", "Evoker", "Devastation", "dps"},
	{19, "Monkey", "Monk", "Windwalker", "dps"},
	{20, "Paly", "Paladin", "Retribution", "dps"},
}

// Fight devuelve el intento sintético (kill, 5 minutos).
func Fight() *model.FightData {
	fd := &model.FightData{
		Report: model.ReportMeta{Code: "demo", Title: "Sample report (synthetic data)", Guild: "Demo"},
		Fight: model.Fight{
			ID: 1, Name: "Test Boss", EncounterID: 9999, Difficulty: 5, Kill: true,
			StartTime: FightStart, EndTime: FightStart + FightMs,
		},
		Abilities: map[int]model.Ability{
			1:             {Name: "Melee", Icon: "inv_sword_04.jpg"},
			AbRaidBlast:   {Name: "Raid Blast", Icon: "spell_fire_selfdestruct.jpg"},
			AbVoidZone:    {Name: "Void Zone", Icon: "spell_shadow_shadowfury.jpg"},
			AbTankBuster:  {Name: "Tank Buster", Icon: "ability_warrior_savageblow.jpg"},
			AbVolley:      {Name: "Shadow Bolt Volley", Icon: "spell_shadow_shadowbolt.jpg"},
			AbPoison:      {Name: "Slow Poison", Icon: "ability_poisonstacking.jpg"},
			AbFiller:      {Name: "Filler Spell", Icon: "inv_misc_questionmark.jpg"},
			AbHealthstone: {Name: "Healthstone", Icon: "inv_stone_04.jpg"},
			AbPotion:      {Name: "Algari Healing Potion", Icon: "inv_potion_175.jpg"},
			871:           {Name: "Shield Wall", Icon: "ability_warrior_shieldwall.jpg"},
			45438:         {Name: "Ice Block", Icon: "spell_frost_frost.jpg"},
			47585:         {Name: "Dispersion", Icon: "spell_priest_dispersion.jpg"},
			184364:        {Name: "Enraged Regeneration", Icon: "ability_warrior_focusedrage.jpg"},
			48792:         {Name: "Icebound Fortitude", Icon: "spell_deathknight_iceboundfortitude.jpg"},
			48707:         {Name: "Anti-Magic Shell", Icon: "spell_deathknight_antimagicshell.jpg"},
			104773:        {Name: "Unending Resolve", Icon: "spell_warlock_unendingresolve.jpg"},
			33206:         {Name: "Pain Suppression", Icon: "spell_holy_painsuppression.jpg"},
			98008:         {Name: "Spirit Link Totem", Icon: "ability_shaman_spiritlink.jpg"},
		},
	}
	maxHP := map[int]int64{}
	role := map[int]string{}
	for _, s := range roster {
		fd.Fight.FriendlyPlayers = append(fd.Fight.FriendlyPlayers, s.id)
		fd.Players = append(fd.Players, model.Player{ID: s.id, Name: s.name, Class: classKey(s.class), Spec: s.spec, Role: s.role})
		maxHP[s.id] = 1_000_000
		if s.role == "tank" {
			maxHP[s.id] = 1_500_000
		}
		role[s.id] = s.role
	}

	at := func(sec float64) int64 { return FightStart + int64(sec*1000) }
	deathAt := map[int]int64{PIDMageDeath: at(90), PIDWarriorDeath: at(150), PIDPriestDeath: at(210)}

	// ---- Casts de relleno (1 cada 1.5s) con pausas controladas ----
	for _, s := range roster {
		step := int64(1500)
		if s.class == "Rogue" || s.class == "Monk" {
			step = 1000 // GCD de 1s
		}
		for ms := int64(0); ms < FightMs; ms += step {
			t := FightStart + ms
			if d, ok := deathAt[s.id]; ok && t >= d {
				break
			}
			if ms >= 240_000 && ms < 250_000 { // transición: toda la raid parada
				continue
			}
			if s.id == PIDRogueDowntime && ms >= 100_000 && ms < 115_000 {
				continue
			}
			fd.Casts = append(fd.Casts, model.Event{Timestamp: t, Type: "cast", SourceID: s.id, AbilityGameID: AbFiller})
		}
	}

	cast := func(pid, ability int, sec float64) {
		fd.Casts = append(fd.Casts, model.Event{Timestamp: at(sec), Type: "cast", SourceID: pid, AbilityGameID: ability})
	}
	// Uso de defensivos / consumibles
	cast(PIDWarriorDeath, AbHealthstone, 120)
	cast(PIDWarriorDeath, AbPotion, 122)
	cast(PIDWarriorDeath, 184364, 145) // Enraged Regeneration
	cast(PIDDKSpike, 48707, 199)       // Anti-Magic Shell en el segundo pico
	cast(1, 871, 45)                   // Shield Wall del tank
	cast(4, 33206, 46)                 // Pain Suppression
	cast(6, 98008, 100)                // Spirit Link Totem

	// ---- Daño ----
	dmg := func(target, ability int, sec float64, amount float64) {
		fd.DamageTaken = append(fd.DamageTaken, model.Event{
			Timestamp: at(sec), Type: "damage", SourceID: 900, TargetID: target, AbilityGameID: ability,
			Amount: amount, HitPoints: maxHP[target] / 2, MaxHitPoints: maxHP[target],
		})
	}
	// Raid Blast: 15% de vida a todos cada 30s
	for sec := 10.0; sec < 300; sec += 30 {
		for _, s := range roster {
			if d, dead := deathAt[s.id]; dead && at(sec) >= d {
				continue
			}
			dmg(s.id, AbRaidBlast, sec, 150_000)
		}
	}
	// Tank Buster cada 45s
	for sec := 15.0; sec < 300; sec += 45 {
		dmg(1, AbTankBuster, sec, 500_000)
	}
	// Void Zone (evitable): siempre Careless + otros 4 en rotación (dps 7..20 salvo los que mueren)
	rot := []int{7, 8, 9, 10, 16, 17, 18, 19, 20}
	for i, sec := 0, 60.0; sec <= 240; i, sec = i+1, sec+60 {
		dmg(PIDCareless, AbVoidZone, sec, 100_000)
		for k := 0; k < 4; k++ {
			dmg(rot[(i*4+k)%len(rot)], AbVoidZone, sec, 100_000)
		}
	}
	// Mago: 3 golpes de 300k en 3s -> muere a los 90s (90% de vida)
	dmg(PIDMageDeath, AbVolley, 87.5, 300_000)
	dmg(PIDMageDeath, AbVolley, 88.5, 300_000)
	dmg(PIDMageDeath, AbVolley, 89.5, 300_000)
	// Guerrero: 700k en 2s -> muere a los 150s (70%), ya usó todo
	dmg(PIDWarriorDeath, AbVolley, 148, 350_000)
	dmg(PIDWarriorDeath, AbVolley, 149.5, 350_000)
	// Sacerdote: 5 ticks de veneno de 80k -> 40% en la ventana, con Dispersion disponible
	for k := 0; k < 5; k++ {
		dmg(PIDPriestDeath, AbPoison, 206+float64(k)*0.9, 80_000)
	}
	// DK: pico de 55% a los 60s (sin usar nada) y a los 200s (usa AMS a los 199s)
	dmg(PIDDKSpike, AbVolley, 61, 280_000)
	dmg(PIDDKSpike, AbVolley, 62, 280_000)
	dmg(PIDDKSpike, AbVolley, 200.5, 280_000)
	dmg(PIDDKSpike, AbVolley, 201.5, 280_000)

	// ---- Muertes ----
	for pid, t := range deathAt {
		ev := model.Event{Timestamp: t, Type: "death", TargetID: pid, KillingAbilityGameID: AbVolley}
		if pid == PIDPriestDeath {
			ev.KillingAbilityGameID = AbPoison
		}
		fd.Deaths = append(fd.Deaths, ev)
	}
	return fd
}

func classKey(c string) string {
	if c == "Demon Hunter" {
		return "DemonHunter"
	}
	return c
}

// Describe devuelve una línea con la plantilla, útil para logs.
func Describe(fd *model.FightData) string {
	return fmt.Sprintf("%s (%d jugadores)", fd.Fight.Name, len(fd.Players))
}
