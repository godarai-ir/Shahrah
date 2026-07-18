import { describe, expect, it } from "vitest";

import { getLocaleDirection } from "./locale-direction";

describe("getLocaleDirection", () => {
  it.each(["fa", "fa-IR", "fa_IR", " FA-ir "])("returns rtl for Persian locale %s", (locale) => {
    expect(getLocaleDirection(locale)).toBe("rtl");
  });

  it.each(["en", "en-US", "en_GB"])("returns ltr for English locale %s", (locale) => {
    expect(getLocaleDirection(locale)).toBe("ltr");
  });

  it("defaults unknown locales to ltr", () => {
    expect(getLocaleDirection("unknown")).toBe("ltr");
  });
});
