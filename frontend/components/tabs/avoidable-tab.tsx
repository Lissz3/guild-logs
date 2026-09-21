import type { AnalysisResult } from "@/lib/types";
import { compact, pct } from "@/lib/format";
import { AbilityLabel, Badge, Muted, Note, PlayerLabel, Table, TD, TH, TR } from "@/components/ui";

export function AvoidableTab({ result }: { result: AnalysisResult }) {
  const players = result.avoidableDamage.filter((p) => p.rows.length);

  return (
    <>
      <Note>
        Non-tanks only. <b>avoidable</b> = ability listed as avoidable in the encounter config. <b>possible</b> = the
        player took far more than the raid median from an ability that hit several players; worth checking by hand,
        it is not a verdict.
      </Note>

      <h3 className="mt-4 mb-2 text-[15px] font-semibold">By ability</h3>
      <Table>
        <thead>
          <tr>
            <TH>Ability</TH>
            <TH num>Total damage</TH>
            <TH num>Players hit</TH>
            <TH num>Median</TH>
            <TH>Most affected</TH>
          </tr>
        </thead>
        <tbody>
          {result.avoidableByAbility.length ? (
            result.avoidableByAbility.map((a) => (
              <TR key={a.abilityId}>
                <TD>
                  <AbilityLabel name={a.ability} icon={a.abilityIcon} abilityId={a.abilityId} />{" "}
                  {a.known ? <Badge tone="bad">config</Badge> : <Badge tone="info">statistical</Badge>}
                </TD>
                <TD num>{compact(a.total)}</TD>
                <TD num>{a.hitters}</TD>
                <TD num>{compact(a.median)}</TD>
                <TD>
                  {a.worst} <Muted>{compact(a.worstAmount)}</Muted>
                </TD>
              </TR>
            ))
          ) : (
            <TR>
              <TD colSpan={5}>
                <Muted>Nothing notable.</Muted>
              </TD>
            </TR>
          )}
        </tbody>
      </Table>

      <h3 className="mt-4 mb-2 text-[15px] font-semibold">By player</h3>
      {players.length ? (
        players.map((p) => (
          <div key={p.playerId} className="mb-4">
            <h4 className="mb-2 font-semibold">
              <PlayerLabel p={p} />
              <span className="ml-1.5 text-[13px] font-normal text-fg-2">
                {compact(p.knownTaken + p.possibleTaken)} of {compact(p.totalTaken)} ({pct(p.avoidablePct, 1)})
              </span>
            </h4>
            <Table>
              <thead>
                <tr>
                  <TH>Ability</TH>
                  <TH num>Damage</TH>
                  <TH num>Hits</TH>
                  <TH num>Raid median</TH>
                  <TH num>× median</TH>
                  <TH num>% of their damage</TH>
                </tr>
              </thead>
              <tbody>
                {p.rows.map((x) => (
                  <TR key={x.abilityId}>
                    <TD>
                      <AbilityLabel name={x.ability} icon={x.abilityIcon} abilityId={x.abilityId} />{" "}
                      {x.known ? <Badge tone="bad">avoidable</Badge> : <Badge tone="info">possible</Badge>}
                    </TD>
                    <TD num>{compact(x.amount)}</TD>
                    <TD num>{x.hits}</TD>
                    <TD num>{compact(x.raidMedian)}</TD>
                    <TD num>{x.ratio ? `${x.ratio.toFixed(1)}×` : "–"}</TD>
                    <TD num>{pct(x.sharePct, 1)}</TD>
                  </TR>
                ))}
              </tbody>
            </Table>
          </div>
        ))
      ) : (
        <Note>Nobody stands out in avoidable damage.</Note>
      )}
    </>
  );
}
