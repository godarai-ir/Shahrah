const PERSIAN_DIGITS = "۰۱۲۳۴۵۶۷۸۹";
const ARABIC_DIGITS = "٠١٢٣٤٥٦٧٨٩";

export function parsePositiveToman(value: string): number | null {
  const normalized = Array.from(value.trim(), (character) => {
    const persianIndex = PERSIAN_DIGITS.indexOf(character);
    if (persianIndex >= 0) return String(persianIndex);

    const arabicIndex = ARABIC_DIGITS.indexOf(character);
    if (arabicIndex >= 0) return String(arabicIndex);

    return character;
  }).join("").replaceAll(",", "").replaceAll("٬", "").replaceAll(" ", "");

  if (normalized === "") return null;

  const amount = Number(normalized);
  return Number.isFinite(amount) && amount > 0 ? amount : null;
}

export function convertTomanToUSD(amountToman: number, usdToIRR: number): number | null {
  if (!Number.isFinite(amountToman) || amountToman <= 0 || !Number.isFinite(usdToIRR) || usdToIRR <= 0) {
    return null;
  }

  const tomanPerUSD = usdToIRR / 10;
  return amountToman / tomanPerUSD;
}

export function formatApproximateUSD(amountUSD: number): string | null {
  if (!Number.isFinite(amountUSD) || amountUSD <= 0) return null;

  const formatted = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amountUSD);

  return `~ ${formatted}`;
}
