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
  FileSearch,
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
import { TablePagination } from "@/components/app-ui/TablePagination";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { ModelDetailsSheet } from "./model-details-sheet";
import { formatCatalogModelDisplayName } from "./formatters";
import { classifyProviderAccountDeleteError } from "./provider-account-delete-error";
import type {
  CatalogModel,
  CatalogModelResponse,
  ModelAvailability,
  ProviderAccountsResponse,
} from "./types";
import {
  parseModelCatalogURLState,
  patchModelCatalogURL,
  type ModelCatalogTab,
} from "./url-state";

const MODEL_PAGE_SIZE = 50;

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
  const [filteredModelTotal, setFilteredModelTotal] = useState(0);
  const [modelOffset, setModelOffset] = useState(0);
  const [loadingAccounts, setLoadingAccounts] = useState(true);
  const [loadingModels, setLoadingModels] = useState(false);
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
      if (["q", "providers", "accounts", "sort", "order"].some((key) => key in patch)) {
        setModelOffset(0);
      }
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

  useEffect(() => {
    if (state.tab !== "models") return;
    let cancelled = false;
    const params = new URLSearchParams({
      limit: String(MODEL_PAGE_SIZE),
      offset: String(modelOffset),
      sort: state.sort,
      order: state.order,
    });
    if (state.q) params.set("q", state.q);
    if (state.providers.length) params.set("providers", state.providers.join(","));
    if (state.accounts.length) params.set("account_ids", state.accounts.join(","));
    void Promise.resolve().then(() => {
      if (!cancelled) setLoadingModels(true);
      return requestJSON<CatalogModelResponse>(`/api/tenant-provider-account-models?${params.toString()}`);
    })
      .then((payload) => {
        if (cancelled) return;
        setModels(payload.data ?? []);
        setFilteredModelTotal(payload.total ?? 0);
      })
      .catch((error) => {
        if (cancelled) return;
        setModels([]);
        setFilteredModelTotal(0);
        setNotice({ kind: "error", message: normalizeError(error, t("errors.loadModels")) });
      })
      .finally(() => {
        if (!cancelled) setLoadingModels(false);
      });
    return () => {
      cancelled = true;
    };
  }, [
    modelOffset,
    reloadKey,
    state.accounts,
    state.order,
    state.providers,
    state.q,
    state.sort,
    state.tab,
    t,
  ]);

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
  const accountOptions = useMemo(
    () =>
      accounts
        .map((account) => {
          const providerMetadata = providerBySlug.get(account.provider);
          const providerName = formatProviderDisplayName(account.provider, providerMetadata?.display_name);
          return {
            value: account.id,
            label: `${account.name} · ${providerName}`,
          };
        })
        .sort((a, b) => a.label.localeCompare(b.label)),
    [accounts, providerBySlug],
  );
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

  const selectedModel = useMemo(
    () =>
      models.find(
        (model) => model.provider === state.modelProvider && model.model === state.model,
      ) ?? null,
    [models, state.model, state.modelProvider],
  );
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
  const modelStart = filteredModelTotal === 0 ? 0 : modelOffset + 1;
  const modelEnd = Math.min(modelOffset + MODEL_PAGE_SIZE, filteredModelTotal);
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

      <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-4 md:flex-row md:items-center md:justify-between">
        <Tabs value={state.tab} onValueChange={changeTab} className="min-w-0 max-w-full">
          <TabsList variant="line" aria-label={t("tabs.label")}>
            <TabsTrigger value="providers">
              <Network data-icon="inline-start" aria-hidden="true" />
              {t("tabs.providers")}
              <Badge variant="secondary">{formattedAccountCount}</Badge>
            </TabsTrigger>
            <TabsTrigger value="models">
              <Boxes data-icon="inline-start" aria-hidden="true" />
              {t("tabs.models")}
              <Badge variant="secondary">{formattedCatalogModelCount}</Badge>
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="flex flex-wrap gap-2">
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
          <Button asChild>
            <Link href="/provider-accounts/new">
              <Plus data-icon="inline-start" aria-hidden="true" />
              {t("actions.addProvider")}
            </Link>
          </Button>
        </div>
      </div>

      <DataListPanel
        aria-label={state.tab === "providers" ? t("providers.ariaLabel") : t("models.ariaLabel")}
        toolbarClassName="@container"
        contentClassName="@container"
        toolbar={
          <div className="flex min-w-0 flex-col gap-3">
            <div className="grid gap-3 @lg:grid-cols-2 @5xl:grid-cols-4">
              <DebouncedTextFilter
                label={t("filters.search")}
                value={state.q}
                onValueChange={(value) => updateURL({ q: value })}
                placeholder={state.tab === "providers" ? t("filters.searchProviders") : t("filters.searchModels")}
              />
              <FilterField label={t("filters.provider")} htmlFor="model-catalog-provider-filter">
                <FilterCombobox
                  id="model-catalog-provider-filter"
                  options={providerOptions.filter((option) => !state.providers.includes(option.value))}
                  value=""
                  onValueChange={(value) => updateURL({ providers: [...state.providers, value] })}
                  placeholder={t("filters.addProvider")}
                  emptyText={t("filters.noProvider")}
                />
              </FilterField>
              <FilterField label={t("filters.account")} htmlFor="model-catalog-account-filter">
                <FilterCombobox
                  id="model-catalog-account-filter"
                  options={accountOptions.filter((option) => !state.accounts.includes(option.value))}
                  value=""
                  onValueChange={(value) => updateURL({ accounts: [...state.accounts, value] })}
                  placeholder={t("filters.addAccount")}
                  emptyText={t("filters.noAccount")}
                />
              </FilterField>
              {state.tab === "providers" ? (
                <div className="grid gap-2 @sm:grid-cols-2">
                  <FilterField label={t("filters.credential")} htmlFor="model-catalog-credential-filter">
                    <FilterCombobox
                      id="model-catalog-credential-filter"
                      options={[
                        { value: "managed", label: t("credential.managed") },
                        { value: "byok", label: t("credential.byok") },
                      ]}
                      value={state.credential}
                      onValueChange={(value) => updateURL({ credential: value })}
                      placeholder={t("filters.any")}
                      emptyText={t("filters.noResult")}
                      searchable={false}
                    />
                  </FilterField>
                  <FilterField label={t("filters.readiness")} htmlFor="model-catalog-readiness-filter">
                    <FilterCombobox
                      id="model-catalog-readiness-filter"
                      options={READINESS_VALUES.map((value) => ({
                        value,
                        label: readinessLabel(t, value),
                      }))}
                      value={state.readiness}
                      onValueChange={(value) => updateURL({ readiness: value })}
                      placeholder={t("filters.any")}
                      emptyText={t("filters.noResult")}
                      searchable={false}
                    />
                  </FilterField>
                </div>
              ) : (
                <div className="grid gap-2 @sm:grid-cols-2">
                  <FilterField label={t("filters.sort")} htmlFor="model-catalog-sort">
                    <FilterCombobox
                      id="model-catalog-sort"
                      options={[
                        "name",
                        "provider",
                        "availability",
                        "context",
                        "input_price",
                        "output_price",
                      ].map((value) => ({ value, label: t(`sort.${value}`) }))}
                      value={state.sort}
                      onValueChange={(value) => updateURL({ sort: value })}
                      placeholder={t("filters.sort")}
                      emptyText={t("filters.noResult")}
                      searchable={false}
                    />
                  </FilterField>
                  <FilterField label={t("filters.order")} htmlFor="model-catalog-order">
                    <FilterCombobox
                      id="model-catalog-order"
                      options={[
                        { value: "asc", label: t("sort.asc") },
                        { value: "desc", label: t("sort.desc") },
                      ]}
                      value={state.order}
                      onValueChange={(value) => updateURL({ order: value })}
                      placeholder={t("filters.order")}
                      emptyText={t("filters.noResult")}
                      searchable={false}
                    />
                  </FilterField>
                </div>
              )}
            </div>
            <ActiveFilterChips
              items={filterChips}
              onRemove={removeFilter}
              getRemoveLabel={(item) => t("filters.remove", { filter: item.label, value: item.value ?? "" })}
            />
          </div>
        }
        toolbarMeta={
          state.tab === "providers"
            ? t("providers.resultCount", { count: filteredAccounts.length })
            : t("models.resultCount", { count: filteredModelTotal })
        }
        loading={state.tab === "providers" ? loadingAccounts : loadingModels}
        loadingLabel={t("loading")}
        loadingRows={7}
        empty={state.tab === "providers" ? filteredAccounts.length === 0 : models.length === 0}
        emptyState={
          <EmptyState
            icon={state.tab === "providers" ? Network : Boxes}
            title={
              noResults || state.q || filterChips.length
                ? t("empty.noResultsTitle")
                : state.tab === "providers"
                  ? t("empty.noProvidersTitle")
                  : t("empty.noModelsTitle")
            }
            description={
              noResults || state.q || filterChips.length
                ? t("empty.noResultsDescription")
                : state.tab === "providers"
                  ? t("empty.noProvidersDescription")
                  : t("empty.noModelsDescription")
            }
            primaryAction={
              state.q || filterChips.length ? (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => updateURL({ q: null, providers: null, accounts: null, credential: null, readiness: null })}
                >
                  {t("actions.clearFilters")}
                </Button>
              ) : (
                <Button asChild>
                  <Link href="/provider-accounts/new">
                    <Plus data-icon="inline-start" aria-hidden="true" />
                    {t("actions.addProvider")}
                  </Link>
                </Button>
              )
            }
          />
        }
        pagination={
          state.tab === "models" && filteredModelTotal > MODEL_PAGE_SIZE ? (
            <TablePagination
              label={t("pagination.label")}
              summary={t("pagination.summary", {
                start: modelStart,
                end: modelEnd,
                total: filteredModelTotal,
              })}
              previousLabel={t("pagination.previous")}
              nextLabel={t("pagination.next")}
              onPrevious={() => setModelOffset((value) => Math.max(0, value - MODEL_PAGE_SIZE))}
              onNext={() => setModelOffset((value) => value + MODEL_PAGE_SIZE)}
              canPrevious={modelOffset > 0}
              canNext={modelEnd < filteredModelTotal}
              disabled={loadingModels}
              page={Math.floor(modelOffset / MODEL_PAGE_SIZE) + 1}
              pageSize={MODEL_PAGE_SIZE}
              pageSizeLabel={t("pagination.pageSize")}
            />
          ) : null
        }
      >
        {state.tab === "providers" ? (
          <ProvidersTable
            accounts={filteredAccounts}
            summaries={summaries}
            providerBySlug={providerBySlug}
            docsHref={docsHref}
            onViewModels={(account) =>
              updateURL({
                tab: "models",
                accounts: [account.id],
                account: null,
                section: null,
              }, "push")
            }
            onTestConnection={(account) => void testConnection(account)}
            onDelete={(account) => {
              setNotice(null);
              setDeleteNotice(null);
              setDeleteTarget(account);
            }}
          />
        ) : (
          <ModelsTable
            models={models}
            providerBySlug={providerBySlug}
            onOpen={(model, availability) =>
              updateURL({
                model_provider: model.provider,
                model: model.model,
                model_account: availability.account_id,
              }, "push")
            }
          />
        )}
      </DataListPanel>

      <ModelDetailsSheet
        model={selectedModel}
        accountID={state.modelAccount}
        onAccountChange={(accountID) => updateURL({ model_account: accountID })}
        onOpenChange={(open) => {
          if (!open) {
            updateURL({ model_provider: null, model: null, model_account: null });
          }
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

type ModelsTableProps = {
  models: CatalogModel[];
  providerBySlug: Map<string, ProviderCatalogProvider>;
  onOpen: (model: CatalogModel, availability: ModelAvailability) => void;
};

function ModelsTable({ models, providerBySlug, onOpen }: ModelsTableProps) {
  const t = useTranslations("modelCatalog");
  return (
    <Table className="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead>{t("columns.model")}</TableHead>
          <TableHead className="hidden @lg:table-cell">{t("columns.provider")}</TableHead>
          <TableHead className="w-28">{t("columns.availability")}</TableHead>
          <TableHead className="hidden @2xl:table-cell">{t("columns.modalities")}</TableHead>
          <TableHead className="hidden @5xl:table-cell">{t("columns.context")}</TableHead>
          <TableHead className="hidden w-32 @2xl:table-cell">{t("columns.pricing")}</TableHead>
          <TableHead className="w-16"><span className="sr-only">{t("columns.actions")}</span></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {models.map((model) => {
          const provider = providerBySlug.get(model.provider);
          const availability = model.availability[0];
          return (
            <TableRow key={`${model.provider}:${model.model}`}>
              <TableCell>
                <button
                  type="button"
                  className="flex min-w-0 flex-col text-start hover:underline"
                  onClick={() => {
                    if (availability) onOpen(model, availability);
                  }}
                >
                  <span className="max-w-52 truncate font-medium sm:max-w-80" dir="ltr" title={formatCatalogModelDisplayName(model.model, model.display_name)}>
                    {formatCatalogModelDisplayName(model.model, model.display_name)}
                  </span>
                  <code className="max-w-52 truncate text-xs text-muted-foreground sm:max-w-80" dir="ltr" title={model.model}>
                    {model.model}
                  </code>
                  <span className="text-xs text-muted-foreground md:hidden" dir="ltr">
                    {formatProviderDisplayName(model.provider, provider?.display_name)}
                  </span>
                </button>
              </TableCell>
              <TableCell className="hidden @lg:table-cell">
                <div className="flex items-center gap-2" dir="ltr">
                  <ProviderLogo
                    provider={model.provider}
                    displayName={formatProviderDisplayName(model.provider, provider?.display_name)}
                    logoURL={provider?.logo_url}
                  />
                  <span>{formatProviderDisplayName(model.provider, provider?.display_name)}</span>
                </div>
              </TableCell>
              <TableCell>
                <AvailabilityPopover model={model} onOpen={onOpen} />
              </TableCell>
              <TableCell className="hidden @2xl:table-cell">
                <div className="flex max-w-60 flex-wrap gap-1">
                  {[...(model.input_modalities ?? []), ...(model.output_modalities ?? [])]
                    .filter((value, index, values) => values.indexOf(value) === index)
                    .map((modality) => (
                      <Badge key={modality} variant="outline" title={modality}>
                        {formatModalityLabel(t, modality)}
                      </Badge>
                    ))}
                </div>
              </TableCell>
              <TableCell className="hidden @5xl:table-cell" dir="ltr">{formatContext(model.context_window, t("values.na"))}</TableCell>
              <TableCell className="hidden @2xl:table-cell">
                <div className="flex flex-col text-xs" dir="ltr">
                  <span>{t("pricing.input")}: {formatPrice(availability?.effective_pricing?.input_usd_per_1m, t("values.na"))}</span>
                  <span>{t("pricing.output")}: {formatPrice(availability?.effective_pricing?.output_usd_per_1m, t("values.na"))}</span>
                </div>
              </TableCell>
              <TableCell>
                {availability ? (
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    aria-label={t("actions.modelDetails", { model: formatCatalogModelDisplayName(model.model, model.display_name) })}
                    onClick={() => onOpen(model, availability)}
                  >
                    <FileSearch aria-hidden="true" />
                  </Button>
                ) : null}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}

function AvailabilityPopover({
  model,
  onOpen,
}: {
  model: CatalogModel;
  onOpen: (model: CatalogModel, availability: ModelAvailability) => void;
}) {
  const t = useTranslations("modelCatalog");
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button type="button" variant="outline" size="sm">
          <Network data-icon="inline-start" aria-hidden="true" />
          {t("availability.count", { count: model.availability.length })}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-80 max-w-full sm:w-96">
        <div className="flex flex-col gap-3">
          <div>
            <h3 className="font-semibold">{t("availability.title")}</h3>
            <p className="text-sm text-muted-foreground">{t("availability.description")}</p>
          </div>
          <div className="flex max-h-72 flex-col gap-2 overflow-y-auto">
            {model.availability.map((availability) => (
              <div
                key={availability.account_id}
                className="flex min-w-0 flex-col gap-2 rounded-lg border border-border p-3 text-start transition-colors hover:bg-accent"
              >
                <button
                  type="button"
                  className="flex min-w-0 flex-col gap-2 text-start"
                  aria-label={t("availability.openAccount", { account: availability.account_name })}
                  onClick={() => onOpen(model, availability)}
                >
                  <span className="flex w-full items-center justify-between gap-2">
                    <span className="truncate font-medium" dir="ltr">{availability.account_name}</span>
                    <ReadinessBadge readiness={availability.readiness} />
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {t(`credential.${availability.credential_source === "managed" ? "managed" : "byok"}`)}
                    {" · "}
                    {availability.pricing_resolved ? t("availability.pricingReady") : t("availability.pricingMissing")}
                  </span>
                </button>
                <div className="flex min-w-0 items-center gap-1" dir="ltr">
                  <code className="min-w-0 flex-1 truncate text-xs text-muted-foreground" title={availability.fqn}>
                    {availability.fqn}
                  </code>
                  <CopyAction
                    value={availability.fqn}
                    copyLabel={t("actions.copyFqn")}
                    copyingLabel={t("actions.copying")}
                    copiedLabel={t("actions.copied")}
                  />
                </div>
                <span className="text-xs text-muted-foreground">
                  {t("pricing.input")}: {formatPrice(availability.effective_pricing?.input_usd_per_1m, t("values.na"))}
                  {" · "}
                  {t("pricing.output")}: {formatPrice(availability.effective_pricing?.output_usd_per_1m, t("values.na"))}
                </span>
              </div>
            ))}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

const READINESS_VALUES: ProviderAccountReadiness[] = [
  "ready",
  "needs_models",
  "needs_credentials",
  "needs_access",
  "pricing_incomplete",
];

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

function formatPrice(value: number | undefined, fallback: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return `$${new Intl.NumberFormat("en-US", { maximumFractionDigits: 6 }).format(value)}`;
}

function formatContext(value: number | undefined, fallback: string): string {
  return typeof value === "number" ? new Intl.NumberFormat("en-US").format(value) : fallback;
}

function formatModalityLabel(t: ReturnType<typeof useTranslations>, value: string): string {
  switch (value.trim().toLowerCase().replaceAll("-", "_")) {
    case "text":
      return t("modalities.text");
    case "image":
      return t("modalities.image");
    case "audio":
      return t("modalities.audio");
    case "video":
      return t("modalities.video");
    case "pdf":
      return t("modalities.pdf");
    case "text_to_speech":
      return t("modalities.textToSpeech");
    default:
      return value.replaceAll("_", " ");
  }
}
