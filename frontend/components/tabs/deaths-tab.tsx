import { useEffect } from "react";
import type { AnalysisResult, DeathReport } from "@/lib/types";
import { compact, mmss, pct } from "@/lib/format";
import { AbilityLabel, AbilityRefList, Badge, Note, PlayerLabel } from "@/components/ui";

export function deathElementId(playerId: number, timeMs: number) {
  return `death-${playerId}-${timeMs}`;
}

export interface DeathHighlight {
  playerId: number;
  timeMs: number;
}

const ICON_SIZE = 24;

function DeathBadges({ d }: { d: DeathReport }) {
  return (
    <>
      {d.flags.includes("avoidable") && <Badge tone="bad">Avoidable</Badge>}
      {d.flags.includes("mechanic_high") ? (
        <Badge tone="warn">Mechanic (high)</Badge>
      ) : (
        d.flags.includes("mechanic") && <Badge tone="warn">Mechanic</Badge>
      )}
      {d.flags.includes("avoidable_damage") && <Badge tone="info">Avoidable damage</Badge>}
      {d.ignored && <Badge tone="ok">After wipe</Badge>}
    </>
  );
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <>
      <span>{label}</span>
      <span>{children}</span>
    </>
  );
}

export function DeathsTab({
  result,
  highlight,
  onHighlightShown,
}: {
  result: AnalysisResult;
  highlight?: DeathHighlight | null;
  onHighlightShown?: () => void;
}) {
  useEffect(() => {
    if (!highlight) return;
    document.getElementById(deathElementId(highlight.playerId, highlight.timeMs))?.scrollIntoView({
      behavior: "smooth",
      block: "center",
    });
    // Keep the highlight visible briefly, then let it fade — otherwise it'd
    // reappear every time this tab remounts (e.g. clicking the Deaths tab
    // again later) instead of just once, right after a summary-row click.
    const t = setTimeout(() => onHighlightShown?.(), 2000);
    return () => clearTimeout(t);
  }, [highlight, onHighlightShown]);

  if (!result.deaths.length) return <Note>No deaths in this fight.</Note>;
  return (
    <>
      {result.deaths.map((d, i) => {
        const isHighlighted = highlight?.playerId === d.playerId && highlight?.timeMs === d.timeMs;
        return (
          <article
            key={`${d.playerId}-${d.timeMs}-${i}`}
            id={deathElementId(d.playerId, d.timeMs)}
            className={`mb-3 rounded-xl border px-5 py-4 transition-colors ${d.ignored ? "opacity-60" : ""} ${
              isHighlighted ? "border-series bg-series/10" : "border-line bg-surface"
            }`}
          >
            <div className="flex flex-wrap items-center gap-3">
              <PlayerLabel p={d} />
              <span>
                <DeathBadges d={d} />
              </span>
              <span className="ml-auto text-lg font-semibold tabular-nums text-fg-2">{mmss(d.timeMs)}</span>
            </div>
            <div className="mt-3 grid grid-cols-1 gap-x-4 gap-y-2 text-[15px] text-fg-2 sm:grid-cols-[220px_1fr] [&>span:nth-child(even)]:text-fg">
              <Row label={`Damage in the last ${d.windowSec}s`}>
                <b>{pct(d.windowPct)} of max HP</b>
              </Row>
              <Row label="Killing blow">
                <b>
                  {d.killingAbility ? (
                    <AbilityLabel
                      name={d.killingAbility}
                      icon={d.killingAbilityIcon}
                      abilityId={d.killingAbilityId}
                      size={ICON_SIZE}
                    />
                  ) : (
                    "?"
                  )}
                </b>
              </Row>
              <Row label="Damage sources">
                {d.topSources.length ? (
                  <span className="flex flex-col gap-1.5">
                    {d.topSources.map((s) => (
                      <span key={s.abilityId} className="inline-flex items-center gap-1.5">
                        <AbilityLabel name={s.ability} icon={s.abilityIcon} abilityId={s.abilityId} size={ICON_SIZE} />
                        <span>
                          {compact(s.amount)} ({pct(s.pctMaxHp)}){s.avoidable ? " ⚠ avoidable" : ""}
                        </span>
                      </span>
                    ))}
                  </span>
                ) : (
                  "–"
                )}
              </Row>
              <Row label="Defensives available">
                <b>
                  <AbilityRefList items={d.defensivesAvailable} size={ICON_SIZE} />
                </b>
              </Row>
              <Row label="Defensives used (recent)">
                <AbilityRefList items={d.defensivesUsed} size={ICON_SIZE} />
              </Row>
              <Row label="HS / potion available">
                <b>
                  <AbilityRefList items={d.consumablesAvailable} size={ICON_SIZE} />
                </b>
              </Row>
            </div>
          </article>
        );
      })}
    </>
  );
}
