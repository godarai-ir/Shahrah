"use client";

import Link from "next/link";
import { useMemo, useState, type ReactNode } from "react";
import { useLocale, useTranslations } from "next-intl";
import { BookOpen, Boxes, Building2, CircleGauge, Plus, Settings2, Trash2 } from "lucide-react";

import { ProviderLogo } from "@/components/app-ui/ProviderLogo";
import type {
  ProviderCatalogProvider,
  TenantProviderAccount,
  TenantProviderAccountSummary,
} from "@/features/provider-accounts/types";
import { buildModelDiscoveryProviderHref, formatProviderDisplayName } from "@/features/provider-accounts/utils";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

import { groupAccountsByProvider } from "./account-groups";
import { formatCatalogModelDisplayName } from "./formatters";
import { formatCapabilityLabel } from "./model-details-sheet";
import {
  filterCatalogModels,
  MODEL_READINESS_SUMMARIES,
  type ModelReadinessSummary,
  type SummarizedCatalogModel,
} from "./readiness-summary";
import type { CatalogModel, ModelPrice } from "./types";

type ModelCatalogExplorerProps = {
  accounts: TenantProviderAccount[];
  docsHref: string;
  summaries: Record<string, TenantProviderAccountSummary>;
  models: CatalogModel[];
  providerBySlug: Map<string, ProviderCatalogProvider>;
  selectedAccountID: string;
  modelAccountID: string;
  selectedModel: CatalogModel | null;
  query: string;
  readinessFilter: ModelReadinessSummary | "";
  loadingAccounts: boolean;
  loadingModels: boolean;
  refreshingModels: boolean;
  modelLoadError: string | null;
  onQueryChange: (query: string) => void;
  onReadinessFilterChange: (readiness: ModelReadinessSummary | "") => void;
  onSelectAccount: (account: TenantProviderAccount) => void;
  onSelectModel: (model: CatalogModel) => void;
  onTestConnection: (account: TenantProviderAccount) => void;
  onDeleteAccount: (account: TenantProviderAccount) => void;
};

const MODEL_PAGE_SIZE = 50;
const VISIBLE_TYPE_COUNT = 2;

const readinessStatus = {
  ready: "success",
  pricing: "warning",
  limited: "info",
  deprecated: "error",
} as const satisfies Record<ModelReadinessSummary, string>;

// Chips read left-to-right from the healthy state; row priority lives in readiness-summary.ts.
const READINESS_CHIP_ORDER = [...MODEL_READINESS_SUMMARIES].reverse();

export function ModelCatalogExplorer({
  accounts,
  docsHref,
  summaries,
  models,
  providerBySlug,
  selectedAccountID,
  modelAccountID,
  selectedModel,
  query,
  readinessFilter,
  loadingAccounts,
  loadingModels,
  refreshingModels,
  modelLoadError,
  onQueryChange,
  onReadinessFilterChange,
  onSelectAccount,
  onSelectModel,
  onTestConnection,
  onDeleteAccount,
}: ModelCatalogExplorerProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();
  const [mobileStep, setMobileStep] = useState("accounts");
  // Seeded from the URL once; the URL follows the input (always via replace), so reading it back
  // mid-typing would only race with in-flight router updates.
  const [search, setSearch] = useState(query);
  const filterScope = `${selectedAccountID}|${search}|${readinessFilter}`;
  const [modelPagination, setModelPagination] = useState({ scope: "", page: 0 });
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const formatCount = (value: number) => numberFormatter.format(value);

  const groups = useMemo(() => groupAccountsByProvider(accounts, providerBySlug), [accounts, providerBySlug]);
  const filtered = useMemo(
    () => filterCatalogModels(models, modelAccountID, search, readinessFilter),
    [modelAccountID, models, readinessFilter, search],
  );
  const modelPageCount = Math.max(1, Math.ceil(filtered.rows.length / MODEL_PAGE_SIZE));
  const effectiveModelPage = modelPagination.scope === filterScope
    ? Math.min(modelPagination.page, modelPageCount - 1)
    : 0;
  const visibleRows = filtered.rows.slice(effectiveModelPage * MODEL_PAGE_SIZE, (effectiveModelPage + 1) * MODEL_PAGE_SIZE);
  const hasFilters = Boolean(search.trim() || readinessFilter);

  const clearFilters = () => {
    setSearch("");
    onQueryChange("");
    onReadinessFilterChange("");
  };

  return (
    <Tabs
      value={mobileStep}
      onValueChange={setMobileStep}
      aria-label={t("explorer.ariaLabel")}
      className="min-w-0 gap-3 md:grid md:grid-cols-[minmax(15rem,20rem)_minmax(0,1fr)] md:items-start"
    >
      <TabsList className="grid w-full grid-cols-2 md:hidden">
        <TabsTrigger value="accounts" className="min-w-0 px-2"><Building2 aria-hidden="true" /><span className="truncate">{t("explorer.accounts")}</span></TabsTrigger>
        <TabsTrigger value="models" className="min-w-0 px-2"><Boxes aria-hidden="true" /><span className="truncate">{t("explorer.models")}</span></TabsTrigger>
      </TabsList>

      <TabsContent value="accounts" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:block">
        <ExplorerColumn
          title={t("explorer.accounts")}
          count={formatCount(accounts.length)}
          loading={loadingAccounts}
          contentClassName="gap-0 px-2 pt-0 md:px-2 md:pt-0"
          action={(
            <Button asChild size="icon-sm">
              <Link href="/provider-accounts/new" aria-label={t("actions.addProvider")} title={t("actions.addProvider")}>
                <Plus aria-hidden="true" />
              </Link>
            </Button>
          )}
        >
          {groups.length > 0 ? groups.map((group) => (
            <section key={group.provider} aria-labelledby={`model-catalog-provider-${group.provider}`} className="flex flex-col gap-0.5 pb-2">
              <h3
                id={`model-catalog-provider-${group.provider}`}
                className="sticky top-0 z-10 flex items-center gap-2 bg-card px-2 pb-1 pt-3 text-xs font-semibold text-muted-foreground"
              >
                <ProviderLogo provider={group.provider} displayName={group.displayName} logoURL={group.metadata?.logo_url} size={18} />
                <span className="min-w-0 flex-1 truncate">{group.displayName}</span>
                <span className="tabular-nums">{formatCount(group.accounts.length)}</span>
              </h3>
              {group.accounts.map((account) => (
                <AccountRow
                  key={account.id}
                  account={account}
                  summary={summaries[account.id]}
                  providerMetadata={group.metadata}
                  docsHref={docsHref}
                  selected={account.id === selectedAccountID}
                  onSelect={() => {
                    onSelectAccount(account);
                    setMobileStep("models");
                  }}
                  onTestConnection={onTestConnection}
                  onDelete={onDeleteAccount}
                />
              ))}
            </section>
          )) : !loadingAccounts ? (
            <EmptyState variant="inline" density="compact" icon={Building2} title={t("explorer.emptyProviders")} description={t("explorer.emptyProvidersDescription")} />
          ) : null}
        </ExplorerColumn>
      </TabsContent>

      <TabsContent value="models" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:block">
        <ExplorerColumn
          title={t("explorer.models")}
          count={formatCount(models.length)}
          loading={loadingModels}
          refreshing={refreshingModels}
          contentClassName="gap-3"
        >
          {modelLoadError ? (
            <OperationNotice
              notice={{ kind: "error", message: modelLoadError }}
              labels={{ success: t("notices.success"), error: t("notices.error"), warning: t("notices.warning"), info: t("notices.info") }}
            />
          ) : selectedAccountID && models.length > 0 ? <>
            <div className="flex flex-col gap-2 lg:flex-row lg:items-center">
              <SearchField
                containerClassName="h-9 lg:max-w-xs"
                placeholder={t("models.searchPlaceholder")}
                value={search}
                onChange={(event) => {
                  setSearch(event.target.value);
                  onQueryChange(event.target.value);
                }}
              />
              <div role="group" aria-label={t("readiness.filterLabel")} className="flex flex-wrap items-center gap-1.5">
                <ReadinessChip
                  label={t("filters.any")}
                  count={formatCount(filtered.queryMatches)}
                  active={!readinessFilter}
                  onClick={() => onReadinessFilterChange("")}
                />
                {READINESS_CHIP_ORDER.map((readiness) => (
                  <ReadinessChip
                    key={readiness}
                    label={t(`readiness.${readiness}`)}
                    count={formatCount(filtered.counts[readiness])}
                    active={readinessFilter === readiness}
                    onClick={() => onReadinessFilterChange(readinessFilter === readiness ? "" : readiness)}
                  />
                ))}
              </div>
            </div>

            {filtered.rows.length > 0 ? (
              <div className="overflow-hidden rounded-lg border border-border">
                <Table className="text-xs lg:text-sm">
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("columns.model")}</TableHead>
                      <TableHead className="hidden lg:table-cell">{t("drawer.capabilities")}</TableHead>
                      <TableHead className="hidden sm:table-cell">{t("models.contextWindow")}</TableHead>
                      <TableHead className="hidden xl:table-cell">{t("models.inputPer1m")}</TableHead>
                      <TableHead className="hidden xl:table-cell">{t("models.outputPer1m")}</TableHead>
                      <TableHead>{t("readiness.filterLabel")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {visibleRows.map((row) => (
                      <ModelRow
                        key={`${row.model.provider}:${row.model.model}`}
                        row={row}
                        selected={selectedModel?.provider === row.model.provider && selectedModel.model === row.model.model}
                        providerMetadata={providerBySlug.get(row.model.provider)}
                        onSelect={() => onSelectModel(row.model)}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <EmptyState
                variant="inline"
                density="compact"
                icon={Boxes}
                title={t("empty.noResultsTitle")}
                description={t("empty.noResultsDescription")}
                primaryAction={hasFilters ? (
                  <Button type="button" variant="outline" size="sm" onClick={clearFilters}>{t("actions.clearFilters")}</Button>
                ) : undefined}
              />
            )}

            {modelPageCount > 1 ? (
              <nav className="flex items-center justify-between gap-2 border-t pt-3" aria-label={t("pagination.label")}>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={effectiveModelPage === 0}
                  onClick={() => setModelPagination({ scope: filterScope, page: Math.max(0, effectiveModelPage - 1) })}
                >
                  {t("pagination.previous")}
                </Button>
                <span className="text-xs tabular-nums text-muted-foreground">
                  {t("pagination.summary", {
                    start: effectiveModelPage * MODEL_PAGE_SIZE + 1,
                    end: Math.min((effectiveModelPage + 1) * MODEL_PAGE_SIZE, filtered.rows.length),
                    total: filtered.rows.length,
                  })}
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={effectiveModelPage >= modelPageCount - 1}
                  onClick={() => setModelPagination({ scope: filterScope, page: Math.min(modelPageCount - 1, effectiveModelPage + 1) })}
                >
                  {t("pagination.next")}
                </Button>
              </nav>
            ) : null}
          </> : !loadingModels ? (
            <EmptyState
              variant="inline"
              density="compact"
              icon={Boxes}
              title={selectedAccountID ? t("explorer.emptyModels") : t("explorer.selectAccount")}
              description={selectedAccountID ? t("explorer.emptyModelsDescription") : t("explorer.selectAccountDescription")}
            />
          ) : null}
        </ExplorerColumn>
      </TabsContent>
    </Tabs>
  );
}

function AccountRow({
  account,
  summary,
  providerMetadata,
  docsHref,
  selected,
  onSelect,
  onTestConnection,
  onDelete,
}: {
  account: TenantProviderAccount;
  summary?: TenantProviderAccountSummary;
  providerMetadata?: ProviderCatalogProvider;
  docsHref: string;
  selected: boolean;
  onSelect: () => void;
  onTestConnection: (account: TenantProviderAccount) => void;
  onDelete: (account: TenantProviderAccount) => void;
}) {
  const t = useTranslations("modelCatalog");
  const accountIdentifier = account.name.includes(":") ? account.name : `${account.provider}:${account.name}`;
  const compactAccountName = accountIdentifier.split(":").at(-1) ?? account.name;

  return (
    <div className={cn("flex min-w-0 items-center rounded-lg", selected && "bg-secondary text-secondary-foreground")}>
      <Button
        type="button"
        variant="ghost"
        className="h-auto min-w-0 flex-1 justify-start gap-2 px-2 py-2 text-start hover:bg-transparent"
        aria-pressed={selected}
        title={accountIdentifier}
        onClick={onSelect}
      >
        <span className="flex min-w-0 flex-1 flex-col items-start gap-0.5">
          <span className="w-full truncate font-medium" dir="ltr">{compactAccountName}</span>
          <span className="text-xs font-normal text-muted-foreground">
            {t("explorer.modelCount", { count: summary?.model_count ?? 0 })}
          </span>
        </span>
        <AccountReadinessBadge readiness={summary?.readiness ?? "needs_models"} />
      </Button>
      <RowActionsMenu
        label={t("actions.openMenu", { account: account.name })}
        triggerSize="icon-sm"
        className="shrink-0"
        stopPropagation
        actions={[
          {
            label: t("actions.manage"),
            icon: <Settings2 />,
            href: `/provider-accounts?${new URLSearchParams({ account: account.id, section: "overview" }).toString()}`,
          },
          {
            label: t("actions.connectionTest"),
            icon: <CircleGauge />,
            disabled: account.use_platform_managed_credentials,
            onSelect: () => onTestConnection(account),
          },
          {
            label: t("actions.docs"),
            icon: <BookOpen />,
            href: providerMetadata ? buildModelDiscoveryProviderHref(providerMetadata.provider) : docsHref,
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
    </div>
  );
}

function ModelRow({
  row,
  selected,
  providerMetadata,
  onSelect,
}: {
  row: SummarizedCatalogModel;
  selected: boolean;
  providerMetadata?: ProviderCatalogProvider;
  onSelect: () => void;
}) {
  const t = useTranslations("modelCatalog");
  const { model, availability, readiness } = row;
  const providerName = formatProviderDisplayName(model.provider, providerMetadata?.display_name);
  const modelName = formatCatalogModelDisplayName(model.model, model.display_name);
  const pricing: ModelPrice | undefined = availability?.effective_pricing ?? model.public_pricing;
  const unknownPrice = readiness === "pricing";
  const types = model.types.map((type) => formatCapabilityLabel(type, t));

  return (
    <TableRow data-state={selected ? "selected" : undefined} className="cursor-pointer" onClick={onSelect}>
      <TableCell className="max-w-0 min-w-48">
        <div className="flex min-w-0 items-center gap-2">
          <ProviderLogo provider={model.provider} displayName={providerName} logoURL={providerMetadata?.logo_url} size={20} />
          <div className="flex min-w-0 flex-1 flex-col">
            <button
              type="button"
              className="truncate text-start font-medium outline-none hover:underline focus-visible:underline"
              dir="ltr"
              title={modelName}
              aria-label={t("actions.modelDetails", { model: modelName })}
              onClick={(event) => {
                event.stopPropagation();
                onSelect();
              }}
            >
              {modelName}
            </button>
            {model.is_deprecated && model.replacement_model ? (
              <span className="truncate text-xs text-muted-foreground">
                ← {t("models.replacedBy", { model: model.replacement_model })}
              </span>
            ) : null}
          </div>
          <span onClick={(event) => event.stopPropagation()} className="shrink-0">
            <CopyAction
              value={availability?.fqn ?? model.model}
              copyLabel={t("actions.copyFqn")}
              copyingLabel={t("actions.copying")}
              copiedLabel={t("actions.copied")}
            />
          </span>
        </div>
      </TableCell>
      <TableCell className="hidden lg:table-cell">
        <div className="flex flex-wrap gap-1">
          {types.slice(0, VISIBLE_TYPE_COUNT).map((type) => <Badge key={type} variant="outline" className="font-normal">{type}</Badge>)}
          {types.length > VISIBLE_TYPE_COUNT ? (
            <Badge variant="outline" className="font-normal" title={types.join(t("listSeparator"))}>
              +{types.length - VISIBLE_TYPE_COUNT}
            </Badge>
          ) : null}
        </div>
      </TableCell>
      <TableCell className="hidden whitespace-nowrap tabular-nums sm:table-cell">
        <bdi dir="ltr">{formatContextWindow(model.context_window, t("values.na"))}</bdi>
      </TableCell>
      <TableCell className="hidden whitespace-nowrap tabular-nums xl:table-cell">
        {unknownPrice ? <span className="text-muted-foreground">{t("models.unknownPrice")}</span> : <bdi dir="ltr">{formatUSD(pricing?.input_usd_per_1m, t("values.na"))}</bdi>}
      </TableCell>
      <TableCell className="hidden whitespace-nowrap tabular-nums xl:table-cell">
        {unknownPrice ? <span className="text-muted-foreground">{t("models.unknownPrice")}</span> : <bdi dir="ltr">{formatUSD(pricing?.output_usd_per_1m, t("values.na"))}</bdi>}
      </TableCell>
      <TableCell>
        <StatusBadge status={readinessStatus[readiness]}>{t(`readiness.${readiness}`)}</StatusBadge>
      </TableCell>
    </TableRow>
  );
}

function ReadinessChip({ label, count, active, onClick }: { label: string; count: string; active: boolean; onClick: () => void }) {
  return (
    <Button
      type="button"
      size="sm"
      variant={active ? "secondary" : "outline"}
      className={cn("h-8 rounded-full px-3", active && "ring-1 ring-primary/30")}
      aria-pressed={active}
      onClick={onClick}
    >
      {label}
      <span className="tabular-nums text-muted-foreground">{count}</span>
    </Button>
  );
}

function ExplorerColumn({
  title,
  count,
  loading,
  refreshing = false,
  action,
  contentClassName,
  children,
}: {
  title: string;
  count: ReactNode;
  loading: boolean;
  refreshing?: boolean;
  action?: ReactNode;
  contentClassName?: string;
  children: ReactNode;
}) {
  return (
    <Card className="flex min-h-96 min-w-0 flex-col shadow-none md:max-h-[calc(100vh-8rem)]">
      <CardHeader className="border-b border-border p-4">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-sm font-semibold leading-6 tracking-tight">{title}</h2>
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{count}</Badge>
            {action}
          </div>
        </div>
      </CardHeader>
      <CardContent
        className={cn(
          "flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-4 pt-4 transition-opacity md:p-4 md:pt-4",
          contentClassName,
          refreshing && "pointer-events-none opacity-60",
        )}
        aria-busy={loading || refreshing}
      >
        {loading ? <ColumnSkeleton /> : children}
      </CardContent>
    </Card>
  );
}

function ColumnSkeleton({ rowCount = 4 }: { rowCount?: number }) {
  const t = useTranslations("modelCatalog");
  return (
    <div className="flex flex-col gap-2 pt-3" role="status" aria-label={t("loading")}>
      {Array.from({ length: rowCount }, (_, index) => <Skeleton key={index} className="h-12 w-full" />)}
    </div>
  );
}

function AccountReadinessBadge({ readiness }: { readiness: TenantProviderAccountSummary["readiness"] }) {
  const t = useTranslations("modelCatalog");
  const status = readiness === "ready"
    ? "success"
    : readiness === "needs_credentials" || readiness === "needs_access"
      ? "error"
      : "warning";
  return <StatusBadge status={status} className="shrink-0">{t(`readiness.${readiness}`)}</StatusBadge>;
}

function formatContextWindow(value: number | undefined, fallback: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return new Intl.NumberFormat("en-US", { notation: "compact", maximumFractionDigits: 1 }).format(value);
}

function formatUSD(value: number | undefined, fallback: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return `$${new Intl.NumberFormat("en-US", { maximumFractionDigits: 4 }).format(value)}`;
}
