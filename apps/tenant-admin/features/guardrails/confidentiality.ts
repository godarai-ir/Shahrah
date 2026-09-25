import type { ConfidentialityLevel } from "./types";
import { CONFIDENTIALITY_LEVELS } from "./form";

/**
 * Mirrors DetectConfidentiality in the gateway so the wizard can preview a sample. A marking only counts when it fills
 * a whole line or a separated segment of a line, the way stamps and sensitivity labels are laid out.
 */
const MARKINGS: Record<string, ConfidentialityLevel> = {
  "به کلی سری": "top_secret",
  "بکلی سری": "top_secret",
  "top secret": "top_secret",
  "خیلی محرمانه": "highly_confidential",
  "highly confidential": "highly_confidential",
  "strictly confidential": "highly_confidential",
  "محرمانه": "confidential",
  "confidential": "confidential",
  "company confidential": "confidential",
  "private and confidential": "confidential",
  "فقط برای استفاده داخلی": "internal",
  "برای استفاده داخلی": "internal",
  "استفاده داخلی": "internal",
  "internal use only": "internal",
  "for internal use only": "internal",
  "internal only": "internal",
};

/** Ordinary words elsewhere, so they only count after a classification label. */
const LABELED_ONLY_MARKINGS: Record<string, ConfidentialityLevel> = {
  "سری": "secret",
  "secret": "secret",
  "restricted": "confidential",
  "داخلی": "internal",
  "internal": "internal",
};

/** Level names that are never a stamp on their own; they only help recognize a listing of the scheme. */
const SCHEME_ONLY_TERMS = new Set(["public", "general", "unclassified", "عمومی", "عادی", "طبقه بندی نشده"]);
const SCHEME_WINDOW = 3;
const SCHEME_NEIGHBORS = 2;

const CLASSIFICATION_LABEL = /^(?:طبقه ?بندی(?: حفاظتی)?|درجه(?: حفاظتی| طبقه ?بندی)?|سطح(?: محرمانگی| دسترسی)|classification|security classification|sensitivity(?: label)?|data classification)\s*[:：-]\s*/u;
const SEPARATORS = /\s*(?:\/\/|[|/،,؛;–—]|\s-\s)\s*/u;
const DECORATION = /^[\s\p{P}\p{S}‌]+|[\s\p{P}\p{S}‌]+$/gu;
const MAX_MARKING_LINE = 120;

function isSchemeTerm(line: string) {
  return Object.hasOwn(MARKINGS, line) || Object.hasOwn(LABELED_ONLY_MARKINGS, line) || SCHEME_ONLY_TERMS.has(line);
}

/** Whether enough other distinct level names stand on nearby lines, as in a policy's level table. */
function inSchemeListing(terms: readonly string[], index: number) {
  const others = new Set<string>();
  for (let j = Math.max(0, index - SCHEME_WINDOW); j <= Math.min(terms.length - 1, index + SCHEME_WINDOW); j++) {
    if (j !== index && terms[j] && terms[j] !== terms[index]) others.add(terms[j]);
  }

  return others.size >= SCHEME_NEIGHBORS;
}

export type ConfidentialityFinding = { level: ConfidentialityLevel; marking: string } | null;

export function normalizeMarkingText(text: string) {
  return text
    .toLowerCase()
    .replace(/[يى]/g, "ی")
    .replace(/ك/g, "ک")
    .replace(/[​‍‎‏⁠﻿]/g, "")
    .split("\n")
    .map((line) => line.replace(/[\s‌]+/g, " ").trim())
    .join("\n");
}

function rank(level: ConfidentialityLevel) {
  return CONFIDENTIALITY_LEVELS.indexOf(level);
}

export function detectConfidentiality(text: string, customMarkings: readonly string[], customLevel: ConfidentialityLevel): ConfidentialityFinding {
  let best: ConfidentialityFinding = null;
  const consider = (level: ConfidentialityLevel, marking: string) => {
    if (!best || rank(level) > rank(best.level)) best = { level, marking };
  };

  const normalized = normalizeMarkingText(text);
  for (const marking of customMarkings.map(normalizeMarkingText)) {
    if (marking && normalized.includes(marking)) consider(customLevel, marking);
  }

  const lines = normalized.split("\n").map((line) => line.replace(DECORATION, "")).filter(Boolean);
  const terms = lines.map((line) => (isSchemeTerm(line) ? line : ""));

  for (const [index, rawLine] of lines.entries()) {
    let line = rawLine;
    if ([...line].length > MAX_MARKING_LINE) continue;

    const label = CLASSIFICATION_LABEL.exec(line);
    if (label) line = line.slice(label[0].length).replace(DECORATION, "");
    if (!label && terms[index] && inSchemeListing(terms, index)) continue;

    for (const rawSegment of line.split(SEPARATORS)) {
      const segment = rawSegment.replace(DECORATION, "");
      // Own-property lookups, so a line such as "constructor" never resolves to an Object.prototype member.
      const level = Object.hasOwn(MARKINGS, segment)
        ? MARKINGS[segment]
        : label && Object.hasOwn(LABELED_ONLY_MARKINGS, segment) ? LABELED_ONLY_MARKINGS[segment] : undefined;
      if (level) consider(level, segment);
    }
  }

  return best;
}
