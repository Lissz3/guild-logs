"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { analyze, fightFromUrl, getConfig, getReport } from "@/lib/api";
import { DIFFICULTY, mmss } from "@/lib/format";
import type { AnalysisOptions, AnalysisResult, ReportMeta } from "@/lib/types";
import { Card, InfoTip } from "@/components/ui";
import { SummaryTab } from "@/components/tabs/summary-tab";
import { DeathsTab, type DeathHighlight } from "@/components/tabs/deaths-tab";
import { AvoidableTab } from "@/components/tabs/avoidable-tab";
import { ActivityTab } from "@/components/tabs/activity-tab";
import { CooldownsTab } from "@/components/tabs/cooldowns-tab";

const TABS = [
  { id: "summary", label: "Summary" },
  { id: "deaths", label: "Deaths" },
  { id: "avoidable", label: "Avoidable damage" },
  { id: "activity", label: "Activity" },
  { id: "cooldowns", label: "Defensives" },
] as const;
type TabId = (typeof TABS)[number]["id"];

const DEFAULT_OPTIONS: AnalysisOptions = {
  window: 5,
  mechanicPct: 60,
  spikePct: 50,
  gapSec: 2.5,
  wipeIgnoreDeaths: 3,
  assumeTalents: false,
};

const input =
  "min-w-0 rounded-md border border-line bg-bg px-2.5 py-2 text-fg outline-none focus:border-series";
const button =
  "rounded-md border border-series bg-series px-3.5 py-2 text-white hover:brightness-110 disabled:opacity-60";
const buttonGhost = "rounded-md border border-series px-3.5 py-2 text-series hover:brightness-110 disabled:opacity-60";
const buttonBig =
  "rounded-md border border-series bg-series px-10 py-3 text-base font-semibold text-white hover:brightness-110 disabled:opacity-60";

function Field({
  label,
  hint,
  children,
  className = "",
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <label className={`flex flex-col gap-1 text-[13px] text-fg-2 ${className}`}>
      <span className="inline-flex items-center gap-1">
        {label}
        {hint && <InfoTip text={hint} />}
      </span>
      {children}
    </label>
  );
}

export function Analyzer() {
  const [codeInput, setCodeInput] = useState("");
  const [status, setStatus] = useState({ msg: "", error: false });
  const [meta, setMeta] = useState<ReportMeta | null>(null);
  const [fightId, setFightId] = useState<number | null>(null);
  const [opts, setOpts] = useState<AnalysisOptions>(DEFAULT_OPTIONS);
  const [result, setResult] = useState<AnalysisResult | null>(null);
  const [tab, setTab] = useState<TabId>("summary");
  const [highlightDeath, setHighlightDeath] = useState<DeathHighlight | null>(null);
  const [busy, setBusy] = useState(false);

  const goToDeath = useCallback((playerId: number, timeMs: number) => {
    setTab("deaths");
    setHighlightDeath({ playerId, timeMs });
  }, []);
  const clearHighlightDeath = useCallback(() => setHighlightDeath(null), []);
  const request = useRef(0); // ignore responses from superseded requests

  useEffect(() => {
    getConfig()
      .then((c) => {
        setOpts((o) => ({
          ...o,
          // ?? keeps the current value when the server response is missing a field
          // (e.g. an older Go build without a newer threshold) instead of blanking
          // the input to undefined, which would break React's controlled inputs.
          window: c.thresholds.windowSec ?? o.window,
          mechanicPct: c.thresholds.mechanicPct ?? o.mechanicPct,
          spikePct: c.thresholds.spikePct ?? o.spikePct,
          gapSec: c.thresholds.gapSec ?? o.gapSec,
          wipeIgnoreDeaths: c.thresholds.wipeIgnoreAfterDeaths ?? o.wipeIgnoreDeaths,
        }));
        if (!c.wclConfigured) {
          setStatus({
            msg: "The server has no Warcraft Logs credentials: only the sample report works.",
            error: false,
          });
        }
      })
      .catch(() =>
        setStatus({
          msg: "Cannot reach the Go server. Is it running (go run ./cmd/guildlogs)?",
          error: true,
        }),
      );
  }, []);

  async function loadReport(input: string) {
    const id = ++request.current;
    setBusy(true);
    setStatus({ msg: "Loading fights…", error: false });
    setMeta(null);
    setResult(null);
    try {
      const m = await getReport(input);
      if (id !== request.current) return;
      setMeta(m);
      const wanted = fightFromUrl(input);
      setFightId(m.fights.find((f) => f.id === wanted)?.id ?? m.fights[0]?.id ?? null);
      setStatus({ msg: `${m.fights.length} encounter fights found.`, error: false });
    } catch (e) {
      if (id === request.current) setStatus({ msg: (e as Error).message, error: true });
    } finally {
      if (id === request.current) setBusy(false);
    }
  }

  async function runAnalysis() {
    if (!meta || fightId === null) return;
    const id = ++request.current;
    setBusy(true);
    setStatus({ msg: "Analyzing… (the first time can take a while: every event is downloaded)", error: false });
    try {
      const r = await analyze(meta.code, fightId, opts);
      if (id !== request.current) return;
      setResult(r);
      setStatus({ msg: "", error: false });
    } catch (e) {
      if (id === request.current) setStatus({ msg: (e as Error).message, error: true });
    } finally {
      if (id === request.current) setBusy(false);
    }
  }

  const num = (key: keyof AnalysisOptions, min: number, max: number, step: number, label: string, hint: string) => (
    <Field label={label} hint={hint}>
      <input
        className={`${input} w-full`}
        type="number"
        min={min}
        max={max}
        step={step}
        value={(opts[key] as number) ?? 0}
        onChange={(e) => setOpts({ ...opts, [key]: Number(e.target.value) })}
      />
    </Field>
  );

  return (
    <>
      <Card>
        <form
          className="flex flex-col items-stretch gap-3 md:flex-row md:items-end"
          onSubmit={(e) => {
            e.preventDefault();
            const v = codeInput.trim();
            if (v) void loadReport(v);
          }}
        >
          <Field label="Warcraft Logs report (URL or code)" className="min-w-[220px] flex-1">
            <input
              className={input}
              type="text"
              value={codeInput}
              onChange={(e) => setCodeInput(e.target.value)}
              placeholder="https://www.warcraftlogs.com/reports/AbCdEf1234567890"
              autoComplete="off"
            />
          </Field>
          <button className={button} type="submit" disabled={busy}>
            Load fights
          </button>
          <button
            className={buttonGhost}
            type="button"
            disabled={busy}
            onClick={() => {
              setCodeInput("demo");
              void loadReport("demo");
            }}
          >
            Try the sample data
          </button>
        </form>
        <div role="status" className={`mt-2.5 min-h-[1.2em] ${status.error ? "text-bad" : "text-fg-2"}`}>
          {status.msg}
        </div>
      </Card>

      {meta && (
        <Card>
          <h2 className="mb-3 text-lg font-semibold">
            {meta.title}
            {meta.guild ? ` · ${meta.guild}` : ""}
          </h2>
          <form
            className="flex flex-col gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              void runAnalysis();
            }}
          >
            <Field label="Fight" className="w-full">
              <select
                className={input}
                value={fightId ?? ""}
                onChange={(e) => setFightId(Number(e.target.value))}
              >
                {meta.fights.map((f) => (
                  <option key={f.id} value={f.id}>
                    #{f.id} {f.name} · {DIFFICULTY[f.difficulty] ?? ""} ·{" "}
                    {f.kill ? "Kill" : `Wipe ${(f.fightPercentage ?? 0).toFixed(0)}%`} · {mmss(f.endTime - f.startTime)}
                  </option>
                ))}
              </select>
            </Field>
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
              {num(
                "window",
                1,
                30,
                1,
                "Lookback window (s)",
                "How many seconds of damage before a death to look at when picking the killing blow's top sources and deciding the verdict.",
              )}
              {num(
                "mechanicPct",
                10,
                200,
                5,
                "Mechanic death (% HP)",
                "A death is flagged as a mechanic death when the player took at least this % of their max HP during the lookback window.",
              )}
              {num(
                "spikePct",
                10,
                200,
                5,
                "Damage spike (% HP)",
                "A window where the player took at least this % of max HP counts as a damage spike, used to check whether a personal defensive was available and went unused.",
              )}
              {num(
                "gapSec",
                1,
                20,
                0.5,
                "Min. gap (s)",
                "In the Activity tab, downtime gaps at least this long are listed individually.",
              )}
              {num(
                "wipeIgnoreDeaths",
                0,
                40,
                1,
                "Ignore deaths after",
                "On a wipe (not a kill), deaths from this Nth one onward are excluded from counts and verdicts — the attempt is assumed already lost by then. 0 or empty: don't filter, count every death.",
              )}
              <label className="flex items-center gap-1.5 self-end pb-2 text-[13px] text-fg-2">
                <input
                  type="checkbox"
                  checked={opts.assumeTalents}
                  onChange={(e) => setOpts({ ...opts, assumeTalents: e.target.checked })}
                />
                <span className="inline-flex items-center gap-1">
                  Assume talents
                  <InfoTip text="Count talent-gated defensives as owned even if the player never used them in this fight." />
                </span>
              </label>
            </div>
            <div className="flex justify-center pt-2">
              <button className={buttonBig} type="submit" disabled={busy || fightId === null}>
                Analyze
              </button>
            </div>
          </form>
        </Card>
      )}

      {result && (
        <div>
          {result.warnings.map((w) => (
            <div key={w} className="mb-2.5 rounded-lg bg-warn-bg px-3 py-2 text-sm text-warn">
              {w}
            </div>
          ))}
          <nav className="mb-4 flex gap-1 overflow-x-auto border-b border-line" role="tablist">
            {TABS.map((t) => (
              <button
                key={t.id}
                role="tab"
                aria-selected={tab === t.id}
                onClick={() => setTab(t.id)}
                className={`border-b-2 px-3.5 py-2.5 whitespace-nowrap ${
                  tab === t.id ? "border-series font-semibold text-fg" : "border-transparent text-fg-2 hover:text-fg"
                }`}
              >
                {t.label}
              </button>
            ))}
          </nav>
          {tab === "summary" && <SummaryTab result={result} onGoToDeath={goToDeath} />}
          {tab === "deaths" && (
            <DeathsTab result={result} highlight={highlightDeath} onHighlightShown={clearHighlightDeath} />
          )}
          {tab === "avoidable" && <AvoidableTab result={result} />}
          {tab === "activity" && <ActivityTab result={result} />}
          {tab === "cooldowns" && <CooldownsTab result={result} />}
        </div>
      )}
    </>
  );
}
