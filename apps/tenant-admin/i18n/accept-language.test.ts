import { describe, expect, it } from "vitest";

import { matchAcceptLanguage } from "./request";

describe("matchAcceptLanguage", () => {
  it("returns null when the browser advertises nothing usable", () => {
    expect(matchAcceptLanguage(null)).toBeNull();
    expect(matchAcceptLanguage("")).toBeNull();
    expect(matchAcceptLanguage("de-DE,fr;q=0.8")).toBeNull();
  });

  it("picks the highest-quality supported language", () => {
    expect(matchAcceptLanguage("de-DE,en-US;q=0.9,fa;q=0.5")).toBe("en");
    expect(matchAcceptLanguage("de-DE,en-US;q=0.4,fa;q=0.8")).toBe("fa");
  });

  it("ignores languages the visitor explicitly refuses", () => {
    expect(matchAcceptLanguage("en;q=0,fa;q=0.3")).toBe("fa");
  });

  it("treats a wildcard as the default locale", () => {
    expect(matchAcceptLanguage("*")).toBe("fa");
  });

  it("matches on the language subtag", () => {
    expect(matchAcceptLanguage("en-GB")).toBe("en");
    expect(matchAcceptLanguage("fa-IR")).toBe("fa");
  });
});
