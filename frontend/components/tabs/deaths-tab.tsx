import type { AnalysisResult, DeathReport } from "@/lib/types";
import { compact, mmss, pct } from "@/lib/format";
import { AbilityLabel, Badge, Note, PlayerLabel } from "@/components/ui";

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

const list = (items: string[]) => (items.length ? items.join(", ") : "none");

export function DeathsTab({ result }: { result: AnalysisResult }) {
  if (!result.deaths.length) return <Note>No deaths in this fight.</Note>;
  return (
    <>
      {result.deaths.map((d, i) => (
        <article
          key={`${d.playerId}-${d.timeMs}-${i}`}
          className={`mb-2.5 rounded-xl border border-line bg-surface px-3.5 py-3 ${d.ignored ? "opacity-60" : ""}`}
        >
          <div className="flex flex-wrap items-baseline gap-2.5">
            <span className="tabular-nums text-fg-2">{mmss(d.timeMs)}</span>
            <PlayerLabel p={d} />
            <span>
              <DeathBadges d={d} />
            </span>
          </div>
          <div className="my-1.5">{d.verdict}</div>
          <div className="grid grid-cols-1 gap-x-3 gap-y-0.5 text-[13px] text-fg-2 sm:grid-cols-[190px_1fr] [&>span:nth-child(even)]:text-fg">
            <Row label="Killing blow">
              <b>
                {d.killingAbility ? (
                  <AbilityLabel name={d.killingAbility} icon={d.killingAbilityIcon} abilityId={d.killingAbilityId} />
                ) : (
                  "?"
                )}
              </b>
            </Row>
            <Row label={`Damage in the last ${d.windowSec}s`}>
              <b>{pct(d.windowPct)} of max HP</b>
            </Row>
            <Row label="Damage sources">
              {d.topSources.length ? (
                <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
                  {d.topSources.map((s) => (
                    <span key={s.abilityId} className="inline-flex items-center gap-1">
                      <AbilityLabel name={s.ability} icon={s.abilityIcon} abilityId={s.abilityId} />
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
              <b>{list(d.defensivesAvailable)}</b>
            </Row>
            <Row label="Defensives used (recent)">{list(d.defensivesUsed)}</Row>
            <Row label="HS / potion available">
              <b>{list(d.consumablesAvailable)}</b>
            </Row>
            <Row label="External / raid CDs">
              {d.externalsAvailable.length
                ? d.externalsAvailable.map((e) => `${e.ability} (${e.caster})`).join(", ")
                : "none"}
            </Row>
          </div>
        </article>
      ))}
    </>
  );
}
