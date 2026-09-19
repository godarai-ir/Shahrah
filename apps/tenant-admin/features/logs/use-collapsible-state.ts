"use client";

import { useCallback, useEffect, useState } from "react";

function getStorage(): Storage | null {
  try {
    return typeof window === "undefined" ? null : window.localStorage;
  } catch {
    return null;
  }
}

export function readCollapsibleState(key: string): boolean | null {
  try {
    const raw = getStorage()?.getItem(key);
    if (raw === "true") return true;
    if (raw === "false") return false;
    return null;
  } catch {
    return null;
  }
}

export function writeCollapsibleState(key: string, value: boolean) {
  try {
    getStorage()?.setItem(key, String(value));
  } catch {
    // Storage can be full or blocked; the section then opens with its default.
  }
}

/**
 * Remembers whether a collapsible section is open. The first render always uses
 * `initial` so server and client markup match; the stored value is applied after mount.
 */
export function useCollapsibleState(key: string, initial = false): [boolean, (value: boolean) => void] {
  const [open, setOpenState] = useState(initial);

  useEffect(() => {
    let disposed = false;
    queueMicrotask(() => {
      if (disposed) return;
      const stored = readCollapsibleState(key);
      if (stored !== null) setOpenState(stored);
    });
    return () => {
      disposed = true;
    };
  }, [key]);

  const setOpen = useCallback((value: boolean) => {
    setOpenState(value);
    writeCollapsibleState(key, value);
  }, [key]);

  return [open, setOpen];
}
