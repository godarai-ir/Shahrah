import { describe, expect, it } from "vitest";

import {
  DEFAULT_LOCALE,
  SUPPORTED_LOCALES,
  getCalendarLocale,
  getIntlLocale,
  isSupportedLocale,
  resolveLocale,
} from "./locale";

describe("supported locales", () => {
  it("exposes Persian first and defaults to it", () => {
    expect(SUPPORTED_LOCALES).toEqual(["fa", "en"]);
    expect(DEFAULT_LOCALE).toBe("fa");
  });

  it.each(["fa", "en"])("accepts %s", (locale) => {
    expect(isSupportedLocale(locale)).toBe(true);
  });

  it.each(["de", "", "fa-IR", null, undefined, 7])("rejects %s as an exact locale", (value) => {
    expect(isSupportedLocale(value)).toBe(false);
  });
});

describe("resolveLocale", () => {
  it.each([
    ["fa", "fa"],
    ["fa-IR", "fa"],
    ["FA_ir", "fa"],
    [" en-US ", "en"],
    ["en", "en"],
  ])("narrows %s to %s", (value, expected) => {
    expect(resolveLocale(value)).toBe(expected);
  });

  it.each([["de-DE"], [""], [null], [undefined], [42]])("falls back to the default for %s", (value) => {
    expect(resolveLocale(value)).toBe(DEFAULT_LOCALE);
  });
});

describe("Intl tags", () => {
  it("uses Persian formatting for fa and US English for en", () => {
    expect(getIntlLocale("fa")).toBe("fa-IR");
    expect(getIntlLocale("en-GB")).toBe("en-US");
  });

  it("keeps the Jalali calendar only for Persian", () => {
    expect(getCalendarLocale("fa-IR")).toBe("fa-IR-u-ca-persian");
    expect(getCalendarLocale("en")).toBe("en-US");
  });
});
