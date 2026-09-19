"use client";

import { useEffect } from "react";

/**
 * Toggles on Cmd+K / Ctrl+K from anywhere, including text inputs. Editors that need the chord
 * themselves (JSON and prompt editors) call preventDefault, which this handler respects.
 */
export function useCommandShortcut(toggle: () => void) {
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      // Browser autofill dispatches keydown events without a key.
      if (typeof event.key !== "string" || event.key.toLowerCase() !== "k" || !(event.metaKey || event.ctrlKey)) return;
      if (event.defaultPrevented) return;
      event.preventDefault();
      toggle();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggle]);
}
