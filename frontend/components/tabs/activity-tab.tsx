import type { AnalysisResult, PlayerActivity } from "@/lib/types";
import { mmss, seconds } from "@/lib/format";
import { Muted, Note, PlayerLabel, Table, TD, TH, TR, UptimeBar } from "@/components/ui";

function Timeline({ a, duration }: { a: PlayerActivity; duration: number }) {
  return (
    <svg
      viewBox={`0 0 ${duration} 10`}
      preserveAspectRatio="none"
      role="img"
      aria-label={`Activity timeline of ${a.player}`}
      className="block h-[22px] w-full"
    >
      <rect x={0} y={0} width={duration} height={10} className="fill-series opacity-35" />
      {a.gaps.map((g) => (
        <rect key={g.startMs} x={g.startMs} y={0} width={g.endMs - g.startMs} height={10} className="fill-gap">
          <title>{`Gap ${mmss(g.startMs)}–${mmss(g.endMs)} (${seconds(g.endMs - g.startMs)})`}</title>
        </rect>
      ))}
      {a.deadFromMs >= 0 && (
        <rect
          x={a.deadFromMs}
          y={0}
          width={Math.max(duration - a.deadFromMs, 1)}
          height={10}
          className="fill-bad"
        >
          <title>{`Dead from ${mmss(a.deadFromMs)}`}</title>
        </rect>
      )}
    </svg>
  );
}

function LegendItem({ className, children }: { className: string; children: React.ReactNode }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className={`inline-block h-2 w-3 rounded-sm ${className}`} />
      {children}
    </span>
  );
}

export function ActivityTab({ result }: { result: AnalysisResult }) {
  const duration = result.fight.durationMs;
  return (
    <>
      <Note>
        Uptime = time spent casting / time alive, approximated with the GCD (no haste or channels). Stretches where
        almost the whole raid is idle (phases, transitions) are not penalized.
      </Note>
      <div className="mb-2.5 flex flex-wrap gap-4 text-[13px] text-fg-2">
        <LegendItem className="bg-series opacity-50">Active</LegendItem>
        <LegendItem className="bg-gap">Gap ≥ minimum (idle)</LegendItem>
        <LegendItem className="bg-bad">Dead</LegendItem>
      </div>
      <Table>
        <thead>
          <tr>
            <TH>Player</TH>
            <TH className="w-[170px]">Uptime</TH>
            <TH num>Downtime</TH>
            <TH num>Excused</TH>
            <TH className="w-[38%]">Timeline ({mmss(duration)})</TH>
          </tr>
        </thead>
        <tbody>
          {result.activity.map((a) => {
            const longest = [...a.gaps].sort((x, y) => y.endMs - y.startMs - (x.endMs - x.startMs)).slice(0, 3);
            return (
              <TR key={a.playerId}>
                <TD>
                  <PlayerLabel p={a} />
                </TD>
                <TD>
                  <UptimeBar value={a.uptimePct} />
                </TD>
                <TD num>{seconds(a.downtimeMs)}</TD>
                <TD num className="text-fg-3">
                  {seconds(a.excusedMs)}
                </TD>
                <TD>
                  <Timeline a={a} duration={duration} />
                  {longest.length > 0 && (
                    <div className="text-xs">
                      <Muted>{longest.map((g) => `${mmss(g.startMs)} (${seconds(g.endMs - g.startMs, 0)})`).join(" · ")}</Muted>
                    </div>
                  )}
                </TD>
              </TR>
            );
          })}
        </tbody>
      </Table>
    </>
  );
}
