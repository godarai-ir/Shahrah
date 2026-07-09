export type CompactDisplayValue = {
  value: string;
  fullValue: string;
  isCompact: boolean;
};

type ExactNumberOptions = {
  locale: string;
  minimumFractionDigits?: number;
  maximumFractionDigits?: number;
};

type CompactNumberOptions = ExactNumberOptions & {
  compactThreshold?: number;
};

type ExactCurrencyOptions = {
  locale: string;
  currency: string;
  minimumFractionDigits?: number;
  maximumFractionDigits?: number;
};

type CompactCurrencyOptions = ExactCurrencyOptions & {
  compactThreshold?: number;
};

const COMPACT_UNITS = [
  { value: 1_000_000_000_000, suffix: "T" },
  { value: 1_000_000_000, suffix: "B" },
  { value: 1_000_000, suffix: "M" },
  { value: 1_000, suffix: "K" },
] as const;

const NUMERIC_PART_TYPES = new Set(["integer", "group", "decimal", "fraction"]);

export function formatExactNumber(
  value: number,
  {
    locale,
    minimumFractionDigits = 0,
    maximumFractionDigits = Number.isInteger(value) ? 0 : 2,
  }: ExactNumberOptions,
) {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(value);
}

export function formatCompactNumber(
  value: number,
  {
    locale,
    compactThreshold = 10_000,
    minimumFractionDigits = 0,
    maximumFractionDigits = Number.isInteger(value) ? 0 : 2,
  }: CompactNumberOptions,
): CompactDisplayValue {
  const fullValue = formatExactNumber(value, {
    locale,
    minimumFractionDigits,
    maximumFractionDigits,
  });

  if (!Number.isFinite(value) || Math.abs(value) < compactThreshold) {
    return { value: fullValue, fullValue, isCompact: false };
  }

  const compactValue = buildCompactMagnitude(value, locale);
  return {
    value: compactValue,
    fullValue,
    isCompact: compactValue !== fullValue,
  };
}

export function formatExactCurrency(
  value: number,
  {
    locale,
    currency,
    minimumFractionDigits = 0,
    maximumFractionDigits = 2,
  }: ExactCurrencyOptions,
) {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(value);
}

export function formatCompactCurrency(
  value: number,
  {
    locale,
    currency,
    compactThreshold = 10_000,
    minimumFractionDigits = 0,
    maximumFractionDigits = 2,
  }: CompactCurrencyOptions,
): CompactDisplayValue {
  const currencyFormatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits,
    maximumFractionDigits,
  });
  const fullValue = currencyFormatter.format(value);

  if (!Number.isFinite(value) || Math.abs(value) < compactThreshold) {
    return { value: fullValue, fullValue, isCompact: false };
  }

  const compactNumber = buildCompactMagnitude(value, locale);
  const compactValue = injectCurrencyAroundNumber(currencyFormatter, value, compactNumber);

  return {
    value: compactValue,
    fullValue,
    isCompact: compactValue !== fullValue,
  };
}

export function appendDisplaySuffix(display: CompactDisplayValue, suffix: string): CompactDisplayValue {
  return {
    value: `${display.value}${suffix}`,
    fullValue: `${display.fullValue}${suffix}`,
    isCompact: display.isCompact,
  };
}

function buildCompactMagnitude(value: number, locale: string) {
  const absoluteValue = Math.abs(value);
  const sign = value < 0 ? -1 : 1;
  let unitIndex = COMPACT_UNITS.findIndex((unit) => absoluteValue >= unit.value);

  if (unitIndex === -1) {
    return formatExactNumber(value, { locale });
  }

  let scaledValue = absoluteValue / COMPACT_UNITS[unitIndex].value;
  let fractionDigits = compactFractionDigits(scaledValue);
  let roundedValue = roundTo(scaledValue, fractionDigits);

  while (roundedValue >= 1000 && unitIndex > 0) {
    unitIndex -= 1;
    scaledValue = absoluteValue / COMPACT_UNITS[unitIndex].value;
    fractionDigits = compactFractionDigits(scaledValue);
    roundedValue = roundTo(scaledValue, fractionDigits);
  }

  const formattedValue = new Intl.NumberFormat(locale, {
    minimumFractionDigits: 0,
    maximumFractionDigits: fractionDigits,
  }).format(sign * roundedValue);

  return `${formattedValue}${COMPACT_UNITS[unitIndex].suffix}`;
}

function compactFractionDigits(value: number) {
  if (value >= 100) return 0;
  return 1;
}

function roundTo(value: number, fractionDigits: number) {
  const factor = 10 ** fractionDigits;
  return Math.round(value * factor) / factor;
}

function injectCurrencyAroundNumber(formatter: Intl.NumberFormat, value: number, compactNumber: string) {
  const parts = formatter.formatToParts(value);
  let prefix = "";
  let suffix = "";
  let seenNumber = false;

  for (const part of parts) {
    if (NUMERIC_PART_TYPES.has(part.type)) {
      seenNumber = true;
      continue;
    }

    if (!seenNumber) prefix += part.value;
    else suffix += part.value;
  }

  return `${prefix}${compactNumber}${suffix}`;
}
