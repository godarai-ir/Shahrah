"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Activity, BookOpen, ChevronDown, ExternalLink, Link2, Plus, Plug, Search, Settings, Trash2, UserCog } from "lucide-react";
import { useTranslations } from "next-intl";

import { EffectiveAccessTester } from "@/app/components/provider-account-wizard/effective-access-tester";
import { ProviderConnectionTestDialog } from "@/app/components/provider-account-wizard/provider-connection-test-dialog";
import { buildPrincipalSelectorLabels } from "@/app/components/provider-account-wizard/principal-labels";
import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import type { PrincipalSearchKind } from "@/app/components/principal-search-input";
import type {
  AccessPrincipalCandidate,
  AccessPrincipalResponse,
  AccountsResponse,
  ProviderCatalogProvider,
  ProvidersResponse,
  TenantProviderAccount,
} from "@/app/components/provider-account-wizard/types";
import {
  getModelDiscoveryCatalogBaseURL,
  buildModelDiscoveryProviderHref,
  normalizeError,
  normalizeProviderSlug,
  requestJSON,
  stringsEqualFold,
  unwrapData,
} from "@/app/components/provider-account-wizard/utils";
import { EmptyState, FormSheet, ModelNameValue, RefreshButton, RowActionsMenu, StatusBadge, TypedConfirmDialog } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

type ProviderView = ProviderCatalogProvider & {
  description?: string;
  docs_url?: string;
  docsUrl?: string;
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type PageDirection = "ltr" | "rtl";

const PROVIDER_DOCS_URLS: Record<string, string> = {
  anthropic: "https://docs.anthropic.com/",
  aws: "https://docs.aws.amazon.com/bedrock/",
  azure: "https://learn.microsoft.com/azure/ai-services/openai/",
  cohere: "https://docs.cohere.com/",
  google: "https://ai.google.dev/gemini-api/docs",
  groq: "https://console.groq.com/docs",
  mistral: "https://docs.mistral.ai/",
  openai: "https://platform.openai.com/api-keys",
  "perplexity-ai": "https://docs.perplexity.ai/",
  vertex: "https://cloud.google.com/vertex-ai/generative-ai/docs",
  xai: "https://docs.x.ai/",
};

function readPageDirection(): PageDirection {
  if (typeof document === "undefined") return "rtl";
  return document.documentElement.getAttribute("dir") === "ltr" ? "ltr" : "rtl";
}

function usePageDirection(): PageDirection {
  const [direction, setDirection] = useState<PageDirection>(() => readPageDirection());

  useEffect(() => {
    const syncDirection = () => setDirection(readPageDirection());
    syncDirection();

    if (typeof MutationObserver === "undefined") return undefined;
    const observer = new MutationObserver(syncDirection);
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["dir"] });
    return () => observer.disconnect();
  }, []);

  return direction;
}

function parseProviderFilters(value?: string) {
  if (!value) return [];
  return value
    .split(",")
    .map((item) => normalizeProviderSlug(item))
    .filter(Boolean);
}

export default function LLMIntegrationsManagement(props: {
  initialProvider?: string;
  initialTab?: "all" | "connected";
  docsHref?: string;
  initialCreatedAccount?: string;
}) {
  const t = useTranslations("providerAccounts");
  const ux = useTranslations("providerAccounts.ux");
  const copy = useTranslations("providerAccounts.integrations");
  const pathname = usePathname();
  const router = useRouter();
  const initialProviderFilters = parseProviderFilters(props.initialProvider);
  const pageDirection = usePageDirection();

  const [activeTab, setActiveTab] = useState<"all" | "connected">(props.initialTab ?? (initialProviderFilters.length > 0 ? "connected" : "all"));
  const [providers, setProviders] = useState<ProviderView[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [query, setQuery] = useState("");
  const [pendingDisconnectAccount, setPendingDisconnectAccount] = useState<TenantProviderAccount | null>(null);
  const [connectionTestAccount, setConnectionTestAccount] = useState<TenantProviderAccount | null>(null);
  const [effectiveAccessAccount, setEffectiveAccessAccount] = useState<TenantProviderAccount | null>(null);
  const [providerFilters, setProviderFilters] = useState<string[]>(initialProviderFilters);
  const [isLoading, setIsLoading] = useState(true);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [successCreatedName, setSuccessCreatedName] = useState(props.initialCreatedAccount?.trim() ?? "");
  const [deletingAccountID, setDeletingAccountID] = useState("");
  const didHandleCreatedAccountRef = useRef(false);

  const loadData = useCallback(async (refreshCatalog = false) => {
    setIsLoading(true);
    setBanner(null);
    try {
      if (refreshCatalog) {
        await requestJSON("/api/provider-catalog/refresh", { method: "POST" });
      }
      const [providersPayload, accountsPayload] = await Promise.all([
        requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
        requestJSON<AccountsResponse>("/api/tenant-provider-accounts"),
      ]);
      const nextProviders = (unwrapData<ProviderView[]>(providersPayload) ?? []).filter((item) => item.enabled);
      setProviders(nextProviders);
      setAccounts(unwrapData<TenantProviderAccount[]>(accountsPayload) ?? []);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, copy("messages.loadFailed")) });
    } finally {
      setIsLoading(false);
    }
  }, [copy]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadData();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadData]);

  useEffect(() => {
    const createdName = props.initialCreatedAccount?.trim();
    if (!createdName || isLoading || didHandleCreatedAccountRef.current) return;
    didHandleCreatedAccountRef.current = true;
    setSuccessCreatedName(createdName);
    setBanner({ kind: "success", message: copy("messages.connectionCreated", { name: createdName }) });
    const params = new URLSearchParams(window.location.search);
    params.delete("created");
    const nextQuery = params.toString();
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }, [copy, isLoading, pathname, props.initialCreatedAccount, router]);

  const providerOptions = useMemo(
    () => providers.map((item) => ({ slug: normalizeProviderSlug(item.provider), label: item.display_name })),
    [providers],
  );
  const principalSelectorLabels = useMemo(() => buildPrincipalSelectorLabels(t, ux), [t, ux]);

  const searchPrincipalCandidates = useCallback(async (kind: PrincipalSearchKind, searchQuery: string) => {
    if (kind === "external_identity") return [];
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${kind}&q=${encodeURIComponent(searchQuery.trim())}&limit=20`,
      );
      return unwrapData<AccessPrincipalCandidate[]>(response) ?? [];
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(
          error,
          kind === "user" ? t("messages.userSearchFailed") : t("messages.managerSearchFailed"),
        ),
      });
      return [];
    }
  }, [t]);

  const accountCountsByProvider = useMemo(() => {
    const counts = new Map<string, number>();
    for (const account of accounts) {
      const slug = normalizeProviderSlug(account.provider);
      counts.set(slug, (counts.get(slug) ?? 0) + 1);
    }
    return counts;
  }, [accounts]);

  const filteredProviders = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    return providers.filter((provider) => {
      if (!normalizedQuery) return true;
      return (
        provider.display_name.toLowerCase().includes(normalizedQuery) ||
        provider.provider.toLowerCase().includes(normalizedQuery) ||
        (provider.description?.toLowerCase().includes(normalizedQuery) ?? false) ||
        provider.auth_modes.some((mode) => mode.toLowerCase().includes(normalizedQuery))
      );
    });
  }, [providers, query]);

  const filteredAccounts = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    return accounts.filter((account) => {
      const providerSlug = normalizeProviderSlug(account.provider);
      if (providerFilters.length > 0 && !providerFilters.includes(providerSlug)) return false;
      if (!normalizedQuery) return true;
      return (
        account.name.toLowerCase().includes(normalizedQuery) ||
        account.provider.toLowerCase().includes(normalizedQuery) ||
        account.auth_mode.toLowerCase().includes(normalizedQuery) ||
        `${account.provider}:${account.name}`.toLowerCase().includes(normalizedQuery)
      );
    });
  }, [accounts, providerFilters, query]);
  const createdAccount = useMemo(
    () => accounts.find((item) => item.name === successCreatedName) ?? null,
    [accounts, successCreatedName],
  );

  async function disconnectAccount(account: TenantProviderAccount) {
    setDeletingAccountID(account.id);
    setBanner(null);
    try {
      await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(account.id)}`, { method: "DELETE" });
      setAccounts((current) => current.filter((item) => item.id !== account.id));
      setBanner({ kind: "success", message: copy("messages.disconnectSuccess") });
      setPendingDisconnectAccount(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, copy("messages.disconnectFailed")) });
    } finally {
      setDeletingAccountID("");
    }
  }

  function syncAddress(nextTab: "all" | "connected", nextProviderFilters: string[]) {
    const params = new URLSearchParams(window.location.search);
    params.delete("account");
    params.delete("created");

    if (nextTab === "connected") {
      params.set("tab", "connected");
      if (nextProviderFilters.length === 0) {
        params.delete("provider");
      } else {
        params.set("provider", nextProviderFilters.join(","));
      }
    } else {
      params.delete("tab");
      params.delete("provider");
    }

    const nextQuery = params.toString();
    router.replace(nextQuery ? `${pathname}?${nextQuery}` : pathname, { scroll: false });
  }

  function changeTab(value: string) {
    const nextTab = value === "connected" ? "connected" : "all";
    setActiveTab(nextTab);
    syncAddress(nextTab, providerFilters);
  }

  function changeProviderFilters(nextProviderFilters: string[]) {
    const normalizedFilters = Array.from(new Set(nextProviderFilters.map((item) => normalizeProviderSlug(item)).filter(Boolean)));
    setProviderFilters(normalizedFilters);
    if (activeTab === "connected") {
      syncAddress("connected", normalizedFilters);
    }
  }

  function openConnectedProvider(providerSlug: string) {
    const nextProviderFilters = [normalizeProviderSlug(providerSlug)].filter(Boolean);
    setActiveTab("connected");
    setProviderFilters(nextProviderFilters);
    setQuery("");
    syncAddress("connected", nextProviderFilters);
  }

  function openCreateProvider(providerSlug: string) {
    router.push(`/provider-accounts/new?provider=${encodeURIComponent(normalizeProviderSlug(providerSlug))}`);
  }

  return (
    <>
    <main className="flex flex-1 px-3 py-3 sm:px-4">
      <section className="flex min-h-[calc(100dvh-1.5rem)] w-full flex-col overflow-hidden rounded-xl border border-border bg-background">
        {banner ? (
          <div className="px-4 pt-3">
            <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
              <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
              <AlertDescription>
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <span>{banner.message}</span>
                  {banner.kind === "success" && createdAccount ? (
                    <div className="flex flex-wrap gap-2">
                      <Button type="button" size="sm" asChild>
                        <Link href={`/provider-accounts?account=${encodeURIComponent(createdAccount.id)}`}>{copy("actions.viewConnection")}</Link>
                      </Button>
                      <Button type="button" size="sm" variant="outline" asChild>
                        <Link href={`/provider-accounts?account=${encodeURIComponent(createdAccount.id)}`}>{copy("actions.manageModels")}</Link>
                      </Button>
                      <Button type="button" size="sm" variant="ghost" asChild>
                        <Link href="/provider-accounts">{copy("actions.backToList")}</Link>
                      </Button>
                    </div>
                  ) : null}
                </div>
              </AlertDescription>
            </Alert>
          </div>
        ) : null}

        <Tabs value={activeTab} onValueChange={changeTab} className="gap-0">
          <div className="border-b border-border px-5 py-4 sm:px-6" dir={pageDirection}>
            <div className="flex items-center gap-3">
              <TabsList dir={pageDirection} className="shrink-0 justify-start gap-2">
                <TabsTrigger value="all">
                  <Plug data-icon="inline-start" />
                  {copy("tabs.all")}
                </TabsTrigger>
                <TabsTrigger value="connected">
                  <Link2 data-icon="inline-start" />
                  {copy("tabs.connected")}
                </TabsTrigger>
              </TabsList>
              <RefreshButton
                label={copy("actions.refresh")}
                loadingLabel={copy("actions.refreshing")}
                isLoading={isLoading}
                onClick={() => void loadData(true)}
              />
              <div className="min-w-0 flex-1">
                <FilterRow
                  activeTab={activeTab}
                  query={query}
                  setQuery={setQuery}
                  providerFilters={providerFilters}
                  setProviderFilters={changeProviderFilters}
                  providerOptions={providerOptions}
                  searchPlaceholder={copy("searchPlaceholder")}
                  providerFilterLabel={copy("providerFilter")}
                  direction={pageDirection}
                />
              </div>
            </div>
          </div>

          <TabsContent value="all" className="px-5 pb-5 pt-4 sm:px-6">
            {isLoading ? (
              <ProviderGridSkeleton direction={pageDirection} />
            ) : filteredProviders.length > 0 ? (
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3" dir={pageDirection}>
                {filteredProviders.map((provider) => (
                    <ProviderCard
                      key={provider.provider}
                      provider={provider}
                      connectLabel={copy("actions.connect")}
                      manageConnectionsLabel={copy("actions.manageConnections")}
                      modelUnitLabel={copy("card.modelUnit")}
                      modelMetricLabel={copy("card.modelMetricLabel")}
                      connectionMetricLabel={copy("card.connectionMetricLabel")}
                      noConnectionLabel={copy("card.noConnection")}
                      createConnectionLabel={copy("card.createConnection")}
                      accountCount={accountCountsByProvider.get(normalizeProviderSlug(provider.provider)) ?? 0}
                      modelCount={provider.model_count}
                      onManageConnections={openConnectedProvider}
                      onConnect={openCreateProvider}
                      connectedStatusLabel={copy("card.activeStatus")}
                    />
                ))}
              </div>
            ) : (
              <EmptyState
                title={copy("empty.providersTitle")}
                description={copy("empty.providersDescription")}
                icon={Plug}
                primaryAction={<Button type="button" onClick={() => void loadData()}>{copy("actions.reload")}</Button>}
              />
            )}
          </TabsContent>

          <TabsContent value="connected" className="px-5 pb-5 pt-4 sm:px-6">
            {isLoading ? (
              <ProviderGridSkeleton direction={pageDirection} />
            ) : filteredAccounts.length > 0 ? (
              <div className="flex flex-wrap items-stretch gap-3" dir={pageDirection} id="connected-integrations">
                {filteredAccounts.map((account) => {
                  const provider = providers.find((item) => stringsEqualFold(item.provider, account.provider));
                  return (
                    <ConnectedIntegrationCard
                      key={account.id}
                      account={account}
                      provider={provider}
                      fqn={`${account.provider}:${account.name}`}
                      editLabel={copy("actions.editIntegration")}
                      manageModelsLabel={copy("actions.manageModels")}
                      testConnectionLabel={copy("actions.testConnection")}
                      testAllowedAccessLabel={copy("actions.testAllowedAccess")}
                      viewDocsLabel={copy("actions.viewDocs")}
                      disconnectLabel={copy("actions.disconnect")}
                      disconnectingLabel={copy("actions.disconnecting")}
                      credentialTypeLabel={copy("card.credentialTypeLabel")}
                      dedicatedCredentialLabel={copy("card.dedicatedCredential")}
                      platformCredentialLabel={copy("card.platformCredential")}
                      onTestConnection={() => setConnectionTestAccount(account)}
                      onTestAllowedAccess={() => setEffectiveAccessAccount(account)}
                      onDisconnect={() => setPendingDisconnectAccount(account)}
                      isDisconnecting={deletingAccountID === account.id}
                    />
                  );
                })}
              </div>
            ) : (
              <div dir={pageDirection}>
                <EmptyState
                  title={copy("empty.connectedTitle")}
                  description={copy("empty.connectedDescription")}
                  icon={Plug}
                  primaryAction={<Button type="button" asChild><Link href="/provider-accounts">{copy("actions.connectProvider")}</Link></Button>}
                  secondaryAction={<Button type="button" variant="outline" asChild><a href="/docs/intro-to-gateway" target="_blank" rel="noreferrer"><BookOpen data-icon="inline-start" />{copy("actions.readDocs")}</a></Button>}
                />
              </div>
            )}
          </TabsContent>
        </Tabs>
      </section>
    </main>
    <TypedConfirmDialog
      open={pendingDisconnectAccount !== null}
      onOpenChange={(open) => {
        if (!open) setPendingDisconnectAccount(null);
      }}
      title={copy("messages.disconnectTitle")}
      description={copy("messages.disconnectConfirm", { name: pendingDisconnectAccount?.name ?? "" })}
      expectedValue={pendingDisconnectAccount?.name ?? ""}
      inputLabel={copy("messages.disconnectConfirmInput", { name: pendingDisconnectAccount?.name ?? "" })}
      cancelLabel={copy("actions.cancel")}
      confirmLabel={copy("actions.disconnect")}
      pendingLabel={copy("actions.disconnecting")}
      onConfirm={() => pendingDisconnectAccount ? disconnectAccount(pendingDisconnectAccount) : undefined}
      disabled={!pendingDisconnectAccount}
      inputDir="auto"
    />
    <ProviderConnectionTestDialog
      open={connectionTestAccount !== null}
      onOpenChange={(open) => {
        if (!open) setConnectionTestAccount(null);
      }}
      account={connectionTestAccount}
    />
    <FormSheet
      open={effectiveAccessAccount !== null}
      onOpenChange={(open) => {
        if (!open) setEffectiveAccessAccount(null);
      }}
      title={ux("accessTester.title")}
      description={ux("accessTester.description", {
        name: effectiveAccessAccount?.name ?? "",
        provider: effectiveAccessAccount?.provider ?? "",
      })}
      side="left"
      size="lg"
    >
      {effectiveAccessAccount ? (
        <EffectiveAccessTester
          account={effectiveAccessAccount}
          onSearchPrincipal={searchPrincipalCandidates}
          principalSelectorLabels={principalSelectorLabels}
        />
      ) : null}
    </FormSheet>
    </>
  );
}

function FilterRow(props: {
  activeTab: "all" | "connected";
  query: string;
  setQuery: (value: string) => void;
  providerFilters: string[];
  setProviderFilters: (value: string[]) => void;
  providerOptions: Array<{ slug: string; label: string }>;
  searchPlaceholder: string;
  providerFilterLabel: string;
  direction: PageDirection;
}) {
  const selectedProviderLabels = props.providerOptions
    .filter((item) => props.providerFilters.includes(item.slug))
    .map((item) => item.label);
  const providerFilterText =
    selectedProviderLabels.length === 0
      ? props.providerFilterLabel
      : selectedProviderLabels.length === 1
        ? selectedProviderLabels[0]
        : `${new Intl.NumberFormat("fa-IR").format(selectedProviderLabels.length)} ${props.providerFilterLabel}`;

  function toggleProvider(slug: string, checked: boolean) {
    if (checked) {
      props.setProviderFilters([...props.providerFilters, slug]);
      return;
    }
    props.setProviderFilters(props.providerFilters.filter((item) => item !== slug));
  }

  return (
    <div className="flex min-w-0 items-center gap-2" dir={props.direction}>
      {props.activeTab === "connected" ? (
        <DropdownMenu dir={props.direction}>
          <DropdownMenuTrigger asChild>
            <Button
              type="button"
              variant="outline"
              className="h-9 shrink-0 justify-between gap-2 px-3 text-sm font-normal lg:w-72"
              aria-label={props.providerFilterLabel}
            >
              <span className="truncate">{providerFilterText}</span>
              <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align={props.direction === "rtl" ? "end" : "start"} className="max-h-80 w-72 overflow-y-auto">
            <DropdownMenuCheckboxItem
              checked={props.providerFilters.length === 0}
              onCheckedChange={() => props.setProviderFilters([])}
              onSelect={(event) => event.preventDefault()}
              className="text-start"
            >
              {props.providerFilterLabel}
            </DropdownMenuCheckboxItem>
            {props.providerOptions.map((item) => (
              <DropdownMenuCheckboxItem
                key={item.slug}
                checked={props.providerFilters.includes(item.slug)}
                onCheckedChange={(checked) => toggleProvider(item.slug, checked === true)}
                onSelect={(event) => event.preventDefault()}
                className="text-start"
              >
                {item.label}
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      ) : null}
      <div className="relative min-w-0 flex-1" dir={props.direction}>
        <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={props.query}
          onChange={(event) => props.setQuery(event.target.value)}
          placeholder={props.searchPlaceholder}
          dir={props.direction}
          className="h-9 ps-9 text-start"
        />
      </div>
    </div>
  );
}

function ProviderCard(props: {
  provider: ProviderView;
  connectLabel: string;
  manageConnectionsLabel: string;
  modelUnitLabel: string;
  modelMetricLabel: string;
  connectionMetricLabel: string;
  noConnectionLabel: string;
  createConnectionLabel: string;
  accountCount: number;
  modelCount?: number | null;
  onManageConnections: (providerSlug: string) => void;
  onConnect: (providerSlug: string) => void;
  connectedStatusLabel: string;
}) {
  const providerSlug = normalizeProviderSlug(props.provider.provider);
  const manageHref = `/provider-accounts?tab=connected&provider=${encodeURIComponent(providerSlug)}`;
  const connectHref = `/provider-accounts/new?provider=${encodeURIComponent(providerSlug)}`;
  const providerHref = `${getModelDiscoveryCatalogBaseURL()}/providers/${encodeURIComponent(providerSlug)}`;
  const providerModelsHref = `${getModelDiscoveryCatalogBaseURL()}/models?provider=${encodeURIComponent(providerSlug)}`;
  const resolvedModelCount = typeof props.modelCount === "number" ? props.modelCount : null;
  const formattedModelCount =
    resolvedModelCount === null ? "-" : new Intl.NumberFormat("fa-IR").format(resolvedModelCount);
  const formattedAccountCount = new Intl.NumberFormat("fa-IR").format(props.accountCount);
  const hasConnections = props.accountCount > 0;

  return (
    <TooltipProvider delayDuration={120}>
      <Card className="flex min-h-56 flex-col overflow-hidden rounded-xl border border-card-border bg-card text-card-foreground shadow-none transition-colors hover:border-primary/35">
        <CardContent className="flex flex-1 flex-col gap-3 px-6 pb-4 pt-6 md:pt-6">
          <div className="flex min-w-0 items-center gap-3">
            <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border border-border bg-background p-2 shadow-sm">
              <ProviderLogo
                provider={props.provider.provider}
                displayName={props.provider.display_name}
                logoURL={props.provider.logo_url}
                size={32}
                decorative
              />
            </div>
            <div className="flex min-w-0 flex-1 items-center justify-between gap-2">
              <h2 className="flex min-h-12 min-w-0 flex-1 items-center gap-1.5 text-lg font-semibold leading-none tracking-normal">
                <span className="truncate" dir="ltr">{props.provider.display_name}</span>
                <a
                  href={providerHref}
                  target="_blank"
                  rel="noreferrer"
                  aria-label={props.provider.display_name}
                  className="inline-flex size-5 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                >
                  <ExternalLink className="size-3" />
                </a>
              </h2>
              <div className="flex shrink-0 items-center gap-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    {hasConnections ? (
                      <Button type="button" size="icon" variant="ghost" asChild className="size-8 rounded-lg text-muted-foreground hover:text-foreground">
                        <Link href={manageHref} onClick={() => props.onManageConnections(providerSlug)} aria-label={props.manageConnectionsLabel}>
                          <UserCog className="size-4" />
                        </Link>
                      </Button>
                    ) : (
                      <Button type="button" size="icon" variant="ghost" disabled className="size-8 rounded-lg text-muted-foreground" aria-label={props.manageConnectionsLabel}>
                        <UserCog className="size-4" />
                      </Button>
                    )}
                  </TooltipTrigger>
                  <TooltipContent>{props.manageConnectionsLabel}</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </div>

          <div className="grid min-h-[4.875rem] grid-cols-2 items-center border-t border-border/70 pt-2.5">
            <div className="text-center">
              <p className="text-xs text-muted-foreground">{props.modelMetricLabel}</p>
              <p className="mt-1 text-base font-bold leading-none">
                {resolvedModelCount !== null ? (
                  <a
                    href={providerModelsHref}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center justify-center gap-1 rounded-sm text-foreground underline-offset-4 hover:text-primary hover:underline"
                  >
                    <bdi>{formattedModelCount}</bdi>{" "}
                    <span>{props.modelUnitLabel}</span>
                  </a>
                ) : (
                  <span className="inline-flex items-center justify-center text-muted-foreground">
                    <bdi>{formattedModelCount}</bdi>
                  </span>
                )}
              </p>
            </div>
            <div className="text-center">
              <p className="text-xs text-muted-foreground">{props.connectionMetricLabel}</p>
              <p className={props.accountCount > 0 ? "mt-1 text-base font-bold leading-none text-[var(--success-soft-foreground)]" : "mt-1 text-base font-bold leading-none text-muted-foreground"}>
                <StatusBadge status={hasConnections ? "active" : "inactive"}>
                  {hasConnections
                    ? `${formattedAccountCount} ${props.connectedStatusLabel}`
                    : props.noConnectionLabel}
                </StatusBadge>
              </p>
            </div>
          </div>
        </CardContent>
        <CardFooter className="bg-card px-6 pb-6 pt-0">
          <Button type="button" className="w-full" asChild>
            <Link
              href={connectHref}
              onClick={(event) => {
                if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
                event.preventDefault();
                props.onConnect(providerSlug);
              }}
            >
              <Plus data-icon="inline-start" />
              {props.createConnectionLabel || props.connectLabel}
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </TooltipProvider>
  );
}

function ConnectedIntegrationCard(props: {
  account: TenantProviderAccount;
  provider?: ProviderView;
  fqn: string;
  editLabel: string;
  manageModelsLabel: string;
  testConnectionLabel: string;
  testAllowedAccessLabel: string;
  viewDocsLabel: string;
  disconnectLabel: string;
  disconnectingLabel: string;
  credentialTypeLabel: string;
  dedicatedCredentialLabel: string;
  platformCredentialLabel: string;
  onTestConnection: () => void;
  onTestAllowedAccess: () => void;
  onDisconnect: () => void;
  isDisconnecting: boolean;
}) {
  const docsURL = getProviderDocsURL(props.provider ?? { provider: props.account.provider });
  const workspaceHref = `/provider-accounts?account=${encodeURIComponent(props.account.id)}`;
  const providerDisplayName = props.provider?.display_name ?? props.account.provider;
  const providerHref = buildModelDiscoveryProviderHref(props.account.provider);
  const canTestStoredCredential =
    props.account.has_credential &&
    props.account.auth_mode !== "no_credential" &&
    !props.account.use_platform_managed_credentials;

  return (
    <Card className="group flex min-h-56 w-full max-w-full flex-col overflow-hidden rounded-xl border border-card-border bg-card text-card-foreground shadow-none transition-colors hover:border-primary/35 sm:w-96">
      <CardContent className="flex flex-1 flex-col gap-3 px-4 pb-4 pt-4 sm:px-6 sm:pt-6 md:pt-6">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border border-border bg-background p-2 shadow-sm">
            <ProviderLogo
              provider={props.account.provider}
              displayName={providerDisplayName}
              logoURL={props.provider?.logo_url}
              size={32}
              decorative
            />
          </div>
          <div className="flex min-w-0 flex-1 items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="flex min-w-0 items-center gap-1">
                <a
                  href={providerHref}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex min-w-0 items-center gap-1 text-xs font-medium text-primary underline-offset-4 hover:underline"
                >
                  <span className="truncate" dir="ltr">{providerDisplayName}</span>
                  <ExternalLink aria-hidden="true" />
                </a>
              </div>
              <CardTitle className="mt-2 truncate text-base" dir="ltr" title={props.account.name}>{props.account.name}</CardTitle>
            </div>
            <RowActionsMenu
              label={props.manageModelsLabel}
              triggerSize="icon-sm"
              actions={[
                { key: "edit-integration", label: props.editLabel, href: workspaceHref, icon: <Settings /> },
                { key: "manage-models", label: props.manageModelsLabel, href: workspaceHref, icon: <Plug /> },
                { key: "test-allowed-access", label: props.testAllowedAccessLabel, onSelect: props.onTestAllowedAccess, icon: <UserCog /> },
                ...(canTestStoredCredential
                  ? [{ key: "test-connection", label: props.testConnectionLabel, onSelect: props.onTestConnection, icon: <Activity /> }]
                  : []),
                ...(docsURL ? [{ key: "view-docs", label: props.viewDocsLabel, href: docsURL, external: true, icon: <ExternalLink /> }] : []),
                {
                  key: "disconnect",
                  label: props.disconnectLabel,
                  pendingLabel: props.disconnectingLabel,
                  pending: props.isDisconnecting,
                  destructive: true,
                  separatorBefore: true,
                  icon: <Trash2 />,
                  onSelect: props.onDisconnect,
                },
              ]}
            />
          </div>
        </div>

        <div className="flex min-h-20 flex-col justify-center border-t border-border/70 pt-3 text-xs text-muted-foreground">
          <span>{props.credentialTypeLabel}</span>
          <span className="mt-1 text-sm font-semibold text-foreground">
            {props.account.has_credential ? props.dedicatedCredentialLabel : props.platformCredentialLabel}
          </span>
        </div>
      </CardContent>
      <CardFooter className="bg-card px-4 pb-4 pt-0 sm:px-6 sm:pb-6">
        <div className="min-w-0 w-full overflow-x-auto rounded-md border border-border bg-background px-3 py-2">
          <ModelNameValue value={props.fqn} />
        </div>
      </CardFooter>
    </Card>
  );
}

function ProviderGridSkeleton(props: { direction: PageDirection }) {
  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3" dir={props.direction}>
      {Array.from({ length: 8 }).map((_, index) => (
        <Card key={`integration-skeleton-${index}`} className="shadow-none">
          <CardContent className="flex h-44 flex-col gap-4 p-4">
            <div className="flex items-start justify-between gap-3">
              <Skeleton className="size-10 rounded-lg" />
              <Skeleton className="h-8 w-28" />
            </div>
            <div className="mt-auto flex flex-col gap-2">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-2/3" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function getProviderDocsURL(provider: Pick<ProviderView, "provider" | "docs_url" | "docsUrl">): string {
  const direct = provider.docs_url?.trim() || provider.docsUrl?.trim();
  if (direct) return direct;
  const slug = normalizeProviderSlug(provider.provider);
  return PROVIDER_DOCS_URLS[slug] ?? "";
}
