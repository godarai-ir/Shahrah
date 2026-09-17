"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import type { TableDensity } from "@/components/ui/table";

export type TablePreferences = { density: TableDensity; hiddenColumns: string[] };

const DEFAULT_PREFERENCES: TablePreferences = { density: "compact", hiddenColumns: [] };

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
      density: parsed.density === "relaxed" || parsed.density === "compact" ? parsed.density : fallback.density,
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
    // Storage can be full or blocked; preferences then last for this session only.
  }
}

export function useTablePreferences(tableId: string, defaults?: Partial<TablePreferences>) {
  const defaultsKey = JSON.stringify(defaults ?? {});
  const [preferences, setPreferences] = useState<TablePreferences>(() => ({ ...DEFAULT_PREFERENCES, ...defaults }));

  // Read after mount so the server render and the first client render agree.
  useEffect(() => {
    setPreferences(readTablePreferences(tableId, JSON.parse(defaultsKey) as Partial<TablePreferences>));
  }, [tableId, defaultsKey]);

  const update = useCallback((next: (current: TablePreferences) => TablePreferences) => {
    setPreferences((current) => {
      const value = next(current);
      writeTablePreferences(tableId, value);
      return value;
    });
  }, [tableId]);

  const setDensity = useCallback((density: TableDensity) => update((current) => ({ ...current, density })), [update]);

  const columnVisibility = useMemo<Record<string, boolean>>(
    () => Object.fromEntries(preferences.hiddenColumns.map((id) => [id, false])),
    [preferences.hiddenColumns],
  );

  const setColumnVisibility = useCallback((visibility: Record<string, boolean>) => {
    update((current) => ({
      ...current,
      hiddenColumns: Object.entries(visibility).filter(([, visible]) => !visible).map(([id]) => id),
    }));
  }, [update]);

  return { density: preferences.density, setDensity, columnVisibility, setColumnVisibility };
}
