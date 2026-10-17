import type { CreditUsageSummary } from "../api";
import type { CreditForecast } from "./api";

const DAY_MS = 24 * 60 * 60 * 1000;

/** Product key of the semantic-cache embedding charge, which the gross cache saving does not net out. */
export const SEMANTIC_CACHE_PRODUCT = "semantic_cache";

/** Whole days until the credits run out, by spending or expiry; 0 when they already have; null without an end. */
export function forecastDaysLeft(forecast: CreditForecast, now: Date): number | null {
  if (!forecast.depletes_at) return null;
  const depletesAt = new Date(forecast.depletes_at).getTime();
  if (Number.isNaN(depletesAt)) return null;
  return Math.max(0, Math.floor((depletesAt - now.getTime()) / DAY_MS));
}

/** Credits the cache saved after subtracting what the semantic cache itself was charged. */
export function netCacheSavedCredits(summary: CreditUsageSummary): number {
  const gross = Number(summary.cache_saved_credits) || 0;
  const cacheCost = Number(summary.by_product.find((item) => item.key === SEMANTIC_CACHE_PRODUCT)?.credits ?? 0) || 0;
  return Math.max(0, gross - cacheCost);
}

/** A row's share of the period total, from 0 to 1. */
export function creditShare(credits: string, total: string): number {
  const value = Number(credits) || 0;
  const sum = Number(total) || 0;
  return sum > 0 ? Math.min(1, value / sum) : 0;
}

/** Days left at or below which the forecast is shown as a warning. */
export const FORECAST_WARNING_DAYS = 7;
