const arabicDigits = '٠١٢٣٤٥٦٧٨٩';
const persianDigits = '۰۱۲۳۴۵۶۷۸۹';

const diacriticsPattern =
  /[\u0610-\u061A\u064B-\u065F\u0670\u06D6-\u06ED]/g;

export function normalizeForSearch(value: string): string {
  return value
    .normalize('NFKC')
    .replace(diacriticsPattern, '')
    .replace(/[يى]/g, 'ی')
    .replace(/ك/g, 'ک')
    .replace(/[ۀة]/g, 'ه')
    .replace(/[أإٱآ]/g, 'ا')
    .replace(/ؤ/g, 'و')
    .replace(/ئ/g, 'ی')
    .replace(/\u0640/g, '')
    .replace(/[\u200c\u200d]/g, ' ')
    .replace(/[٠-٩۰-۹]/g, digit => {
      const arabicIndex = arabicDigits.indexOf(digit);
      if (arabicIndex >= 0) return String(arabicIndex);
      return String(persianDigits.indexOf(digit));
    })
    .toLowerCase()
    .replace(/[^\p{L}\p{N}._:/+-]+/gu, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

export function tokenizeSearchQuery(query: string): string[] {
  const normalized = normalizeForSearch(query);
  if (!normalized) return [];

  return [...new Set(normalized.split(' ').filter(Boolean))];
}

