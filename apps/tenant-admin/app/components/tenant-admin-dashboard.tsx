"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
  status: string;
  created_at: string;
};

type WalletTransactionListResponse = {
  data?: WalletTransaction[];
};

type KeyListResponse = {
  data?: TenantKey[];
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type TenantAdminDashboardProps = {
  initialHost: string;
  initialWallet: Wallet | null;
  initialKeys: TenantKey[];
  initialModels: TenantModel[];
  initialError: string | null;
  initialTransactions: WalletTransaction[];
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

  const [keys, setKeys] = useState<TenantKey[]>(props.initialKeys);
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
  const [isCreatingModel, setIsCreatingModel] = useState(false);

  const [keyAlias, setKeyAlias] = useState("tenant-dashboard-key");
  const [keyModelsInput, setKeyModelsInput] = useState("gpt-4o-mini");
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
    return slug || "default";
  }, [props.initialHost]);

  const userIdentity =
    props.currentUser.email ?? props.currentUser.mobile ?? props.currentUser.id;

  const loadDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const [walletData, keysData, txData] = await Promise.all([
        requestJSON<Wallet>("/api/wallet"),
        requestJSON<KeyListResponse>("/api/keys"),
        requestJSON<WalletTransactionListResponse>("/api/wallet/transactions?limit=20"),
      ]);

      setWallet(walletData);
      setKeys(keysData.data ?? []);
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

  async function handleCreateKey(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);
    setLastCreatedKey(null);

    const cleanAlias = keyAlias.trim();
    const selectedModels = keyModelsInput
      .split(",")
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

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("subtitle")}</CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{tenantSlug}.localhost</Badge>
            <Badge variant="secondary">
              {t("currentUser")}: {userIdentity}
            </Badge>
            <Button type="button" variant="outline" onClick={() => void loadDashboard()} disabled={isLoading}>
              {isLoading ? t("refreshing") : common("refresh")}
            </Button>
          </div>
        </CardHeader>
      </Card>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("messages.errorTitle") : t("messages.successTitle")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      {lastCreatedKey ? (
        <Alert className="border-primary/35 bg-primary/5">
          <AlertTitle>{t("keyCreatedTitle")}</AlertTitle>
          <AlertDescription>
            <code className="mt-2 block overflow-x-auto rounded bg-muted px-3 py-2 text-xs font-mono">
              {lastCreatedKey}
            </code>
          </AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-4 md:grid-cols-3">
        <SummaryCard
          title={t("stats.balance")}
          value={wallet ? formatIRR(wallet.balance_irr) : common("none")}
          subtitle={t("stats.balanceHint")}
        />
        <SummaryCard
          title={t("stats.pending")}
          value={wallet ? formatIRR(wallet.pending_irr) : common("none")}
          subtitle={t("stats.pendingHint")}
        />
        <SummaryCard
          title={t("stats.keys")}
          value={numberFormatter.format(keys.length)}
          subtitle={t("stats.keysHint")}
        />
      </section>

      <section className="grid gap-6 xl:grid-cols-2">
        <Card id="keys">
          <CardHeader>
            <CardTitle>{t("keys.title")}</CardTitle>
            <CardDescription>{t("keys.description")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <form className="grid gap-3" onSubmit={handleCreateKey}>
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
                  value={keyModelsInput}
                  onChange={(event) => setKeyModelsInput(event.target.value)}
                  placeholder={t("keys.modelsPlaceholder")}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="keyMode">Key type</Label>
                <select
                  id="keyMode"
                  value={keyMode}
                  onChange={(event) =>
                    setKeyMode(
                      event.target.value === "service_account"
                        ? "service_account"
                        : "virtual_key",
                    )
                  }
                  className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="virtual_key">Virtual key</option>
                  <option value="service_account">Service account (team)</option>
                </select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="keyTeamID">Team ID (required for service account)</Label>
                <Input
                  id="keyTeamID"
                  value={keyTeamID}
                  onChange={(event) => setKeyTeamID(event.target.value)}
                  placeholder="UUID team_id"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="keyOrgID">Organization ID (optional)</Label>
                <Input
                  id="keyOrgID"
                  value={keyOrgID}
                  onChange={(event) => setKeyOrgID(event.target.value)}
                  placeholder="UUID organization_id"
                />
              </div>

              <Button type="submit" disabled={isCreatingKey} className="w-fit">
                {isCreatingKey ? t("keys.creating") : t("keys.create")}
              </Button>
            </form>

            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("keys.table.name")}</TableHead>
                  <TableHead>{t("keys.table.role")}</TableHead>
                  <TableHead>Scope</TableHead>
                  <TableHead>{t("keys.table.models")}</TableHead>
                  <TableHead>{t("keys.table.key")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-muted-foreground">
                      {t("keys.table.empty")}
                    </TableCell>
                  </TableRow>
                ) : (
                  keys.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{item.key_alias}</TableCell>
                      <TableCell>{item.role}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {item.team_id ? `team:${item.team_id}` : "tenant"}
                      </TableCell>
                      <TableCell>
                        {(item.models ?? []).length > 0
                          ? item.models.join("، ")
                          : common("none")}
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {item.key_name ?? t("keys.table.hidden")}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("models.title")}</CardTitle>
            <CardDescription>{t("models.description")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
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
                  {isCreatingModel ? t("models.creating") : t("models.create")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </section>
    </div>
  );
}

function SummaryCard(props: { title: string; value: string; subtitle: string }) {
  return (
    <Card>
      <CardHeader className="gap-3">
        <CardDescription>{props.title}</CardDescription>
        <CardTitle className="text-2xl">{props.value}</CardTitle>
        <CardDescription>{props.subtitle}</CardDescription>
      </CardHeader>
    </Card>
  );
}

function InputField(props: { id: string; label: string; value: string; onChange: (value: string) => void; placeholder: string }) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={props.id}>{props.label}</Label>
      <Input id={props.id} value={props.value} onChange={(event) => props.onChange(event.target.value)} placeholder={props.placeholder} />
    </div>
  );
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
