import { newDate as newJalaliDate } from "date-fns-jalali";

const TEHRAN_UTC_OFFSET_MILLISECONDS = (3 * 60 + 30) * 60 * 1000;

const persianDatePartsFormatter = new Intl.DateTimeFormat("en-US-u-ca-persian", {
  timeZone: "Asia/Tehran",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

export function getCurrentJalaliMonthRange(reference = new Date()) {
  const parts = persianDatePartsFormatter.formatToParts(reference);
  const year = Number(parts.find((part) => part.type === "year")?.value ?? 0);
  const month = Number(parts.find((part) => part.type === "month")?.value ?? 0);
  const localMonthStart = newJalaliDate(year, month - 1, 1);
  const monthStartUTC = Date.UTC(
    localMonthStart.getFullYear(),
    localMonthStart.getMonth(),
    localMonthStart.getDate(),
  ) - TEHRAN_UTC_OFFSET_MILLISECONDS;

  return { from: new Date(monthStartUTC), to: new Date(reference) };
}
