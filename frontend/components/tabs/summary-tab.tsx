"use client";

import { useMemo, useState } from "react";
import type { AnalysisResult, PlayerSummary } from "@/lib/types";
import { compact, DIFFICULTY, mmss, pct } from "@/lib/format";
import { Badge, Muted, Note, PlayerLabel, Table, Tile, TD, TH, TR } from "@/components/ui";

type SortKey = keyof PlayerSummary | "score" | "firstDeathMs";

const score = (p: PlayerSummary) =>
  p.avoidableDeaths * 10 + p.mechanicDeaths * 3 + p.missedSpikes + p.avoidablePct / 10;

export function SummaryTab({
  result,
  onGoToDeath,
}: {
  result: AnalysisResult;
  onGoToDeath?: (playerId: number, timeMs: number) => void;
}) {
  const [sort, setSort] = useState<{ key: SortKey; dir: 1 | -1 }>({ key: "score", dir: -1 });
  const s = result.summary;

  const rows = useMemo(() => {
    // Players who never died sort to the end when sorting by first death.
    const value = (p: PlayerSummary): string | number => {
      if (sort.key === "score") return score(p);
      if (sort.key === "firstDeathMs") return p.deathTimestamps[0] ?? Infinity;
      return p[sort.key] as string | number;
    };
    return [...s.players].sort((a, b) => {
      const x = value(a);
      const y = value(b);
      const c = typeof x === "string" ? x.localeCompare(String(y)) : x - (y as number);
      return c * sort.dir;
    });
  }, [s.players, sort]);

  const header = (key: SortKey, label: string, num = true) => (
    <TH
      num={num}
      className="cursor-pointer select-none"
      onClick={() =>
        setSort((cur) => ({
          key,
          dir: cur.key === key ? (cur.dir === 1 ? -1 : 1) : key === "player" ? 1 : -1,
        }))
      }
    >
      {label}
      {sort.key === key ? (sort.dir > 0 ? " ▲" : " ▼") : ""}
    </TH>
  );

  return (
    <>
      <div className="mb-4 grid grid-cols-[repeat(auto-fit,minmax(170px,1fr))] gap-3">
        <Tile value={s.countedDeaths} label="Deaths (counted)" />
        <Tile value={s.avoidableDeaths} label="Avoidable deaths" />
        <Tile value={s.mechanicDeaths} label="Mechanic deaths" />
        <Tile value={pct(s.avgUptimePct)} label="Average uptime" />
        <Tile value={compact(s.totalAvoidableDamage)} label="Avoidable damage (fight)" />
      </div>
      <Note>
        {result.fight.name} · {DIFFICULTY[result.fight.difficulty] ?? ""} · {result.fight.kill ? "Kill" : "Wipe"} ·{" "}
        {mmss(result.fight.durationMs)}. Sorted by “most to review”; click a column header to re-sort.
      </Note>
      <Table>
        <thead>
          <tr>
            {header("player", "Player", false)}
            {header("deaths", "Deaths")}
            {header("firstDeathMs", "Death timer")}
            {header("avoidableDeaths", "Avoidable")}
            {header("mechanicDeaths", "Mechanic")}
            {header("uptimePct", "Uptime")}
            {header("avoidablePct", "Avoidable dmg")}
            {header("cooldownUsePct", "CD use")}
            {header("missedSpikes", "Spikes w/o CD")}
          </tr>
        </thead>
        <tbody>
          {rows.map((p) => {
            const times = p.deathTimestamps;
            const jumpable = times.length > 0 && !!onGoToDeath;
            return (
              <TR
                key={p.playerId}
                className={jumpable ? "cursor-pointer hover:bg-track/50" : undefined}
                onClick={jumpable ? () => onGoToDeath!(p.playerId, times[0]) : undefined}
                title={jumpable ? "Jump to the first death in the Deaths tab" : undefined}
              >
                <TD>
                  <PlayerLabel p={p} />
                </TD>
                <TD num>{p.deaths || "–"}</TD>
                <TD num className="text-fg-3">
                  {times.length ? (
                    times.map((t, i) => (
                      <span key={t}>
                        {i > 0 && ", "}
                        <button
                          type="button"
                          className="underline decoration-dotted underline-offset-2 hover:text-fg"
                          title="Jump to this death in the Deaths tab"
                          onClick={(e) => {
                            e.stopPropagation();
                            onGoToDeath?.(p.playerId, t);
                          }}
                        >
                          {mmss(t)}
                        </button>
                      </span>
                    ))
                  ) : (
                    "–"
                  )}
                </TD>
                <TD num>{p.avoidableDeaths ? <Badge tone="bad">{p.avoidableDeaths}</Badge> : "–"}</TD>
                <TD num>{p.mechanicDeaths ? <Badge tone="warn">{p.mechanicDeaths}</Badge> : "–"}</TD>
                <TD num>{pct(p.uptimePct, 1)}</TD>
                <TD num>
                  {p.role === "tank" ? (
                    <Muted>n/a</Muted>
                  ) : p.avoidableDamage ? (
                    <>
                      {compact(p.avoidableDamage)} <Muted>({pct(p.avoidablePct, 1)})</Muted>
                    </>
                  ) : (
                    "–"
                  )}
                </TD>
                <TD num>{p.cooldownUsePct ? pct(p.cooldownUsePct) : <Muted>–</Muted>}</TD>
                <TD num>{p.missedSpikes || "–"}</TD>
              </TR>
            );
          })}
        </tbody>
      </Table>
    </>
  );
}
