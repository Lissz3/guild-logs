package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"guildlogs/internal/model"
)

// analyzeDeaths marca cada muerte como:
//   - evitable:  el jugador tenía defensivo(s) personales y/o HS/poción disponibles;
//   - mecánica:  recibió >= mechanicPct % de su HP máx en la ventana previa a morir.
func (a *Analyzer) analyzeDeaths() []DeathReport {
	var deaths []model.Event
	for _, e := range a.fd.Deaths {
		if _, ok := a.byID[e.TargetID]; ok && e.Type == "death" {
			deaths = append(deaths, e)
		}
	}
	sort.SliceStable(deaths, func(i, j int) bool { return deaths[i].Timestamp < deaths[j].Timestamp })

	// En wipes, las muertes tras el "punto de no retorno" no cuentan.
	ignoreFrom := len(deaths)
	if !a.fd.Fight.Kill && len(a.players) > 0 {
		n := int(math.Ceil(a.t.WipeIgnoreAfterPct / 100 * float64(len(a.players))))
		if n < 2 {
			n = 2
		}
		if n < ignoreFrom {
			ignoreFrom = n
		}
	}

	out := make([]DeathReport, 0, len(deaths))
	for i, e := range deaths {
		out = append(out, a.analyzeDeath(e, i >= ignoreFrom))
	}
	return out
}

func (a *Analyzer) analyzeDeath(e model.Event, ignored bool) DeathReport {
	p := a.byID[e.TargetID]
	t := e.Timestamp
	r := DeathReport{
		PlayerID: p.ID, Player: p.Name, Class: p.Class, Spec: p.Spec, Role: p.Role,
		TimeMs: a.rel(t), WindowSec: a.t.WindowSec, Ignored: ignored,
		DefensivesAvailable: []string{}, DefensivesUsed: []string{},
		ConsumablesAvailable: []string{}, ExternalsAvailable: []ExternalOption{}, TopSources: []AbilityDamage{},
		Flags: []string{},
	}
	if e.KillingAbilityGameID != 0 {
		r.KillingAbility = a.abilityName(e.KillingAbilityGameID)
		r.KillingAbilityIcon = a.abilityIcon(e.KillingAbilityGameID)
		r.KillingAbilityID = e.KillingAbilityGameID
	}

	// --- Daño en la ventana previa ---
	winStart := t - a.ms(a.t.WindowSec)
	bySrc := map[int]*AbilityDamage{}
	var total, avoidable float64
	for _, ev := range a.dmg[p.ID] {
		if ev.Timestamp <= winStart || ev.Timestamp > t {
			continue
		}
		v := eff(ev)
		if v == 0 {
			continue
		}
		total += v
		ad := bySrc[ev.AbilityGameID]
		if ad == nil {
			ad = &AbilityDamage{
				Ability: a.abilityName(ev.AbilityGameID), AbilityID: ev.AbilityGameID,
				AbilityIcon: a.abilityIcon(ev.AbilityGameID), Avoidable: a.avoidableKnown[ev.AbilityGameID],
			}
			bySrc[ev.AbilityGameID] = ad
		}
		ad.Amount += v
		ad.Hits++
		if ad.Avoidable {
			avoidable += v
		}
	}
	for _, ad := range bySrc {
		ad.PctMaxHP = a.pct(ad.Amount, p.ID)
		r.TopSources = append(r.TopSources, *ad)
	}
	sort.Slice(r.TopSources, func(i, j int) bool { return r.TopSources[i].Amount > r.TopSources[j].Amount })
	if len(r.TopSources) > 5 {
		r.TopSources = r.TopSources[:5]
	}
	r.WindowPct = a.pct(total, p.ID)
	if total > 0 {
		r.AvoidableShare = avoidable / total * 100
	}
	if r.KillingAbility == "" && len(r.TopSources) > 0 {
		r.KillingAbility = r.TopSources[0].Ability
		r.KillingAbilityIcon = r.TopSources[0].AbilityIcon
		r.KillingAbilityID = r.TopSources[0].AbilityID
	}

	// --- Defensivos propios: disponibles / usados ---
	lookback := a.ms(a.t.UsedLookbackSec)
	for _, d := range a.owned[p.ID] {
		if d.Kind != "personal" {
			continue
		}
		if a.defUsedBetween(p.ID, d, t-lookback, t) {
			r.DefensivesUsed = append(r.DefensivesUsed, d.Name)
		}
		if a.defAvailable(p.ID, d, t) {
			r.DefensivesAvailable = append(r.DefensivesAvailable, d.Name)
		}
	}
	r.ConsumablesAvailable = append(r.ConsumablesAvailable, a.consumablesAvailable(p.ID, t)...)

	// --- Externos / CDs de raid disponibles de otros jugadores vivos ---
	for _, o := range a.players {
		if o.ID == p.ID || !a.isAlive(o.ID, t) {
			continue
		}
		for _, d := range a.owned[o.ID] {
			if (d.Kind == "external" || d.Kind == "raid") && a.defAvailable(o.ID, d, t) {
				r.ExternalsAvailable = append(r.ExternalsAvailable, ExternalOption{Ability: d.Name, Caster: o.Name, Kind: d.Kind})
			}
		}
	}

	// --- Flags ---
	if r.WindowPct >= a.t.MechanicHighPct {
		r.Flags = append(r.Flags, "mechanic_high")
	} else if r.WindowPct >= a.t.MechanicPct {
		r.Flags = append(r.Flags, "mechanic")
	}
	if len(r.DefensivesAvailable)+len(r.ConsumablesAvailable) > 0 {
		r.Flags = append(r.Flags, "avoidable")
	}
	if r.AvoidableShare >= 50 {
		r.Flags = append(r.Flags, "avoidable_damage")
	}
	r.Verdict = a.verdict(r)
	return r
}

func (a *Analyzer) verdict(r DeathReport) string {
	var parts []string
	avail := append(append([]string{}, r.DefensivesAvailable...), r.ConsumablesAvailable...)
	mech := contains(r.Flags, "mechanic") || contains(r.Flags, "mechanic_high")
	switch {
	case contains(r.Flags, "avoidable") && mech:
		parts = append(parts, fmt.Sprintf("Mechanic death (%.0f%% of max HP in %.0fs) with cooldowns left unused: %s.",
			r.WindowPct, r.WindowSec, strings.Join(avail, ", ")))
	case contains(r.Flags, "avoidable"):
		parts = append(parts, fmt.Sprintf("Avoidable death: had %s available.", strings.Join(avail, ", ")))
	case mech:
		parts = append(parts, fmt.Sprintf("Mechanic death (%.0f%% of max HP in %.0fs) with no personal defensives available.",
			r.WindowPct, r.WindowSec))
	default:
		parts = append(parts, "No defensives available and no clear damage spike.")
	}
	if contains(r.Flags, "avoidable_damage") {
		parts = append(parts, fmt.Sprintf("%.0f%% of the damage came from avoidable abilities.", r.AvoidableShare))
	}
	if len(r.ExternalsAvailable) > 0 && contains(r.Flags, "mechanic_high") {
		parts = append(parts, "External/raid cooldowns were available.")
	}
	if r.Ignored {
		parts = append(parts, "(After the wipe point: not counted.)")
	}
	return strings.Join(parts, " ")
}
