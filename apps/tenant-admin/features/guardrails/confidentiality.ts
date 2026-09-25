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

const CLASSIFICATION_LABEL = /^(?:طبقه ?بندی(?: حفاظتی)?|درجه(?: حفاظتی| طبقه ?بندی)?|سطح(?: محرمانگی| دسترسی)|classification|security classification|sensitivity(?: label)?|data classification)\s*[:：-]\s*/u;
const SEPARATORS = /\s*(?:\/\/|[|/،,؛;–—]|\s-\s)\s*/u;
const DECORATION = /^[\s\p{P}\p{S}‌]+|[\s\p{P}\p{S}‌]+$/gu;
const MAX_MARKING_LINE = 120;

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

  for (const rawLine of normalized.split("\n")) {
    let line = rawLine.replace(DECORATION, "");
    if (!line || [...line].length > MAX_MARKING_LINE) continue;

    const label = CLASSIFICATION_LABEL.exec(line);
    if (label) line = line.slice(label[0].length).replace(DECORATION, "");

    for (const rawSegment of line.split(SEPARATORS)) {
      const segment = rawSegment.replace(DECORATION, "");
      const level = MARKINGS[segment] ?? (label ? LABELED_ONLY_MARKINGS[segment] : undefined);
      if (level) consider(level, segment);
    }
  }

  return best;
}
