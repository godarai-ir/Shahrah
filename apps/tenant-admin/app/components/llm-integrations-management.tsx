"use client";

import Link from "next/link";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { BookOpen, ChevronDown, Copy, ExternalLink, Link2, Plus, Plug, Search, Settings, Trash2, Unplug, UserCog } from "lucide-react";
import { useTranslations } from "next-intl";

import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import type {
  AccountsResponse,
  ProviderCatalogProvider,
  ProviderModelsResponse,
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
import { EmptyState, RowActionsMenu, TypedConfirmDialog } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
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
import { cn } from "@/lib/utils";

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

const INITIAL_PROVIDER_BATCH = 9;
const PROVIDER_BATCH_SIZE = 6;
const MODEL_COUNT_REQUEST_TIMEOUT_MS = 4000;

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
}) {
  const t = useTranslations("providerAccounts");
  const copy = useTranslations("providerAccounts.integrations");
  const pathname = usePathname();
  const router = useRouter();
  const initialProviderFilters = parseProviderFilters(props.initialProvider);
  const pageDirection = usePageDirection();

  const [activeTab, setActiveTab] = useState<"all" | "connected">(props.initialTab ?? (initialProviderFilters.length > 0 ? "connected" : "all"));
  const [providers, setProviders] = useState<ProviderView[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [modelCounts, setModelCounts] = useState<Record<string, number | null>>({});
  const [query, setQuery] = useState("");
  const [pendingDisconnectAccount, setPendingDisconnectAccount] = useState<TenantProviderAccount | null>(null);
  const [providerFilters, setProviderFilters] = useState<string[]>(initialProviderFilters);
  const [providerPagination, setProviderPagination] = useState<{ signature: string; count: number }>({
    signature: "",
    count: INITIAL_PROVIDER_BATCH,
  });
  const [isLoading, setIsLoading] = useState(true);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [deletingAccountID, setDeletingAccountID] = useState("");
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const pendingModelCountSlugsRef = useRef<Set<string>>(new Set());

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setBanner(null);
    try {
      const [providersPayload, accountsPayload] = await Promise.all([
        requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
        requestJSON<AccountsResponse>("/api/tenant-provider-accounts"),
      ]);
      const nextProviders = (unwrapData<ProviderView[]>(providersPayload) ?? []).filter((item) => item.enabled);
      const nextProviderSlugs = new Set(nextProviders.map((item) => normalizeProviderSlug(item.provider)));
      pendingModelCountSlugsRef.current.clear();
      setProviders(nextProviders);
      setAccounts(unwrapData<TenantProviderAccount[]>(accountsPayload) ?? []);
      setModelCounts((current) =>
        Object.fromEntries(Object.entries(current).filter(([slug]) => nextProviderSlugs.has(slug))),
      );
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

  const providerPaginationSignature = useMemo(
    () => `${activeTab}:${query.trim().toLowerCase()}:${providers.map((item) => normalizeProviderSlug(item.provider)).join("|")}`,
    [activeTab, providers, query],
  );

  const providerOptions = useMemo(
    () => providers.map((item) => ({ slug: normalizeProviderSlug(item.provider), label: item.display_name })),
    [providers],
  );

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

  const visibleProviderCount =
    providerPagination.signature === providerPaginationSignature ? providerPagination.count : INITIAL_PROVIDER_BATCH;

  const visibleProviders = useMemo(
    () => filteredProviders.slice(0, visibleProviderCount),
    [filteredProviders, visibleProviderCount],
  );

  const hasMoreProviders = visibleProviders.length < filteredProviders.length;

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

  useEffect(() => {
    if (activeTab !== "all" || !hasMoreProviders) return undefined;
    const target = loadMoreRef.current;
    if (!target || typeof IntersectionObserver === "undefined") return undefined;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setProviderPagination((current) => {
            const baseCount = current.signature === providerPaginationSignature ? current.count : INITIAL_PROVIDER_BATCH;
            return {
              signature: providerPaginationSignature,
              count: Math.min(baseCount + PROVIDER_BATCH_SIZE, filteredProviders.length),
            };
          });
        }
      },
      { rootMargin: "160px 0px" },
    );

    observer.observe(target);
    return () => observer.disconnect();
  }, [activeTab, filteredProviders.length, hasMoreProviders, providerPaginationSignature, visibleProviderCount]);

  useEffect(() => {
    if (visibleProviders.length === 0) return undefined;

    for (const provider of visibleProviders) {
      const slug = normalizeProviderSlug(provider.provider);
      if (!slug || modelCounts[slug] !== undefined || pendingModelCountSlugsRef.current.has(slug)) continue;

      pendingModelCountSlugsRef.current.add(slug);
      void loadProviderModelCount(slug)
        .then((count) => {
          setModelCounts((current) => (current[slug] === undefined ? { ...current, [slug]: count } : current));
        })
        .finally(() => {
          pendingModelCountSlugsRef.current.delete(slug);
        });
    }
    return undefined;
  }, [modelCounts, visibleProviders]);

  async function copyIdentifier(account: TenantProviderAccount) {
    const value = `${account.provider}:${account.name}`;
    await copyTextToClipboard(value);
    setBanner({ kind: "success", message: copy("messages.identifierCopied") });
  }

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

  return (
    <>
    <main className="flex flex-1 px-3 py-3 sm:px-4">
      <section className="flex min-h-[calc(100dvh-1.5rem)] w-full flex-col overflow-hidden rounded-xl border border-border bg-background">
        {banner ? (
          <div className="px-4 pt-3">
            <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
              <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
              <AlertDescription>{banner.message}</AlertDescription>
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
              <div className="space-y-3">
                <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3" dir={pageDirection}>
                  {visibleProviders.map((provider) => (
                    <ProviderCard
                      key={provider.provider}
                      provider={provider}
                      connectLabel={copy("actions.connect")}
                      manageConnectionsLabel={copy("actions.manageConnections")}
                      modelUnitLabel={copy("card.modelUnit")}
                      modelMetricLabel={copy("card.modelMetricLabel")}
                      connectionMetricLabel={copy("card.connectionMetricLabel")}
                      activeAccountCountLabel={copy("card.activeAccountCount", {
                        count: accountCountsByProvider.get(normalizeProviderSlug(provider.provider)) ?? 0,
                      })}
                      noConnectionLabel={copy("card.noConnection")}
                      createConnectionLabel={copy("card.createConnection")}
                      accountCount={accountCountsByProvider.get(normalizeProviderSlug(provider.provider)) ?? 0}
                      modelCount={modelCounts[normalizeProviderSlug(provider.provider)]}
                      onManageConnections={openConnectedProvider}
                    />
                  ))}
                </div>
                {hasMoreProviders ? <div ref={loadMoreRef} className="h-4 w-full" aria-hidden="true" /> : null}
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
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3" dir={pageDirection}>
                {filteredAccounts.map((account) => {
                  const provider = providers.find((item) => stringsEqualFold(item.provider, account.provider));
                  return (
                    <ConnectedIntegrationCard
                      key={account.id}
                      account={account}
                      provider={provider}
                      fqn={`${account.provider}:${account.name}`}
                      copyLabel={copy("actions.copyIdentifier")}
                      editLabel={copy("actions.editIntegration")}
                      manageModelsLabel={copy("actions.manageModels")}
                      viewDocsLabel={copy("actions.viewDocs")}
                      disconnectLabel={copy("actions.disconnect")}
                      disconnectingLabel={copy("actions.disconnecting")}
                      credentialModeLabel={copy("card.credentialModeLabel")}
                      authModeLabel={copy("card.authModeLabel")}
                      providerCredentialLabel={copy("card.providerCredential")}
                      virtualKeyOnlyLabel={copy("card.virtualKeyOnly")}
                      description={copy("integrationDescription", {
                        name: account.name,
                        provider: provider?.display_name ?? account.provider,
                        authMode: formatAuthMode(account.auth_mode),
                      })}
                      onCopy={() => void copyIdentifier(account)}
                      onDisconnect={() => setPendingDisconnectAccount(account)}
                      isDisconnecting={deletingAccountID === account.id}
                    />
                  );
                })}
              </div>
            ) : (
              <EmptyState
                title={copy("empty.connectedTitle")}
                description={copy("empty.connectedDescription")}
                icon={Plug}
                primaryAction={<Button type="button" asChild><Link href="/provider-accounts/new">{copy("actions.connectProvider")}</Link></Button>}
                secondaryAction={<Button type="button" variant="outline" asChild><a href={PROVIDER_DOCS_URLS.openai} target="_blank" rel="noreferrer"><BookOpen data-icon="inline-start" />{copy("actions.readDocs")}</a></Button>}
              />
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
          className={cn("h-9 ps-9", props.direction === "rtl" ? "text-right" : "text-left")}
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
  activeAccountCountLabel: string;
  noConnectionLabel: string;
  createConnectionLabel: string;
  accountCount: number;
  modelCount?: number | null;
  onManageConnections: (providerSlug: string) => void;
}) {
  const providerSlug = normalizeProviderSlug(props.provider.provider);
  const manageHref = `/provider-accounts?tab=connected&provider=${encodeURIComponent(providerSlug)}`;
  const connectHref = `/provider-accounts/new?provider=${encodeURIComponent(providerSlug)}`;
  const providerHref = `${getModelDiscoveryCatalogBaseURL()}/providers/${encodeURIComponent(providerSlug)}`;
  const providerModelsHref = `${getModelDiscoveryCatalogBaseURL()}/models?provider=${encodeURIComponent(providerSlug)}`;
  const isModelCountLoading = props.modelCount === undefined;
  const resolvedModelCount = typeof props.modelCount === "number" ? props.modelCount : null;
  const formattedModelCount =
    resolvedModelCount === null ? "-" : new Intl.NumberFormat("fa-IR").format(resolvedModelCount);
  const formattedAccountCount = new Intl.NumberFormat("fa-IR").format(props.accountCount);
  const hasConnections = props.accountCount > 0;

  return (
    <TooltipProvider delayDuration={120}>
      <Card className="surface-elevated min-h-[13.75rem] overflow-hidden rounded-xl border shadow-none transition-colors hover:border-primary/35">
        <CardContent className="flex h-full flex-col gap-3 !p-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border border-border bg-background p-2 shadow-sm">
              <ProviderLogo
                provider={props.provider.provider}
                displayName={props.provider.display_name}
                logoURL={props.provider.logo_url}
                size={32}
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
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button type="button" size="icon" asChild className="size-8 rounded-lg bg-primary text-primary-foreground shadow-none hover:bg-primary-hover">
                      <Link href={connectHref} aria-label={props.createConnectionLabel || props.connectLabel}>
                        <Plus className="size-4" />
                      </Link>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>{props.createConnectionLabel || props.connectLabel}</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </div>

          <div className="grid min-h-[4.875rem] grid-cols-2 items-center border-t border-border/70 pt-2.5">
            <div className="text-center">
              <p className="text-xs text-muted-foreground">{props.modelMetricLabel}</p>
              <p className="mt-1 text-base font-bold leading-none" dir="rtl">
                {isModelCountLoading ? (
                  <span className="inline-block h-5 w-14 animate-pulse rounded bg-muted align-middle" aria-label="loading" />
                ) : resolvedModelCount !== null ? (
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
              <p className={props.accountCount > 0 ? "mt-1 text-base font-bold leading-none text-[var(--success-soft-foreground)]" : "mt-1 text-base font-bold leading-none text-muted-foreground"} dir="rtl">
                {props.accountCount > 0 ? (
                  `${formattedAccountCount} ${props.activeAccountCountLabel.replace(/[0-9۰-۹٠-٩,٬]+/g, "").trim()}`
                ) : (
                  <span className="inline-flex size-6 items-center justify-center rounded-full border border-border bg-muted/40 text-muted-foreground" aria-label={props.noConnectionLabel} title={props.noConnectionLabel}>
                    <Unplug className="size-3.5" />
                  </span>
                )}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </TooltipProvider>
  );
}

function ConnectedIntegrationCard(props: {
  account: TenantProviderAccount;
  provider?: ProviderView;
  fqn: string;
  copyLabel: string;
  editLabel: string;
  manageModelsLabel: string;
  viewDocsLabel: string;
  disconnectLabel: string;
  disconnectingLabel: string;
  credentialModeLabel: string;
  authModeLabel: string;
  providerCredentialLabel: string;
  virtualKeyOnlyLabel: string;
  description: string;
  onCopy: () => void;
  onDisconnect: () => void;
  isDisconnecting: boolean;
}) {
  const docsURL = getProviderDocsURL(props.provider ?? { provider: props.account.provider });
  const workspaceHref = `/provider-accounts?account=${encodeURIComponent(props.account.id)}`;
  const providerDisplayName = props.provider?.display_name ?? props.account.provider;
  const providerHref = buildModelDiscoveryProviderHref(props.account.provider);

  return (
    <Card className="surface-elevated group overflow-hidden rounded-xl border border-card-border shadow-none transition-colors hover:border-primary/35">
      <CardHeader className="flex flex-row items-start justify-between gap-3 p-4 pb-3">
        <div className="flex min-w-0 items-start gap-3">
          <div className="flex size-11 shrink-0 items-center justify-center rounded-lg border border-border bg-background p-2 shadow-sm">
            <ProviderLogo
              provider={props.account.provider}
              displayName={providerDisplayName}
              logoURL={props.provider?.logo_url}
              size={28}
            />
          </div>
          <div className="min-w-0">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
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
            <CardTitle className="mt-2 truncate text-base">{props.account.name}</CardTitle>
            <CardDescription className="mt-2 line-clamp-3 leading-6">
              {props.description}
            </CardDescription>
          </div>
        </div>
        <RowActionsMenu
          label={props.manageModelsLabel}
          triggerSize="icon-sm"
          actions={[
            { key: "edit-integration", label: props.editLabel, href: workspaceHref, icon: <Settings /> },
            { key: "manage-models", label: props.manageModelsLabel, href: workspaceHref, icon: <Plug /> },
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
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-3 px-4 pb-3 pt-0">
        <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
          <div className="rounded-lg border border-border bg-muted/20 px-3 py-2">
            <span className="block">{props.credentialModeLabel}</span>
            <span className="mt-1 block truncate font-medium text-foreground">
              {props.account.allows_virtual_keys_only ? props.virtualKeyOnlyLabel : props.providerCredentialLabel}
            </span>
          </div>
          <div className="rounded-lg border border-border bg-muted/20 px-3 py-2">
            <span className="block">{props.authModeLabel}</span>
            <span className="mt-1 block truncate font-mono text-foreground" dir="ltr">{props.account.auth_mode}</span>
          </div>
        </div>
      </CardContent>
      <CardFooter className="border-t border-border/70 bg-muted/20 p-3">
        <Button type="button" size="sm" variant="outline" onClick={props.onCopy} className="min-w-0 flex-1 justify-between font-mono" dir="ltr">
          <span className="truncate" dir="ltr">{props.fqn}</span>
          <Copy data-icon="inline-end" />
        </Button>
      </CardFooter>
    </Card>
  );
}

function formatAuthMode(value: string): string {
  return value.replaceAll("_", " ");
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

async function loadProviderModelCount(providerSlug: string): Promise<number | null> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), MODEL_COUNT_REQUEST_TIMEOUT_MS);
  try {
    const payload = await requestJSON<ProviderModelsResponse>(
      `/api/provider-catalog/providers/${encodeURIComponent(providerSlug)}/models`,
      { signal: controller.signal },
    );
    return unwrapData<ProviderModelsResponse["data"]>(payload)?.length ?? 0;
  } catch {
    return null;
  } finally {
    window.clearTimeout(timeout);
  }
}
