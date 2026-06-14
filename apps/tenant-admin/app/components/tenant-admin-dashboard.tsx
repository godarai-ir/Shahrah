"use client";

import { type ComponentProps, FormEvent, useCallback, useMemo, useState } from "react";
import {
  Activity,
  Bot,
  Check,
  Copy,
  KeyRound,
  type LucideIcon,
  Plus,
  RefreshCcw,
  Server,
  ShieldCheck,
  WalletCards,
} from "lucide-react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
  type PrincipalSearchOption,
} from "@/app/components/principal-search-input";
import { extractTenantSlugFromHost } from "@/lib/tenant";

type TenantKey = {
  id: string;
  key?: string;
  key_alias: string;
  key_name?: string;
  role: string;
  models: string[];
  team_id?: string;
  organization_id?: string;
  spend: number;
  blocked: boolean;
  expires_at?: string;
};

type TenantModel = {
  id: string;
  provider?: string;
  aliases?: string[];
};

type Wallet = {
  id: string;
  tenant_id: string;
  balance_irr: number;
  pending_irr: number;
};

type WalletTransaction = {
  id: string;
  type: string;
  amount_irr: number;
  amount_usd?: number | null;
  status: string;
  created_at: string;
};

type KeyListResponse = {
  data?: TenantKey[];
};
type ModelListResponse = {
  data?: TenantModel[];
};

type AccessPrincipalResponse = {
  data?: PrincipalSearchOption[];
};
type AllowedModelResponse = {
  data?: Array<{ id: string }>;
};
type WalletTransactionListResponse = {
  data?: WalletTransaction[];
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type DashboardLabels = {
  copied: string;
  copyValue: string;
  keyType: string;
  virtualKey: string;
  serviceAccount: string;
  teamID: string;
  organizationID: string;
  scope: string;
  aliases: string;
};

type TenantAdminDashboardProps = {
  initialHost: string;
  initialWallet: Wallet | null;
  initialKeys: TenantKey[];
  initialModels: TenantModel[];
  initialTransactions: WalletTransaction[];
  initialAllowedModels: string[];
  initialError: string | null;
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

const numberFormatter = new Intl.NumberFormat("fa-IR");

export default function TenantAdminDashboard(props: TenantAdminDashboardProps) {
  const t = useTranslations("dashboard");
  const common = useTranslations("common");
  const principalsT = useTranslations("providerAccounts");
  const principalsUx = useTranslations("providerAccounts.ux");
  const principalSelectorLabels = buildDashboardPrincipalSelectorLabels(principalsT, principalsUx);
  const labels = buildDashboardLabels(t);

  const [keys, setKeys] = useState<TenantKey[]>(props.initialKeys);
  const [models, setModels] = useState<TenantModel[]>(props.initialModels);
  const [wallet, setWallet] = useState<Wallet | null>(props.initialWallet);
  const [transactions, setTransactions] = useState<WalletTransaction[]>(props.initialTransactions);
  const [lastCreatedKey, setLastCreatedKey] = useState<string | null>(null);
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );

  const [isLoading, setIsLoading] = useState(false);
  const [isCreatingKey, setIsCreatingKey] = useState(false);
  const [isLoadingAllowedModels, setIsLoadingAllowedModels] = useState(false);
  const [isCreatingModel, setIsCreatingModel] = useState(false);

  const [keyAlias, setKeyAlias] = useState("tenant-dashboard-key");
  const [keyModelQuery, setKeyModelQuery] = useState("");
  const [keySelectedModels, setKeySelectedModels] = useState<string[]>([]);
  const [keyAllowedModels, setKeyAllowedModels] = useState<string[]>(props.initialAllowedModels);
  const [keyMode, setKeyMode] = useState<"virtual_key" | "service_account">(
    "virtual_key",
  );
  const [keyTeamID, setKeyTeamID] = useState("");
  const [keyOrgID, setKeyOrgID] = useState("");

  const [modelName, setModelName] = useState("gpt-4o-mini");
  const [provider, setProvider] = useState("openai");
  const [providerModel, setProviderModel] = useState("gpt-4o-mini");
  const [apiBase, setApiBase] = useState("http://provider.example/v1");
  const [apiKey, setApiKey] = useState("sk-provider");
  const [rpm, setRPM] = useState("250");
  const [tpm, setTPM] = useState("150000");
  const [weight, setWeight] = useState("1");

  const tenantSlug = useMemo(() => {
    const slug = extractTenantSlugFromHost(props.initialHost);
    return slug || props.initialHost;
  }, [props.initialHost]);

  const userIdentity =
    props.currentUser.email ?? props.currentUser.mobile ?? props.currentUser.id;
  const filteredAllowedModels = useMemo(() => {
    const q = keyModelQuery.trim().toLowerCase();
    if (!q) return keyAllowedModels;
    return keyAllowedModels.filter((item) => item.toLowerCase().includes(q));
  }, [keyAllowedModels, keyModelQuery]);
  const dashboardHealth = useMemo(() => {
    if (wallet === null) return { variant: "warning" as const, label: common("notSet") };
    if (wallet.balance_irr <= 0) return { variant: "destructive" as const, label: t("stats.balance") };
    if (keys.length === 0) return { variant: "warning" as const, label: t("keys.title") };
    if (models.length === 0) return { variant: "warning" as const, label: t("models.title") };
    return { variant: "success" as const, label: t("messages.successTitle") };
  }, [common, keys.length, models.length, t, wallet]);

  const loadAllowedModelsForScope = useCallback(async (mode: "virtual_key" | "service_account", teamID: string) => {
    const keyType = mode === "service_account" ? "vat" : "pat";
    if (keyType === "vat" && !teamID.trim()) {
      setKeyAllowedModels([]);
      setKeySelectedModels([]);
      return;
    }
    setIsLoadingAllowedModels(true);
    try {
      const query = new URLSearchParams();
      query.set("key_type", keyType);
      if (keyType === "vat") query.set("team_id", teamID.trim());
      const response = await requestJSON<AllowedModelResponse>(`/api/allowed-models?${query.toString()}`);
      const items = (response.data ?? []).map((item) => item.id);
      setKeyAllowedModels(items);
      setKeySelectedModels((prev) => prev.filter((item) => items.includes(item)));
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsLoadingAllowedModels(false);
    }
  }, [common]);

  const loadDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const [walletData, keysData, modelsData, txData] = await Promise.all([
        requestJSON<Wallet>("/api/wallet"),
        requestJSON<KeyListResponse>("/api/keys"),
        requestJSON<ModelListResponse>("/api/models"),
        requestJSON<WalletTransactionListResponse>("/api/wallet/transactions?limit=20"),
      ]);

      setWallet(walletData);
      setKeys(keysData.data ?? []);
      setModels(modelsData.data ?? []);
      setTransactions(txData.data ?? []);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsLoading(false);
    }
  }, [common]);

  const searchPrincipals = useCallback(async (kind: PrincipalSearchKind, query: string) => {
    if (kind === "external_identity") return [];
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query.trim())}&limit=20`,
      );
      return response.data ?? [];
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
      return [];
    }
  }, [common]);

  async function handleCreateKey(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);
    setLastCreatedKey(null);

    const cleanAlias = keyAlias.trim();
    const selectedModels = keySelectedModels
      .map((item) => item.trim())
      .filter(Boolean);

    if (!cleanAlias || selectedModels.length === 0) {
      setBanner({
        kind: "error",
        message: t("keys.messages.validation"),
      });
      return;
    }
    if (keyMode === "service_account" && !keyTeamID.trim()) {
      setBanner({
        kind: "error",
        message: "For Service Account keys, team_id is required.",
      });
      return;
    }

    setIsCreatingKey(true);
    try {
      const created = await requestJSON<TenantKey>("/api/keys", {
        method: "POST",
        body: JSON.stringify({
          key_alias: cleanAlias,
          models: selectedModels,
          role: keyMode,
          team_id: keyTeamID.trim() || undefined,
          organization_id: keyOrgID.trim() || undefined,
        }),
      });

      setLastCreatedKey(created.key ?? null);
      setKeySelectedModels([]);
      await loadDashboard();
      setBanner({
        kind: "success",
        message: t("keys.messages.success"),
      });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsCreatingKey(false);
    }
  }

  async function handleCreateModel(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);

    if (
      !modelName.trim() ||
      !provider.trim() ||
      !providerModel.trim() ||
      !apiBase.trim() ||
      !apiKey.trim()
    ) {
      setBanner({
        kind: "error",
        message: t("models.messages.validation"),
      });
      return;
    }

    setIsCreatingModel(true);
    try {
      await requestJSON("/api/models", {
        method: "POST",
        body: JSON.stringify({
          model_name: modelName.trim(),
          provider: provider.trim(),
          model: providerModel.trim(),
          api_base: apiBase.trim(),
          api_key: apiKey.trim(),
          rpm: toNumberOrUndefined(rpm),
          tpm: toNumberOrUndefined(tpm),
          weight: toNumberOrUndefined(weight),
        }),
      });

      await loadDashboard();
      setBanner({
        kind: "success",
        message: t("models.messages.success"),
      });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, common("requestError")),
      });
    } finally {
      setIsCreatingModel(false);
    }
  }

  const isServiceAccountMode = keyMode === "service_account";

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <Badge variant={dashboardHealth.variant}>{dashboardHealth.label}</Badge>
            <Badge variant="outline" dir="ltr" className="font-mono">
              {tenantSlug}.localhost
            </Badge>
          </div>
          <h1 className="text-2xl font-semibold tracking-normal">{t("title")}</h1>
          <p className="mt-1 max-w-2xl text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2 sm:justify-end">
          <Badge variant="secondary" className="max-w-full justify-start">
            <span className="shrink-0">{t("currentUser")}:</span>
            <span dir="ltr" className="min-w-0 truncate font-mono">
              {userIdentity}
            </span>
          </Badge>
          <Button
            type="button"
            variant="outline"
            onClick={() => void loadDashboard()}
            disabled={isLoading}
          >
            <RefreshCcw aria-hidden="true" />
            {isLoading ? t("refreshing") : common("refresh")}
          </Button>
        </div>
      </header>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "success"}>
          <AlertTitle>{banner.kind === "error" ? t("messages.errorTitle") : t("messages.successTitle")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      {lastCreatedKey ? (
        <Alert variant="info">
          <AlertTitle>{t("keyCreatedTitle")}</AlertTitle>
          <AlertDescription>
            <code dir="ltr" className="mt-2 block overflow-x-auto rounded-md border bg-muted px-3 py-2 text-xs font-mono">
              {lastCreatedKey}
            </code>
          </AlertDescription>
        </Alert>
      ) : null}

      <section aria-label="Tenant summary" className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <SummaryCard
          icon={WalletCards}
          title={t("stats.balance")}
          value={wallet ? formatIRR(wallet.balance_irr) : common("none")}
          subtitle={t("stats.balanceHint")}
          tone={wallet && wallet.balance_irr > 0 ? "success" : "warning"}
        />
        <SummaryCard
          icon={Activity}
          title={t("stats.pending")}
          value={wallet ? formatIRR(wallet.pending_irr) : common("none")}
          subtitle={t("stats.pendingHint")}
          tone={wallet && wallet.pending_irr > 0 ? "warning" : "muted"}
        />
        <SummaryCard
          icon={KeyRound}
          title={t("stats.keys")}
          value={numberFormatter.format(keys.length)}
          subtitle={t("stats.keysHint")}
          tone={keys.length > 0 ? "success" : "warning"}
        />
        <SummaryCard
          icon={Bot}
          title={t("models.title")}
          value={numberFormatter.format(models.length)}
          subtitle={t("models.description")}
          tone={models.length > 0 ? "success" : "warning"}
        />
      </section>

      <Tabs defaultValue="overview" className="gap-4">
        <TabsList variant="line" className="w-full justify-start overflow-x-auto border-b border-border">
          <TabsTrigger value="overview">{t("title")}</TabsTrigger>
          <TabsTrigger value="keys">{t("keys.title")}</TabsTrigger>
          <TabsTrigger value="models">{t("models.title")}</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="flex flex-col gap-6">
          <section className="grid gap-4 xl:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]">
            <DashboardPanel>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <ShieldCheck className="size-4 text-primary" aria-hidden="true" />
                  {t("title")}
                </CardTitle>
                <CardDescription>{t("subtitle")}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-3">
                <ReadinessItem
                  label={t("stats.balance")}
                  value={wallet ? formatIRR(wallet.balance_irr) : common("none")}
                  state={wallet && wallet.balance_irr > 0 ? "success" : "warning"}
                  statusLabel={wallet && wallet.balance_irr > 0 ? t("messages.successTitle") : common("notSet")}
                />
                <ReadinessItem
                  label={t("keys.title")}
                  value={numberFormatter.format(keys.length)}
                  state={keys.length > 0 ? "success" : "warning"}
                  statusLabel={keys.length > 0 ? t("messages.successTitle") : common("notSet")}
                />
                <ReadinessItem
                  label={t("models.title")}
                  value={numberFormatter.format(models.length)}
                  state={models.length > 0 ? "success" : "warning"}
                  statusLabel={models.length > 0 ? t("messages.successTitle") : common("notSet")}
                />
              </CardContent>
            </DashboardPanel>

            <TransactionsTable transactions={transactions} t={t} common={common} labels={labels} />
          </section>
        </TabsContent>

        <TabsContent value="keys" className="grid gap-4 xl:grid-cols-[minmax(320px,0.8fr)_minmax(0,1.2fr)]">
          <DashboardPanel id="keys">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Plus className="size-4 text-primary" aria-hidden="true" />
                {t("keys.create")}
              </CardTitle>
              <CardDescription>{t("keys.description")}</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              <form className="grid gap-4" onSubmit={handleCreateKey}>
                <div className="grid gap-2">
                  <Label htmlFor="keyAlias">{t("keys.name")}</Label>
                  <Input
                    id="keyAlias"
                    value={keyAlias}
                    onChange={(event) => setKeyAlias(event.target.value)}
                    placeholder={t("keys.namePlaceholder")}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="keyModels">{t("keys.models")}</Label>
                  <Input
                    id="keyModels"
                    value={keyModelQuery}
                    onChange={(event) => setKeyModelQuery(event.target.value)}
                    placeholder={t("keys.modelsPlaceholder")}
                  />
                  <div className="flex max-h-56 min-h-24 flex-wrap content-start gap-2 overflow-auto rounded-md border border-border bg-muted/30 p-2">
                    {isLoadingAllowedModels ? (
                      <p className="text-xs text-muted-foreground">{common("loading")}</p>
                    ) : filteredAllowedModels.length === 0 ? (
                      <p className="text-xs text-muted-foreground">{common("none")}</p>
                    ) : (
                      filteredAllowedModels.map((modelID) => {
                        const isActive = keySelectedModels.includes(modelID);
                        return (
                          <Button
                            key={modelID}
                            type="button"
                            size="sm"
                            variant={isActive ? "default" : "outline"}
                            onClick={() => {
                              if (isActive) {
                                setKeySelectedModels((prev) => prev.filter((item) => item !== modelID));
                                return;
                              }
                              setKeySelectedModels((prev) => Array.from(new Set([...prev, modelID])));
                            }}
                            className="max-w-full font-mono text-xs"
                            dir="ltr"
                          >
                            {modelID}
                          </Button>
                        );
                      })
                    )}
                  </div>
                </div>
                <div className="flex items-center justify-between gap-4 rounded-md border border-border bg-muted/30 p-3">
                  <div className="flex flex-col gap-1">
                    <Label htmlFor="keyMode" className="cursor-pointer">
                      {labels.keyType}
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      {isServiceAccountMode ? labels.serviceAccount : labels.virtualKey}
                    </p>
                  </div>
                  <Switch
                    id="keyMode"
                    checked={isServiceAccountMode}
                    onCheckedChange={(checked) => {
                      const nextMode = checked ? "service_account" : "virtual_key";
                      setKeyMode(nextMode);
                      void loadAllowedModelsForScope(nextMode, keyTeamID);
                    }}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="keyTeamID">{labels.teamID}</Label>
                  <PrincipalEntitySelect
                    id="keyTeamID"
                    selected={
                      keyTeamID.trim()
                        ? {
                            type: "team",
                            principal_id: keyTeamID.trim(),
                            display_name: keyTeamID.trim(),
                          }
                        : null
                    }
                    onSelectedChange={(selection) => {
                      const value = selection?.type === "team" ? (selection.principal_id ?? "") : "";
                      setKeyTeamID(value);
                      if (keyMode === "service_account") {
                        void loadAllowedModelsForScope("service_account", value);
                      }
                    }}
                    onSearch={searchPrincipals}
                    labels={principalSelectorLabels}
                    kinds={["team"]}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="keyOrgID">{labels.organizationID}</Label>
                  <Input
                    id="keyOrgID"
                    value={keyOrgID}
                    onChange={(event) => setKeyOrgID(event.target.value)}
                    placeholder="UUID organization_id"
                  />
                </div>

                <Button type="submit" disabled={isCreatingKey} className="w-fit">
                  <KeyRound aria-hidden="true" />
                  {isCreatingKey ? t("keys.creating") : t("keys.create")}
                </Button>
              </form>
            </CardContent>
          </DashboardPanel>

          <KeyInventoryTable keys={keys} t={t} common={common} labels={labels} />
        </TabsContent>

        <TabsContent value="models" className="grid gap-4 xl:grid-cols-[minmax(320px,0.8fr)_minmax(0,1.2fr)]">
          <DashboardPanel>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Server className="size-4 text-primary" aria-hidden="true" />
                {t("models.create")}
              </CardTitle>
              <CardDescription>{t("models.description")}</CardDescription>
            </CardHeader>
            <CardContent>
              <form className="grid gap-3 sm:grid-cols-2" onSubmit={handleCreateModel}>
                <InputField id="modelName" label={t("models.displayName")} value={modelName} onChange={setModelName} placeholder="gpt-4o-mini" />
                <InputField id="provider" label={t("models.provider")} value={provider} onChange={setProvider} placeholder="openai" />
                <InputField id="providerModel" label={t("models.providerModel")} value={providerModel} onChange={setProviderModel} placeholder="gpt-4o-mini" />
                <InputField id="apiBase" label={t("models.apiBase")} value={apiBase} onChange={setApiBase} placeholder="http://provider.example/v1" />
                <InputField id="apiKey" label={t("models.apiKey")} value={apiKey} onChange={setApiKey} placeholder="sk-provider" />
                <InputField id="rpm" label={t("models.rpm")} value={rpm} onChange={setRPM} placeholder="250" />
                <InputField id="tpm" label={t("models.tpm")} value={tpm} onChange={setTPM} placeholder="150000" />
                <InputField id="weight" label={t("models.weight")} value={weight} onChange={setWeight} placeholder="1" />

                <div className="sm:col-span-2">
                  <Button type="submit" disabled={isCreatingModel}>
                    <Plus aria-hidden="true" />
                    {isCreatingModel ? t("models.creating") : t("models.create")}
                  </Button>
                </div>
              </form>
            </CardContent>
          </DashboardPanel>

          <ModelInventoryTable models={models} t={t} common={common} labels={labels} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function SummaryCard(props: {
  title: string;
  value: string;
  subtitle: string;
  icon: LucideIcon;
  tone: "success" | "warning" | "muted";
}) {
  const Icon = props.icon;
  const toneClassName = {
    success: "bg-success/10 text-success",
    warning: "bg-warning/10 text-warning",
    muted: "bg-muted text-muted-foreground",
  }[props.tone];

  return (
    <DashboardPanel>
      <CardHeader className="gap-3 p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <CardDescription className="truncate">{props.title}</CardDescription>
            <CardTitle className="mt-2 text-2xl">{props.value}</CardTitle>
          </div>
          <span className={`flex size-9 shrink-0 items-center justify-center rounded-lg ${toneClassName}`}>
            <Icon className="size-4" aria-hidden="true" />
          </span>
        </div>
        <CardDescription>{props.subtitle}</CardDescription>
      </CardHeader>
    </DashboardPanel>
  );
}

function DashboardPanel(props: ComponentProps<typeof Card>) {
  return <Card {...props} className={`rounded-lg ${props.className ?? ""}`} />;
}

function InputField(props: { id: string; label: string; value: string; onChange: (value: string) => void; placeholder: string }) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input
        id={props.id}
        value={props.value}
        onChange={(event) => props.onChange(event.target.value)}
        placeholder={props.placeholder}
        dir={isTechnicalInput(props.id) ? "ltr" : undefined}
        className={isTechnicalInput(props.id) ? "font-mono" : undefined}
      />
    </div>
  );
}

function ReadinessItem(props: { label: string; value: string; state: "success" | "warning"; statusLabel: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-border bg-background px-3 py-2">
      <div className="min-w-0">
        <p className="text-sm font-medium">{props.label}</p>
        <p className="mt-0.5 truncate text-xs text-muted-foreground">{props.value}</p>
      </div>
      <Badge variant={props.state}>{props.statusLabel}</Badge>
    </div>
  );
}

function TransactionsTable(props: {
  transactions: WalletTransaction[];
  t: ReturnType<typeof useTranslations>;
  common: ReturnType<typeof useTranslations>;
  labels: DashboardLabels;
}) {
  return (
    <DashboardPanel>
      <CardHeader>
        <CardTitle className="text-base">{props.t("transactions.title")}</CardTitle>
        <CardDescription>{props.t("transactions.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-start">{props.t("transactions.columns.id")}</TableHead>
              <TableHead className="text-start">{props.t("transactions.columns.type")}</TableHead>
              <TableHead className="text-start">{props.t("transactions.columns.status")}</TableHead>
              <TableHead className="text-end">{props.t("transactions.columns.amount")}</TableHead>
              <TableHead className="text-start">{props.t("transactions.columns.createdAt")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.transactions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="py-8 text-center text-muted-foreground">
                  {props.t("transactions.empty")}
                </TableCell>
              </TableRow>
            ) : (
              props.transactions.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <CopyableCode value={item.id} labels={props.labels} className="max-w-48" />
                  </TableCell>
                  <TableCell>{item.type || props.common("none")}</TableCell>
                  <TableCell>
                    <TransactionStatusBadge status={item.status} />
                  </TableCell>
                  <TableCell className="text-end tabular-nums">{formatIRR(item.amount_irr)}</TableCell>
                  <TableCell>{formatDateTime(item.created_at)}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </DashboardPanel>
  );
}

function KeyInventoryTable(props: {
  keys: TenantKey[];
  t: ReturnType<typeof useTranslations>;
  common: ReturnType<typeof useTranslations>;
  labels: DashboardLabels;
}) {
  return (
    <DashboardPanel>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <KeyRound className="size-4 text-primary" aria-hidden="true" />
          {props.t("keys.title")}
        </CardTitle>
        <CardDescription>{props.t("keys.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-start">{props.t("keys.table.name")}</TableHead>
              <TableHead className="text-start">{props.t("keys.table.role")}</TableHead>
              <TableHead className="text-start">{props.labels.scope}</TableHead>
              <TableHead className="text-start">{props.t("keys.table.models")}</TableHead>
              <TableHead className="text-start">{props.t("keys.table.key")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.keys.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="py-8 text-center text-muted-foreground">
                  {props.t("keys.table.empty")}
                </TableCell>
              </TableRow>
            ) : (
              props.keys.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-medium">{item.key_alias}</TableCell>
                  <TableCell>
                    <Badge variant={item.blocked ? "destructive" : "secondary"}>{item.role}</Badge>
                  </TableCell>
                  <TableCell>
                    <CopyableCode value={item.team_id ? `team:${item.team_id}` : "tenant"} labels={props.labels} className="max-w-48" />
                  </TableCell>
                  <TableCell>
                    <ModelListValue models={item.models ?? []} emptyLabel={props.common("none")} />
                  </TableCell>
                  <TableCell>
                    {item.key_name ? (
                      <CopyableCode value={item.key_name} labels={props.labels} className="max-w-48" />
                    ) : (
                      <span className="text-muted-foreground">{props.t("keys.table.hidden")}</span>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </DashboardPanel>
  );
}

function ModelInventoryTable(props: {
  models: TenantModel[];
  t: ReturnType<typeof useTranslations>;
  common: ReturnType<typeof useTranslations>;
  labels: DashboardLabels;
}) {
  return (
    <DashboardPanel>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Bot className="size-4 text-primary" aria-hidden="true" />
          {props.t("models.title")}
        </CardTitle>
        <CardDescription>{props.t("models.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-start">{props.t("models.displayName")}</TableHead>
              <TableHead className="text-start">{props.t("models.provider")}</TableHead>
              <TableHead className="text-start">{props.labels.aliases}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.models.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} className="py-8 text-center text-muted-foreground">
                  {props.common("none")}
                </TableCell>
              </TableRow>
            ) : (
              props.models.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <CopyableCode value={item.id} labels={props.labels} className="max-w-64" />
                  </TableCell>
                  <TableCell>{item.provider ? <Badge variant="outline">{item.provider}</Badge> : props.common("none")}</TableCell>
                  <TableCell>
                    <ModelListValue models={item.aliases ?? []} emptyLabel={props.common("none")} />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </DashboardPanel>
  );
}

function CopyableCode(props: { value: string; labels: DashboardLabels; className?: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(props.value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      setCopied(false);
    }
  }

  return (
    <span className="flex min-w-0 items-center gap-2">
      <code dir="ltr" className={`block min-w-0 truncate font-mono text-xs ${props.className ?? ""}`}>
        {props.value}
      </code>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="h-7 w-7 min-w-0 shrink-0 rounded-md"
        aria-label={copied ? props.labels.copied : props.labels.copyValue}
        onClick={() => void handleCopy()}
      >
        {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
      </Button>
    </span>
  );
}

function TransactionStatusBadge({ status }: { status: string }) {
  const normalized = status.toLowerCase();
  if (["success", "succeeded", "paid", "completed"].includes(normalized)) {
    return <Badge variant="success">{status}</Badge>;
  }
  if (["failed", "error", "rejected", "canceled", "cancelled"].includes(normalized)) {
    return <Badge variant="destructive">{status}</Badge>;
  }
  if (["pending", "processing"].includes(normalized)) {
    return <Badge variant="warning">{status}</Badge>;
  }
  return <Badge variant="outline">{status}</Badge>;
}

function ModelListValue({ models, emptyLabel }: { models: string[]; emptyLabel: string }) {
  if (models.length === 0) return <span className="text-muted-foreground">{emptyLabel}</span>;
  return (
    <div className="flex max-w-sm flex-wrap gap-1.5">
      {models.slice(0, 3).map((model) => (
        <Badge key={model} variant="outline" dir="ltr" className="max-w-full truncate font-mono">
          {model}
        </Badge>
      ))}
      {models.length > 3 ? <Badge variant="secondary">+{numberFormatter.format(models.length - 3)}</Badge> : null}
    </div>
  );
}

function isTechnicalInput(id: string): boolean {
  return ["modelName", "provider", "providerModel", "apiBase", "apiKey", "rpm", "tpm", "weight"].includes(id);
}

function buildDashboardLabels(t: ReturnType<typeof useTranslations>): DashboardLabels {
  const isEnglish = t("title").trim().toLowerCase().includes("dashboard");
  return {
    copied: isEnglish ? "Copied" : "کپی شد",
    copyValue: isEnglish ? "Copy value" : "کپی مقدار",
    keyType: isEnglish ? "Key type" : "نوع کلید",
    virtualKey: isEnglish ? "Virtual key" : "کلید مجازی",
    serviceAccount: isEnglish ? "Service account (team)" : "اکانت سرویس (تیم)",
    teamID: isEnglish ? "Team ID (required for service account)" : "شناسه تیم (برای اکانت سرویس الزامی است)",
    organizationID: isEnglish ? "Organization ID (optional)" : "شناسه سازمان (اختیاری)",
    scope: isEnglish ? "Scope" : "دامنه",
    aliases: isEnglish ? "Aliases" : "نام‌های مستعار",
  };
}

function buildDashboardPrincipalSelectorLabels(
  t: ReturnType<typeof useTranslations>,
  ux: ReturnType<typeof useTranslations>,
): PrincipalEntitySelectLabels {
  const isEnglish = t("principal.user").trim().toLowerCase() === "user";
  return {
    users: isEnglish ? "Users" : "کاربران",
    teams: isEnglish ? "Teams" : "تیم‌ها",
    orgUnits: isEnglish ? "Org units" : "واحدهای سازمانی",
    virtualAccounts: isEnglish ? "Virtual accounts" : "اکانت‌های مجازی",
    externalIdentities: isEnglish ? "External identities" : "هویت‌های خارجی",
    everyone: t("principal.everyone"),
    searchPlaceholder: t("principal.searchPlaceholder"),
    triggerPlaceholder: t("principal.searchPlaceholder"),
    loading: ux("actions.loading"),
    empty: ux("effectiveModels.empty"),
  };
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, {
    ...options,
    headers,
    cache: "no-store",
  });

  const text = await response.text();
  const data = parseResponsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessageFromPayload(data) ?? `خطای درخواست (${response.status})`);
  }
  return data as T;
}

function parseResponsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessageFromPayload(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) return error.message;
  return fallback;
}

function toNumberOrUndefined(value: string): number | undefined {
  const raw = value.trim();
  if (!raw) return undefined;
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function formatIRR(value: number): string {
  return `${numberFormatter.format(Math.round(value))} ریال`;
}

function formatDateTime(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  return new Intl.DateTimeFormat("fa-IR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(timestamp));
}

