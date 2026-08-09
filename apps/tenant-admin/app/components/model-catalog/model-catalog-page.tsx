"use client";

import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  BookOpen,
  Boxes,
  CircleGauge,
  Eye,
  KeyRound,
  Layers3,
  Network,
  Plus,
  ShieldCheck,
  Trash2,
} from "lucide-react";

import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import type {
  ProviderAccountReadiness,
  ProviderCatalogProvider,
  ProvidersResponse,
  TenantProviderAccount,
  TenantProviderAccountSummary,
} from "@/app/components/provider-account-wizard/types";
import {
  buildModelDiscoveryProviderHref,
  formatProviderDisplayName,
  normalizeError,
  requestJSON,
} from "@/app/components/provider-account-wizard/utils";
import { ActiveFilterChips, type ActiveFilterChip } from "@/components/app-ui/ActiveFilterChips";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DataListPanel } from "@/components/app-ui/DataListPanel";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DebouncedTextFilter, FilterField } from "@/components/app-ui/FilterField";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { ModelCatalogExplorer } from "./model-catalog-explorer";
import { classifyProviderAccountDeleteError } from "./provider-account-delete-error";
import type {
  CatalogModel,
  CatalogModelResponse,
  ProviderAccountsResponse,
} from "./types";
import {
  parseModelCatalogURLState,
  patchModelCatalogURL,
  type ModelCatalogTab,
} from "./url-state";

type ModelCatalogPageProps = {
  docsHref: string;
};

export function ModelCatalogPage({ docsHref }: ModelCatalogPageProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();
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
  const [models, setModels] = useState<CatalogModel[]>([]);
  const [catalogModelTotal, setCatalogModelTotal] = useState(0);
  const [loadingAccounts, setLoadingAccounts] = useState(true);
  const [loadingModels, setLoadingModels] = useState(true);
  const [loadedModelScope, setLoadedModelScope] = useState("");
  const [modelLoadError, setModelLoadError] = useState<string | null>(null);
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
    Promise.allSettled([
      requestJSON<ProviderAccountsResponse>("/api/tenant-provider-accounts"),
      requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
      requestJSON<CatalogModelResponse>("/api/tenant-provider-account-models?limit=1"),
    ])
      .then(([accountResult, providerResult, modelResult]) => {
        if (cancelled) return;
        const failures: string[] = [];
        if (accountResult.status === "fulfilled") {
          setAccounts(accountResult.value.data ?? []);
          setSummaries(accountResult.value.summaries ?? {});
        } else {
          failures.push(normalizeError(accountResult.reason, t("errors.loadProviders")));
        }
        if (providerResult.status === "fulfilled") {
          setProviders(providerResult.value.data ?? []);
        } else {
          failures.push(t("errors.loadProviderMetadata"));
        }
        if (modelResult.status === "fulfilled") {
          setCatalogModelTotal(modelResult.value.total ?? 0);
        } else {
          failures.push(t("errors.loadModelCount"));
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
  const providerOptions = useMemo(() => {
    const values = new Map<string, string>();
    for (const provider of providers) {
      if (provider.enabled) values.set(
        provider.provider,
        formatProviderDisplayName(provider.provider, provider.display_name),
      );
    }
    for (const account of accounts) {
      if (!values.has(account.provider)) values.set(account.provider, account.provider);
    }
    return [...values.entries()]
      .map(([value, label]) => ({ value, label }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [accounts, providers]);
  const accountProviderSlugs = useMemo(
    () => [...new Set(accounts.map((account) => account.provider))]
      .sort((a, b) => formatProviderDisplayName(a, providerBySlug.get(a)?.display_name)
        .localeCompare(formatProviderDisplayName(b, providerBySlug.get(b)?.display_name))),
    [accounts, providerBySlug],
  );
  const selectedProvider = accountProviderSlugs.includes(state.modelProvider)
    ? state.modelProvider
    : accountProviderSlugs[0] ?? "";
  const selectedProviderAccounts = useMemo(
    () => accounts.filter((account) => account.provider === selectedProvider),
    [accounts, selectedProvider],
  );
  const selectedAccountID = selectedProviderAccounts.some((account) => account.id === state.modelAccount)
    ? state.modelAccount
    : selectedProviderAccounts[0]?.id ?? "";
  const selectedModelScope = selectedProvider && selectedAccountID
    ? `${selectedProvider}:${selectedAccountID}`
    : "";
  const modelsPending = loadingModels || loadedModelScope !== selectedModelScope;

  useEffect(() => {
    if (state.tab !== "models" || loadingAccounts) return;
    if (selectedProvider !== state.modelProvider) {
      updateURL({ model_provider: selectedProvider || null, model_account: null, model: null });
      return;
    }
    if (selectedAccountID !== state.modelAccount) {
      updateURL({ model_account: selectedAccountID || null, model: null });
    }
  }, [loadingAccounts, selectedAccountID, selectedProvider, state.modelAccount, state.modelProvider, state.tab, updateURL]);

  useEffect(() => {
    if (state.tab !== "models") return;
    if (!selectedProvider || !selectedAccountID) {
      setModels([]);
      setLoadedModelScope("");
      setModelLoadError(null);
      setLoadingModels(false);
      return;
    }
    let cancelled = false;
    const controller = new AbortController();
    setLoadingModels(true);
    setModelLoadError(null);
    requestAllAccountModels(selectedProvider, selectedAccountID, controller.signal)
      .then((accountModels) => {
        if (cancelled) return;
        setModels(accountModels);
        setLoadedModelScope(selectedModelScope);
      })
      .catch((error) => {
        if (cancelled) return;
        setModels([]);
        setLoadedModelScope(selectedModelScope);
        setModelLoadError(normalizeError(error, t("errors.loadModels")));
      })
      .finally(() => {
        if (!cancelled) setLoadingModels(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [reloadKey, selectedAccountID, selectedModelScope, selectedProvider, state.tab, t]);
  const filteredAccounts = useMemo(() => {
    const query = state.q.toLocaleLowerCase();
    return accounts.filter((account) => {
      const summary = summaries[account.id];
      if (query && !`${account.name} ${account.provider}`.toLocaleLowerCase().includes(query)) return false;
      if (state.providers.length && !state.providers.includes(account.provider)) return false;
      if (state.accounts.length && !state.accounts.includes(account.id)) return false;
      if (state.credential) {
        const credential = account.use_platform_managed_credentials ? "managed" : "byok";
        if (credential !== state.credential) return false;
      }
      if (state.readiness && summary?.readiness !== state.readiness) return false;
      return true;
    });
  }, [accounts, state.accounts, state.credential, state.providers, state.q, state.readiness, summaries]);

  const accountModels = useMemo(
    () => models.filter(
      (model) => model.provider === selectedProvider
        && model.availability.some((availability) => availability.account_id === selectedAccountID),
    ),
    [models, selectedAccountID, selectedProvider],
  );
  const selectedModel = useMemo(
    () => accountModels.find((model) => model.model === state.model) ?? accountModels[0] ?? null,
    [accountModels, state.model],
  );

  useEffect(() => {
    if (state.tab !== "models" || modelsPending) return;
    const nextModel = selectedModel?.model ?? "";
    if (nextModel !== state.model) updateURL({ model: nextModel || null });
  }, [modelsPending, selectedModel, state.model, state.tab, updateURL]);
  const filterChips = useMemo<ActiveFilterChip[]>(() => {
    const chips: ActiveFilterChip[] = [];
    for (const provider of state.providers) {
      chips.push({
        key: `provider:${provider}`,
        label: t("filters.provider"),
        value: formatProviderDisplayName(provider, providerBySlug.get(provider)?.display_name),
      });
    }
    for (const accountID of state.accounts) {
      const account = accounts.find((item) => item.id === accountID);
      chips.push({
        key: `account:${accountID}`,
        label: t("filters.account"),
        value: account
          ? `${account.name} · ${formatProviderDisplayName(account.provider, providerBySlug.get(account.provider)?.display_name)}`
          : accountID,
      });
    }
    if (state.credential) {
      chips.push({
        key: "credential",
        label: t("filters.credential"),
        value: t(`credential.${state.credential === "managed" ? "managed" : "byok"}`),
      });
    }
    if (state.readiness) {
      chips.push({
        key: "readiness",
        label: t("filters.readiness"),
        value: readinessLabel(t, state.readiness as ProviderAccountReadiness),
      });
    }
    return chips;
  }, [accounts, providerBySlug, state.accounts, state.credential, state.providers, state.readiness, t]);

  function removeFilter(key: string) {
    if (key.startsWith("provider:")) {
      updateURL({ providers: state.providers.filter((value) => value !== key.slice("provider:".length)) });
      return;
    }
    if (key.startsWith("account:")) {
      updateURL({ accounts: state.accounts.filter((value) => value !== key.slice("account:".length)) });
      return;
    }
    updateURL({ [key]: null });
  }

  function changeTab(tab: string) {
    const nextTab = tab as ModelCatalogTab;
    if (nextTab === "models") setLoadingModels(true);
    updateURL({
      tab: nextTab,
      account: null,
      section: null,
      model_provider: null,
      model: null,
      model_account: null,
    }, "push");
  }

  async function testConnection(account: TenantProviderAccount) {
    setNotice(null);
    try {
      const response = await requestJSON<{ success?: boolean; message?: string }>(
        `/api/tenant-provider-accounts/${encodeURIComponent(account.id)}/test-credentials`,
        { method: "POST", body: JSON.stringify({}) },
      );
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
    await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(deleteTarget.id)}`, {
      method: "DELETE",
    });
    showFeedbackToast({ kind: "success", title: t("notices.deleted", { account: deleteTarget.name }) });
    setDeleteNotice(null);
    setDeleteTarget(null);
    setReloadKey((value) => value + 1);
  }

  const noResults = (state.q || filterChips.length > 0) && filteredAccounts.length === 0;
  const formattedAccountCount = new Intl.NumberFormat(locale).format(accounts.length);
  const formattedCatalogModelCount = new Intl.NumberFormat(locale).format(catalogModelTotal);

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

      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <Tabs value={state.tab} onValueChange={changeTab} className="min-w-0 flex-1">
          <TabsList variant="line" aria-label={t("tabs.label")} className="grid h-11 w-full grid-cols-2 gap-1 p-0 md:w-auto md:min-w-96">
            <TabsTrigger value="providers" className="h-11 min-w-0 justify-center px-2 sm:px-4">
              <Network data-icon="inline-start" aria-hidden="true" />
              <span className="truncate">{t("tabs.providers")}</span>
              <Badge variant="secondary" className="shrink-0">{formattedAccountCount}</Badge>
            </TabsTrigger>
            <TabsTrigger value="models" className="h-11 min-w-0 justify-center px-2 sm:px-4">
              <Boxes data-icon="inline-start" aria-hidden="true" />
              <span className="truncate">{t("tabs.models")}</span>
              <Badge variant="secondary" className="shrink-0">{formattedCatalogModelCount}</Badge>
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="flex w-full items-center gap-2 md:w-auto md:shrink-0">
          <RefreshButton
            label={t("actions.refresh")}
            loadingLabel={t("loading")}
            isLoading={loadingAccounts || (state.tab === "models" && loadingModels)}
            onClick={() => {
              setLoadingAccounts(true);
              if (state.tab === "models") setLoadingModels(true);
              setReloadKey((value) => value + 1);
            }}
          />
          <Button asChild className="min-w-0 flex-1 md:flex-none">
            <Link href="/provider-accounts/new">
              <Plus data-icon="inline-start" aria-hidden="true" />
              {t("actions.addProvider")}
            </Link>
          </Button>
        </div>
      </div>

      {state.tab === "providers" ? (
        <DataListPanel
          aria-label={t("providers.ariaLabel")}
          toolbarClassName="@container rounded-none border-0 bg-transparent p-0 shadow-none"
          contentClassName="@container border-card-border shadow-xs"
          toolbar={
            <div className="flex min-w-0 flex-col gap-4">
              <div className="grid gap-3 @lg:grid-cols-2">
                <DebouncedTextFilter
                  label={t("filters.search")}
                  hideLabel
                  value={state.q}
                  onValueChange={(value) => updateURL({ q: value })}
                  placeholder={t("filters.searchProviders")}
                />
                <FilterField label={t("filters.provider")} htmlFor="model-catalog-provider-filter" hideLabel>
                  <FilterCombobox
                    id="model-catalog-provider-filter"
                    options={providerOptions.filter((option) => !state.providers.includes(option.value))}
                    value=""
                    onValueChange={(value) => updateURL({ providers: [...state.providers, value] })}
                    placeholder={t("filters.providers")}
                    emptyText={t("filters.noProvider")}
                  />
                </FilterField>
              </div>
              {filterChips.length > 0 ? (
                <div className="border-t border-border/60 pt-3">
                  <ActiveFilterChips items={filterChips} onRemove={removeFilter} getRemoveLabel={(item) => t("filters.remove", { filter: item.label, value: item.value ?? "" })} />
                </div>
              ) : null}
            </div>
          }
          loading={loadingAccounts}
          loadingLabel={t("loading")}
          loadingRows={7}
          empty={filteredAccounts.length === 0}
          emptyState={
            <EmptyState
              icon={Network}
              title={noResults || state.q || filterChips.length ? t("empty.noResultsTitle") : t("empty.noProvidersTitle")}
              description={noResults || state.q || filterChips.length ? t("empty.noResultsDescription") : t("empty.noProvidersDescription")}
              primaryAction={state.q || filterChips.length ? (
                <Button type="button" variant="outline" onClick={() => updateURL({ q: null, providers: null, accounts: null, credential: null, readiness: null })}>
                  {t("actions.clearFilters")}
                </Button>
              ) : (
                <Button asChild><Link href="/provider-accounts/new"><Plus data-icon="inline-start" aria-hidden="true" />{t("actions.addProvider")}</Link></Button>
              )}
            />
          }
        >
          <ProvidersTable
            accounts={filteredAccounts}
            summaries={summaries}
            providerBySlug={providerBySlug}
            docsHref={docsHref}
            onViewModels={(account) => {
              setLoadingModels(true);
              updateURL({
                tab: "models",
                q: null,
                providers: null,
                accounts: null,
                credential: null,
                readiness: null,
                account: null,
                section: null,
                model_provider: account.provider,
                model_account: account.id,
                model: null,
              }, "push");
            }}
            onTestConnection={(account) => void testConnection(account)}
            onDelete={(account) => {
              setNotice(null);
              setDeleteNotice(null);
              setDeleteTarget(account);
            }}
          />
        </DataListPanel>
      ) : (
        <ModelCatalogExplorer
          accounts={accounts}
          docsHref={docsHref}
          summaries={summaries}
          models={accountModels}
          providerBySlug={providerBySlug}
          selectedProvider={selectedProvider}
          selectedAccountID={selectedAccountID}
          selectedModel={selectedModel}
          loadingAccounts={loadingAccounts}
          loadingModels={modelsPending}
          modelLoadError={modelLoadError}
          onSelectProvider={(provider) => {
            setLoadingModels(true);
            updateURL({ model_provider: provider, model_account: null, model: null }, "push");
          }}
          onSelectAccount={(accountID) => {
            setLoadingModels(true);
            updateURL({ model_provider: selectedProvider, model_account: accountID, model: null }, "push");
          }}
          onSelectModel={(model) => updateURL({ model_provider: selectedProvider, model_account: selectedAccountID, model: model.model }, "push")}
          onTestConnection={(account) => void testConnection(account)}
          onDeleteAccount={(account) => {
            setNotice(null);
            setDeleteNotice(null);
            setDeleteTarget(account);
          }}
        />
      )}

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
    const payload = await requestJSON<CatalogModelResponse>(
      `/api/tenant-provider-account-models?${params.toString()}`,
      { signal },
    );
    const page = payload.data ?? [];
    models.push(...page);
    total = payload.total ?? models.length;
    offset += page.length;
    if (page.length === 0) break;
  } while (offset < total);

  return models;
}

type ProvidersTableProps = {
  accounts: TenantProviderAccount[];
  summaries: Record<string, TenantProviderAccountSummary>;
  providerBySlug: Map<string, ProviderCatalogProvider>;
  docsHref: string;
  onViewModels: (account: TenantProviderAccount) => void;
  onTestConnection: (account: TenantProviderAccount) => void;
  onDelete: (account: TenantProviderAccount) => void;
};

function ProvidersTable({
  accounts,
  summaries,
  providerBySlug,
  docsHref,
  onViewModels,
  onTestConnection,
  onDelete,
}: ProvidersTableProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();

  return (
    <Table className="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead>{t("columns.name")}</TableHead>
          <TableHead className="hidden w-48 @4xl:table-cell">{t("columns.connectionIdentifier")}</TableHead>
          <TableHead className="hidden w-20 @sm:table-cell">{t("columns.models")}</TableHead>
          <TableHead className="hidden w-32 @2xl:table-cell">{t("columns.credential")}</TableHead>
          <TableHead className="w-36">{t("columns.readiness")}</TableHead>
          <TableHead className="hidden w-44 @5xl:table-cell">{t("columns.updated")}</TableHead>
          <TableHead className="w-16"><span className="sr-only">{t("columns.actions")}</span></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {accounts.map((account) => {
          const provider = providerBySlug.get(account.provider);
          const summary = summaries[account.id];
          const detailsHref = `/provider-accounts?${new URLSearchParams({ account: account.id, section: "overview" }).toString()}`;
          return (
            <TableRow key={account.id}>
              <TableCell>
                <Link href={detailsHref} className="flex min-w-0 items-center gap-3 font-medium hover:underline">
                  <ProviderLogo
                    provider={account.provider}
                    displayName={formatProviderDisplayName(account.provider, provider?.display_name)}
                    logoURL={provider?.logo_url}
                    size={24}
                  />
                  <span className="min-w-0">
                    <span className="block truncate" dir="ltr">{account.name}</span>
                    <span className="block truncate text-xs font-normal text-muted-foreground" dir="ltr">
                      {formatProviderDisplayName(account.provider, provider?.display_name)}
                    </span>
                  </span>
                </Link>
              </TableCell>
              <TableCell className="hidden @4xl:table-cell">
                <div className="flex min-w-0 items-center gap-1" dir="ltr">
                  <code className="truncate font-mono text-xs" title={`${account.provider}:${account.name}`}>
                    {account.provider}:{account.name}
                  </code>
                  <CopyAction
                    value={`${account.provider}:${account.name}`}
                    copyLabel={t("actions.copyConnectionIdentifier")}
                    copyingLabel={t("actions.copying")}
                    copiedLabel={t("actions.copied")}
                  />
                </div>
              </TableCell>
              <TableCell className="hidden @sm:table-cell">{summary?.model_count ?? 0}</TableCell>
              <TableCell className="hidden @2xl:table-cell">
                <StatusBadge status="neutral">
                  {account.use_platform_managed_credentials ? t("credential.managed") : t("credential.byok")}
                </StatusBadge>
              </TableCell>
              <TableCell>
                <ReadinessBadge readiness={summary?.readiness ?? "needs_models"} />
              </TableCell>
              <TableCell className="hidden whitespace-nowrap text-muted-foreground @5xl:table-cell">
                <DateTimeValue value={account.updated_at} locale={locale} fallback={t("values.na")} />
              </TableCell>
              <TableCell>
                <RowActionsMenu
                  label={t("actions.openMenu", { account: account.name })}
                  dir={locale === "fa" ? "rtl" : "ltr"}
                  actions={[
                    {
                      label: t("actions.viewModels"),
                      icon: <Layers3 />,
                      onSelect: () => onViewModels(account),
                    },
                    {
                      label: t("actions.details"),
                      icon: <Eye />,
                      href: detailsHref,
                    },
                    {
                      label: t("actions.effectiveAccess"),
                      icon: <ShieldCheck />,
                      href: `/provider-accounts?${new URLSearchParams({ account: account.id, section: "access" }).toString()}`,
                    },
                    {
                      label: t("actions.connectionTest"),
                      icon: <CircleGauge />,
                      onSelect: () => onTestConnection(account),
                    },
                    {
                      label: t("actions.docs"),
                      icon: <BookOpen />,
                      href: provider ? buildModelDiscoveryProviderHref(provider.provider) : docsHref,
                      external: true,
                    },
                    {
                      label: t("actions.delete"),
                      icon: <Trash2 />,
                      destructive: true,
                      separatorBefore: true,
                      onSelect: () => onDelete(account),
                    },
                  ]}
                />
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}

function ReadinessBadge({ readiness }: { readiness: ProviderAccountReadiness }) {
  const t = useTranslations("modelCatalog");
  const status = readiness === "ready"
    ? "success"
    : readiness === "needs_credentials" || readiness === "needs_access"
      ? "error"
      : "warning";
  const icon = readiness === "needs_credentials"
    ? <KeyRound data-icon="inline-start" aria-hidden="true" />
    : readiness === "needs_access"
      ? <ShieldCheck data-icon="inline-start" aria-hidden="true" />
      : readiness === "needs_models"
        ? <Layers3 data-icon="inline-start" aria-hidden="true" />
        : readiness === "pricing_incomplete"
          ? <CircleGauge data-icon="inline-start" aria-hidden="true" />
          : null;
  return (
    <StatusBadge status={status}>
      {icon}
      {readinessLabel(t, readiness)}
    </StatusBadge>
  );
}

function readinessLabel(
  t: ReturnType<typeof useTranslations>,
  readiness: ProviderAccountReadiness,
): string {
  return t(`readiness.${readiness}`);
}
