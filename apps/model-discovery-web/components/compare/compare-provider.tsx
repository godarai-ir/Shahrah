"use client";

import { createContext, useCallback, useContext, useMemo, useState, useSyncExternalStore } from "react";
import { useTranslations } from "next-intl";
import {
  dedupeCompareIds,
  getStoredCompareIdsServerSnapshot,
  getStoredCompareIdsSnapshot,
  MAX_COMPARE_MODELS,
  parseCompareIds,
  subscribeStoredCompareIds,
  type CompareModelId,
  writeStoredCompareIds,
} from "@/lib/compare-models";

type CompareContextValue = {
  selectedIds: CompareModelId[];
  selectedCount: number;
  canAdd: boolean;
  notice: string | null;
  add: (id: CompareModelId) => boolean;
  remove: (id: CompareModelId) => void;
  clear: () => void;
  setFromIds: (ids: CompareModelId[]) => void;
  isSelected: (id: CompareModelId) => boolean;
};

const CompareContext = createContext<CompareContextValue | null>(null);

export function CompareProvider({ children }: { children: React.ReactNode }) {
  const t = useTranslations("Compare");
  const selectedIds = useSyncExternalStore(
    subscribeStoredCompareIds,
    getStoredCompareIdsSnapshot,
    getStoredCompareIdsServerSnapshot,
  );
  const [notice, setNotice] = useState<string | null>(null);

  const persist = useCallback((ids: CompareModelId[]) => {
    const next = dedupeCompareIds(ids).slice(0, MAX_COMPARE_MODELS);
    writeStoredCompareIds(next);
    return next;
  }, []);

  const add = useCallback((id: CompareModelId) => {
    const normalized = id.trim();
    if (!normalized) return false;
    if (selectedIds.includes(normalized)) return true;
    if (selectedIds.length >= MAX_COMPARE_MODELS) {
      setNotice(t("limitNotice", { count: MAX_COMPARE_MODELS }));
      return false;
    }
    persist([...selectedIds, normalized]);
    setNotice(null);
    return true;
  }, [persist, selectedIds, t]);

  const remove = useCallback((id: CompareModelId) => {
    persist(selectedIds.filter((x) => x !== id));
  }, [persist, selectedIds]);

  const clear = useCallback(() => {
    persist([]);
    setNotice(null);
  }, [persist]);

  const setFromIds = useCallback((ids: CompareModelId[]) => {
    persist(ids);
    setNotice(null);
  }, [persist]);

  const value = useMemo<CompareContextValue>(
    () => ({
      selectedIds,
      selectedCount: selectedIds.length,
      canAdd: selectedIds.length < MAX_COMPARE_MODELS,
      notice,
      add,
      remove,
      clear,
      setFromIds,
      isSelected: (id) => selectedIds.includes(id),
    }),
    [add, clear, notice, remove, selectedIds, setFromIds],
  );

  return <CompareContext.Provider value={value}>{children}</CompareContext.Provider>;
}

export function useCompareModels() {
  const context = useContext(CompareContext);
  if (!context) {
    throw new Error("useCompareModels must be used within CompareProvider");
  }
  return context;
}

export function useCompareIdsFromQuery(idsParam: string | null | undefined): CompareModelId[] {
  return parseCompareIds(idsParam);
}

