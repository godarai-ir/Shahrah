"use client";

import { useCallback, useEffect, useState } from "react";

import { getAPIErrorMessage } from "@/lib/http/api-client";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { listPrompts, listPromptUsers } from "@/features/prompts/api";

import type {
  PromptTranslationFn,
  TenantAdminUserLookup,
} from "@/lib/prompt-builder";
import type { Prompt } from "@/lib/prompt-builder";

export function usePromptRegistryList(t: PromptTranslationFn) {
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState("");
  const [pageSize, setPageSize] = useState(25);
  const [pageOffset, setPageOffset] = useState(0);
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [promptTotal, setPromptTotal] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [userLabelByID, setUserLabelByID] = useState<Record<string, string>>({});

  const loadPrompts = useCallback(
    async (nextQuery: string, offset: number, limit: number) => {
      setIsLoading(true);
      try {
        const payload = await listPrompts(nextQuery, offset, limit);
        const rows = payload.data ?? [];
        setPrompts(rows);
        setPromptTotal(payload.total ?? rows.length);
        setErrorMessage(null);
        await loadUserLabels(rows);
      } catch (error) {
        setErrorMessage(getAPIErrorMessage(error, t("messages.loadPromptsError")));
      } finally {
        setIsLoading(false);
      }
    },
    [t],
  );

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadPrompts("", 0, 25);
  }, [loadPrompts]);

  return {
    errorMessage,
    isLoading,
    loadPrompts,
    pageOffset,
    pageSize,
    promptTotal,
    prompts,
    query,
    setPageOffset,
    setPageSize,
    setQuery,
    userLabelByID,
  };

  async function loadUserLabels(rows: Prompt[]) {
    const userIDs = [...new Set(rows.map((item) => (item.created_by ?? "").trim()).filter(Boolean))];
    if (userIDs.length === 0) {
      setUserLabelByID({});
      return;
    }
    try {
      const payload = await listPromptUsers();
      const users = payload.data ?? [];
      setUserLabelByID(
        Object.fromEntries(
          userIDs.map((userID) => {
            const match = users.find((item) => item.id === userID);
            return [userID, resolveUserLabel(match, userID)];
          }),
        ),
      );
    } catch {
      setUserLabelByID(Object.fromEntries(userIDs.map((userID) => [userID, userID])));
    }
  }
}

function resolveUserLabel(user: TenantAdminUserLookup | undefined, fallbackID: string) {
  if (!user) return fallbackID;
  return user.display_name?.trim() || user.email?.trim() || (mobileAuthUIEnabled ? user.mobile?.trim() : "") || user.id || fallbackID;
}
