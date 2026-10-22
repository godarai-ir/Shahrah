"use client";

import { useCallback, useMemo, useSyncExternalStore } from "react";

export type TablePreferences = { hiddenColumns: string[] };

const DEFAULT_PREFERENCES: TablePreferences = { hiddenColumns: [] };

export function tablePreferencesKey(tableId: string) {
  return `table-prefs:${tableId}`;
}

function getStorage(): Storage | null {
  try {
    return typeof window === "undefined" ? null : window.localStorage;
  } catch {
    return null;
  }
}

export function readTablePreferences(tableId: string, defaults?: Partial<TablePreferences>): TablePreferences {
  const fallback = { ...DEFAULT_PREFERENCES, ...defaults };
  try {
    const raw = getStorage()?.getItem(tablePreferencesKey(tableId));
    if (!raw) return fallback;
    const parsed = JSON.parse(raw) as Partial<TablePreferences>;
    return {
      hiddenColumns: Array.isArray(parsed.hiddenColumns)
        ? parsed.hiddenColumns.filter((id): id is string => typeof id === "string")
        : fallback.hiddenColumns,
    };
  } catch {
    return fallback;
  }
}

export function writeTablePreferences(tableId: string, preferences: TablePreferences) {
  try {
    getStorage()?.setItem(tablePreferencesKey(tableId), JSON.stringify(preferences));
  } catch {
    // Storage can be full or blocked; the table then keeps its defaults.
  }
}

const listeners = new Set<() => void>();

function subscribe(listener: () => void) {
  listeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    listeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

function readRaw(tableId: string): string | null {
  try {
    return getStorage()?.getItem(tablePreferencesKey(tableId)) ?? null;
  } catch {
    return null;
  }
}

export function useTablePreferences(tableId: string, defaults?: Partial<TablePreferences>) {
  const defaultsKey = JSON.stringify(defaults ?? {});
  // The server snapshot is always null, so the first client render matches SSR before storage is read.
  const raw = useSyncExternalStore(subscribe, () => readRaw(tableId), () => null);

  const preferences = useMemo<TablePreferences>(() => {
    const parsedDefaults = JSON.parse(defaultsKey) as Partial<TablePreferences>;
    return raw === null ? { ...DEFAULT_PREFERENCES, ...parsedDefaults } : readTablePreferences(tableId, parsedDefaults);
  }, [raw, tableId, defaultsKey]);

  const update = useCallback((value: TablePreferences) => {
    writeTablePreferences(tableId, value);
    listeners.forEach((listener) => listener());
  }, [tableId]);

  const columnVisibility = useMemo<Record<string, boolean>>(
    () => Object.fromEntries(preferences.hiddenColumns.map((id) => [id, false])),
    [preferences.hiddenColumns],
  );

  const setColumnVisibility = useCallback((visibility: Record<string, boolean>) => {
    update({
      ...preferences,
      hiddenColumns: Object.entries(visibility).filter(([, visible]) => !visible).map(([id]) => id),
    });
  }, [preferences, update]);

  return { columnVisibility, setColumnVisibility };
}
