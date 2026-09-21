package analysis

import (
	"sort"
)

// spike es un pico de daño recibido por un jugador.
type spike struct {
	start, end int64 // absolutos
	pct        float64
	top        int // abilityID con más daño
}

// findSpikes recorre el daño recibido con una ventana deslizante y devuelve los
// picos >= spikePct del HP máx, fusionando ventanas que se solapan o están cerca.
func (a *Analyzer) findSpikes(pid int) []spike {
	evs := a.dmg[pid]
	hp := float64(a.maxHP[pid])
	if hp <= 0 || len(evs) == 0 {
		return nil
	}
	win := a.ms(a.t.WindowSec)
	limit := a.t.SpikePct / 100 * hp

	var raw []spike
	l := 0
	var sum float64
	for r := range evs {
		sum += eff(evs[r])
		for evs[r].Timestamp-evs[l].Timestamp >= win {
			sum -= eff(evs[l])
			l++
		}
		if sum >= limit {
			byAbil := map[int]float64{}
			for i := l; i <= r; i++ {
				byAbil[evs[i].AbilityGameID] += eff(evs[i])
			}
			top, best := 0, -1.0
			for id, v := range byAbil {
				if v > best {
					top, best = id, v
				}
			}
			raw = append(raw, spike{start: evs[l].Timestamp, end: evs[r].Timestamp, pct: sum / hp * 100, top: top})
		}
	}
	// Fusionar: dentro de un mismo cluster (<= windowSec entre ventanas) se queda el pico máximo.
	var out []spike
	for _, s := range raw {
		if n := len(out); n > 0 && s.start <= out[n-1].end+win {
			last := &out[n-1]
			if s.pct > last.pct {
				last.pct, last.top = s.pct, s.top
			}
			if s.start < last.start {
				last.start = s.start
			}
			if s.end > last.end {
				last.end = s.end
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

func (a *Analyzer) analyzeCooldowns() []PlayerCooldowns {
	var out []PlayerCooldowns
	fightMs := a.end - a.start

	for _, p := range a.players {
		pc := PlayerCooldowns{PlayerID: p.ID, Player: p.Name, Class: p.Class, Spec: p.Spec, Role: p.Role,
			Cooldowns: []CooldownUse{}, Spikes: []Spike{}}

		// Tiempo vivo aproximado: hasta la primera muerte (los defensivos no se pueden usar muerto).
		aliveMs := fightMs
		if ds := a.deaths[p.ID]; len(ds) > 0 {
			aliveMs = ds[0] - a.start
		}

		defs := a.owned[p.ID]
		missedByDef := map[string]int{}

		// Picos de daño y oportunidades perdidas (solo defensivos personales).
		for _, sp := range a.findSpikes(p.ID) {
			s := Spike{
				StartMs: a.rel(sp.start), EndMs: a.rel(sp.end), Pct: sp.pct,
				Top: a.abilityName(sp.top), TopIcon: a.abilityIcon(sp.top), TopID: sp.top,
				Available: []string{}, Used: []string{},
			}
			for _, ds := range a.deaths[p.ID] {
				if ds >= sp.start && ds <= sp.end+1000 {
					s.Died = true
				}
			}
			for _, d := range defs {
				if d.Kind != "personal" {
					continue
				}
				// disponible al inicio de la ventana, y no usado desde 2s antes hasta el final
				if a.defUsedBetween(p.ID, d, sp.start-2000, sp.end) {
					s.Used = append(s.Used, d.Name)
					continue
				}
				if a.defAvailable(p.ID, d, sp.start) {
					s.Available = append(s.Available, d.Name)
					missedByDef[d.Name]++
				}
			}
			pc.Spikes = append(pc.Spikes, s)
		}

		// Usos posibles vs reales (personales, externos y de raid del propio jugador).
		var usedTot, possTot int
		for _, d := range defs {
			times := a.defCastTimes(p.ID, d)
			cu := CooldownUse{Name: d.Name, Kind: d.Kind, Cooldown: d.Cooldown, UsedAtMs: []int64{}}
			cd := a.ms(d.Cooldown)
			if cd > 0 {
				cu.Possible = d.MaxCharges() + int(aliveMs/cd)
			}
			for _, t := range times {
				if t <= a.start+aliveMs {
					cu.Used++
					cu.UsedAtMs = append(cu.UsedAtMs, a.rel(t))
				}
			}
			if cu.Used > cu.Possible {
				cu.Possible = cu.Used
			}
			cu.Missed = missedByDef[d.Name]
			pc.Cooldowns = append(pc.Cooldowns, cu)
			if d.Kind == "personal" {
				usedTot += cu.Used
				possTot += cu.Possible
			}
		}
		if possTot > 0 {
			pc.UsePct = float64(usedTot) / float64(possTot) * 100
		}
		sort.Slice(pc.Cooldowns, func(i, j int) bool { return pc.Cooldowns[i].Kind < pc.Cooldowns[j].Kind })
		out = append(out, pc)
	}
	sort.Slice(out, func(i, j int) bool { return len(out[i].Spikes) > len(out[j].Spikes) })
	return out
}
