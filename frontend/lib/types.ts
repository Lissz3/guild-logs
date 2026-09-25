// Types mirroring the JSON returned by the Go API (internal/analysis/types.go
// and internal/model/model.go). All times are milliseconds from the start of
// the fight unless noted.

export interface Fight {
  id: number;
  name: string;
  encounterID: number;
  difficulty: number;
  kill: boolean;
  startTime: number; // absolute (report-relative) ms
  endTime: number;
  fightPercentage: number;
}

export interface ReportMeta {
  code: string;
  title: string;
  guild: string;
  fights: Fight[];
}

export interface ServerConfig {
  thresholds: {
    windowSec: number;
    mechanicPct: number;
    spikePct: number;
    gapSec: number;
    wipeIgnoreAfterDeaths: number;
  };
  wclConfigured: boolean;
}

export type Role = "tank" | "healer" | "dps";

export interface PlayerRef {
  playerId: number;
  player: string;
  class: string;
  spec: string;
  role: Role;
}

export interface AbilityRef {
  name: string;
  abilityId: number;
  abilityIcon: string;
  /** "item" links to wowhead.com/item=; anything else (usually absent) links to wowhead.com/spell=. */
  abilityKind?: string;
}

export interface AbilityDamage {
  ability: string;
  abilityId: number;
  abilityIcon: string;
  amount: number;
  pctMaxHp: number;
  hits: number;
  avoidable: boolean;
}

export type DeathFlag = "avoidable" | "mechanic" | "mechanic_high" | "avoidable_damage";

export interface DeathReport extends PlayerRef {
  timeMs: number;
  killingAbility: string;
  killingAbilityIcon: string;
  killingAbilityId: number;
  windowPct: number;
  windowSec: number;
  topSources: AbilityDamage[];
  avoidableShare: number;
  defensivesAvailable: AbilityRef[];
  defensivesUsed: AbilityRef[];
  consumablesAvailable: AbilityRef[];
  flags: DeathFlag[];
  verdict: string;
  ignored: boolean;
}

export interface AvoidableRow {
  ability: string;
  abilityId: number;
  abilityIcon: string;
  amount: number;
  hits: number;
  raidMedian: number;
  ratio: number;
  sharePct: number;
  known: boolean;
}

export interface PlayerAvoidable extends PlayerRef {
  totalTaken: number;
  knownTaken: number;
  possibleTaken: number;
  avoidablePct: number;
  rows: AvoidableRow[];
}

export interface AbilityAvoidable {
  ability: string;
  abilityId: number;
  abilityIcon: string;
  known: boolean;
  total: number;
  hitters: number;
  median: number;
  worst: string;
  worstAmount: number;
}

export interface Gap {
  startMs: number;
  endMs: number;
}

export interface PlayerActivity extends PlayerRef {
  uptimePct: number;
  aliveMs: number;
  activeMs: number;
  excusedMs: number;
  downtimeMs: number;
  casts: number;
  gaps: Gap[];
  deadFromMs: number; // -1 if still alive
}

export interface CooldownUse {
  name: string;
  abilityId: number;
  abilityIcon: string;
  kind: "personal" | "external" | "raid";
  cooldown: number;
  possible: number;
  used: number;
  usedAtMs: number[];
  missedOpportunities: number;
}

export interface Spike {
  startMs: number;
  endMs: number;
  pct: number;
  top: string;
  topIcon: string;
  topAbilityId: number;
  available: AbilityRef[];
  used: AbilityRef[];
  died: boolean;
}

export interface PlayerCooldowns extends PlayerRef {
  cooldowns: CooldownUse[];
  spikes: Spike[];
  usePct: number;
}

export interface PlayerSummary extends PlayerRef {
  deaths: number;
  deathTimestamps: number[];
  avoidableDeaths: number;
  mechanicDeaths: number;
  uptimePct: number;
  avoidableDamage: number;
  avoidablePct: number;
  cooldownUsePct: number;
  missedSpikes: number;
}

export interface Summary {
  deaths: number;
  countedDeaths: number;
  avoidableDeaths: number;
  mechanicDeaths: number;
  avgUptimePct: number;
  totalAvoidableDamage: number;
  players: PlayerSummary[];
}

export interface AnalysisResult {
  reportCode: string;
  reportTitle: string;
  fight: { id: number; name: string; kill: boolean; difficulty: number; durationMs: number };
  summary: Summary;
  deaths: DeathReport[];
  avoidableDamage: PlayerAvoidable[];
  avoidableByAbility: AbilityAvoidable[];
  activity: PlayerActivity[];
  cooldowns: PlayerCooldowns[];
  warnings: string[];
}

export interface AnalysisOptions {
  window: number;
  mechanicPct: number;
  spikePct: number;
  gapSec: number;
  wipeIgnoreDeaths: number;
  assumeTalents: boolean;
}
