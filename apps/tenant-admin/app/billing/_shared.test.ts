import { describe, expect, it } from "vitest";

import {
  buildBankTransferPayload,
  buildBillingMonthlyPath,
  buildBillingSummaryPath,
  formatDateOnly,
  getCurrentJalaliDateValue,
  getCurrentJalaliMonthRange,
  getCurrentJalaliMonthSelection,
} from "./_shared";

describe("billing jalali helpers", () => {
  it("derives the current Jalali month in Asia/Tehran", () => {
    const reference = new Date("2026-06-21T20:35:00Z");

    expect(getCurrentJalaliMonthSelection(reference)).toEqual({ year: 1405, month: 4 });
    expect(getCurrentJalaliDateValue(reference)).toBe("1405-04-01");
    expect(getCurrentJalaliMonthRange(reference)).toEqual({
      from: new Date("2026-06-21T20:30:00.000Z"),
      to: reference,
    });
  });

  it("builds explicit Jalali billing report queries", () => {
    expect(buildBillingSummaryPath("current_month")).toBe("/api/billing/reports/summary?preset=current_month");
    expect(buildBillingMonthlyPath({ year: 1405, month: 4 })).toBe("/api/billing/reports/monthly?jalali_year=1405&jalali_month=4");
  });

  it("normalizes Jalali bank transfer payloads", () => {
    expect(buildBankTransferPayload({
      amountIRR: 250000,
      trackingNumber: " ABC-123 ",
      bankTransactionDateJalali: "۱۴۰۵-۰۴-۱۷",
    })).toEqual({
      amount_irr: 250000,
      tracking_number: "ABC-123",
      bank_transaction_date_jalali: "1405-04-17",
      receipt_file_url: undefined,
    });
  });

  it("formats visible dates in the Persian calendar", () => {
    expect(formatDateOnly("2026-07-08T08:30:00Z")).toContain("۱۴۰۵");
  });
});
