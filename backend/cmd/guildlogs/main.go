// Command guildlogs: web server that analyzes Warcraft Logs reports.
//
// Settings come from config.json (or the file given with -config) and can be
// overridden with environment variables: PORT, WCL_CLIENT_ID, WCL_CLIENT_SECRET,
// GUILDLOGS_CACHE and GUILDLOGS_CONFIG. See README.md.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"guildlogs/internal/analysis"
	"guildlogs/internal/config"
	"guildlogs/internal/demo"
	"guildlogs/internal/model"
	"guildlogs/internal/settings"
	"guildlogs/internal/wcl"
)

type server struct {
	cfg *config.Config
	wcl *wcl.Client // nil si no hay credenciales
}

func main() {
	cfgPath := flag.String("config", "", "path to the settings JSON (default: ./"+settings.DefaultFile+" if it exists)")
	flag.Parse()

	set, loaded, err := settings.Load(*cfgPath, os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if loaded != "" {
		log.Printf("Settings loaded from %s", loaded)
	} else {
		log.Printf("No %s found: using environment variables and defaults.", settings.DefaultFile)
	}

	cfg, err := config.Load(set.AnalysisConfig)
	if err != nil {
		log.Fatal(err)
	}
	s := &server{cfg: cfg}
	if set.HasCredentials() {
		s.wcl = wcl.New(set.WCLClientID, set.WCLClientSecret, set.CacheDir)
	} else {
		log.Println("WARNING: no Warcraft Logs credentials (wclClientId / wclClientSecret); only the \"demo\" report will work.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/report", s.handleReport)
	mux.HandleFunc("/api/analyze", s.handleAnalyze)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "Guild Logs API. The web UI is the Next.js app in ./frontend (default: http://localhost:3000).\n")
	})

	log.Printf("Listening on http://localhost:%s", set.Port)
	log.Fatal(http.ListenAndServe(":"+set.Port, mux))
}

var codeRe = regexp.MustCompile(`reports/([A-Za-z0-9]+)`)

// reportCode acepta un código suelto o una URL completa de Warcraft Logs.
func reportCode(in string) string {
	if m := codeRe.FindStringSubmatch(in); m != nil {
		return m[1]
	}
	return in
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"thresholds":     s.cfg.Thresholds,
		"wclConfigured":  s.wcl != nil,
		"avoidableKnown": len(s.cfg.Avoidable.Global) + len(s.cfg.Avoidable.ByEncounter),
	})
}

var errNoWCL = errors.New("the server has no Warcraft Logs credentials (wclClientId / wclClientSecret); use the code \"demo\"")

func (s *server) handleReport(w http.ResponseWriter, r *http.Request) {
	code := reportCode(r.URL.Query().Get("code"))
	if code == "" {
		writeErr(w, 400, errors.New("missing report code"))
		return
	}
	if code == "demo" {
		fd := demo.Fight()
		fd.Report.Fights = []model.Fight{fd.Fight}
		writeJSON(w, 200, fd.Report)
		return
	}
	if s.wcl == nil {
		writeErr(w, 503, errNoWCL)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	meta, _, err := s.wcl.ReportMeta(ctx, code)
	if err != nil {
		writeErr(w, 502, err)
		return
	}
	writeJSON(w, 200, meta)
}

func (s *server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := reportCode(q.Get("code"))
	fightID, err := strconv.Atoi(q.Get("fight"))
	if code == "" || err != nil {
		writeErr(w, 400, errors.New("missing code and fight"))
		return
	}

	opt := analysis.Options{Thresholds: s.cfg.Thresholds, AssumeTalents: q.Get("assumeTalents") == "1"}
	t := &opt.Thresholds
	setF := func(name string, dst *float64) {
		if v := q.Get(name); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
				*dst = f
			}
		}
	}
	setF("window", &t.WindowSec)
	setF("mechanicPct", &t.MechanicPct)
	setF("mechanicHighPct", &t.MechanicHighPct)
	setF("spikePct", &t.SpikePct)
	setF("gapSec", &t.GapSec)

	var fd *model.FightData
	if code == "demo" {
		fd = demo.Fight()
	} else {
		if s.wcl == nil {
			writeErr(w, 503, errNoWCL)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()
		fd, err = s.wcl.LoadFight(ctx, code, fightID)
		if err != nil {
			writeErr(w, 502, err)
			return
		}
	}
	writeJSON(w, 200, analysis.Analyze(s.cfg, fd, opt))
}
