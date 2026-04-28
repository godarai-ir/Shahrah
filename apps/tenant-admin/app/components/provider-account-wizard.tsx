"use client";

import Image from "next/image";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { CheckCircle2, Circle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type ProviderAuthMode =
  | "api_key"
  | "azure_entra_cert"
  | "aws_sigv4"
  | "google_service_account"
  | "no_credential";

type ModelType =
  | "chat"
  | "completion"
  | "embedding"
  | "realtime"
  | "text_to_speech"
  | "audio_transcription"
  | "image_generation"
  | "moderation"
  | "rerank"
  | "search";

type PrincipalType = "user" | "team" | "everyone";
type WizardStep = 1 | 2 | 3;

type AccessPrincipal = {
  type: PrincipalType;
  principal_id?: string;
  display_name?: string;
};

type AccessPrincipalCandidate = {
  type: PrincipalType;
  principal_id: string;
  display_name: string;
  secondary?: string;
};

type ProviderCatalogProvider = {
  provider: string;
  display_name: string;
  enabled: boolean;
  auth_modes: ProviderAuthMode[];
  supports_bring_your_own_key: boolean;
  supports_virtual_keys: boolean;
};

type ProviderCatalogModel = {
  provider: string;
  model: string;
  display_name: string;
  types: ModelType[];
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

type TenantProviderAccount = {
  id: string;
  provider: string;
  name: string;
  auth_mode: ProviderAuthMode;
  has_credential: boolean;
  allows_virtual_keys_only: boolean;
};

type TenantProviderAccountModel = {
  model: string;
  enabled: boolean;
  types: ModelType[];
  pricing_mode: "public" | "private";
  private_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
  settings?: Record<string, unknown>;
};

type AccountDetailsResponse = {
  account: TenantProviderAccount;
  models: TenantProviderAccountModel[];
  access: {
    managers: AccessPrincipal[];
    users: AccessPrincipal[];
  };
};

type ProvidersResponse = { data?: ProviderCatalogProvider[] };
type ProviderModelsResponse = { data?: ProviderCatalogModel[] };
type AccountsResponse = { data?: TenantProviderAccount[] };
type AccessPrincipalResponse = { data?: AccessPrincipalCandidate[] };

type Banner = {
  kind: "success" | "error";
  message: string;
};

type ModelDraft = {
  selected: boolean;
  types: ModelType[];
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
};

export default function ProviderAccountWizard() {
  const t = useTranslations("providerAccounts");
  const [currentStep, setCurrentStep] = useState<WizardStep>(1);

  const [isLoading, setIsLoading] = useState(false);
  const [isSavingModels, setIsSavingModels] = useState(false);
  const [isSavingAccess, setIsSavingAccess] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);
  const [modelSearchQuery, setModelSearchQuery] = useState("");

  const [selectedAccountID, setSelectedAccountID] = useState("");
  const [provider, setProvider] = useState("");
  const [name, setName] = useState("");
  const [authMode, setAuthMode] = useState<ProviderAuthMode>("api_key");
  const [allowVirtualOnly, setAllowVirtualOnly] = useState(false);

  const [apiKey, setAPIKey] = useState("");
  const [secretRef, setSecretRef] = useState("");
  const [azureTenantID, setAzureTenantID] = useState("");
  const [azureClientID, setAzureClientID] = useState("");
  const [awsAccessKeyID, setAWSAccessKeyID] = useState("");
  const [awsSecretAccessKey, setAWSSecretAccessKey] = useState("");
  const [awsRegion, setAWSRegion] = useState("us-east-1");
  const [googleServiceAccountJSON, setGoogleServiceAccountJSON] = useState("");

  const [modelDrafts, setModelDrafts] = useState<Record<string, ModelDraft>>({});
  const [managers, setManagers] = useState<AccessPrincipal[]>([]);
  const [users, setUsers] = useState<AccessPrincipal[]>([{ type: "everyone", display_name: "everyone" }]);

  const [managerKind, setManagerKind] = useState<"user" | "team">("user");
  const [managerQuery, setManagerQuery] = useState("");
  const [managerResults, setManagerResults] = useState<AccessPrincipalCandidate[]>([]);

  const [userKind, setUserKind] = useState<"user" | "team">("user");
  const [userQuery, setUserQuery] = useState("");
  const [userResults, setUserResults] = useState<AccessPrincipalCandidate[]>([]);

  const steps = [
    { step: 1 as const, title: t("step1.title") },
    { step: 2 as const, title: t("step2.title") },
    { step: 3 as const, title: t("step3.title") },
  ];

  const selectedProvider = useMemo(
    () => providers.find((item) => item.provider === provider) ?? null,
    [providers, provider],
  );
  const visibleProviders = useMemo(
    () => providers.filter((item) => item.enabled && stringsEqualFold(item.provider, "azure") === false),
    [providers],
  );

  const filteredProviderModels = useMemo(() => {
    const query = modelSearchQuery.trim().toLowerCase();
    if (!query) return providerModels;
    return providerModels.filter((model) => {
      const display = (model.display_name || "").toLowerCase();
      const key = model.model.toLowerCase();
      return display.includes(query) || key.includes(query);
    });
  }, [providerModels, modelSearchQuery]);

  const hasSelectedProvider = Boolean(provider);
  const canGoStep3 = Boolean(selectedAccountID);

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setBanner(null);
    try {
      const [providerData, accountData] = await Promise.all([
        requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
        requestJSON<AccountsResponse>("/api/tenant-provider-accounts"),
      ]);

      const providerItems = providerData.data ?? [];
      const accountItems = accountData.data ?? [];
      setProviders(providerItems);
      setAccounts(accountItems);

      if (providerItems.length > 0) {
        setProvider((prev) => prev || providerItems[0].provider);
        setName((prev) => prev || providerItems[0].provider);
        setAuthMode((prev) => prev || providerItems[0].auth_modes?.[0] || "api_key");
      }
      if (accountItems.length > 0) {
        setSelectedAccountID((prev) => prev || accountItems[0].id);
      }
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.initialLoadFailed")) });
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadInitial();
  }, [loadInitial]);

  const handleLoadProviderModels = useCallback(
    async (nextProvider: string) => {
      const normalized = nextProvider.trim().toLowerCase();
      if (!normalized) {
        setProviderModels([]);
        setModelDrafts({});
        return;
      }
      try {
        const response = await requestJSON<ProviderModelsResponse>(
          `/api/provider-catalog/providers/${encodeURIComponent(normalized)}/models`,
        );
        const items = response.data ?? [];
        setProviderModels(items);
        setModelDrafts((prev) => buildDefaultDrafts(items, prev));
      } catch (error) {
        setProviderModels([]);
        setModelDrafts({});
        setBanner({
          kind: "error",
          message: normalizeError(error, t("messages.providerModelsLoadFailed")),
        });
      }
    },
    [t],
  );

  useEffect(() => {
    if (!provider) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void handleLoadProviderModels(provider);
  }, [provider, handleLoadProviderModels]);

  function handleSelectProvider(item: ProviderCatalogProvider) {
    setProvider(item.provider);
    setAuthMode(item.auth_modes?.[0] ?? "api_key");
    if (!name.trim()) setName(item.provider);
    setCurrentStep(2);
  }

  async function handleCreateAccount() {
    setBanner(null);
    if (!provider || !name.trim()) {
      setBanner({ kind: "error", message: t("messages.validationProviderName") });
      return;
    }

    const providerConfig: Record<string, unknown> = {};
    const credentials: Record<string, unknown> = {};

    if (authMode === "api_key") {
      if (apiKey.trim()) credentials.api_key = apiKey.trim();
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    }
    if (authMode === "azure_entra_cert") {
      providerConfig.azure_tenant_id = azureTenantID.trim() || undefined;
      providerConfig.azure_client_id = azureClientID.trim() || undefined;
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    }
    if (authMode === "aws_sigv4") {
      providerConfig.aws_region = awsRegion.trim() || undefined;
      providerConfig.aws_access_key_id = awsAccessKeyID.trim() || undefined;
      providerConfig.aws_secret_access_key = awsSecretAccessKey.trim() || undefined;
    }
    if (authMode === "google_service_account") {
      providerConfig.google_service_account_json = googleServiceAccountJSON.trim() || undefined;
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    }

    const payload: Record<string, unknown> = {
      provider,
      name: name.trim(),
      auth_mode: authMode,
      allow_virtual_keys_only: allowVirtualOnly,
      provider_config: providerConfig,
    };
    if (Object.keys(credentials).length > 0) payload.credentials = credentials;

    setIsLoading(true);
    try {
      const created = await requestJSON<TenantProviderAccount>(
        "/api/tenant-provider-accounts",
        { method: "POST", body: JSON.stringify(payload) },
      );
      setSelectedAccountID(created.id);
      await loadInitial();
      await handleLoadProviderModels(created.provider);
      setCurrentStep(3);
      setBanner({ kind: "success", message: t("messages.accountCreateSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accountCreateFailed")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleLoadAccount(accountID: string) {
    if (!accountID) return;
    setIsLoading(true);
    setBanner(null);
    try {
      const details = await requestJSON<AccountDetailsResponse>(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}`,
      );
      setProvider(details.account.provider);
      setName(details.account.name);
      setAuthMode(details.account.auth_mode);
      setAllowVirtualOnly(details.account.allows_virtual_keys_only);
      setManagers(details.access.managers ?? []);
      setUsers(details.access.users ?? []);
      await handleLoadProviderModels(details.account.provider);
      setModelDrafts((prev) => hydrateDraftsFromAccount(details.models ?? [], prev));
      setBanner({ kind: "success", message: t("messages.accountLoadSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accountLoadFailed")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleSaveModels() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }

    const payloadModels = Object.entries(modelDrafts)
      .filter(([, draft]) => draft.selected)
      .map(([model, draft]) => ({
        model,
        enabled: true,
        types: draft.types,
        pricing_mode: draft.pricingMode,
        private_pricing:
          draft.pricingMode === "private"
            ? {
                input_usd_per_1m: toNumberOrUndefined(draft.inputPrice),
                output_usd_per_1m: toNumberOrUndefined(draft.outputPrice),
              }
            : undefined,
        settings: {},
      }));

    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: t("messages.selectAtLeastOneModel") });
      return;
    }

    setIsSavingModels(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/models`,
        { method: "PUT", body: JSON.stringify({ models: payloadModels }) },
      );
      setBanner({ kind: "success", message: t("messages.modelsSaveSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.modelsSaveFailed")) });
    } finally {
      setIsSavingModels(false);
    }
  }

  async function handleSaveAccess() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }
    if (managers.length === 0) {
      setBanner({ kind: "error", message: t("messages.managerRequired") });
      return;
    }

    setIsSavingAccess(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/access`,
        { method: "PUT", body: JSON.stringify({ access: { managers, users } }) },
      );
      setBanner({ kind: "success", message: t("messages.accessSaveSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accessSaveFailed")) });
    } finally {
      setIsSavingAccess(false);
    }
  }

  async function handleTestCredentials() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }
    setIsTesting(true);
    try {
      const result = await requestJSON<{ success: boolean; message?: string }>(
        `/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/test-credentials`,
        { method: "POST", body: JSON.stringify({}) },
      );
      setBanner({
        kind: result.success ? "success" : "error",
        message: result.message ?? (result.success ? t("messages.credentialValid") : t("messages.credentialInvalid")),
      });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.credentialTestFailed")) });
    } finally {
      setIsTesting(false);
    }
  }

  async function searchManagerPrincipals() {
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${managerKind}&q=${encodeURIComponent(managerQuery.trim())}&limit=20`,
      );
      setManagerResults(response.data ?? []);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.managerSearchFailed")) });
    }
  }

  async function searchUserPrincipals() {
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${userKind}&q=${encodeURIComponent(userQuery.trim())}&limit=20`,
      );
      setUserResults(response.data ?? []);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.userSearchFailed")) });
    }
  }

  return (
    <Card id="tenant-provider-accounts" className="border-border/70 shadow-sm">
      <CardHeader className="rounded-t-xl bg-gradient-to-l from-background to-muted/30">
        <CardTitle className="text-2xl">{t("title")}</CardTitle>
        <CardDescription className="text-base">{t("description")}</CardDescription>
      </CardHeader>

      <CardContent className="space-y-6 pt-6">
        {banner ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

        <section className="rounded-xl border bg-muted/20 p-4">
          <div className="grid gap-3 md:grid-cols-3">
            {steps.map((item) => {
              const isActive = currentStep === item.step;
              const isDone = currentStep > item.step;
              return (
                <button
                  key={item.step}
                  type="button"
                  className={`rounded-lg border p-3 text-start transition ${
                    isActive
                      ? "border-primary bg-primary/10"
                      : isDone
                        ? "border-primary/40 bg-primary/5"
                        : "border-border bg-background"
                  }`}
                  onClick={() => setCurrentStep(item.step)}
                >
                  <div className="mb-2 flex items-center gap-2">
                    {isDone ? <CheckCircle2 className="h-4 w-4 text-primary" /> : <Circle className="h-4 w-4 text-muted-foreground" />}
                    <span className="text-xs text-muted-foreground">{t("wizard.stepLabel", { number: item.step })}</span>
                  </div>
                  <div className="font-semibold">{item.title}</div>
                </button>
              );
            })}
          </div>
        </section>

        {currentStep === 1 ? (
          <section className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold">{t("step1.title")}</h3>
              <p className="text-sm text-muted-foreground">{t("step1.description")}</p>
            </div>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {visibleProviders.map((item) => {
                const isSelected = provider === item.provider;
                return (
                  <button
                    key={item.provider}
                    type="button"
                    className={`group rounded-xl border p-3 text-start transition ${
                      isSelected ? "border-primary bg-primary/10 shadow-sm" : "border-border hover:border-primary/40 hover:bg-muted/30"
                    }`}
                    onClick={() => handleSelectProvider(item)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg border bg-background">
                        <Image src={getProviderLogoPath(item.provider)} alt={item.display_name} width={24} height={24} className="object-contain" />
                      </div>
                      {isSelected ? <Badge variant="secondary">{t("successTitle")}</Badge> : null}
                    </div>
                    <div className="mt-3">
                      <div className="font-semibold">{item.display_name}</div>
                      <div className="text-xs text-muted-foreground">{item.provider}</div>
                    </div>
                  </button>
                );
              })}
            </div>
          </section>
        ) : null}

        {currentStep === 2 ? (
          <section className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold">{t("step2.title")}</h3>
              <p className="text-sm text-muted-foreground">{t("step2.connectionDescription")}</p>
            </div>

            <Card className="border-border/70">
              <CardHeader><CardTitle className="text-base">{t("fields.accountName")}</CardTitle></CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-3">
                <div className="grid gap-2">
                  <Label htmlFor="providerAccountName">{t("fields.accountName")}</Label>
                  <Input id="providerAccountName" value={name} onChange={(event) => setName(event.target.value)} placeholder="azure-openai-main" />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="authMode">{t("fields.authMode")}</Label>
                  <select
                    id="authMode"
                    className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                    value={authMode}
                    onChange={(event) => setAuthMode(event.target.value as ProviderAuthMode)}
                  >
                    {(selectedProvider?.auth_modes ?? ["api_key"]).map((mode) => (
                      <option key={mode} value={mode}>{mode}</option>
                    ))}
                  </select>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="allowVirtual">{t("fields.virtualKeysMode")}</Label>
                  <label className="flex h-10 items-center gap-2 rounded-md border border-input px-3 text-sm">
                    <input id="allowVirtual" type="checkbox" checked={allowVirtualOnly} onChange={(event) => setAllowVirtualOnly(event.target.checked)} />
                    {t("fields.allowVirtualOnly")}
                  </label>
                </div>
              </CardContent>
            </Card>

            <Card className="border-border/70">
              <CardHeader><CardTitle className="text-base">{t("fields.authMode")}</CardTitle></CardHeader>
              <CardContent className="space-y-4">
                {authMode === "api_key" ? (
                  <div className="grid gap-3 md:grid-cols-2">
                    <Field label={t("fields.apiKeyOptional")} id="providerAPIKey">
                      <Input id="providerAPIKey" value={apiKey} onChange={(event) => setAPIKey(event.target.value)} placeholder="sk-..." />
                    </Field>
                    <Field label={t("fields.secretRefOptional")} id="providerSecretRef">
                      <Input id="providerSecretRef" value={secretRef} onChange={(event) => setSecretRef(event.target.value)} placeholder="vault://..." />
                    </Field>
                  </div>
                ) : null}

                {authMode === "azure_entra_cert" ? (
                  <div className="grid gap-3 md:grid-cols-3">
                    <Field label={t("fields.azureTenantID")} id="azureTenantID">
                      <Input id="azureTenantID" value={azureTenantID} onChange={(event) => setAzureTenantID(event.target.value)} placeholder="tenant-id" />
                    </Field>
                    <Field label={t("fields.azureClientID")} id="azureClientID">
                      <Input id="azureClientID" value={azureClientID} onChange={(event) => setAzureClientID(event.target.value)} placeholder="client-id" />
                    </Field>
                    <Field label={t("fields.secretRefOptional")} id="providerSecretRefAzure">
                      <Input id="providerSecretRefAzure" value={secretRef} onChange={(event) => setSecretRef(event.target.value)} placeholder="vault://azure-cert" />
                    </Field>
                  </div>
                ) : null}

                {authMode === "aws_sigv4" ? (
                  <div className="grid gap-3 md:grid-cols-3">
                    <Field label={t("fields.awsAccessKeyID")} id="awsAccessKeyID">
                      <Input id="awsAccessKeyID" value={awsAccessKeyID} onChange={(event) => setAWSAccessKeyID(event.target.value)} placeholder="AKIA..." />
                    </Field>
                    <Field label={t("fields.awsSecretAccessKey")} id="awsSecretAccessKey">
                      <Input id="awsSecretAccessKey" value={awsSecretAccessKey} onChange={(event) => setAWSSecretAccessKey(event.target.value)} placeholder="secret" />
                    </Field>
                    <Field label={t("fields.awsRegion")} id="awsRegion">
                      <Input id="awsRegion" value={awsRegion} onChange={(event) => setAWSRegion(event.target.value)} placeholder="us-east-1" />
                    </Field>
                  </div>
                ) : null}

                {authMode === "google_service_account" ? (
                  <div className="grid gap-3">
                    <Field label={t("fields.googleServiceAccountJSON")} id="googleServiceAccountJSON">
                      <textarea
                        id="googleServiceAccountJSON"
                        value={googleServiceAccountJSON}
                        onChange={(event) => setGoogleServiceAccountJSON(event.target.value)}
                        placeholder='{"type":"service_account", ...}'
                        className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm"
                      />
                    </Field>
                    <Field label={t("fields.secretRefOptional")} id="providerSecretRefGoogle">
                      <Input id="providerSecretRefGoogle" value={secretRef} onChange={(event) => setSecretRef(event.target.value)} placeholder="vault://gcp-service-account" />
                    </Field>
                  </div>
                ) : null}
              </CardContent>
            </Card>

            <div className="flex justify-end">
              <Button type="button" onClick={() => void handleCreateAccount()} disabled={isLoading || !hasSelectedProvider}>
                {isLoading ? t("creating") : t("createProviderAccount")}
              </Button>
            </div>
          </section>
        ) : null}

        {currentStep === 3 ? (
          <section className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold">{t("step3.title")}</h3>
              <p className="text-sm text-muted-foreground">{t("step3.description")}</p>
            </div>

            <section className="grid gap-3 rounded-xl border bg-muted/10 p-3 md:grid-cols-2 lg:grid-cols-4">
              <Field label={t("existingAccount")} id="providerAccountSelect">
                <select id="providerAccountSelect" className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={selectedAccountID} onChange={(event) => setSelectedAccountID(event.target.value)}>
                  <option value="">{t("selectAccount")}</option>
                  {accounts.map((item) => <option key={item.id} value={item.id}>{item.name} ({item.provider})</option>)}
                </select>
              </Field>
              <div className="flex items-end"><Button type="button" variant="outline" onClick={() => void handleLoadAccount(selectedAccountID)} disabled={!selectedAccountID || isLoading}>{t("loadAccount")}</Button></div>
              <div className="flex items-end"><Button type="button" variant="outline" onClick={() => void loadInitial()} disabled={isLoading}>{isLoading ? t("refreshing") : t("refresh")}</Button></div>
              <div className="flex items-end"><Button type="button" variant="outline" onClick={() => void handleTestCredentials()} disabled={isTesting || !selectedAccountID}>{isTesting ? t("testing") : t("testCredentials")}</Button></div>
            </section>

            <div className="grid gap-6 xl:grid-cols-2">
              <Card className="border-border/70">
                <CardHeader><CardTitle className="text-base">{t("saveSelectedModels")}</CardTitle></CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <Input
                      value={modelSearchQuery}
                      onChange={(event) => setModelSearchQuery(event.target.value)}
                      placeholder={t("step3.searchPlaceholder")}
                      className="max-w-sm"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        const allSelected = filteredProviderModels.length > 0 && filteredProviderModels.every((model) => {
                          const draft = modelDrafts[model.model];
                          return Boolean(draft?.selected);
                        });
                        setModelDrafts((prev) => {
                          const next = { ...prev };
                          for (const model of filteredProviderModels) {
                            const draft = next[model.model] ?? {
                              selected: false,
                              types: model.types ?? [],
                              pricingMode: "public" as const,
                              inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
                              outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
                            };
                            next[model.model] = { ...draft, selected: !allSelected };
                          }
                          return next;
                        });
                      }}
                      disabled={filteredProviderModels.length === 0}
                    >
                      {t("step3.toggleSelectAll")}
                    </Button>
                  </div>

                  {filteredProviderModels.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("step3.empty")}</p>
                  ) : (
                    <div className="overflow-hidden rounded-lg border">
                      <div className="grid grid-cols-[48px_1fr_160px_160px] border-b bg-muted/30 px-3 py-2 text-xs font-semibold text-muted-foreground">
                        <div />
                        <div>{t("step3.columns.model")}</div>
                        <div>{t("step3.columns.inputCost")}</div>
                        <div>{t("step3.columns.outputCost")}</div>
                      </div>
                      {filteredProviderModels.map((model) => {
                        const draft = modelDrafts[model.model] ?? {
                          selected: false,
                          types: model.types ?? [],
                          pricingMode: "public" as const,
                          inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
                          outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
                        };
                        return (
                          <div key={model.model} className="grid grid-cols-[48px_1fr_160px_160px] items-center border-b px-3 py-3 last:border-b-0">
                            <div>
                              <input
                                type="checkbox"
                                checked={draft.selected}
                                onChange={(event) =>
                                  setModelDrafts((prev) => ({
                                    ...prev,
                                    [model.model]: { ...draft, selected: event.target.checked },
                                  }))
                                }
                              />
                            </div>
                            <div className="min-w-0">
                              <div className="truncate font-medium">{model.display_name || model.model}</div>
                              <div className="flex flex-wrap items-center gap-2">
                                <span className="text-xs text-muted-foreground">{model.model}</span>
                                {(model.types ?? []).map((type) => (
                                  <Badge key={type} variant="secondary">{type}</Badge>
                                ))}
                              </div>
                            </div>
                            <div className="text-sm">{formatCost(model.public_pricing?.input_usd_per_1m)}</div>
                            <div className="text-sm">{formatCost(model.public_pricing?.output_usd_per_1m)}</div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                  <Button type="button" onClick={() => void handleSaveModels()} disabled={isSavingModels || !selectedAccountID}>{isSavingModels ? t("saving") : t("saveSelectedModels")}</Button>
                </CardContent>
              </Card>

              <Card className="border-border/70">
                <CardHeader><CardTitle className="text-base">{t("step3.accessTitle")}</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <PrincipalPicker title={t("step3.managersTitle")} kind={managerKind} setKind={setManagerKind} query={managerQuery} setQuery={setManagerQuery} results={managerResults} onSearch={searchManagerPrincipals} selected={managers} setSelected={setManagers} allowEveryone={false} t={t} />
                  <PrincipalPicker title={t("step3.usersTitle")} kind={userKind} setKind={setUserKind} query={userQuery} setQuery={setUserQuery} results={userResults} onSearch={searchUserPrincipals} selected={users} setSelected={setUsers} allowEveryone t={t} />
                  <Button type="button" onClick={() => void handleSaveAccess()} disabled={isSavingAccess || !selectedAccountID}>{isSavingAccess ? t("saving") : t("saveAccessControl")}</Button>
                </CardContent>
              </Card>
            </div>
          </section>
        ) : null}

        <section className="sticky bottom-0 z-10 flex flex-wrap items-center justify-between gap-2 rounded-xl border bg-background/95 p-3 backdrop-blur">
          <Button type="button" variant="outline" onClick={() => setCurrentStep((prev) => (prev > 1 ? ((prev - 1) as WizardStep) : prev))} disabled={currentStep === 1}>
            {t("wizard.back")}
          </Button>
          {currentStep === 2 ? (
            <Button type="button" onClick={() => setCurrentStep(3)} disabled={!canGoStep3}>
              {t("wizard.next")}
            </Button>
          ) : null}
          {currentStep === 3 ? <div className="text-sm text-muted-foreground">{t("wizard.finalHint")}</div> : null}
        </section>
      </CardContent>
    </Card>
  );
}

function Field(props: { label: string; id: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={props.id}>{props.label}</Label>
      {props.children}
    </div>
  );
}

function PrincipalPicker(props: {
  title: string;
  kind: "user" | "team";
  setKind: (value: "user" | "team") => void;
  query: string;
  setQuery: (value: string) => void;
  results: AccessPrincipalCandidate[];
  onSearch: () => Promise<void>;
  selected: AccessPrincipal[];
  setSelected: (value: AccessPrincipal[]) => void;
  allowEveryone: boolean;
  t: ReturnType<typeof useTranslations>;
}) {
  function addPrincipal(item: AccessPrincipalCandidate) {
    const exists = props.selected.some(
      (current) =>
        current.type === item.type &&
        (current.principal_id ?? "") === (item.principal_id ?? ""),
    );
    if (exists) return;
    props.setSelected([
      ...props.selected,
      { type: item.type, principal_id: item.principal_id, display_name: item.display_name },
    ]);
  }

  function removePrincipal(item: AccessPrincipal) {
    props.setSelected(
      props.selected.filter(
        (current) =>
          !(
            current.type === item.type &&
            (current.principal_id ?? "") === (item.principal_id ?? "")
          ),
      ),
    );
  }

  return (
    <div className="space-y-3 rounded-lg border p-3">
      <h4 className="font-medium">{props.title}</h4>
      <div className="grid gap-2 sm:grid-cols-3">
        <select className="h-9 rounded-md border border-input bg-background px-3 text-sm" value={props.kind} onChange={(event) => props.setKind(event.target.value as "user" | "team")}>
          <option value="user">{props.t("principal.user")}</option>
          <option value="team">{props.t("principal.team")}</option>
        </select>
        <Input value={props.query} onChange={(event) => props.setQuery(event.target.value)} placeholder={props.t("principal.searchPlaceholder")} />
        <Button type="button" variant="outline" onClick={() => void props.onSearch()}>{props.t("principal.search")}</Button>
      </div>

      {props.allowEveryone ? (
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            const exists = props.selected.some((item) => item.type === "everyone");
            if (!exists) {
              props.setSelected([...props.selected, { type: "everyone", display_name: props.t("principal.everyone") }]);
            }
          }}
        >
          {props.t("principal.addEveryone")}
        </Button>
      ) : null}

      <div className="space-y-2">
        {props.results.map((item) => (
          <div key={`${item.type}-${item.principal_id}`} className="flex items-center justify-between rounded border px-2 py-1">
            <div>
              <div className="text-sm font-medium">{item.display_name}</div>
              <div className="text-xs text-muted-foreground">
                {principalTypeLabel(item.type, props.t)} {item.secondary ? `- ${item.secondary}` : ""}
              </div>
            </div>
            <Button type="button" size="sm" variant="outline" onClick={() => addPrincipal(item)}>
              {props.t("principal.add")}
            </Button>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        {props.selected.map((item) => (
          <button key={`${item.type}-${item.principal_id ?? "everyone"}`} type="button" className="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs" onClick={() => removePrincipal(item)} title={props.t("principal.remove")}>
            <span>{item.display_name || item.principal_id || item.type}</span>
            <span>x</span>
          </button>
        ))}
      </div>
    </div>
  );
}

function buildDefaultDrafts(
  models: ProviderCatalogModel[],
  current: Record<string, ModelDraft>,
): Record<string, ModelDraft> {
  const next: Record<string, ModelDraft> = {};
  for (const model of models) {
    next[model.model] =
      current[model.model] ?? {
        selected: false,
        types: model.types ?? [],
        pricingMode: "public",
        inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
        outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
      };
  }
  return next;
}

function hydrateDraftsFromAccount(
  models: TenantProviderAccountModel[],
  current: Record<string, ModelDraft>,
): Record<string, ModelDraft> {
  const next = { ...current };
  for (const item of models) {
    next[item.model] = {
      selected: item.enabled,
      types: item.types ?? [],
      pricingMode: item.pricing_mode ?? "public",
      inputPrice: item.private_pricing?.input_usd_per_1m?.toString() ?? "",
      outputPrice: item.private_pricing?.output_usd_per_1m?.toString() ?? "",
    };
  }
  return next;
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");

  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const data = parseResponsePayload(text);
  if (!response.ok) throw new Error(extractMessageFromPayload(data) ?? `Request failed (${response.status})`);
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
  if (!Number.isFinite(parsed)) return undefined;
  return parsed;
}

function principalTypeLabel(type: PrincipalType, t: ReturnType<typeof useTranslations>): string {
  if (type === "user") return t("principal.user");
  if (type === "team") return t("principal.team");
  return t("principal.everyone");
}

function getProviderLogoPath(provider: string): string {
  const key = provider.trim().toLowerCase();
  if (key.includes("openai")) return "/providers/openai.svg";
  if (key.includes("gemini") || key.includes("google")) return "/providers/google.svg";
  if (key.includes("vertex")) return "/providers/vertex.svg";
  if (key.includes("azure")) return "/providers/azure.svg";
  if (key.includes("aws") || key.includes("bedrock") || key.includes("sagemaker")) return "/providers/aws.svg";
  if (key.includes("anthropic")) return "/providers/anthropic.svg";
  if (key.includes("cohere")) return "/providers/cohere.svg";
  if (key.includes("groq")) return "/providers/groq.svg";
  if (key.includes("mistral")) return "/providers/mistral.svg";
  return "/globe.svg";
}

function stringsEqualFold(a: string, b: string): boolean {
  return a.trim().toLowerCase() === b.trim().toLowerCase();
}

function formatCost(value?: number): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return "-";
  return `$${value} / 1M`;
}
