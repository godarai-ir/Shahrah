import { afterEach, describe, expect, it, vi } from "vitest";

import { readTablePreferences, tablePreferencesKey, writeTablePreferences } from "./use-table-preferences";

function stubLocalStorage() {
  const store = new Map<string, string>();
  vi.stubGlobal("window", {
    localStorage: {
      getItem: (key: string) => store.get(key) ?? null,
      setItem: (key: string, value: string) => void store.set(key, value),
    },
  });
  return store;
}

describe("table preferences storage", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("writes and reads preferences under the table key", () => {
    const store = stubLocalStorage();
    writeTablePreferences("users", { hiddenColumns: ["usage"] });

    expect(store.has(tablePreferencesKey("users"))).toBe(true);
    expect(readTablePreferences("users")).toEqual({ hiddenColumns: ["usage"] });
  });

  it("falls back to defaults for missing or corrupt values", () => {
    const store = stubLocalStorage();
    expect(readTablePreferences("users", { hiddenColumns: ["usage"] })).toEqual({ hiddenColumns: ["usage"] });

    store.set(tablePreferencesKey("users"), "{not json");
    expect(readTablePreferences("users")).toEqual({ hiddenColumns: [] });
  });

  it("does not throw when localStorage is unavailable (SSR)", () => {
    expect(typeof window).toBe("undefined");
    expect(() => writeTablePreferences("users", { hiddenColumns: [] })).not.toThrow();
    expect(readTablePreferences("users")).toEqual({ hiddenColumns: [] });
  });

  it("does not throw when localStorage access is blocked", () => {
    vi.stubGlobal("window", {
      get localStorage(): Storage {
        throw new Error("SecurityError");
      },
    });
    expect(() => writeTablePreferences("users", { hiddenColumns: [] })).not.toThrow();
    expect(readTablePreferences("users").hiddenColumns).toEqual([]);
  });
});
