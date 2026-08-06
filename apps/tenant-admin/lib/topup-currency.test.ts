import { describe, expect, it } from "vitest";

import { convertTomanToUSD, formatApproximateUSD, parsePositiveToman } from "./topup-currency";

describe("top-up currency helpers", () => {
  it("parses localized positive toman amounts", () => {
    expect(parsePositiveToman("۱٬۰۰۰٬۰۰۰")).toBe(1_000_000);
    expect(parsePositiveToman("1,000,000")).toBe(1_000_000);
  });

  it("rejects empty, non-numeric, zero, and negative amounts", () => {
    expect(parsePositiveToman("")).toBeNull();
    expect(parsePositiveToman("invalid")).toBeNull();
    expect(parsePositiveToman("0")).toBeNull();
    expect(parsePositiveToman("-1")).toBeNull();
  });

  it("converts the IRR quote to toman before calculating USD", () => {
    expect(convertTomanToUSD(10_000_000, 2_000_000)).toBe(50);
  });

  it("formats the result as an approximate dollar amount", () => {
    expect(formatApproximateUSD(12.345)).toBe("~ $12.35");
    expect(formatApproximateUSD(0)).toBeNull();
  });
});
