"use client";

import { useMemo, useState } from "react";
import type { AnalysisResult, PlayerSummary } from "@/lib/types";
import { DIFFICULTY, mmss, pct } from "@/lib/format";
import { Badge, Muted, Note, PlayerLabel, Table, Tile, TD, TH, TR } from "@/components/ui";

type SortKey = keyof PlayerSummary | "score";

const score = (p: PlayerSummary) =>
  p.avoidableDeaths * 10 + p.mechanicDeaths * 3 + p.missedSpikes + p.avoidablePct / 10;

export function SummaryTab({ result }: { result: AnalysisResult }) {
  const [sort, setSort] = useState<{ key: SortKey; dir: 1 | -1 }>({ key: "score", dir: -1 });
  const s = result.summary;

  const rows = useMemo(() => {
    const value = (p: PlayerSummary): string | number =>
      sort.key === "score" ? score(p) : (p[sort.key] as string | number);
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
            {header("avoidableDeaths", "Avoidable")}
            {header("mechanicDeaths", "Mechanic")}
            {header("uptimePct", "Uptime")}
            {header("avoidablePct", "Avoidable dmg")}
            {header("cooldownUsePct", "CD use")}
            {header("missedSpikes", "Spikes w/o CD")}
          </tr>
        </thead>
        <tbody>
          {rows.map((p) => (
            <TR key={p.playerId}>
              <TD>
                <PlayerLabel p={p} />
              </TD>
              <TD num>{p.deaths || "–"}</TD>
              <TD num>{p.avoidableDeaths ? <Badge tone="bad">{p.avoidableDeaths}</Badge> : "–"}</TD>
              <TD num>{p.mechanicDeaths ? <Badge tone="warn">{p.mechanicDeaths}</Badge> : "–"}</TD>
              <TD num>{pct(p.uptimePct, 1)}</TD>
              <TD num>{p.role === "tank" ? <Muted>n/a</Muted> : pct(p.avoidablePct, 1)}</TD>
              <TD num>{p.cooldownUsePct ? pct(p.cooldownUsePct) : <Muted>–</Muted>}</TD>
              <TD num>{p.missedSpikes || "–"}</TD>
            </TR>
          ))}
        </tbody>
      </Table>
    </>
  );
}
