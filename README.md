# Guild Logs

A Warcraft Logs report analyzer for our guild, in the spirit of WoWAnalyzer. Go backend (standard library only) and a Next.js frontend.

## What it analyzes

| Analysis | How it is computed |
|---|---|
| **Avoidable deaths** | At the moment of death, the player had personal defensive(s) and/or Healthstone/potion available (based on their cast history and the cooldown). |
| **Mechanic deaths** | The player took >= 60% (configurable; >= 80% is flagged "high") of their max HP in the N seconds before dying (default 5). |
| **Avoidable damage** | Non-tanks only. Abilities listed as avoidable per encounter in the config, plus statistical detection: damage >= 2x the raid median on an ability that hit at least 4 players. |
| **Uptime / downtime** | Casts + GCD merged over the time the player was alive. Gaps >= 2.5s are listed. Stretches where almost the whole raid is idle (phases/transitions) are not penalized. |
| **Defensives used vs possible** | Actual uses vs possible uses given the cooldown, plus **damage spikes** (>= 50% HP within the window) where a personal defensive was available and went unused. |

## Usage

Requirements: Go 1.22 or newer, and Node.js 20.9 or newer with [pnpm](https://pnpm.io/installation) (`npm install -g pnpm`) for the web UI.

The app is two processes: the **Go API** (port 8080) does the Warcraft Logs work and the analysis, and the **Next.js frontend** (port 3000) is the UI. The browser only talks to Next.js, which proxies `/api/*` to Go.

1. Create a Warcraft Logs API client at https://www.warcraftlogs.com/api/clients to get a client ID and secret.
2. Copy `config.example.json` to `config.json` and fill it in:

   ```json
   {
     "port": "8080",
     "wclClientId": "PASTE_YOUR_CLIENT_ID",
     "wclClientSecret": "PASTE_YOUR_CLIENT_SECRET",
     "cacheDir": ".cache",
     "analysisConfig": ""
   }
   ```

3. Start the Go API (terminal 1):

   ```bash
   go run ./cmd/guildlogs          # API on http://localhost:8080
   ```

   Optional, while developing: use [Air](https://github.com/air-verse/air) to rebuild and restart the API every time you save a `.go` file or a JSON config:

   ```bash
   go install github.com/air-verse/air@latest   # once; recent Air versions ask for Go 1.25+
   air                                          # uses .air.toml
   ```

   Make sure Go's bin folder is on your `PATH` (`%USERPROFILE%\go\bin` on Windows, `$(go env GOPATH)/bin` elsewhere). `.air.toml` is set up for Windows; the two lines to change on macOS/Linux are in its header comment.

4. Start the frontend (terminal 2), then open http://localhost:3000:

   ```bash
   cd frontend
   pnpm install                    # first time only
   pnpm dev                        # development, hot reload
   # or, for a production build:
   pnpm build && pnpm start
   ```

If the Go API is not on `http://localhost:8080`, set `GUILDLOGS_API_URL` for the frontend (see `frontend/.env.example`).

http://localhost:8080 only shows a short text note: the UI lives in the Next.js app.

Paste a report URL or code into the page, pick a fight, and hit "Analyze". Without credentials, only the `demo` report (synthetic data) works, which is handy for trying out the UI.

`config.json` holds your secret and is in `.gitignore`: do not commit it.

### Server settings

The server reads `config.json` from the working directory, or another file passed with `-config path/to/file.json`. A missing default file is fine; a missing file passed with `-config` is an error, and unknown keys are rejected so typos do not go unnoticed. Environment variables override the file:

| File key | Environment variable | Default |
|---|---|---|
| `port` | `PORT` | `8080` |
| `wclClientId` | `WCL_CLIENT_ID` | none |
| `wclClientSecret` | `WCL_CLIENT_SECRET` | none |
| `cacheDir` | `GUILDLOGS_CACHE` | `.cache` (downloaded events are stored on disk, so a report is only fetched once) |
| `analysisConfig` | `GUILDLOGS_CONFIG` | embedded `default.json` (a path here fully replaces it) |

Run the tests with `go test ./...`.

## Project layout

```
.air.toml               Air live-reload config (optional dev tool)
cmd/guildlogs/          HTTP server (endpoints: /api/config, /api/report, /api/analyze)
internal/settings/      Server settings loader (config.json + environment overrides)
internal/wcl/           Warcraft Logs API v2 client (OAuth, GraphQL, pagination, disk cache)
internal/analysis/      Analyzers: deaths, avoidable damage, activity, cooldowns
internal/config/        Embedded default.json + loader
internal/demo/          Deterministic synthetic fight (demo report and test fixture)
internal/model/         Shared types
frontend/               Next.js (App Router, TypeScript, Tailwind CSS) UI, managed with pnpm
```

## Configuration (`internal/config/default.json`)

* `defensives`: defensives per class/spec, with `kind` = `personal` | `external` | `raid`, cooldown, charges, and `talent` (whether it depends on a talent).
* `avoidable.byEncounter`: `{"<encounterID>": [abilityID, ...]}` with the avoidable mechanics of each boss. **This is what improves accuracy the most**: fill it in every raid tier.
* `thresholds`, `consumables`, `activity`.

**Review spell IDs and cooldowns every season.** Defensives are matched by id *or* by English name, so an id change will not break anything, but cooldowns and talents do change between patches.

## Known limitations

* Talents are not read from the log: a talent-gated defensive only counts as "owned" if the player used it during the fight (or if you tick "Assume talents").
* Cooldown reductions and procs that reset cooldowns are not modeled.
* Healthstone/potion: everyone is assumed to carry a potion, and a Healthstone is assumed to be available if there is a Warlock in the raid; one use per combat.
* Uptime is approximated with the class GCD; haste and long channels are not considered (a channeler may show false downtime).
* The GraphQL queries in `internal/wcl` were written against the public v2 API schema and have not yet been run against the live API. If Warcraft Logs changes or renames a field, that package is the only place to touch.
