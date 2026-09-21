export function mmss(ms: number): string {
  const s = Math.max(0, Math.round(ms / 1000));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}

export function compact(n: number): string {
  const a = Math.abs(n);
  if (a >= 1e6) return `${(n / 1e6).toFixed(2)}M`;
  if (a >= 1e3) return `${(n / 1e3).toFixed(0)}k`;
  return String(Math.round(n));
}

export function pct(n: number | undefined, digits = 0): string {
  return `${(n ?? 0).toFixed(digits)}%`;
}

export function seconds(ms: number, digits = 1): string {
  return `${(ms / 1000).toFixed(digits)}s`;
}

export const DIFFICULTY: Record<number, string> = { 1: "LFR", 3: "Normal", 4: "Heroic", 5: "Mythic" };
export const ROLE_LABEL = { tank: "Tank", healer: "Healer", dps: "DPS" } as const;
export const KIND_LABEL = { personal: "Personal", external: "External", raid: "Raid" } as const;
