import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useCommandShortcut } from "./use-command-shortcut";

describe("useCommandShortcut", () => {
  it("toggles on Ctrl+K and Cmd+K", () => {
    const toggle = vi.fn();
    renderHook(() => useCommandShortcut(toggle));

    window.dispatchEvent(new KeyboardEvent("keydown", { key: "k", ctrlKey: true }));
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "K", metaKey: true }));

    expect(toggle).toHaveBeenCalledTimes(2);
  });

  it("ignores keydown events without a key, as dispatched by browser autofill", () => {
    const toggle = vi.fn();
    renderHook(() => useCommandShortcut(toggle));

    const autofill = new Event("keydown") as KeyboardEvent;
    expect(() => window.dispatchEvent(autofill)).not.toThrow();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "k" }));

    expect(toggle).not.toHaveBeenCalled();
  });
});
