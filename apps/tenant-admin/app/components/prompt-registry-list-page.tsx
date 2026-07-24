"use client";

import { useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";

import { normalizeError, requestJSON } from "@/app/components/provider-account-wizard/utils";
import { ConfirmDialog, OperationNotice, PageContainer } from "@/components/app-ui";

import { PromptRegistryList } from "./prompt-registry-list";
import type { Notice, Prompt } from "./prompt-management-shared";
import { usePromptRegistryList } from "./use-prompt-registry-list";

export default function PromptRegistryListPage() {
  const t = useTranslations("prompts");
  const locale = useLocale();
  const router = useRouter();
  const [notice, setNotice] = useState<Notice>(null);
  const [pendingDeletePrompt, setPendingDeletePrompt] = useState<Prompt | null>(null);
  const {
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
  } = usePromptRegistryList(t);

  const pageDir = locale === "fa" ? "rtl" : "ltr";

  return (
    <PageContainer padding="none" spacing="compact" dir={pageDir}>
      <OperationNotice
        notice={errorMessage ? { kind: "error", message: errorMessage } : null}
        labels={{ success: t("notice.successTitle"), error: t("notice.errorTitle"), warning: t("notice.errorTitle"), info: t("notice.successTitle") }}
      />
      <OperationNotice
        notice={notice}
        labels={{ success: t("notice.successTitle"), error: t("notice.errorTitle"), warning: t("notice.errorTitle"), info: t("notice.successTitle") }}
      />

      <PromptRegistryList
        isLoading={isLoading}
        locale={locale}
        onClearSearch={() => {
          setQuery("");
          setPageOffset(0);
          void loadPrompts("", 0, pageSize);
        }}
        onDelete={setPendingDeletePrompt}
        onNextPage={() => {
          const nextOffset = pageOffset + pageSize;
          setPageOffset(nextOffset);
          void loadPrompts(query, nextOffset, pageSize);
        }}
        onPreviousPage={() => {
          const nextOffset = Math.max(pageOffset - pageSize, 0);
          setPageOffset(nextOffset);
          void loadPrompts(query, nextOffset, pageSize);
        }}
        onRefresh={() => {
          setPageOffset(0);
          void loadPrompts(query, 0, pageSize);
        }}
        onRowClick={(prompt) => router.push(`/prompts/${prompt.id}`)}
        onRowsPerPageChange={(next) => {
          setPageSize(next);
          setPageOffset(0);
          void loadPrompts(query, 0, next);
        }}
        onSearchKeyDown={(key) => {
          if (key === "Enter") {
            setPageOffset(0);
            void loadPrompts(query, 0, pageSize);
          }
        }}
        onSearchValueChange={setQuery}
        pageOffset={pageOffset}
        pageSize={pageSize}
        promptTotal={promptTotal}
        prompts={prompts}
        query={query}
        t={t}
        userLabelByID={userLabelByID}
      />
      <ConfirmDialog open={pendingDeletePrompt !== null} onOpenChange={(open) => { if (!open) setPendingDeletePrompt(null); }} title={t("actions.delete")} description={t("messages.deletePromptConfirm", { name: pendingDeletePrompt?.name ?? "" })} cancelLabel={t("actions.cancel")} confirmLabel={t("actions.delete")} pendingLabel={t("actions.delete")} onConfirm={() => pendingDeletePrompt ? deletePrompt(pendingDeletePrompt) : undefined} />
    </PageContainer>
  );

  async function deletePrompt(prompt: Prompt) {
    try {
      await requestJSON(`/api/prompts/${encodeURIComponent(prompt.id)}`, { method: "DELETE" });
      setNotice({ kind: "success", message: t("messages.promptDeleted") });
      const nextOffset = prompts.length === 1 ? Math.max(pageOffset - pageSize, 0) : pageOffset;
      setPageOffset(nextOffset);
      await loadPrompts(query, nextOffset, pageSize);
      setPendingDeletePrompt(null);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.deletePromptError")) });
    }
  }
}
