import { describe, expect, it } from "vitest";

import { appendDisplaySuffix, formatCompactCurrency, formatCompactIrr, formatCompactNumber } from "./compact-number";

describe("formatCompactNumber", () => {
  it("keeps smaller values exact", () => {
    expect(formatCompactNumber(9_999, { locale: "en-US" })).toEqual({
      value: "9,999",
      fullValue: "9,999",
      isCompact: false,
    });
  });

  it("formats large values with familiar suffixes", () => {
    expect(formatCompactNumber(12_340, { locale: "en-US" }).value).toBe("12.3K");
    expect(formatCompactNumber(120_000, { locale: "en-US" }).value).toBe("120K");
    expect(formatCompactNumber(1_250_000, { locale: "en-US" }).value).toBe("1.3M");
  });

  it("rolls over to the next suffix after rounding", () => {
    expect(formatCompactNumber(999_950, { locale: "en-US" }).value).toBe("1M");
  });

  it("uses localized digits while keeping compact suffixes", () => {
    expect(formatCompactNumber(12_340, { locale: "fa-IR" }).value).toBe("۱۲٫۳K");
  });
});

describe("formatCompactCurrency", () => {
  it("keeps smaller values exact", () => {
    expect(formatCompactCurrency(52.125, { locale: "en-US", currency: "USD", maximumFractionDigits: 3 })).toEqual({
      value: "$52.125",
      fullValue: "$52.125",
      isCompact: false,
    });
  });

  it("formats large currency values compactly", () => {
    expect(formatCompactCurrency(12_340, { locale: "en-US", currency: "USD", maximumFractionDigits: 2 }).value).toBe("$12.3K");
  });

  it("rounds IRR values without fractional digits", () => {
    expect(formatCompactCurrency(852.85, { locale: "fa-IR", currency: "IRR" }).fullValue).not.toMatch(/[٫.]/);
    expect(formatCompactCurrency(12_340, { locale: "en-US", currency: "IRR" }).value).not.toMatch(/[٫.]/);
  });
});

describe("formatCompactIrr", () => {
  it("keeps IRR exact displays rounded to whole rials", () => {
    expect(formatCompactIrr(852.85, { locale: "fa-IR" })).toEqual({
      value: "۸۵۳ IRR",
      fullValue: "۸۵۳ IRR",
      isCompact: false,
    });
  });

  it("keeps IRR compact displays free of fractional digits", () => {
    expect(formatCompactIrr(12_340, { locale: "en-US" })).toEqual({
      value: "12K IRR",
      fullValue: "12,340 IRR",
      isCompact: true,
    });
  });
});

describe("appendDisplaySuffix", () => {
  it("adds suffixes to both compact and exact values", () => {
    expect(
      appendDisplaySuffix(
        { value: "12.3K", fullValue: "12,340", isCompact: true },
        " requests",
      ),
    ).toEqual({
      value: "12.3K requests",
      fullValue: "12,340 requests",
      isCompact: true,
    });
  });
});
