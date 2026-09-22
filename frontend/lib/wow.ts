// WoW class colors (Blizzard standard) and per-spec icons, used to decorate
// player names across the tabs. Class/spec strings match what the Go API
// forwards from Warcraft Logs (PascalCase, no spaces, e.g. "DemonHunter",
// "BeastMastery"). Icons are served from the same CDN as ability icons
// (see iconUrl in ui.tsx) and fall back to a generic icon if a filename is
// ever wrong — see AbilityLabel's onError handler.

export const CLASS_COLORS: Record<string, string> = {
  Warrior: "#C79C6E",
  Paladin: "#F58CBA",
  Hunter: "#ABD473",
  Rogue: "#FFF569",
  Priest: "#FFFFFF",
  DeathKnight: "#C41F3B",
  Shaman: "#0070DE",
  Mage: "#69CCF0",
  Warlock: "#9482C9",
  Monk: "#00FF96",
  Druid: "#FF7D0A",
  DemonHunter: "#A330C9",
  Evoker: "#33937F",
};

const SPEC_ICON: Record<string, Record<string, string>> = {
  Warrior: {
    Arms: "ability_warrior_savageblow.jpg",
    Fury: "ability_warrior_innerrage.jpg",
    Protection: "ability_warrior_defensivestance.jpg",
  },
  Paladin: {
    Holy: "spell_holy_holybolt.jpg",
    Protection: "ability_paladin_shieldofthetemplar.jpg",
    Retribution: "spell_holy_auraoflight.jpg",
  },
  Hunter: {
    BeastMastery: "ability_hunter_bestialdiscipline.jpg",
    Marksmanship: "ability_hunter_focusedaim.jpg",
    Survival: "ability_hunter_camouflage.jpg",
  },
  Rogue: {
    Assassination: "ability_rogue_deadlybrew.jpg",
    Outlaw: "ability_rogue_waylay.jpg",
    Subtlety: "ability_stealth.jpg",
  },
  Priest: {
    Discipline: "spell_holy_powerwordshield.jpg",
    Holy: "spell_holy_guardianspirit.jpg",
    Shadow: "spell_shadow_shadowwordpain.jpg",
  },
  DeathKnight: {
    Blood: "spell_deathknight_bloodpresence.jpg",
    Frost: "spell_deathknight_frostpresence.jpg",
    Unholy: "spell_deathknight_unholypresence.jpg",
  },
  Shaman: {
    Elemental: "spell_nature_lightning.jpg",
    Enhancement: "spell_shaman_improvedstormstrike.jpg",
    Restoration: "spell_nature_magicimmunity.jpg",
  },
  Mage: {
    Arcane: "spell_holy_magicalsentry.jpg",
    Fire: "spell_fire_firebolt02.jpg",
    Frost: "spell_frost_frostbolt02.jpg",
  },
  Warlock: {
    Affliction: "spell_shadow_deathcoil.jpg",
    Demonology: "spell_shadow_metamorphosis.jpg",
    Destruction: "spell_shadow_rainoffire.jpg",
  },
  Monk: {
    Brewmaster: "spell_monk_brewmaster_spec.jpg",
    Mistweaver: "spell_monk_mistweaver_spec.jpg",
    Windwalker: "spell_monk_windwalker_spec.jpg",
  },
  Druid: {
    Balance: "spell_nature_starfall.jpg",
    Feral: "ability_druid_catform.jpg",
    Guardian: "ability_racial_bearform.jpg",
    Restoration: "spell_nature_healingtouch.jpg",
  },
  DemonHunter: {
    Havoc: "ability_demonhunter_specdps.jpg",
    Vengeance: "ability_demonhunter_spectank.jpg",
  },
  Evoker: {
    Devastation: "ability_evoker_eternitysurge.jpg",
    Preservation: "classicon_evoker_preservation.jpg",
    Augmentation: "classicon_evoker_augmentation.jpg",
  },
};

export function specIcon(className: string, spec: string): string | undefined {
  return SPEC_ICON[className]?.[spec];
}
