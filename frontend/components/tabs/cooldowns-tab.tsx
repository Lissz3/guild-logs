import type { AnalysisResult } from "@/lib/types";
import { KIND_LABEL, mmss, pct } from "@/lib/format";
import { AbilityLabel, AbilityRefList, Badge, Muted, Note, PlayerLabel, Table, TD, TH, TR } from "@/components/ui";

export function CooldownsTab({ result }: { result: AnalysisResult }) {
  const list = result.cooldowns.filter((p) => p.cooldowns.length || p.spikes.length);

  return (
    <>
      <Note>
        “Possible” assumes the cooldown is ready at the pull and used again as soon as it recharges (no cooldown
        reductions). A <b>spike</b> is a window with damage ≥ the configured threshold; it counts as “unused” when the
        player had personal defensives available and used none of them.
      </Note>
      {list.length === 0 && <Note>No cooldown data for this fight.</Note>}
      {list.map((p) => (
        <div key={p.playerId} className="mb-4">
          <h3 className="mb-2 font-semibold">
            <PlayerLabel p={p} />
            {p.usePct > 0 && (
              <span className="ml-1.5 text-[13px] font-normal text-fg-2">
                personal defensive use {pct(p.usePct)}
              </span>
            )}
          </h3>

          {p.cooldowns.length > 0 && (
            <Table>
              <thead>
                <tr>
                  <TH>Cooldown</TH>
                  <TH>Type</TH>
                  <TH num>CD</TH>
                  <TH num>Used / possible</TH>
                  <TH num>Unused at spikes</TH>
                  <TH>When</TH>
                </tr>
              </thead>
              <tbody>
                {p.cooldowns.map((c) => (
                  <TR key={c.name}>
                    <TD>
                      <AbilityLabel name={c.name} icon={c.abilityIcon} abilityId={c.abilityId} />
                    </TD>
                    <TD>{KIND_LABEL[c.kind]}</TD>
                    <TD num>{c.cooldown}s</TD>
                    <TD num>
                      {c.used} / {c.possible}
                    </TD>
                    <TD num>{c.missedOpportunities ? <Badge tone="bad">{c.missedOpportunities}</Badge> : "–"}</TD>
                    <TD>{c.usedAtMs.length ? c.usedAtMs.map(mmss).join(", ") : <Muted>never</Muted>}</TD>
                  </TR>
                ))}
              </tbody>
            </Table>
          )}

          {p.spikes.length > 0 && (
            <div className="mt-2">
              <Table>
                <thead>
                  <tr>
                    <TH>Spike</TH>
                    <TH num>% HP</TH>
                    <TH>Main damage</TH>
                    <TH>Available, unused</TH>
                    <TH>Used</TH>
                  </tr>
                </thead>
                <tbody>
                  {p.spikes.map((s) => (
                    <TR key={s.startMs}>
                      <TD>
                        {mmss(s.startMs)} {s.died && <Badge tone="bad">died</Badge>}
                      </TD>
                      <TD num>{pct(s.pct)}</TD>
                      <TD>
                        <AbilityLabel name={s.top} icon={s.topIcon} abilityId={s.topAbilityId} />
                      </TD>
                      <TD>
                        <b>
                          <AbilityRefList items={s.available} empty="–" />
                        </b>
                      </TD>
                      <TD>
                        <AbilityRefList items={s.used} empty="–" />
                      </TD>
                    </TR>
                  ))}
                </tbody>
              </Table>
            </div>
          )}
        </div>
      ))}
    </>
  );
}
