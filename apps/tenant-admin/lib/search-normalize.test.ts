import { describe, expect, it } from "vitest";

import { matchesSearch, normalizeSearchText } from "./search-normalize";

describe("normalizeSearchText", () => {
  it("treats ZWNJ, a space, and no separator the same", () => {
    const expected = normalizeSearchText("لاگها");
    expect(normalizeSearchText("لاگ‌ها")).toBe(expected);
    expect(normalizeSearchText("لاگ ها")).toBe(expected);
  });

  it("folds Arabic kaf and ye into their Persian forms", () => {
    expect(normalizeSearchText("كيف")).toBe(normalizeSearchText("کیف"));
  });

  it("maps Persian and Arabic digits to ASCII", () => {
    expect(normalizeSearchText("۱۲۳")).toBe("123");
    expect(normalizeSearchText("١٢٣")).toBe("123");
  });

  it("drops diacritics and tatweel", () => {
    expect(normalizeSearchText("مـُدل")).toBe(normalizeSearchText("مدل"));
  });
});

describe("matchesSearch", () => {
  it("matches any candidate by substring, case-insensitively", () => {
    expect(matchesSearch(["لاگ‌ها", "logs"], "LOG")).toBe(true);
    expect(matchesSearch(["لاگ‌ها"], "لاگ ها")).toBe(true);
    expect(matchesSearch(["کاربران"], "logs")).toBe(false);
  });

  it("matches everything for an empty query", () => {
    expect(matchesSearch(["x"], "  ")).toBe(true);
  });
});
