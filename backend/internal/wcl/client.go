// Package wcl es un cliente mínimo de la API v2 (GraphQL) de Warcraft Logs.
//
// NOTA: las consultas se han escrito a partir del esquema público conocido de
// la API v2. Si Warcraft Logs cambia algún campo, todo lo que toca la API está
// aislado en este paquete.
package wcl

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"guildlogs/internal/model"
)

const (
	tokenURL   = "https://www.warcraftlogs.com/oauth/token"
	graphqlURL = "https://www.warcraftlogs.com/api/v2/client"
)

// Client habla con la API de Warcraft Logs usando client-credentials.
type Client struct {
	id, secret string
	http       *http.Client
	cacheDir   string

	mu      sync.Mutex
	token   string
	expires time.Time
}

// New crea un cliente. cacheDir puede estar vacío para desactivar la caché en disco.
func New(clientID, clientSecret, cacheDir string) *Client {
	return &Client{
		id:       clientID,
		secret:   clientSecret,
		http:     &http.Client{Timeout: 60 * time.Second},
		cacheDir: cacheDir,
	}
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expires) {
		return c.token, nil
	}
	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.id, c.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var t struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &t); err != nil {
		return "", err
	}
	c.token = t.AccessToken
	c.expires = time.Now().Add(time.Duration(t.ExpiresIn-60) * time.Second)
	return c.token, nil
}

// gql ejecuta una consulta GraphQL y decodifica data en out.
func (c *Client) gql(ctx context.Context, query string, vars map[string]any, out any) error {
	tok, err := c.getToken(ctx)
	if err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql %d: %s", resp.StatusCode, truncate(string(body), 300))
	}
	var env struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if len(env.Errors) > 0 {
		return errors.New("warcraftlogs: " + env.Errors[0].Message)
	}
	return json.Unmarshal(env.Data, out)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// cached lee/escribe un resultado JSON en disco. Los reportes cerrados son
// inmutables, así que se cachean indefinidamente.
func (c *Client) cached(key string, out any, fetch func() (any, error)) error {
	if c.cacheDir != "" {
		sum := sha1.Sum([]byte(key))
		path := filepath.Join(c.cacheDir, hex.EncodeToString(sum[:])+".json")
		if b, err := os.ReadFile(path); err == nil {
			if json.Unmarshal(b, out) == nil {
				return nil
			}
		}
		v, err := fetch()
		if err != nil {
			return err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_ = os.MkdirAll(c.cacheDir, 0o755)
		_ = os.WriteFile(path, b, 0o644)
		return json.Unmarshal(b, out)
	}
	v, err := fetch()
	if err != nil {
		return err
	}
	b, _ := json.Marshal(v)
	return json.Unmarshal(b, out)
}

// ---- Metadatos del reporte ----

const metaQuery = `
query($code: String!) {
  reportData {
    report(code: $code) {
      title
      startTime
      guild { name }
      fights(killType: Encounters) {
        id name encounterID difficulty kill startTime endTime fightPercentage friendlyPlayers
      }
      masterData {
        abilities { gameID name icon }
      }
    }
  }
}`

type metaResp struct {
	ReportData struct {
		Report struct {
			Title     string `json:"title"`
			StartTime int64  `json:"startTime"`
			Guild     *struct {
				Name string `json:"name"`
			} `json:"guild"`
			Fights     []model.Fight `json:"fights"`
			MasterData struct {
				Abilities []struct {
					GameID int    `json:"gameID"`
					Name   string `json:"name"`
					Icon   string `json:"icon"`
				} `json:"abilities"`
			} `json:"masterData"`
		} `json:"report"`
	} `json:"reportData"`
}

// ReportMeta devuelve título, peleas de encuentro y nombre/icono de habilidades.
func (c *Client) ReportMeta(ctx context.Context, code string) (model.ReportMeta, map[int]model.Ability, error) {
	var r metaResp
	err := c.cached("meta:"+code, &r, func() (any, error) {
		var out metaResp
		if err := c.gql(ctx, metaQuery, map[string]any{"code": code}, &out); err != nil {
			return nil, err
		}
		return out, nil
	})
	if err != nil {
		return model.ReportMeta{}, nil, err
	}
	rep := r.ReportData.Report
	meta := model.ReportMeta{Code: code, Title: rep.Title, StartTime: rep.StartTime, Fights: rep.Fights}
	if rep.Guild != nil {
		meta.Guild = rep.Guild.Name
	}
	abil := make(map[int]model.Ability, len(rep.MasterData.Abilities))
	for _, a := range rep.MasterData.Abilities {
		abil[a.GameID] = model.Ability{Name: a.Name, Icon: a.Icon}
	}
	return meta, abil, nil
}

// ---- Jugadores / roles ----

const playersQuery = `
query($code: String!, $fight: Int!) {
  reportData { report(code: $code) { playerDetails(fightIDs: [$fight]) } }
}`

type playerDetailsResp struct {
	ReportData struct {
		Report struct {
			PlayerDetails json.RawMessage `json:"playerDetails"`
		} `json:"report"`
	} `json:"reportData"`
}

type pdEntry struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Specs []struct {
		Spec  string `json:"spec"`
		Count int    `json:"count"`
	} `json:"specs"`
}

// Players devuelve los jugadores de un intento con clase, spec y rol.
func (c *Client) Players(ctx context.Context, code string, fightID int) ([]model.Player, error) {
	var r playerDetailsResp
	key := fmt.Sprintf("players:%s:%d", code, fightID)
	err := c.cached(key, &r, func() (any, error) {
		var out playerDetailsResp
		if err := c.gql(ctx, playersQuery, map[string]any{"code": code, "fight": fightID}, &out); err != nil {
			return nil, err
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return ParsePlayerDetails(r.ReportData.Report.PlayerDetails)
}

// ParsePlayerDetails interpreta el JSON de playerDetails ({data:{playerDetails:{tanks,healers,dps}}}).
func ParsePlayerDetails(raw json.RawMessage) ([]model.Player, error) {
	var wrap struct {
		Data struct {
			PlayerDetails map[string][]pdEntry `json:"playerDetails"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, fmt.Errorf("playerDetails: %w", err)
	}
	var players []model.Player
	for role, list := range map[string][]pdEntry(wrap.Data.PlayerDetails) {
		r := map[string]string{"tanks": "tank", "healers": "healer", "dps": "dps"}[role]
		if r == "" {
			continue
		}
		for _, e := range list {
			spec, best := "", -1
			for _, s := range e.Specs {
				if s.Count > best {
					spec, best = s.Spec, s.Count
				}
			}
			players = append(players, model.Player{ID: e.ID, Name: e.Name, Class: e.Type, Spec: spec, Role: r})
		}
	}
	return players, nil
}

// ---- Eventos ----

const eventsQuery = `
query($code: String!, $fight: Int!, $start: Float!, $end: Float!, $dt: EventDataType!) {
  reportData {
    report(code: $code) {
      events(fightIDs: [$fight], startTime: $start, endTime: $end, dataType: $dt, limit: 10000, includeResources: true) {
        data
        nextPageTimestamp
      }
    }
  }
}`

type eventsResp struct {
	ReportData struct {
		Report struct {
			Events struct {
				Data              []model.Event `json:"data"`
				NextPageTimestamp *float64      `json:"nextPageTimestamp"`
			} `json:"events"`
		} `json:"report"`
	} `json:"reportData"`
}

// Events descarga todos los eventos de un tipo (Deaths, DamageTaken, Casts...)
// paginando con nextPageTimestamp.
func (c *Client) Events(ctx context.Context, code string, f model.Fight, dataType string) ([]model.Event, error) {
	var all []model.Event
	key := fmt.Sprintf("events:%s:%d:%s", code, f.ID, dataType)
	err := c.cached(key, &all, func() (any, error) {
		var acc []model.Event
		start := float64(f.StartTime)
		for {
			var r eventsResp
			vars := map[string]any{"code": code, "fight": f.ID, "start": start, "end": float64(f.EndTime), "dt": dataType}
			if err := c.gql(ctx, eventsQuery, vars, &r); err != nil {
				return nil, err
			}
			ev := r.ReportData.Report.Events
			acc = append(acc, ev.Data...)
			if ev.NextPageTimestamp == nil {
				break
			}
			start = *ev.NextPageTimestamp
		}
		return acc, nil
	})
	return all, err
}

// LoadFight reúne todos los datos de un intento.
func (c *Client) LoadFight(ctx context.Context, code string, fightID int) (*model.FightData, error) {
	meta, abil, err := c.ReportMeta(ctx, code)
	if err != nil {
		return nil, err
	}
	var fight *model.Fight
	for i := range meta.Fights {
		if meta.Fights[i].ID == fightID {
			fight = &meta.Fights[i]
		}
	}
	if fight == nil {
		return nil, fmt.Errorf("fight %d does not exist in the report", fightID)
	}
	players, err := c.Players(ctx, code, fightID)
	if err != nil {
		return nil, err
	}
	fd := &model.FightData{Report: meta, Fight: *fight, Players: players, Abilities: abil}
	if fd.Deaths, err = c.Events(ctx, code, *fight, "Deaths"); err != nil {
		return nil, err
	}
	if fd.DamageTaken, err = c.Events(ctx, code, *fight, "DamageTaken"); err != nil {
		return nil, err
	}
	if fd.Casts, err = c.Events(ctx, code, *fight, "Casts"); err != nil {
		return nil, err
	}
	return fd, nil
}
