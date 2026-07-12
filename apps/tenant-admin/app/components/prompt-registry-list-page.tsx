"use client";

import { useState } from "react";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { Plus } from "lucide-react";

import { normalizeError, requestJSON } from "@/app/components/provider-account-wizard/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

import { PromptRegistryList } from "./prompt-registry-list";
import type { Notice, Prompt } from "./prompt-management-shared";
import { usePromptRegistryList } from "./use-prompt-registry-list";

export default function PromptRegistryListPage() {
  const t = useTranslations("prompts");
  const locale = useLocale();
  const router = useRouter();
  const [notice, setNotice] = useState<Notice>(null);
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
    <div className="flex flex-col gap-5" dir={pageDir}>
      <section className="flex flex-col gap-3 rounded-xl border bg-card p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t("header.title")}</h2>
            <p className="text-sm text-muted-foreground">{t("header.description")}</p>
          </div>
          <Button type="button" onClick={() => router.push("/prompts/new")}>
            <Plus data-icon="inline-start" />
            {t("actions.createPrompt")}
          </Button>
        </div>

        {errorMessage ? (
          <Alert variant="destructive">
            <AlertTitle>{t("notice.errorTitle")}</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        ) : null}
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("notice.errorTitle") : t("notice.successTitle")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <PromptRegistryList
          isLoading={isLoading}
          locale={locale}
          onCopyFqn={(prompt) => void copyFqn(prompt)}
          onCreateVersion={(prompt) => router.push(`/prompts/${prompt.id}/edit?intent=create-version`)}
          onDelete={(prompt) => void deletePrompt(prompt)}
          onEditLatestVersion={(prompt) => router.push(`/prompts/${prompt.id}/edit?intent=edit-latest`)}
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
          onTryPrompt={(prompt) => router.push(`/prompts/${prompt.id}/edit?intent=try`)}
          pageOffset={pageOffset}
          pageSize={pageSize}
          promptTotal={promptTotal}
          prompts={prompts}
          query={query}
          t={t}
          userLabelByID={userLabelByID}
        />
      </section>
    </div>
  );

  async function copyFqn(prompt: Prompt) {
    const fqn = prompt.latest_version_detail?.fqn?.trim();
    if (!fqn) {
      setNotice({ kind: "error", message: t("messages.fqnUnavailable") });
      return;
    }
    await copyTextToClipboard(fqn);
    setNotice({ kind: "success", message: t("messages.fqnCopied") });
  }

  async function deletePrompt(prompt: Prompt) {
    const confirmed = window.confirm(t("messages.deletePromptConfirm", { name: prompt.name }));
    if (!confirmed) return;
    try {
      await requestJSON(`/api/prompts/${encodeURIComponent(prompt.id)}`, { method: "DELETE" });
      setNotice({ kind: "success", message: t("messages.promptDeleted") });
      const nextOffset = prompts.length === 1 ? Math.max(pageOffset - pageSize, 0) : pageOffset;
      setPageOffset(nextOffset);
      await loadPrompts(query, nextOffset, pageSize);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.deletePromptError")) });
    }
  }
}
