"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { BookOpen, ChevronDown, Copy, ExternalLink, Link2, MoreVertical, Network, Plug, Search, Settings, Trash2, Unplug, UserCog } from "lucide-react";
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
  normalizeError,
  normalizeProviderSlug,
  requestJSON,
  stringsEqualFold,
  unwrapData,
} from "@/app/components/provider-account-wizard/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  const [modelCounts, setModelCounts] = useState<Record<string, number>>({});
  const [query, setQuery] = useState("");
  const [providerFilters, setProviderFilters] = useState<string[]>(initialProviderFilters);
  const [isLoading, setIsLoading] = useState(true);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [deletingAccountID, setDeletingAccountID] = useState("");

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setBanner(null);
    try {
      const [providersPayload, accountsPayload] = await Promise.all([
        requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
        requestJSON<AccountsResponse>("/api/tenant-provider-accounts"),
      ]);
      const nextProviders = (unwrapData<ProviderView[]>(providersPayload) ?? []).filter((item) => item.enabled);
      setProviders(nextProviders);
      setAccounts(unwrapData<TenantProviderAccount[]>(accountsPayload) ?? []);
      void loadProviderModelCounts(nextProviders).then(setModelCounts);
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

  async function copyIdentifier(account: TenantProviderAccount) {
    const value = `${account.provider}:${account.name}`;
    await navigator.clipboard.writeText(value);
    setBanner({ kind: "success", message: copy("messages.identifierCopied") });
  }

  async function disconnectAccount(account: TenantProviderAccount) {
    const confirmed = window.confirm(copy("messages.disconnectConfirm", { name: account.name }));
    if (!confirmed) return;
    setDeletingAccountID(account.id);
    setBanner(null);
    try {
      await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(account.id)}`, { method: "DELETE" });
      setAccounts((current) => current.filter((item) => item.id !== account.id));
      setBanner({ kind: "success", message: copy("messages.disconnectSuccess") });
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
    <main className="flex flex-1 px-3 py-3 sm:px-4">
      <section className="flex min-h-[calc(100dvh-1.5rem)] w-full flex-col overflow-hidden rounded-xl border border-border bg-background">
        <header className="border-b border-border px-5 py-5 sm:px-6" dir={pageDirection}>
          <div className="flex max-w-5xl flex-col gap-1.5 text-start">
            <div className="flex min-w-0 items-center gap-2">
              <Network className="size-4 text-muted-foreground" />
              <h1 className="text-2xl font-semibold leading-8 tracking-normal">{copy("pageTitle")}</h1>
            </div>
            <p className="text-sm leading-6 text-muted-foreground">{copy("pageDescription")}</p>
          </div>
        </header>

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
            <div className="flex flex-col gap-3">
              <TabsList dir={pageDirection} className="justify-start gap-2">
                <TabsTrigger value="all">
                  <Plug data-icon="inline-start" />
                  {copy("tabs.all")}
                </TabsTrigger>
                <TabsTrigger value="connected">
                  <Link2 data-icon="inline-start" />
                  {copy("tabs.connected")}
                </TabsTrigger>
              </TabsList>

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
                    activeAccountCountLabel={copy("card.activeAccountCount", {
                      count: accountCountsByProvider.get(normalizeProviderSlug(provider.provider)) ?? 0,
                    })}
                    noConnectionLabel={copy("card.noConnection")}
                    createConnectionLabel={copy("card.createConnection")}
                    accountCount={accountCountsByProvider.get(normalizeProviderSlug(provider.provider)) ?? 0}
                    modelCount={modelCounts[normalizeProviderSlug(provider.provider)]}
                    descriptionFallback={copy("providerDescriptionFallback")}
                    onManageConnections={openConnectedProvider}
                  />
                ))}
              </div>
            ) : (
              <EmptyState
                title={copy("empty.providersTitle")}
                description={copy("empty.providersDescription")}
                actionLabel={copy("actions.reload")}
                onAction={() => void loadData()}
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
                      descriptionFallback={copy("integrationDescriptionFallback")}
                      onCopy={() => void copyIdentifier(account)}
                      onDisconnect={() => void disconnectAccount(account)}
                      isDisconnecting={deletingAccountID === account.id}
                    />
                  );
                })}
              </div>
            ) : (
              <EmptyState
                title={copy("empty.connectedTitle")}
                description={copy("empty.connectedDescription")}
                actionLabel={copy("actions.connectProvider")}
                actionHref="/provider-accounts/new"
                secondaryLabel={copy("actions.readDocs")}
                secondaryHref={PROVIDER_DOCS_URLS.openai}
              />
            )}
          </TabsContent>
        </Tabs>
      </section>
    </main>
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
    <div className="flex flex-col gap-2 lg:flex-row" dir={props.direction}>
      {props.activeTab === "connected" ? (
        <DropdownMenu dir={props.direction}>
          <DropdownMenuTrigger asChild>
            <Button
              type="button"
              variant="outline"
              className="h-9 justify-between gap-2 px-3 text-sm font-normal lg:w-72"
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
  modelCount?: number;
  descriptionFallback: string;
  onManageConnections: (providerSlug: string) => void;
}) {
  const providerSlug = normalizeProviderSlug(props.provider.provider);
  const description = getProviderDescription(props.provider, props.descriptionFallback);
  const manageHref = `/provider-accounts?tab=connected&provider=${encodeURIComponent(providerSlug)}`;
  const connectHref = `/provider-accounts/new?provider=${encodeURIComponent(providerSlug)}`;
  const providerHref = `${getModelDiscoveryCatalogBaseURL()}/providers/${encodeURIComponent(providerSlug)}`;
  const providerModelsHref = `${getModelDiscoveryCatalogBaseURL()}/models?provider=${encodeURIComponent(providerSlug)}`;
  const formattedModelCount = props.modelCount === undefined ? "-" : new Intl.NumberFormat("fa-IR").format(props.modelCount);
  const formattedAccountCount = new Intl.NumberFormat("fa-IR").format(props.accountCount);

  return (
    <Card className="surface-elevated min-h-[16.5rem] overflow-hidden rounded-xl border shadow-none transition-colors hover:border-primary/35">
      <CardContent className="flex h-full flex-col gap-3 !p-4">
        <div className="flex min-w-0 items-start gap-3">
          <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border border-border bg-background p-2 shadow-sm">
            <ProviderLogo
              provider={props.provider.provider}
              displayName={props.provider.display_name}
              logoURL={props.provider.logo_url}
              size={32}
            />
          </div>
          <h2 className="flex min-w-0 flex-1 items-center gap-1.5 pt-0.5 text-base font-semibold leading-6 tracking-normal">
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
        </div>

        <p className="line-clamp-3 min-h-[3.75rem] w-full text-right text-xs leading-5 text-muted-foreground" dir="rtl">
          {description}
        </p>

        <div className="grid min-h-[4.875rem] grid-cols-2 items-center border-y border-border/70 py-2.5">
          <div className="text-center">
            <p className="text-xs text-muted-foreground">{props.modelMetricLabel}</p>
            <p className="mt-1 text-base font-bold leading-none" dir="rtl">
              {props.modelCount === undefined ? (
                <span className="inline-block h-5 w-14 animate-pulse rounded bg-muted align-middle" aria-label="loading" />
              ) : (
                <a
                  href={providerModelsHref}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center justify-center gap-1 rounded-sm text-foreground underline-offset-4 hover:text-primary hover:underline"
                >
                  <bdi>{formattedModelCount}</bdi>{" "}
                  <span>{props.modelUnitLabel}</span>
                </a>
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

        <div className="grid grid-cols-2 gap-2">
          {props.accountCount > 0 ? (
            <Button type="button" size="sm" asChild className="surface-primary-soft surface-primary-soft-hover h-8 rounded-md border text-xs font-semibold shadow-none">
              <Link href={manageHref} dir="ltr" onClick={() => props.onManageConnections(providerSlug)}>
                <span className="truncate">{props.manageConnectionsLabel}</span>
                <UserCog className="size-4" />
              </Link>
            </Button>
          ) : (
            <span aria-hidden="true" />
          )}
          <Button
            type="button"
            size="sm"
            asChild
            className="h-8 rounded-md bg-primary text-xs font-semibold text-primary-foreground shadow-none hover:bg-primary-hover"
          >
            <Link href={connectHref} dir="ltr">
              <span className="truncate">{props.createConnectionLabel || props.connectLabel}</span>
              <Plug className="size-4" />
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
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
  descriptionFallback: string;
  onCopy: () => void;
  onDisconnect: () => void;
  isDisconnecting: boolean;
}) {
  const docsURL = getProviderDocsURL(props.provider ?? { provider: props.account.provider });
  const workspaceHref = `/provider-accounts?account=${encodeURIComponent(props.account.id)}`;

  return (
    <Card className="surface-elevated min-h-52 overflow-hidden rounded-xl border shadow-none">
      <CardContent className="flex h-full flex-col p-0">
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div className="flex items-start justify-between gap-3">
            <ProviderLogo
              provider={props.account.provider}
              displayName={props.provider?.display_name ?? props.account.provider}
              logoURL={props.provider?.logo_url}
            />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button type="button" variant="ghost" size="icon" aria-label={props.manageModelsLabel} className="size-8">
                  <MoreVertical />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem asChild>
                  <Link href={workspaceHref}>
                    <Settings data-icon="inline-start" />
                    {props.editLabel}
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link href={workspaceHref}>
                    <Plug data-icon="inline-start" />
                    {props.manageModelsLabel}
                  </Link>
                </DropdownMenuItem>
                {docsURL ? (
                  <DropdownMenuItem asChild>
                    <a href={docsURL} target="_blank" rel="noreferrer">
                      <ExternalLink data-icon="inline-start" />
                      {props.viewDocsLabel}
                    </a>
                  </DropdownMenuItem>
                ) : null}
                <DropdownMenuItem onSelect={props.onDisconnect} disabled={props.isDisconnecting}>
                  <Trash2 data-icon="inline-start" />
                  {props.isDisconnecting ? props.disconnectingLabel : props.disconnectLabel}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold">{props.account.name}</h2>
            <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted-foreground">
              {props.provider?.description?.trim() || props.descriptionFallback}
            </p>
          </div>
        </div>
        <div className="flex items-center justify-between gap-3 border-t border-border/70 bg-muted/20 px-4 py-3">
          <div className="flex min-w-0 items-center gap-2 text-sm text-muted-foreground">
            <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary">I</span>
            <span className="truncate" dir="ltr">{props.account.provider}</span>
          </div>
          <Button type="button" size="sm" variant="outline" onClick={props.onCopy} className="max-w-40 font-mono">
            <span className="truncate" dir="ltr">{props.fqn}</span>
            <Copy data-icon="inline-end" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function EmptyState(props: {
  title: string;
  description: string;
  actionLabel: string;
  actionHref?: string;
  onAction?: () => void;
  secondaryLabel?: string;
  secondaryHref?: string;
}) {
  return (
    <section className="flex flex-col items-center gap-3 rounded-lg border border-dashed p-8 text-center">
      <div className="rounded-full border bg-muted/30 p-3">
        <Plug className="size-5 text-muted-foreground" />
      </div>
      <div className="flex max-w-md flex-col gap-1">
        <h2 className="text-lg font-semibold [unicode-bidi:plaintext]" dir="auto">{props.title}</h2>
        <p className="text-sm leading-6 text-muted-foreground [unicode-bidi:plaintext]" dir="auto">{props.description}</p>
      </div>
      <div className="flex flex-wrap items-center justify-center gap-2">
        {props.actionHref ? (
          <Button type="button" asChild>
            <Link href={props.actionHref}>{props.actionLabel}</Link>
          </Button>
        ) : (
          <Button type="button" onClick={props.onAction}>{props.actionLabel}</Button>
        )}
        {props.secondaryHref && props.secondaryLabel ? (
          <Button type="button" variant="outline" asChild>
            <a href={props.secondaryHref} target="_blank" rel="noreferrer">
              <BookOpen data-icon="inline-start" />
              {props.secondaryLabel}
            </a>
          </Button>
        ) : null}
      </div>
    </section>
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

async function loadProviderModelCounts(providers: ProviderView[]): Promise<Record<string, number>> {
  const pairs = await Promise.all(
    providers.map(async (provider) => {
      const slug = normalizeProviderSlug(provider.provider);
      try {
        const payload = await requestJSON<ProviderModelsResponse>(
          `/api/provider-catalog/providers/${encodeURIComponent(slug)}/models`,
        );
        return [slug, unwrapData<ProviderModelsResponse["data"]>(payload)?.length ?? 0] as const;
      } catch {
        return [slug, 0] as const;
      }
    }),
  );
  return Object.fromEntries(pairs);
}

function getProviderDescription(provider: ProviderView, fallback: string): string {
  const description = provider.description?.trim();
  if (description) return description;
  const displayName = provider.display_name.trim() || provider.provider;
  return fallback.replace("this provider", displayName).replace("این Provider", displayName);
}
