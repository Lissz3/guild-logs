import type { AnalysisOptions, AnalysisResult, ReportMeta, ServerConfig } from "./types";

// All requests go through Next.js, which proxies /api/* to the Go server.

async function get<T>(path: string, params?: Record<string, string>): Promise<T> {
  const qs = params ? `?${new URLSearchParams(params)}` : "";
  const res = await fetch(`/api/${path}${qs}`);
  const body: unknown = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = (body as { error?: string }).error;
    throw new Error(msg ?? `Error ${res.status}`);
  }
  return body as T;
}

export function getConfig(): Promise<ServerConfig> {
  return get<ServerConfig>("config");
}

export async function getReport(codeOrUrl: string): Promise<ReportMeta> {
  const meta = await get<ReportMeta>("report", { code: codeOrUrl });
  // Go encodes empty slices as null; keep the UI code simple.
  return { ...meta, fights: meta.fights ?? [] };
}

export async function analyze(code: string, fight: number, o: AnalysisOptions): Promise<AnalysisResult> {
  const r = await get<AnalysisResult>("analyze", {
    code,
    fight: String(fight),
    window: String(o.window),
    mechanicPct: String(o.mechanicPct),
    spikePct: String(o.spikePct),
    gapSec: String(o.gapSec),
    assumeTalents: o.assumeTalents ? "1" : "0",
  });
  return {
    ...r,
    warnings: r.warnings ?? [],
    deaths: r.deaths ?? [],
    avoidableDamage: r.avoidableDamage ?? [],
    avoidableByAbility: r.avoidableByAbility ?? [],
    activity: r.activity ?? [],
    cooldowns: r.cooldowns ?? [],
    summary: { ...r.summary, players: r.summary?.players ?? [] },
  };
}

/** Extracts the fight number from a Warcraft Logs URL such as ...?fight=12. */
export function fightFromUrl(input: string): number | null {
  const m = /[?&#]fight=(\d+)/.exec(input);
  return m ? Number(m[1]) : null;
}
