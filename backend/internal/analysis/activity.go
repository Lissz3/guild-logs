package analysis

import (
	"sort"

	"guildlogs/internal/model"
)

func (a *Analyzer) gcdMs(p model.Player) int64 {
	g := a.cfg.Activity.DefaultGcd
	if v, ok := a.cfg.Activity.GcdByClass[p.Class]; ok {
		g = v
	}
	if g <= 0 {
		g = 1.5
	}
	return a.ms(g)
}

func (a *Analyzer) ignoredForActivity(abilityID int) bool {
	name := a.abilityName(abilityID)
	for _, re := range a.ignoreActRe {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

// activityIntervals convierte los casts de un jugador en intervalos "ocupado".
// Un cast instantáneo ocupa un GCD; si hay begincast, ocupa hasta el cast.
func (a *Analyzer) activityIntervals(p model.Player) []interval {
	gcd := a.gcdMs(p)
	pending := map[int]int64{} // habilidad -> inicio del begincast
	var iv []interval
	for _, e := range a.casts[p.ID] {
		if a.ignoredForActivity(e.AbilityGameID) {
			continue
		}
		switch e.Type {
		case "begincast":
			pending[e.AbilityGameID] = e.Timestamp
		case "cast":
			s := e.Timestamp
			end := s + gcd
			if b, ok := pending[e.AbilityGameID]; ok {
				delete(pending, e.AbilityGameID)
				if s-b <= 15000 { // begincast válido
					s = b
					if s+gcd > end {
						end = s + gcd
					}
					if e.Timestamp > end {
						end = e.Timestamp
					}
				}
			}
			iv = append(iv, interval{s, end})
		}
	}
	return mergeIntervals(iv)
}

// aliveIntervals: vivo desde el inicio hasta la muerte; tras morir, vuelve a
// estar vivo en el siguiente cast (resurrección).
func (a *Analyzer) aliveIntervals(p model.Player) ([]interval, int64) {
	alive := []interval{}
	cur := a.start
	var deadFrom int64 = -1
	for _, d := range a.deaths[p.ID] {
		if d > cur {
			alive = append(alive, interval{cur, d})
		}
		if deadFrom < 0 {
			deadFrom = a.rel(d)
		}
		cur = a.end // muerto hasta que se demuestre lo contrario
		for _, c := range a.casts[p.ID] {
			if c.Timestamp > d {
				cur = c.Timestamp
				break
			}
		}
		if cur >= a.end {
			// sigue muerto hasta el final
			return alive, deadFrom
		}
	}
	if cur < a.end {
		alive = append(alive, interval{cur, a.end})
	}
	// si resucitó, deadFrom indica la primera muerte pero ya no sigue muerto
	return mergeIntervals(alive), deadFrom
}

func (a *Analyzer) analyzeActivity() []PlayerActivity {
	type data struct {
		p      model.Player
		active []interval
		alive  []interval
		dead   int64
	}
	all := make([]data, 0, len(a.players))
	for _, p := range a.players {
		al, dead := a.aliveIntervals(p)
		all = append(all, data{p: p, active: a.activityIntervals(p), alive: al, dead: dead})
	}

	// Ventanas en las que casi toda la raid está inactiva (fases sin objetivo,
	// transiciones...): esa inactividad no se penaliza.
	idle := a.raidIdleWindows(func(yield func(active, alive []interval)) {
		for _, d := range all {
			yield(d.active, d.alive)
		}
	})

	out := make([]PlayerActivity, 0, len(all))
	for _, d := range all {
		aliveMs := totalLen(d.alive)
		activeMs := totalLen(intersect(d.active, d.alive))
		gaps := subtract(d.alive, d.active)
		excused := totalLen(intersect(gaps, idle))
		real := subtract(gaps, idle)

		pa := PlayerActivity{
			PlayerID: d.p.ID, Player: d.p.Name, Class: d.p.Class, Spec: d.p.Spec, Role: d.p.Role,
			AliveMs: aliveMs, ActiveMs: activeMs, ExcusedMs: excused, DeadFromMs: d.dead,
			Casts: 0, Gaps: []Gap{},
		}
		for _, e := range a.casts[d.p.ID] {
			if e.Type == "cast" && !a.ignoredForActivity(e.AbilityGameID) {
				pa.Casts++
			}
		}
		if denom := aliveMs - excused; denom > 0 {
			pa.UptimePct = float64(activeMs) / float64(denom) * 100
			pa.DowntimeMs = denom - activeMs
		}
		var listed []interval
		for _, g := range real {
			if g.length() >= a.ms(a.t.GapSec) {
				listed = append(listed, g)
			}
		}
		sort.Slice(listed, func(i, j int) bool { return listed[i].length() > listed[j].length() })
		if len(listed) > 12 {
			listed = listed[:12]
		}
		sort.Slice(listed, func(i, j int) bool { return listed[i].s < listed[j].s })
		for _, g := range listed {
			pa.Gaps = append(pa.Gaps, Gap{StartMs: a.rel(g.s), EndMs: a.rel(g.e)})
		}
		out = append(out, pa)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UptimePct < out[j].UptimePct })
	return out
}

// raidIdleWindows devuelve los intervalos (en bins de 1s) en los que menos del
// 25% de los jugadores vivos estaba activo.
func (a *Analyzer) raidIdleWindows(each func(yield func(active, alive []interval))) []interval {
	dur := a.end - a.start
	if dur <= 0 {
		return nil
	}
	n := int(dur/1000) + 1
	activeCnt := make([]int, n)
	aliveCnt := make([]int, n)
	each(func(active, alive []interval) {
		for b := 0; b < n; b++ {
			mid := a.start + int64(b)*1000 + 500
			if inAny(alive, mid) {
				aliveCnt[b]++
				if inAny(active, mid) {
					activeCnt[b]++
				}
			}
		}
	})
	var idle []interval
	for b := 0; b < n; b++ {
		if aliveCnt[b] >= 5 && float64(activeCnt[b])/float64(aliveCnt[b]) < 0.25 {
			idle = append(idle, interval{a.start + int64(b)*1000, a.start + int64(b+1)*1000})
		}
	}
	return mergeIntervals(idle)
}

func inAny(iv []interval, t int64) bool {
	i := sort.Search(len(iv), func(i int) bool { return iv[i].e > t })
	return i < len(iv) && iv[i].s <= t
}
