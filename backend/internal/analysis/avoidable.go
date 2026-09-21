package analysis

import (
	"sort"
)

type dmgAgg struct {
	amount float64
	hits   int
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	c := append([]float64(nil), v...)
	sort.Float64s(c)
	n := len(c)
	if n%2 == 1 {
		return c[n/2]
	}
	return (c[n/2-1] + c[n/2]) / 2
}

// analyzeAvoidable compara, por habilidad, el daño que ha recibido cada
// jugador que no es tank con la mediana de la raid.
//
//   - "Known": la habilidad está en la lista de evitables de la config (por encuentro).
//   - "Posible": no está en la lista, pero el jugador ha recibido >= outlierRatio veces
//     la mediana de quienes fueron golpeados, y eso es >= outlierMinSharePct de su daño total.
func (a *Analyzer) analyzeAvoidable() ([]PlayerAvoidable, []AbilityAvoidable) {
	// habilidad -> jugador -> daño
	perAbility := map[int]map[int]*dmgAgg{}
	totals := map[int]float64{}
	for _, p := range a.players {
		if p.Role == "tank" {
			continue
		}
		for _, ev := range a.dmg[p.ID] {
			if a.ignoreDmg[ev.AbilityGameID] {
				continue
			}
			v := eff(ev)
			if v == 0 {
				continue
			}
			totals[p.ID] += v
			m := perAbility[ev.AbilityGameID]
			if m == nil {
				m = map[int]*dmgAgg{}
				perAbility[ev.AbilityGameID] = m
			}
			g := m[p.ID]
			if g == nil {
				g = &dmgAgg{}
				m[p.ID] = g
			}
			g.amount += v
			g.hits++
		}
	}

	const minHitters = 4
	medians := map[int]float64{}
	var byAbility []AbilityAvoidable
	for id, m := range perAbility {
		vals := make([]float64, 0, len(m))
		var sum, worstAmt float64
		var worst string
		for pid, g := range m {
			vals = append(vals, g.amount)
			sum += g.amount
			if g.amount > worstAmt {
				worstAmt, worst = g.amount, a.byID[pid].Name
			}
		}
		med := median(vals)
		medians[id] = med
		known := a.avoidableKnown[id]
		if known || len(vals) >= minHitters {
			byAbility = append(byAbility, AbilityAvoidable{
				Ability: a.abilityName(id), AbilityID: id, AbilityIcon: a.abilityIcon(id), Known: known,
				Total: sum, Hitters: len(vals), Median: med, Worst: worst, WorstAmt: worstAmt,
			})
		}
	}

	flagged := map[int]bool{}
	var players []PlayerAvoidable
	for _, p := range a.players {
		if p.Role == "tank" {
			continue
		}
		pa := PlayerAvoidable{PlayerID: p.ID, Player: p.Name, Class: p.Class, Spec: p.Spec, Role: p.Role,
			TotalTaken: totals[p.ID], Rows: []AvoidableRow{}}
		for id, m := range perAbility {
			g := m[p.ID]
			if g == nil {
				continue
			}
			known := a.avoidableKnown[id]
			med := medians[id]
			share := 0.0
			if totals[p.ID] > 0 {
				share = g.amount / totals[p.ID] * 100
			}
			possible := !known && len(m) >= minHitters && med > 0 &&
				g.amount >= a.t.OutlierRatio*med && share >= a.t.OutlierMinSharePct
			if !known && !possible {
				continue
			}
			flagged[id] = true
			ratio := 0.0
			if med > 0 {
				ratio = g.amount / med
			}
			pa.Rows = append(pa.Rows, AvoidableRow{
				Ability: a.abilityName(id), AbilityID: id, AbilityIcon: a.abilityIcon(id), Amount: g.amount, Hits: g.hits,
				RaidMedian: med, Ratio: ratio, SharePct: share, Known: known,
			})
			if known {
				pa.KnownTaken += g.amount
			} else {
				pa.PossibleTaken += g.amount
			}
		}
		sort.Slice(pa.Rows, func(i, j int) bool { return pa.Rows[i].Amount > pa.Rows[j].Amount })
		if len(pa.Rows) > 8 {
			pa.Rows = pa.Rows[:8]
		}
		if pa.TotalTaken > 0 {
			pa.AvoidablePct = (pa.KnownTaken + pa.PossibleTaken) / pa.TotalTaken * 100
		}
		players = append(players, pa)
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].KnownTaken+players[i].PossibleTaken > players[j].KnownTaken+players[j].PossibleTaken
	})
	// Solo se listan habilidades marcadas como evitables o con algún jugador atípico.
	kept := byAbility[:0]
	for _, ab := range byAbility {
		if ab.Known || flagged[ab.AbilityID] {
			kept = append(kept, ab)
		}
	}
	byAbility = kept
	sort.Slice(byAbility, func(i, j int) bool { return byAbility[i].Total > byAbility[j].Total })
	if len(byAbility) > 25 {
		byAbility = byAbility[:25]
	}
	return players, byAbility
}
