package analysis

import (
	"strings"
	"testing"

	"guildlogs/internal/config"
	"guildlogs/internal/demo"
)

func run(t *testing.T) *Result {
	t.Helper()
	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	return Analyze(cfg, demo.Fight(), Options{Thresholds: cfg.Thresholds})
}

func findDeath(t *testing.T, r *Result, pid int) DeathReport {
	t.Helper()
	for _, d := range r.Deaths {
		if d.PlayerID == pid {
			return d
		}
	}
	t.Fatalf("no hay muerte del jugador %d", pid)
	return DeathReport{}
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func hasAbility(list []AbilityRef, name string) bool {
	for _, x := range list {
		if x.Name == name {
			return true
		}
	}
	return false
}

func TestDeathMageMechanicAndAvoidable(t *testing.T) {
	d := findDeath(t, run(t), demo.PIDMageDeath)
	if d.WindowPct < 85 || d.WindowPct > 95 {
		t.Errorf("windowPct = %.1f, quería ~90", d.WindowPct)
	}
	for _, f := range []string{"mechanic_high", "avoidable"} {
		if !has(d.Flags, f) {
			t.Errorf("falta flag %q: %v", f, d.Flags)
		}
	}
	if !hasAbility(d.DefensivesAvailable, "Ice Block") {
		t.Errorf("Ice Block debería estar disponible: %v", d.DefensivesAvailable)
	}
	if !hasAbility(d.ConsumablesAvailable, "Healthstone") || !hasAbility(d.ConsumablesAvailable, "Healing Potion") {
		t.Errorf("HS y poción deberían estar disponibles: %v", d.ConsumablesAvailable)
	}
	if d.KillingAbility != "Shadow Bolt Volley" {
		t.Errorf("killing ability = %q", d.KillingAbility)
	}
	if len(d.ExternalsAvailable) == 0 {
		t.Error("debería haber externos/CDs de raid disponibles")
	}
}

func TestDeathWarriorUsedEverything(t *testing.T) {
	d := findDeath(t, run(t), demo.PIDWarriorDeath)
	if !has(d.Flags, "mechanic") {
		t.Errorf("should be a mechanic death: %v (%.1f%%)", d.Flags, d.WindowPct)
	}
	if has(d.Flags, "avoidable") {
		t.Errorf("no debería ser evitable: disp=%v cons=%v", d.DefensivesAvailable, d.ConsumablesAvailable)
	}
	if !hasAbility(d.DefensivesUsed, "Enraged Regeneration") {
		t.Errorf("Enraged Regeneration debería figurar como usado: %v", d.DefensivesUsed)
	}
}

func TestDeathPriestSlowAvoidable(t *testing.T) {
	d := findDeath(t, run(t), demo.PIDPriestDeath)
	if has(d.Flags, "mechanic") || has(d.Flags, "mechanic_high") {
		t.Errorf("no should be a mechanic death: %v (%.1f%%)", d.Flags, d.WindowPct)
	}
	if !has(d.Flags, "avoidable") || !hasAbility(d.DefensivesAvailable, "Dispersion") {
		t.Errorf("debería ser evitable con Dispersion: %v %v", d.Flags, d.DefensivesAvailable)
	}
}

func TestSummaryCounts(t *testing.T) {
	r := run(t)
	s := r.Summary
	if s.Deaths != 3 || s.CountedDeaths != 3 || s.AvoidableDeaths != 2 || s.MechanicDeaths != 2 {
		t.Errorf("resumen inesperado: %+v", s)
	}
}

func TestWipeIgnoresLateDeaths(t *testing.T) {
	cfg, _ := config.Load("")
	fd := demo.Fight()
	fd.Fight.Kill = false
	th := cfg.Thresholds
	th.WipeIgnoreAfterPct = 10 // 20 jugadores -> cuentan las 2 primeras muertes
	r := Analyze(cfg, fd, Options{Thresholds: th})
	ign := 0
	for _, d := range r.Deaths {
		if d.Ignored {
			ign++
		}
	}
	if ign != 1 || r.Summary.CountedDeaths != 2 {
		t.Errorf("ignoradas=%d contadas=%d", ign, r.Summary.CountedDeaths)
	}
}

func TestAvoidableOutlier(t *testing.T) {
	r := run(t)
	var careless *PlayerAvoidable
	for i := range r.AvoidableDamage {
		if r.AvoidableDamage[i].PlayerID == demo.PIDCareless {
			careless = &r.AvoidableDamage[i]
		}
	}
	if careless == nil || len(careless.Rows) == 0 {
		t.Fatalf("Careless debería tener filas de daño evitable: %+v", careless)
	}
	if careless.Rows[0].Ability != "Void Zone" || careless.Rows[0].Known {
		t.Errorf("fila inesperada: %+v", careless.Rows[0])
	}
	if r.AvoidableDamage[0].PlayerID != demo.PIDCareless {
		t.Errorf("Careless debería ser el peor: %s", r.AvoidableDamage[0].Player)
	}
	// Raid Blast lo recibe toda la raid por igual: no debe marcarse
	for _, row := range careless.Rows {
		if row.Ability == "Raid Blast" {
			t.Error("Raid Blast no es evitable")
		}
	}
}

func TestAvoidableKnownFromConfig(t *testing.T) {
	cfg, _ := config.Load("")
	cfg.Avoidable.ByEncounter = map[string][]int{"9999": {demo.AbVoidZone}}
	r := Analyze(cfg, demo.Fight(), Options{Thresholds: cfg.Thresholds})
	var found bool
	for _, p := range r.AvoidableDamage {
		if p.KnownTaken > 0 {
			found = true
		}
	}
	if !found {
		t.Error("con la habilidad en config debería haber daño 'known'")
	}
}

func TestActivityDowntimeAndExcusedPhase(t *testing.T) {
	r := run(t)
	for _, a := range r.Activity {
		switch a.PlayerID {
		case demo.PIDRogueDowntime:
			if a.UptimePct > 96 || a.UptimePct < 90 {
				t.Errorf("uptime del pícaro = %.1f, quería 90-96", a.UptimePct)
			}
			if len(a.Gaps) == 0 || a.Gaps[0].EndMs-a.Gaps[0].StartMs < 13_000 {
				t.Errorf("debería listar un hueco de ~15s: %+v", a.Gaps)
			}
		case demo.PIDMageDeath, demo.PIDWarriorDeath, demo.PIDPriestDeath:
			if a.DeadFromMs < 0 {
				t.Errorf("%s debería figurar como muerto", a.Player)
			}
		default:
			if a.UptimePct < 99 {
				t.Errorf("%s uptime = %.1f (la transición de 10s debería estar excusada)", a.Player, a.UptimePct)
			}
		}
	}
}

func TestCooldownMissedOpportunities(t *testing.T) {
	r := run(t)
	for _, pc := range r.Cooldowns {
		if pc.PlayerID != demo.PIDDKSpike {
			continue
		}
		if len(pc.Spikes) < 2 {
			t.Fatalf("esperaba 2 picos, hay %d", len(pc.Spikes))
		}
		missed := map[string]int{}
		for _, c := range pc.Cooldowns {
			missed[c.Name] = c.Missed
		}
		if missed["Icebound Fortitude"] != 2 {
			t.Errorf("Icebound Fortitude debería estar perdido 2 veces: %v", missed)
		}
		if missed["Anti-Magic Shell"] != 1 {
			t.Errorf("AMS debería estar perdido 1 vez (usado en el 2º pico): %v", missed)
		}
		return
	}
	t.Fatal("no hay cooldowns del DK")
}

func TestChargesAvailable(t *testing.T) {
	cd := int64(10_000)
	if !chargesAvailable(nil, cd, 1, 0) {
		t.Error("sin casts debería estar disponible")
	}
	if chargesAvailable([]int64{1000}, cd, 1, 5000) {
		t.Error("en CD a los 5s")
	}
	if !chargesAvailable([]int64{1000}, cd, 1, 11_000) {
		t.Error("recargado a los 11s")
	}
	if !chargesAvailable([]int64{1000}, cd, 2, 5000) {
		t.Error("con 2 cargas queda 1")
	}
	if chargesAvailable([]int64{1000, 2000}, cd, 2, 5000) {
		t.Error("2 cargas gastadas")
	}
	if !chargesAvailable([]int64{1000, 2000}, cd, 2, 11_500) {
		t.Error("una carga vuelve a los 11s")
	}
}

func TestVerdictsAreEnglish(t *testing.T) {
	d := findDeath(t, run(t), demo.PIDMageDeath)
	if !strings.Contains(d.Verdict, "Mechanic death") {
		t.Errorf("veredicto: %q", d.Verdict)
	}
}

func TestAvoidableByAbilityOnlyFlagged(t *testing.T) {
	r := run(t)
	for _, a := range r.AvoidableByAbil {
		if a.Ability == "Raid Blast" {
			t.Error("Raid Blast golpea a todos por igual y no debería listarse")
		}
	}
	if len(r.AvoidableByAbil) != 1 || r.AvoidableByAbil[0].Ability != "Void Zone" {
		t.Errorf("esperaba solo Void Zone: %+v", r.AvoidableByAbil)
	}
}
