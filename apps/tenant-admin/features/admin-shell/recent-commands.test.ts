import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  RECENT_COMMANDS_STORAGE_KEY,
  pushRecentCommandID,
  readRecentCommandIDs,
  resolveRecentCommands,
} from "./recent-commands";

describe("recent commands", () => {
  let store: Map<string, string>;

  beforeEach(() => {
    store = new Map();
    vi.stubGlobal("window", {
      localStorage: {
        getItem: (key: string) => store.get(key) ?? null,
        setItem: (key: string, value: string) => store.set(key, value),
      },
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("keeps the newest five, newest first", () => {
    for (const id of ["a", "b", "c", "d", "e", "f"]) pushRecentCommandID(id);
    expect(readRecentCommandIDs()).toEqual(["f", "e", "d", "c", "b"]);
  });

  it("moves a repeated id to the front instead of duplicating it", () => {
    pushRecentCommandID("a");
    pushRecentCommandID("b");
    pushRecentCommandID("a");
    expect(readRecentCommandIDs()).toEqual(["a", "b"]);
  });

  it("ignores corrupt storage", () => {
    store.set(RECENT_COMMANDS_STORAGE_KEY, "{not json");
    expect(readRecentCommandIDs()).toEqual([]);
    store.set(RECENT_COMMANDS_STORAGE_KEY, JSON.stringify(["a", 3, null]));
    expect(readRecentCommandIDs()).toEqual(["a"]);
  });

  it("silently drops ids that no longer resolve to a permitted command", () => {
    const byId = new Map([["nav:logs", { id: "nav:logs" }]]);
    expect(resolveRecentCommands(["nav:gone", "nav:logs"], byId)).toEqual([{ id: "nav:logs" }]);
  });
});
