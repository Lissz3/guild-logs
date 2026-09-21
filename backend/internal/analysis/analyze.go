package analysis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"guildlogs/internal/config"
	"guildlogs/internal/model"
)

// Analyzer mantiene los índices precalculados de un intento.
type Analyzer struct {
	cfg   *config.Config
	fd    *model.FightData
	opt   Options
	t     config.Thresholds
	start int64
	end   int64

	players []model.Player
	byID    map[int]model.Player
	casts   map[int][]model.Event // cast + begincast por jugador, ordenados
	dmg     map[int][]model.Event // daño recibido por jugador, ordenado
	maxHP   map[int]int64
	deaths  map[int][]int64 // timestamps absolutos de muerte por jugador

	owned    map[int][]config.Defensive
	defCasts map[string][]int64
	consumRe []*regexp.Regexp
	classes  map[string]bool

	avoidableKnown map[int]bool
	ignoreDmg      map[int]bool
	ignoreActRe    []*regexp.Regexp
	warnings       []string
}

// Analyze ejecuta todos los análisis sobre un intento.
func Analyze(cfg *config.Config, fd *model.FightData, opt Options) *Result {
	a := newAnalyzer(cfg, fd, opt)

	deaths := a.analyzeDeaths()
	avoidPlayers, avoidAbilities := a.analyzeAvoidable()
	activity := a.analyzeActivity()
	cds := a.analyzeCooldowns()

	res := &Result{
		ReportCode:  fd.Report.Code,
		ReportTitle: fd.Report.Title,
		Fight: FightInfo{
			ID: fd.Fight.ID, Name: fd.Fight.Name, Kill: fd.Fight.Kill,
			Difficulty: fd.Fight.Difficulty, DurationMs: fd.Fight.DurationMs(),
		},
		Options:         opt,
		Deaths:          deaths,
		AvoidableDamage: avoidPlayers,
		AvoidableByAbil: avoidAbilities,
		Activity:        activity,
		Cooldowns:       cds,
		Warnings:        a.warnings,
	}
	res.Summary = a.summarize(res)
	return res
}

func newAnalyzer(cfg *config.Config, fd *model.FightData, opt Options) *Analyzer {
	a := &Analyzer{
		cfg: cfg, fd: fd, opt: opt, t: opt.Thresholds,
		start: fd.Fight.StartTime, end: fd.Fight.EndTime,
		byID:     map[int]model.Player{},
		casts:    map[int][]model.Event{},
		dmg:      map[int][]model.Event{},
		maxHP:    map[int]int64{},
		deaths:   map[int][]int64{},
		owned:    map[int][]config.Defensive{},
		defCasts: map[string][]int64{},
		classes:  map[string]bool{},

		avoidableKnown: cfg.AvoidableFor(fd.Fight.EncounterID),
		ignoreDmg:      map[int]bool{},
	}
	for _, id := range cfg.IgnoreDmg {
		a.ignoreDmg[id] = true
	}
	for _, c := range cfg.Consumables {
		a.consumRe = append(a.consumRe, regexp.MustCompile("(?i)"+c.Pattern))
	}
	for _, p := range cfg.Activity.IgnorePatterns {
		a.ignoreActRe = append(a.ignoreActRe, regexp.MustCompile("(?i)"+regexp.QuoteMeta(p)))
	}

	a.players = append(a.players, fd.Players...)
	sort.Slice(a.players, func(i, j int) bool { return a.players[i].Name < a.players[j].Name })
	for _, p := range a.players {
		a.byID[p.ID] = p
		a.classes[p.Class] = true
	}

	for _, e := range fd.Casts {
		if _, ok := a.byID[e.SourceID]; !ok {
			continue
		}
		if e.Type == "cast" || e.Type == "begincast" {
			a.casts[e.SourceID] = append(a.casts[e.SourceID], e)
		}
	}
	for _, e := range fd.DamageTaken {
		if _, ok := a.byID[e.TargetID]; !ok || e.Type != "damage" {
			continue
		}
		a.dmg[e.TargetID] = append(a.dmg[e.TargetID], e)
		if e.MaxHitPoints > a.maxHP[e.TargetID] {
			a.maxHP[e.TargetID] = e.MaxHitPoints
		}
	}
	for _, e := range fd.Deaths {
		if e.Type == "death" {
			if _, ok := a.byID[e.TargetID]; ok {
				a.deaths[e.TargetID] = append(a.deaths[e.TargetID], e.Timestamp)
			}
		}
	}
	for id := range a.casts {
		evs := a.casts[id]
		sort.SliceStable(evs, func(i, j int) bool { return evs[i].Timestamp < evs[j].Timestamp })
	}
	for id := range a.dmg {
		evs := a.dmg[id]
		sort.SliceStable(evs, func(i, j int) bool { return evs[i].Timestamp < evs[j].Timestamp })
	}
	for id := range a.deaths {
		ts := a.deaths[id]
		sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
	}

	for _, p := range a.players {
		for _, d := range cfg.Defensives {
			if a.owns(p, d) {
				a.owned[p.ID] = append(a.owned[p.ID], d)
			}
		}
	}

	if len(fd.DamageTaken) > 0 && len(a.maxHP) == 0 {
		a.warnings = append(a.warnings, "Damage events do not include max HP (includeResources): HP percentages are not reliable.")
	}
	if len(fd.Players) == 0 {
		a.warnings = append(a.warnings, "Could not determine the players/roles of this fight.")
	}
	return a
}

func (a *Analyzer) abilityName(id int) string {
	if ab, ok := a.fd.Abilities[id]; ok && ab.Name != "" {
		return ab.Name
	}
	return fmt.Sprintf("#%d", id)
}

func (a *Analyzer) abilityIcon(id int) string {
	return a.fd.Abilities[id].Icon
}

func (a *Analyzer) rel(ts int64) int64 { return ts - a.start }

func (a *Analyzer) ms(sec float64) int64 { return int64(sec * 1000) }

func contains(list []string, s string) bool {
	for _, x := range list {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

// matchesDef indica si un evento de cast corresponde a un defensivo (por id o por nombre).
func (a *Analyzer) matchesDef(e model.Event, d config.Defensive) bool {
	if e.AbilityGameID == d.ID {
		return true
	}
	return strings.EqualFold(a.abilityName(e.AbilityGameID), d.Name)
}

// defCastTimes devuelve los instantes (absolutos) en que el jugador lanzó el defensivo.
func (a *Analyzer) defCastTimes(pid int, d config.Defensive) []int64 {
	key := fmt.Sprintf("%d|%s", pid, d.Name)
	if v, ok := a.defCasts[key]; ok {
		return v
	}
	var out []int64
	for _, e := range a.casts[pid] {
		if e.Type == "cast" && a.matchesDef(e, d) {
			out = append(out, e.Timestamp)
		}
	}
	a.defCasts[key] = out
	return out
}

// owns decide si un jugador "tiene" el defensivo (clase, spec y, si es talento, uso previo).
func (a *Analyzer) owns(p model.Player, d config.Defensive) bool {
	if p.Class != d.Class {
		return false
	}
	if len(d.Specs) > 0 && p.Spec != "" && !contains(d.Specs, p.Spec) {
		return false
	}
	if d.Talent && !a.opt.AssumeTalents && len(a.defCastTimes(p.ID, d)) == 0 {
		return false
	}
	return true
}

// chargesAvailable simula el recargo de cargas y dice si hay al menos una en el instante t.
// casts debe estar ordenado. Se asume que el CD estaba completo al empezar el intento.
func chargesAvailable(casts []int64, cdMs int64, maxCharges int, t int64) bool {
	cur := maxCharges
	var next int64
	regen := func(now int64) {
		for cur < maxCharges && next <= now {
			cur++
			if cur < maxCharges {
				next += cdMs
			}
		}
	}
	for _, c := range casts {
		if c > t {
			break
		}
		regen(c)
		if cur == maxCharges {
			next = c + cdMs
		}
		if cur > 0 {
			cur--
		}
	}
	regen(t)
	return cur > 0
}

func (a *Analyzer) defAvailable(pid int, d config.Defensive, t int64) bool {
	return chargesAvailable(a.defCastTimes(pid, d), a.ms(d.Cooldown), d.MaxCharges(), t)
}

func (a *Analyzer) defUsedBetween(pid int, d config.Defensive, from, to int64) bool {
	for _, c := range a.defCastTimes(pid, d) {
		if c >= from && c <= to {
			return true
		}
	}
	return false
}

// isAlive: un jugador está muerto desde su muerte hasta que vuelve a lanzar algo.
func (a *Analyzer) isAlive(pid int, t int64) bool {
	var lastDeath int64 = -1
	for _, d := range a.deaths[pid] {
		if d <= t {
			lastDeath = d
		}
	}
	if lastDeath < 0 {
		return true
	}
	for _, c := range a.casts[pid] {
		if c.Timestamp > lastDeath && c.Timestamp <= t {
			return true
		}
	}
	return false
}

// eff es el daño efectivo de un evento (incluye lo absorbido, descuenta el overkill).
func eff(e model.Event) float64 {
	v := e.Amount + e.Absorbed - e.Overkill
	if v < 0 {
		return 0
	}
	return v
}

func (a *Analyzer) pct(dmg float64, pid int) float64 {
	hp := a.maxHP[pid]
	if hp <= 0 {
		return 0
	}
	return dmg / float64(hp) * 100
}

// consumableUsedBefore: los consumibles (HS/poción) se asumen de un uso por combate.
func (a *Analyzer) consumableUsedBefore(pid, idx int, t int64) bool {
	for _, e := range a.casts[pid] {
		if e.Timestamp > t {
			break
		}
		if e.Type == "cast" && a.consumRe[idx].MatchString(a.abilityName(e.AbilityGameID)) {
			return true
		}
	}
	return false
}

func (a *Analyzer) consumablesAvailable(pid int, t int64) []string {
	var out []string
	for i, c := range a.cfg.Consumables {
		if c.RequiresClassInRaid != "" && !a.classes[c.RequiresClassInRaid] {
			continue
		}
		if !a.consumableUsedBefore(pid, i, t) {
			out = append(out, c.Name)
		}
	}
	return out
}

func (a *Analyzer) summarize(r *Result) Summary {
	s := Summary{Deaths: len(r.Deaths)}
	byPlayer := map[int]*PlayerSummary{}
	for _, p := range a.players {
		byPlayer[p.ID] = &PlayerSummary{PlayerID: p.ID, Player: p.Name, Class: p.Class, Spec: p.Spec, Role: p.Role}
	}
	for _, d := range r.Deaths {
		if d.Ignored {
			continue
		}
		s.CountedDeaths++
		ps := byPlayer[d.PlayerID]
		ps.Deaths++
		if contains(d.Flags, "avoidable") {
			s.AvoidableDeaths++
			ps.AvoidableDeaths++
		}
		if contains(d.Flags, "mechanic") || contains(d.Flags, "mechanic_high") {
			s.MechanicDeaths++
			ps.MechanicDeaths++
		}
	}
	var sum float64
	var n int
	for _, x := range r.Activity {
		byPlayer[x.PlayerID].UptimePct = x.UptimePct
		if x.AliveMs > 0 {
			sum += x.UptimePct
			n++
		}
	}
	if n > 0 {
		s.AvgUptimePct = sum / float64(n)
	}
	for _, x := range r.AvoidableDamage {
		byPlayer[x.PlayerID].AvoidablePct = x.AvoidablePct
	}
	for _, x := range r.Cooldowns {
		ps := byPlayer[x.PlayerID]
		ps.CooldownUsePct = x.UsePct
		for _, c := range x.Cooldowns {
			ps.MissedSpikes += c.Missed
		}
	}
	for _, p := range a.players {
		s.Players = append(s.Players, *byPlayer[p.ID])
	}
	return s
}
