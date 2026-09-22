import type { ReactNode } from "react";
import type { AbilityRef, PlayerRef } from "@/lib/types";

type Tone = "bad" | "warn" | "info" | "ok";

const TONES: Record<Tone, string> = {
  bad: "bg-bad-bg text-bad",
  warn: "bg-warn-bg text-warn",
  info: "bg-info-bg text-info",
  ok: "bg-ok-bg text-ok",
};

export function Badge({ tone, children }: { tone: Tone; children: ReactNode }) {
  return (
    <span className={`mr-1.5 inline-block rounded-full px-2 py-px text-xs font-semibold ${TONES[tone]}`}>
      {children}
    </span>
  );
}

export function Card({ children, className = "" }: { children: ReactNode; className?: string }) {
  return <section className={`mb-4 rounded-xl border border-line bg-surface p-4 ${className}`}>{children}</section>;
}

export function Tile({ value, label }: { value: ReactNode; label: string }) {
  return (
    <div className="rounded-xl border border-line bg-surface px-3.5 py-3">
      <div className="text-3xl font-semibold">{value}</div>
      <div className="text-sm text-fg-2">{label}</div>
    </div>
  );
}

export function Note({ children }: { children: ReactNode }) {
  return <p className="mb-3 text-sm text-fg-2">{children}</p>;
}

export function Muted({ children }: { children: ReactNode }) {
  return <span className="text-fg-3">{children}</span>;
}

// ---- Ability icons ----

const ICON_BASE = "https://assets.rpglogs.com/img/warcraft/abilities/";
const FALLBACK_ICON = "inv_misc_questionmark.jpg";

function iconUrl(icon?: string) {
  return `${ICON_BASE}${icon || FALLBACK_ICON}`;
}

/**
 * Ability icon + name. When abilityId is known, wraps in a Wowhead spell
 * link: the wowhead tooltip script (see app/layout.tsx) turns hovering it
 * into the real Wowhead tooltip (full description, cooldown, etc.) — WCL's
 * API only exposes name/icon, not ability text.
 */
export function AbilityLabel({ name, icon, abilityId }: { name: string; icon?: string; abilityId?: number }) {
  const content = (
    <>
      <img
        src={iconUrl(icon)}
        alt=""
        width={20}
        height={20}
        loading="lazy"
        className="h-5 w-5 shrink-0 rounded border border-line align-middle"
        onError={(e) => {
          const img = e.currentTarget;
          if (img.src !== iconUrl(undefined)) img.src = iconUrl(undefined);
        }}
      />
      <span>{name}</span>
    </>
  );
  if (!abilityId) {
    return <span className="inline-flex items-center gap-1.5">{content}</span>;
  }
  return (
    <a
      href={`https://www.wowhead.com/spell=${abilityId}`}
      target="_blank"
      rel="noreferrer noopener"
      className="inline-flex items-center gap-1.5 text-fg no-underline hover:underline"
    >
      {content}
    </a>
  );
}

/** Small "?" badge; hovering (or focusing, for keyboard/touch) shows an explanatory tooltip. */
export function InfoTip({ text }: { text: string }) {
  return (
    <span
      tabIndex={0}
      // Stop clicks from reaching an ancestor <label>/<button> — this badge is a
      // hover/focus affordance only, it must never trigger a parent control.
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
      }}
      className="group/tip relative inline-flex h-3.5 w-3.5 cursor-help items-center justify-center rounded-full border border-line text-[10px] leading-none text-fg-3 outline-none"
    >
      ?
      <span className="pointer-events-none absolute bottom-full left-1/2 z-10 mb-1.5 w-max max-w-[240px] -translate-x-1/2 rounded-md border border-line bg-surface px-2 py-1.5 text-xs font-normal normal-case whitespace-normal text-fg opacity-0 shadow-md transition-opacity group-hover/tip:opacity-100 group-focus/tip:opacity-100">
        {text}
      </span>
    </span>
  );
}

/** A row of AbilityLabel chips, for lists of player abilities (defensives, consumables...). */
export function AbilityRefList({ items, empty = "none" }: { items: AbilityRef[]; empty?: string }) {
  if (!items.length) return <Muted>{empty}</Muted>;
  return (
    <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
      {items.map((x, i) => (
        <AbilityLabel key={x.abilityId || `${x.name}-${i}`} name={x.name} icon={x.abilityIcon} abilityId={x.abilityId} />
      ))}
    </span>
  );
}

export function PlayerLabel({ p }: { p: Pick<PlayerRef, "player" | "spec" | "class"> }) {
  return (
    <>
      <b>{p.player}</b>{" "}
      <span className="text-fg-3">
        {p.spec} {p.class}
      </span>
    </>
  );
}

// ---- Tables ----

export function Table({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-xl border border-line bg-surface">
      <table className="w-full border-collapse">{children}</table>
    </div>
  );
}

export function TH({
  children,
  num,
  className = "",
  ...rest
}: React.ThHTMLAttributes<HTMLTableCellElement> & { num?: boolean }) {
  return (
    <th
      className={`border-b border-line px-2.5 py-2 text-xs font-semibold tracking-wide text-fg-2 uppercase ${
        num ? "text-right" : "text-left"
      } ${className}`}
      {...rest}
    >
      {children}
    </th>
  );
}

export function TD({
  children,
  num,
  className = "",
  ...rest
}: React.TdHTMLAttributes<HTMLTableCellElement> & { num?: boolean }) {
  return (
    <td
      className={`border-b border-line px-2.5 py-2 align-top text-sm group-last:border-b-0 ${
        num ? "text-right tabular-nums" : "text-left"
      } ${className}`}
      {...rest}
    >
      {children}
    </td>
  );
}

export function TR({ children }: { children: ReactNode }) {
  return <tr className="group">{children}</tr>;
}

export function UptimeBar({ value }: { value: number }) {
  return (
    <div className="flex items-center gap-2">
      <div className="relative h-2.5 min-w-[90px] flex-1 rounded-full bg-track">
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-series"
          style={{ width: `${Math.min(Math.max(value, 0), 100)}%` }}
        />
      </div>
      <b className="tabular-nums">{value.toFixed(1)}%</b>
    </div>
  );
}
