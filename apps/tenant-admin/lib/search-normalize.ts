/**
 * Folds the variants Persian keyboards produce (Arabic ye/kaf, ZWNJ, diacritics, Persian/Arabic digits)
 * so that "لاگ‌ها", "لاگ ها" and "لاگها" compare equal.
 */
export function normalizeSearchText(value: string): string {
  return value
    .toLowerCase()
    .replace(/[يى]/g, "ی")
    .replace(/ك/g, "ک")
    .replace(/[ً-ْٰـ]/g, "")
    .replace(/[‌‏‎]/g, "")
    .replace(/[۰-۹]/g, (digit) => String(digit.charCodeAt(0) - 0x06f0))
    .replace(/[٠-٩]/g, (digit) => String(digit.charCodeAt(0) - 0x0660))
    .replace(/\s+/g, "")
    .trim();
}

/** Substring match on normalized text; an empty query matches everything. */
export function matchesSearch(candidates: readonly string[], query: string): boolean {
  const needle = normalizeSearchText(query);
  if (!needle) return true;
  return candidates.some((candidate) => normalizeSearchText(candidate).includes(needle));
}
