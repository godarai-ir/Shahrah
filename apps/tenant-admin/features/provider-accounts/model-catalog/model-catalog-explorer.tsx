"use client";

import { useMemo, useState, type ReactNode } from "react";
import { useLocale, useTranslations } from "next-intl";
import { BookOpen, Boxes, Building2, CircleGauge, Layers3, Network, Settings2, Trash2 } from "lucide-react";

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
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

import { formatCatalogModelDisplayName } from "./formatters";
import { ModelProfile } from "./model-details-sheet";
import type { CatalogModel } from "./types";

type ModelCatalogExplorerProps = {
  accounts: TenantProviderAccount[];
  docsHref: string;
  summaries: Record<string, TenantProviderAccountSummary>;
  models: CatalogModel[];
  providerBySlug: Map<string, ProviderCatalogProvider>;
  selectedProvider: string;
  selectedAccountID: string;
  selectedModel: CatalogModel | null;
  loadingAccounts: boolean;
  loadingModels: boolean;
  modelLoadError: string | null;
  onSelectProvider: (provider: string) => void;
  onSelectAccount: (accountID: string) => void;
  onSelectModel: (model: CatalogModel) => void;
  onTestConnection: (account: TenantProviderAccount) => void;
  onDeleteAccount: (account: TenantProviderAccount) => void;
};

const MODEL_PAGE_SIZE = 50;

export function ModelCatalogExplorer({
  accounts,
  docsHref,
  summaries,
  models,
  providerBySlug,
  selectedProvider,
  selectedAccountID,
  selectedModel,
  loadingAccounts,
  loadingModels,
  modelLoadError,
  onSelectProvider,
  onSelectAccount,
  onSelectModel,
  onTestConnection,
  onDeleteAccount,
}: ModelCatalogExplorerProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();
  const [mobileStep, setMobileStep] = useState("providers");
  const [modelPagination, setModelPagination] = useState({ accountID: "", page: 0 });
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const formatCount = (value: number) => numberFormatter.format(value);
  const providers = useMemo(() => {
    const counts = new Map<string, number>();
    for (const account of accounts) counts.set(account.provider, (counts.get(account.provider) ?? 0) + 1);
    return [...counts.entries()]
      .map(([provider, count]) => ({ provider, count }))
      .sort((a, b) => formatProviderDisplayName(a.provider, providerBySlug.get(a.provider)?.display_name)
        .localeCompare(formatProviderDisplayName(b.provider, providerBySlug.get(b.provider)?.display_name)));
  }, [accounts, providerBySlug]);
  const providerAccounts = useMemo(
    () => accounts.filter((account) => account.provider === selectedProvider),
    [accounts, selectedProvider],
  );
  const selectedAvailability = useMemo(
    () => selectedModel?.availability.find((availability) => availability.account_id === selectedAccountID)
      ?? selectedModel?.availability[0]
      ?? null,
    [selectedAccountID, selectedModel],
  );
  const modelPageCount = Math.max(1, Math.ceil(models.length / MODEL_PAGE_SIZE));
  const effectiveModelPage = modelPagination.accountID === selectedAccountID
    ? Math.min(modelPagination.page, modelPageCount - 1)
    : 0;
  const visibleModels = useMemo(
    () => models.slice(effectiveModelPage * MODEL_PAGE_SIZE, (effectiveModelPage + 1) * MODEL_PAGE_SIZE),
    [effectiveModelPage, models],
  );

  return (
    <Tabs
      value={mobileStep}
      onValueChange={setMobileStep}
      aria-label={t("explorer.ariaLabel")}
      className="min-w-0 gap-3 md:grid md:grid-cols-3"
    >
      <TabsList className="grid w-full grid-cols-4 md:hidden">
        <TabsTrigger value="providers" className="min-w-0 px-2"><Network aria-hidden="true" /><span className="truncate">{t("explorer.providers")}</span></TabsTrigger>
        <TabsTrigger value="accounts" className="min-w-0 px-2"><Building2 aria-hidden="true" /><span className="truncate">{t("explorer.accounts")}</span></TabsTrigger>
        <TabsTrigger value="models" className="min-w-0 px-2"><Boxes aria-hidden="true" /><span className="truncate">{t("explorer.models")}</span></TabsTrigger>
        <TabsTrigger value="profile" className="min-w-0 px-2"><Layers3 aria-hidden="true" /><span className="truncate">{t("explorer.profile")}</span></TabsTrigger>
      </TabsList>

      <TabsContent value="providers" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:block">
        <ExplorerColumn title={t("explorer.providers")} count={formatCount(providers.length)} loading={loadingAccounts}>
        {providers.length > 0 ? providers.map(({ provider, count }) => {
          const metadata = providerBySlug.get(provider);
          const displayName = formatProviderDisplayName(provider, metadata?.display_name);
          const selected = provider === selectedProvider;
          const providerAccount = accounts.find((account) => account.provider === provider && account.id === selectedAccountID)
            ?? accounts.find((account) => account.provider === provider);
          return (
            <div
              key={provider}
              className={cn("flex min-w-0 items-center rounded-lg", selected && "bg-secondary text-secondary-foreground")}
            >
              <Button
                type="button"
                variant="ghost"
                className="h-auto min-w-0 flex-1 justify-start px-3 py-3 text-start"
                aria-pressed={selected}
                onClick={() => {
                  onSelectProvider(provider);
                  setMobileStep("accounts");
                }}
              >
                <ProviderLogo provider={provider} displayName={displayName} logoURL={metadata?.logo_url} size={28} />
                <span className="min-w-0 flex-1 truncate">{displayName}</span>
                <Badge variant="outline" className="shrink-0">{formatCount(count)}</Badge>
              </Button>
              <RowActionsMenu
                label={t("actions.openMenu", { account: displayName })}
                triggerSize="icon-sm"
                className="shrink-0"
                stopPropagation
                actions={[
                  {
                    label: t("actions.manage"),
                    icon: <Settings2 />,
                    href: providerAccount
                      ? `/provider-accounts?${new URLSearchParams({ account: providerAccount.id, section: "overview" }).toString()}`
                      : undefined,
                    disabled: !providerAccount,
                  },
                  {
                    label: t("actions.connectionTest"),
                    icon: <CircleGauge />,
                    disabled: !providerAccount || providerAccount.use_platform_managed_credentials,
                    onSelect: providerAccount ? () => onTestConnection(providerAccount) : undefined,
                  },
                  {
                    label: t("actions.docs"),
                    icon: <BookOpen />,
                    href: metadata ? buildModelDiscoveryProviderHref(metadata.provider) : docsHref,
                    external: true,
                  },
                  {
                    label: t("actions.delete"),
                    icon: <Trash2 />,
                    destructive: true,
                    separatorBefore: true,
                    disabled: !providerAccount,
                    onSelect: providerAccount ? () => onDeleteAccount(providerAccount) : undefined,
                  },
                ]}
              />
            </div>
          );
        }) : !loadingAccounts ? (
          <EmptyState variant="inline" density="compact" icon={Network} title={t("explorer.emptyProviders")} description={t("explorer.emptyProvidersDescription")} />
        ) : null}
        </ExplorerColumn>
      </TabsContent>

      <TabsContent value="accounts" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:block">
        <ExplorerColumn title={t("explorer.accounts")} count={formatCount(providerAccounts.length)} loading={loadingAccounts}>
        {selectedProvider && providerAccounts.length > 0 ? providerAccounts.map((account) => {
          const summary = summaries[account.id];
          const selected = account.id === selectedAccountID;
          const accountIdentifier = account.name.includes(":")
            ? account.name
            : `${account.provider}:${account.name}`;
          const compactAccountName = accountIdentifier.split(":").at(-1) ?? account.name;
          return (
            <Button
              key={account.id}
              type="button"
              variant={selected ? "secondary" : "ghost"}
              className="h-auto w-full justify-start px-3 py-3 text-start"
              aria-pressed={selected}
              onClick={() => {
                setModelPagination({ accountID: account.id, page: 0 });
                onSelectAccount(account.id);
                setMobileStep("models");
              }}
            >
              <Building2 data-icon="inline-start" aria-hidden="true" />
              <span className="flex min-w-0 flex-1 items-center gap-1">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <span className="min-w-0 truncate" dir="ltr">
                        <span className="xl:hidden">{compactAccountName}</span>
                        <span className="hidden xl:inline">{accountIdentifier}</span>
                      </span>
                    </TooltipTrigger>
                    <TooltipContent dir="ltr">{accountIdentifier}</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                <span className="shrink-0 text-xs font-normal text-muted-foreground">
                  ({formatCount(summary?.model_count ?? 0)})
                </span>
              </span>
              <AccountReadinessBadge readiness={summary?.readiness ?? "needs_models"} />
            </Button>
          );
        }) : !loadingAccounts ? (
          <EmptyState
            variant="inline"
            density="compact"
            icon={Building2}
            title={selectedProvider ? t("explorer.emptyAccounts") : t("explorer.selectProvider")}
            description={selectedProvider ? t("explorer.emptyAccountsDescription") : t("explorer.selectProviderDescription")}
          />
        ) : null}
        </ExplorerColumn>
      </TabsContent>

      <TabsContent value="models" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:block">
        <ExplorerColumn title={t("explorer.models")} count={formatCount(models.length)} loading={loadingModels}>
        {modelLoadError ? (
          <OperationNotice
            notice={{ kind: "error", message: modelLoadError }}
            labels={{ success: t("notices.success"), error: t("notices.error"), warning: t("notices.warning"), info: t("notices.info") }}
          />
        ) : selectedAccountID && models.length > 0 ? <>
          {visibleModels.map((model) => {
          const metadata = providerBySlug.get(model.provider);
          const displayName = formatProviderDisplayName(model.provider, metadata?.display_name);
          const modelName = formatCatalogModelDisplayName(model.model, model.display_name);
          const availability = model.availability.find((item) => item.account_id === selectedAccountID) ?? model.availability[0];
          const selected = selectedModel?.provider === model.provider && selectedModel.model === model.model;
          return (
            <div key={`${model.provider}:${model.model}`} className={cn("flex min-w-0 items-center rounded-lg", selected && "bg-secondary text-secondary-foreground")}>
              <Button
                type="button"
                variant="ghost"
                className="h-auto min-w-0 flex-1 justify-start px-3 py-3 text-start"
                aria-pressed={selected}
                onClick={() => {
                  onSelectModel(model);
                  setMobileStep("profile");
                }}
              >
                <ProviderLogo provider={model.provider} displayName={displayName} logoURL={metadata?.logo_url} size={24} />
                <span className="min-w-0 flex-1 truncate" dir="ltr" title={modelName}>{modelName}</span>
              </Button>
              <CopyAction
                value={availability?.fqn ?? model.model}
                copyLabel={t("actions.copyFqn")}
                copyingLabel={t("actions.copying")}
                copiedLabel={t("actions.copied")}
              />
            </div>
          );
          })}
          {modelPageCount > 1 ? (
            <nav className="mt-3 flex items-center justify-between gap-2 border-t pt-3" aria-label={t("pagination.label")}>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={effectiveModelPage === 0}
                onClick={() => setModelPagination({
                  accountID: selectedAccountID,
                  page: Math.max(0, effectiveModelPage - 1),
                })}
              >
                {t("pagination.previous")}
              </Button>
              <span className="text-xs tabular-nums text-muted-foreground">
                {t("pagination.summary", {
                  start: effectiveModelPage * MODEL_PAGE_SIZE + 1,
                  end: Math.min((effectiveModelPage + 1) * MODEL_PAGE_SIZE, models.length),
                  total: models.length,
                })}
              </span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={effectiveModelPage >= modelPageCount - 1}
                onClick={() => setModelPagination({
                  accountID: selectedAccountID,
                  page: Math.min(modelPageCount - 1, effectiveModelPage + 1),
                })}
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

      <TabsContent value="profile" forceMount className="mt-0 hidden min-w-0 data-[state=active]:block md:col-span-3 md:block">
      <Card className="flex min-h-96 min-w-0 flex-col shadow-none md:max-h-screen">
        <CardHeader className="border-b border-border p-4">
          <div className="flex min-w-0 items-center justify-between gap-2">
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              {selectedModel ? (
                <div className="flex min-w-0 items-center gap-1">
                  <h2 className="truncate text-lg font-semibold tracking-tight" dir="ltr" title={formatCatalogModelDisplayName(selectedModel.model, selectedModel.display_name)}>
                    {formatCatalogModelDisplayName(selectedModel.model, selectedModel.display_name)}
                  </h2>
                  {selectedAvailability ? (
                    <CopyAction
                      value={selectedAvailability.fqn}
                      copyLabel={t("actions.copyFqn")}
                      copyingLabel={t("actions.copying")}
                      copiedLabel={t("actions.copied")}
                    />
                  ) : null}
                </div>
              ) : null}
            </div>
            {selectedModel?.is_deprecated ? <StatusBadge status="warning">{t("status.deprecated")}</StatusBadge> : null}
          </div>
        </CardHeader>
        <CardContent className="min-h-0 flex-1 overflow-y-auto p-4 pt-4 md:p-4 md:pt-4">
          {loadingModels ? <ColumnSkeleton rowCount={5} /> : selectedModel ? (
            <ModelProfile model={selectedModel} accountID={selectedAccountID} showAccountSelector={false} />
          ) : (
            <EmptyState variant="inline" density="compact" icon={Layers3} title={t("explorer.selectModel")} description={t("explorer.selectModelDescription")} />
          )}
        </CardContent>
      </Card>
      </TabsContent>
    </Tabs>
  );
}

function ExplorerColumn({
  title,
  count,
  loading,
  children,
}: {
  title: string;
  count: ReactNode;
  loading: boolean;
  children: ReactNode;
}) {
  return (
    <Card className="flex min-h-96 min-w-0 flex-col shadow-none md:max-h-screen">
      <CardHeader className="border-b border-border p-4">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-sm">{title}</CardTitle>
          <Badge variant="secondary">{count}</Badge>
        </div>
      </CardHeader>
      <CardContent
        className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-4 pt-4 md:p-4 md:pt-4"
        aria-busy={loading}
      >
        {loading ? <ColumnSkeleton /> : children}
      </CardContent>
    </Card>
  );
}

function ColumnSkeleton({ rowCount = 4 }: { rowCount?: number }) {
  const t = useTranslations("modelCatalog");
  return (
    <div className="flex flex-col gap-2" role="status" aria-label={t("loading")}>
      {Array.from({ length: rowCount }, (_, index) => <Skeleton key={index} className="h-14 w-full" />)}
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
  return <StatusBadge status={status}>{t(`readiness.${readiness}`)}</StatusBadge>;
}
