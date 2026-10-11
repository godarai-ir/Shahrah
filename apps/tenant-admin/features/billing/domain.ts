import { format as formatJalali, newDate as newJalaliDate, subMonths as subJalaliMonths } from "date-fns-jalali";

import { DEFAULT_LOCALE, getIntlLocale, resolveLocale, type AppLocale } from "@/lib/locale";

export { getCurrentJalaliMonthRange } from "@/lib/date/jalali-month";

export const FINANCIAL_CALENDAR = "jalali";
export const FINANCIAL_TIME_ZONE = "Asia/Tehran";

export type BillingSummaryPreset =
  | "current_month"
  | "current_quarter"
  | "current_year"
  | "last_3_months"
  | "last_6_months"
  | "last_12_months";

export type JalaliMonthSelection = {
  year: number;
  month: number;
};

export type WalletPayload = {
  id?: string;
  wallet_id?: string;
  tenant_id?: string;
  balance_irr?: number;
  ledger_balance_irr?: number;
  available_balance_irr?: number;
  reserved_irr?: number;
  committed_unsettled_irr?: number;
  pending_irr?: number;
  sync_lag_seconds?: number;
  sync_status?: string;
  window_id?: string;
  as_of?: string;
};

export type WalletTransaction = {
  id: string;
  type: string;
  amount_irr: number;
  balance_after_irr?: number;
  status: string;
  created_at: string;
};

export type BillingProviderUsage = {
  provider_name: string;
  request_count: number;
  input_tokens: number;
  output_tokens: number;
  raw_cost_irr: number;
  discount_irr: number;
  cache_savings_irr: number;
  final_cost_irr: number;
  share: number;
};

export type BillingUsageSummary = {
  calendar: typeof FINANCIAL_CALENDAR;
  preset: BillingSummaryPreset;
  start: string;
  end: string;
  total_usage_irr: number;
  discount_irr: number;
  cache_savings_irr: number;
  gift_credit_irr: number;
  providers: BillingProviderUsage[];
};

export type MonthlyUsageDetail = {
  calendar: typeof FINANCIAL_CALENDAR;
  jalali_year: number;
  jalali_month: number;
  start: string;
  end: string;
  total_usage_irr: number;
  gift_credit_irr: number;
  discount_irr: number;
  cache_savings_irr: number;
  providers: BillingProviderUsage[];
};

export type BankTransferRequest = {
  id: string;
  public_id?: string;
  amount_irr: number;
  tracking_number: string;
  bank_transaction_at: string;
  receipt_file_url?: string;
  status: string;
  rejection_reason?: string;
  created_at: string;
};

export type BillingProfile = {
  id?: string;
  profile_type: string;
  data: Record<string, unknown>;
};

export type Invoice = {
  id: string;
  invoice_number: string;
  status: string;
  period_start: string;
  period_end: string;
  raw_usage_irr: number;
  discount_irr: number;
  gift_credit_irr: number;
  cache_savings_irr: number;
  final_amount_irr: number;
  net_adjustment_irr: number;
  created_at: string;
};

export type DiscountRedemption = {
  id: string;
  code: string;
  discount_type: string;
  value_irr?: number;
  percent?: number;
  status: string;
  used_amount_irr: number;
  created_at: string;
};

export type BankTransferPayload = {
  amount_irr: number;
  tracking_number: string;
  bank_transaction_date_jalali: string;
  receipt_file_url?: string;
};

type JalaliTodayParts = JalaliMonthSelection & {
  day: number;
};

/**
 * Billing periods are legally anchored to the Jalali calendar, so both locales
 * keep that calendar and only the script and month names change.
 */
function financialCalendarLocale(locale: string) {
  return resolveLocale(locale) === "fa" ? "fa-IR-u-ca-persian" : "en-US-u-ca-persian";
}

const numberFormatters = new Map<string, Intl.NumberFormat>();
const dateTimeFormatters = new Map<string, Intl.DateTimeFormat>();
const dateFormatters = new Map<string, Intl.DateTimeFormat>();
const monthFormatters = new Map<string, Intl.DateTimeFormat>();

function cached<T>(cache: Map<string, T>, key: string, create: () => T): T {
  const existing = cache.get(key);
  if (existing) return existing;
  const created = create();
  cache.set(key, created);
  return created;
}

function financialNumberFormatter(locale: string) {
  const intlLocale = getIntlLocale(locale);
  return cached(numberFormatters, intlLocale, () => new Intl.NumberFormat(intlLocale));
}

const creditNumberFormatters = new Map<string, Intl.NumberFormat>();

function creditNumberFormatter(locale: string) {
  const intlLocale = getIntlLocale(locale);
  return cached(creditNumberFormatters, intlLocale, () => new Intl.NumberFormat(intlLocale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }));
}

function financialDateTimeFormatter(locale: string) {
  const calendarLocale = financialCalendarLocale(locale);
  return cached(dateTimeFormatters, calendarLocale, () => new Intl.DateTimeFormat(calendarLocale, {
    timeZone: FINANCIAL_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }));
}

function financialDateFormatter(locale: string) {
  const calendarLocale = financialCalendarLocale(locale);
  return cached(dateFormatters, calendarLocale, () => new Intl.DateTimeFormat(calendarLocale, {
    timeZone: FINANCIAL_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }));
}

function financialMonthFormatter(locale: string) {
  const calendarLocale = financialCalendarLocale(locale);
  return cached(monthFormatters, calendarLocale, () => new Intl.DateTimeFormat(calendarLocale, {
    timeZone: FINANCIAL_TIME_ZONE,
    year: "numeric",
    month: "long",
  }));
}

const persianTodayPartsFormatter = new Intl.DateTimeFormat("en-US-u-ca-persian", {
  timeZone: FINANCIAL_TIME_ZONE,
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

/** Rounded rial amount without a unit; callers append the localized unit. */
export function formatIRRAmount(value: number | undefined | null, locale: AppLocale = DEFAULT_LOCALE) {
  return financialNumberFormatter(locale).format(Math.round(Number(value ?? 0)));
}

/** Rounded toman amount without a unit; callers append the localized unit. */
export function formatTomanAmount(valueIRR: number | undefined | null, locale: AppLocale = DEFAULT_LOCALE) {
  return financialNumberFormatter(locale).format(Math.round(Number(valueIRR ?? 0) / 10));
}

export function formatDate(value: string | undefined | null, locale: AppLocale = DEFAULT_LOCALE) {
  if (!value) return "-";
  return financialDateTimeFormatter(locale).format(new Date(value));
}

export function formatDateOnly(value: string | undefined | null, locale: AppLocale = DEFAULT_LOCALE) {
  if (!value) return "-";
  return financialDateFormatter(locale).format(new Date(value));
}

export function getCurrentJalaliMonthSelection(reference = new Date()): JalaliMonthSelection {
  const { year, month } = getCurrentJalaliDateParts(reference);
  return { year, month };
}

export function getCurrentJalaliDateValue(reference = new Date()) {
  const { year, month, day } = getCurrentJalaliDateParts(reference);
  return `${year}-${pad(month)}-${pad(day)}`;
}

export function getJalaliMonthKey(selection: JalaliMonthSelection) {
  return `${selection.year}-${pad(selection.month)}`;
}

export function parseJalaliMonthKey(value: string): JalaliMonthSelection | null {
  const normalized = normalizeJalaliInput(value);
  const match = /^(\d{4})-(\d{2})$/.exec(normalized);
  if (!match) return null;

  const year = Number(match[1]);
  const month = Number(match[2]);
  if (!Number.isInteger(year) || !Number.isInteger(month) || month < 1 || month > 12) {
    return null;
  }

  return { year, month };
}

export function buildRecentJalaliMonthOptions(
  count = 12,
  reference = new Date(),
  locale: AppLocale = DEFAULT_LOCALE,
) {
  const current = getCurrentJalaliMonthSelection(reference);
  const currentMonthDate = newJalaliDate(current.year, current.month - 1, 1);

  return Array.from({ length: count }, (_, index) => {
    const monthDate = subJalaliMonths(currentMonthDate, index);
    const key = formatJalali(monthDate, "yyyy-MM");
    return {
      value: key,
      label: formatJalaliMonthLabel(monthDate, locale),
    };
  });
}

export function buildBillingSummaryPath(preset: BillingSummaryPreset) {
  return `/api/billing/reports/summary?preset=${encodeURIComponent(preset)}`;
}

export function buildControlPlaneBillingSummaryPath(preset: BillingSummaryPreset) {
  return `/api/v1/billing/reports/summary?preset=${encodeURIComponent(preset)}`;
}

export function buildControlPlaneCreditSummaryPath(preset: BillingSummaryPreset) {
  return `/api/v1/billing/credits/summary?preset=${encodeURIComponent(preset)}`;
}

/**
 * Formats a decimal credit amount (wire format "4.50") with exactly two
 * decimals and locale digits; callers append the localized unit.
 */
export function formatCreditAmount(value: string | number | undefined | null, locale: AppLocale = DEFAULT_LOCALE) {
  const amount = Number(value ?? 0);
  return creditNumberFormatter(locale).format(Number.isFinite(amount) ? amount : 0);
}

export function buildBillingMonthlyPath(selection: JalaliMonthSelection) {
  const params = new URLSearchParams({
    jalali_year: String(selection.year),
    jalali_month: String(selection.month),
  });
  return `/api/billing/reports/monthly?${params.toString()}`;
}

export function buildControlPlaneBillingMonthlyPath(selection: JalaliMonthSelection) {
  const params = new URLSearchParams({
    jalali_year: String(selection.year),
    jalali_month: String(selection.month),
  });
  return `/api/v1/billing/reports/monthly?${params.toString()}`;
}

export function buildBankTransferPayload(input: {
  amountIRR: number;
  trackingNumber: string;
  bankTransactionDateJalali: string;
  receiptFileURL?: string;
}): BankTransferPayload {
  return {
    amount_irr: input.amountIRR,
    tracking_number: input.trackingNumber.trim(),
    bank_transaction_date_jalali: normalizeJalaliInput(input.bankTransactionDateJalali),
    receipt_file_url: input.receiptFileURL,
  };
}

export function normalizeJalaliDateInput(value: string) {
  const normalized = normalizeJalaliInput(value);
  const parts = normalized.split("-");
  if (parts.length !== 3) return normalized;

  const [year, month, day] = parts;
  if (!year || !month || !day) return normalized;
  return `${year.padStart(4, "0")}-${month.padStart(2, "0")}-${day.padStart(2, "0")}`;
}

export function formatSelectedJalaliMonth(value: string, locale: AppLocale = DEFAULT_LOCALE) {
  const selection = parseJalaliMonthKey(value);
  if (!selection) return value;
  return formatJalaliMonthLabel(newJalaliDate(selection.year, selection.month - 1, 1), locale);
}

export function statusTone(status: string): "success" | "warning" | "error" | "pending" | "neutral" {
  if (["successful", "completed", "approved", "paid", "used", "active", "issued"].includes(status)) return "success";
  if (["pending", "pending_review", "draft"].includes(status)) return "pending";
  if (["failed", "rejected", "cancelled", "expired"].includes(status)) return "error";
  if (status === "refunded") return "warning";
  return "neutral";
}

function getCurrentJalaliDateParts(reference: Date): JalaliTodayParts {
  const parts = persianTodayPartsFormatter.formatToParts(reference);
  const year = Number(parts.find((part) => part.type === "year")?.value ?? 0);
  const month = Number(parts.find((part) => part.type === "month")?.value ?? 0);
  const day = Number(parts.find((part) => part.type === "day")?.value ?? 0);

  return { year, month, day };
}

function formatJalaliMonthLabel(value: Date, locale: AppLocale) {
  const parts = financialMonthFormatter(locale).formatToParts(value);
  const month = parts.find((part) => part.type === "month")?.value ?? "";
  const year = parts.find((part) => part.type === "year")?.value ?? "";
  return `${month} ${year}`.trim();
}

function normalizeJalaliInput(value: string) {
  return value
    .trim()
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)));
}

function pad(value: number) {
  return String(value).padStart(2, "0");
}
