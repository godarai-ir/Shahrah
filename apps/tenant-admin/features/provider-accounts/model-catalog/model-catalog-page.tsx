"use client";

import { useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import type {
  ProviderCatalogProvider,
  TenantProviderAccount,
  TenantProviderAccountSummary,
} from "@/features/provider-accounts/types";
import { normalizeError } from "@/features/provider-accounts/utils";
import {
  deleteProviderAccount,
  getCatalogModels,
  getProviderAccounts,
  getProviderCatalog,
  testProviderAccountCredentials,
} from "@/features/provider-accounts/api";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";

import { groupAccountsByProvider } from "./account-groups";
import { ModelCatalogExplorer } from "./model-catalog-explorer";
import { ModelDetailsSheet } from "./model-details-sheet";
import { classifyProviderAccountDeleteError } from "./provider-account-delete-error";
import { isModelReadinessSummary } from "./readiness-summary";
import type { CatalogModel } from "./types";
import { parseModelCatalogURLState, patchModelCatalogURL } from "./url-state";

type ModelCatalogPageProps = {
  docsHref: string;
};

export function ModelCatalogPage({ docsHref }: ModelCatalogPageProps) {
  const t = useTranslations("modelCatalog");
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryString = searchParams.toString();
  const state = useMemo(
    () => parseModelCatalogURLState(new URLSearchParams(queryString)),
    [queryString],
  );

  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [summaries, setSummaries] = useState<Record<string, TenantProviderAccountSummary>>({});
  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [modelLoad, setModelLoad] = useState<{
    scope: string;
    pendingScope: string;
    models: CatalogModel[];
    error: string | null;
  }>({ scope: "", pendingScope: "", models: [], error: null });
  const [loadingAccounts, setLoadingAccounts] = useState(true);
  const [notice, setNotice] = useState<OperationNoticeValue | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<OperationNoticeValue | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<TenantProviderAccount | null>(null);
  const [reloadKey, setReloadKey] = useState(0);

  const localizeDeleteError = useCallback(
    (error: unknown): string => {
      switch (classifyProviderAccountDeleteError(error)) {
        case "activeAccessKeys":
          return t("errors.deleteActiveAccessKeys");
        case "forbidden":
          return t("errors.deleteForbidden");
        case "notFound":
          return t("errors.deleteNotFound");
        default:
          return t("errors.deleteConflict");
      }
    },
    [t],
  );

  const updateURL = useCallback(
    (
      patch: Record<string, string | string[] | null | undefined>,
      mode: "push" | "replace" = "replace",
    ) => {
      const next = patchModelCatalogURL(new URLSearchParams(queryString), patch);
      const href = next ? `${pathname}?${next}` : pathname;
      if (mode === "push") router.push(href);
      else router.replace(href);
    },
    [pathname, queryString, router],
  );

  useEffect(() => {
    let cancelled = false;
    Promise.allSettled([getProviderAccounts(), getProviderCatalog()])
      .then(([accountResult, providerResult]) => {
        if (cancelled) return;
        const failures: string[] = [];
        if (accountResult.status === "fulfilled") {
          setAccounts(accountResult.value.data ?? []);
          setSummaries(accountResult.value.summaries ?? {});
        } else {
          failures.push(normalizeError(accountResult.reason, t("errors.loadProviders")));
        }
        if (providerResult.status === "fulfilled") {
          setProviders(providerResult.value);
        } else {
          failures.push(t("errors.loadProviderMetadata"));
        }
        if (failures.length > 0) {
          setNotice({
            kind: accountResult.status === "rejected" ? "error" : "warning",
            message: failures.join(" "),
          });
        }
      })
      .finally(() => {
        if (!cancelled) setLoadingAccounts(false);
      });
    return () => {
      cancelled = true;
    };
  }, [reloadKey, t]);

  const providerBySlug = useMemo(
    () => new Map(providers.map((provider) => [provider.provider, provider])),
    [providers],
  );
  // Default to the first account in the same provider-grouped order the explorer renders.
  const firstAccount = useMemo(
    () => groupAccountsByProvider(accounts, providerBySlug)[0]?.accounts[0] ?? null,
    [accounts, providerBySlug],
  );
  const selectedAccount = accounts.find((account) => account.id === state.modelAccount) ?? firstAccount;
  const selectedProvider = selectedAccount?.provider ?? "";
  const selectedAccountID = selectedAccount?.id ?? "";
  const readinessFilter = isModelReadinessSummary(state.readiness) ? state.readiness : "";
  const selectedModelScope = selectedProvider && selectedAccountID
    ? `${selectedProvider}:${selectedAccountID}`
    : "";
  const modelsPending = modelLoad.scope !== selectedModelScope
    || modelLoad.pendingScope === selectedModelScope;
  const loadingModels = (loadingAccounts || modelsPending) && modelLoad.models.length === 0;
  const refreshingModels = !loadingAccounts && modelsPending && modelLoad.models.length > 0;
  const modelAccountID = modelLoad.scope
    ? modelLoad.scope.slice(modelLoad.scope.indexOf(":") + 1)
    : selectedAccountID;

  useEffect(() => {
    if (loadingAccounts) return;
    if (selectedAccountID !== state.modelAccount) {
      updateURL({ model_provider: selectedProvider || null, model_account: selectedAccountID || null, model: null });
      return;
    }
    if (selectedProvider !== state.modelProvider) {
      updateURL({ model_provider: selectedProvider || null });
    }
  }, [loadingAccounts, selectedAccountID, selectedProvider, state.modelAccount, state.modelProvider, updateURL]);

  useEffect(() => {
    if (!selectedProvider || !selectedAccountID) {
      let cancelled = false;
      queueMicrotask(() => {
        if (cancelled) return;
        setModelLoad({ scope: "", pendingScope: "", models: [], error: null });
      });
      return () => {
        cancelled = true;
      };
    }
    let cancelled = false;
    const controller = new AbortController();
    queueMicrotask(() => {
      if (cancelled) return;
      setModelLoad((current) => ({
        ...current,
        pendingScope: selectedModelScope,
        error: null,
      }));
    });
    requestAllAccountModels(selectedProvider, selectedAccountID, controller.signal)
      .then((accountModels) => {
        if (cancelled) return;
        setModelLoad({
          scope: selectedModelScope,
          pendingScope: "",
          models: accountModels,
          error: null,
        });
      })
      .catch((error) => {
        if (cancelled) return;
        setModelLoad({
          scope: selectedModelScope,
          pendingScope: "",
          models: [],
          error: normalizeError(error, t("errors.loadModels")),
        });
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [reloadKey, selectedAccountID, selectedModelScope, selectedProvider, t]);

  const accountModels = useMemo(
    () => modelLoad.models,
    [modelLoad.models],
  );
  // A model is only selected when the URL names one; the selection is what opens the details sheet.
  const selectedModel = useMemo(
    () => (state.model ? accountModels.find((model) => model.model === state.model) ?? null : null),
    [accountModels, state.model],
  );

  useEffect(() => {
    if (modelsPending || !state.model || selectedModel) return;
    updateURL({ model: null });
  }, [modelsPending, selectedModel, state.model, updateURL]);

  async function testConnection(account: TenantProviderAccount) {
    setNotice(null);
    try {
      const response = await testProviderAccountCredentials(account.id);
      setNotice({
        kind: response.success === false ? "error" : "success",
        message: response.message || t("notices.connectionTestSucceeded", { account: account.name }),
      });
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("errors.connectionTest")) });
    }
  }

  async function deleteAccount() {
    if (!deleteTarget) return;
    await deleteProviderAccount(deleteTarget.id);
    showFeedbackToast({ kind: "success", title: t("notices.deleted", { account: deleteTarget.name }) });
    setDeleteNotice(null);
    setDeleteTarget(null);
    setReloadKey((value) => value + 1);
  }

  return (
    <PageContainer as="main" size="full">
      <OperationNotice
        notice={notice}
        labels={{
          success: t("notices.success"),
          error: t("notices.error"),
          warning: t("notices.warning"),
          info: t("notices.info"),
        }}
      />

      <ModelCatalogExplorer
        accounts={accounts}
        docsHref={docsHref}
        summaries={summaries}
        models={accountModels}
        providerBySlug={providerBySlug}
        selectedAccountID={selectedAccountID}
        modelAccountID={modelAccountID}
        selectedModel={selectedModel}
        query={state.q}
        readinessFilter={readinessFilter}
        loadingAccounts={loadingAccounts}
        loadingModels={loadingModels}
        refreshingModels={refreshingModels}
        modelLoadError={modelLoad.error}
        onQueryChange={(q) => updateURL({ q })}
        onReadinessFilterChange={(readiness) => updateURL({ readiness })}
        onSelectAccount={(account) => {
          updateURL({ model_provider: account.provider, model_account: account.id, model: null }, "push");
        }}
        onSelectModel={(model) => updateURL({ model_provider: selectedProvider, model_account: selectedAccountID, model: model.model }, "push")}
        onTestConnection={(account) => void testConnection(account)}
        onDeleteAccount={(account) => {
          setNotice(null);
          setDeleteNotice(null);
          setDeleteTarget(account);
        }}
      />

      <ModelDetailsSheet
        model={selectedModel}
        accountID={modelAccountID}
        onOpenChange={(open) => {
          if (!open) updateURL({ model: null }, "push");
        }}
      />

      <TypedConfirmDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setDeleteNotice(null);
            setDeleteTarget(null);
          }
        }}
        title={t("delete.title")}
        description={deleteTarget ? t("delete.description", { account: deleteTarget.name }) : ""}
        expectedValue={deleteTarget?.name ?? ""}
        inputLabel={t("delete.inputLabel")}
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.delete")}
        pendingLabel={t("actions.deleting")}
        warning={deleteTarget ? (
          <OperationNotice
            notice={{
              kind: "warning",
              message: t("delete.modelCount", {
                count: summaries[deleteTarget.id]?.model_count ?? 0,
              }),
            }}
            labels={{
              success: t("notices.success"),
              error: t("notices.error"),
              warning: t("notices.warning"),
              info: t("notices.info"),
            }}
          />
        ) : null}
        feedback={deleteNotice ? (
          <OperationNotice
            notice={deleteNotice}
            labels={{
              success: t("notices.success"),
              error: t("notices.error"),
              warning: t("notices.warning"),
              info: t("notices.info"),
            }}
          />
        ) : null}
        onConfirm={deleteAccount}
        onConfirmError={(error) =>
          setDeleteNotice({ kind: "error", message: localizeDeleteError(error) })
        }
      />
    </PageContainer>
  );
}

async function requestAllAccountModels(
  provider: string,
  accountID: string,
  signal: AbortSignal,
): Promise<CatalogModel[]> {
  const pageSize = 100;
  const models: CatalogModel[] = [];
  let offset = 0;
  let total = 0;

  do {
    const params = new URLSearchParams({
      limit: String(pageSize),
      offset: String(offset),
      sort: "name",
      order: "asc",
      providers: provider,
      account_ids: accountID,
    });
    const payload = await getCatalogModels(params.toString(), signal);
    const page = payload.data ?? [];
    models.push(...page);
    total = payload.total ?? models.length;
    offset += page.length;
    if (page.length === 0) break;
  } while (offset < total);

  return models;
}
