// @vitest-environment jsdom
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { readCollapsibleState, useCollapsibleState, writeCollapsibleState } from "./use-collapsible-state";

const KEY = "logs.rangeSummaryOpen";

describe("useCollapsibleState", () => {
  afterEach(() => window.localStorage.clear());

  it("renders the default first and applies the stored value after mount", async () => {
    window.localStorage.setItem(KEY, "true");
    const seen: boolean[] = [];

    const { result } = renderHook(() => {
      const state = useCollapsibleState(KEY);
      seen.push(state[0]);
      return state;
    });

    expect(seen[0]).toBe(false);
    await waitFor(() => expect(result.current[0]).toBe(true));
  });

  it("stays on the default when nothing is stored", async () => {
    const { result } = renderHook(() => useCollapsibleState(KEY, true));

    await act(async () => {});
    expect(result.current[0]).toBe(true);
  });

  it("persists changes made through the setter", () => {
    const { result } = renderHook(() => useCollapsibleState(KEY));

    act(() => result.current[1](true));

    expect(result.current[0]).toBe(true);
    expect(window.localStorage.getItem(KEY)).toBe("true");
  });

  it("ignores unknown stored values", () => {
    window.localStorage.setItem(KEY, "yes");
    expect(readCollapsibleState(KEY)).toBeNull();

    writeCollapsibleState(KEY, false);
    expect(readCollapsibleState(KEY)).toBe(false);
  });
});
