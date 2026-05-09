"use client";

import Image from "next/image";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { Copy, Eye, KeyRound, Layers3, ShieldCheck } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PrincipalSearchInput, type PrincipalSearchOption } from "@/app/components/principal-search-input";

type ProviderAuthMode =
  | "api_key"
  | "azure_entra_cert"
  | "aws_sigv4"
  | "google_service_account"
  | "no_credential";

type ModelType = string;

type PrincipalType = "user" | "team" | "everyone";
type SetupStep = 1 | 2 | 3 | 4 | 5;
type WorkspaceTab = "models" | "access" | "credentials" | "observability";
type WorkspaceMode = "overview" | "setup" | "workspace";

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

type EffectiveModelView = {
  model: string;
  types?: string[];
  effective_capabilities?: string[];
  pricing_mode: "public" | "private";
  pricing_resolved: boolean;
  allowed: boolean;
  denied_reason?: string;
};

type EffectiveModelsResponse = {
  data?: EffectiveModelView[];
};

type ProvidersResponse = { data?: ProviderCatalogProvider[] };
type ProviderModelsResponse = { data?: ProviderCatalogModel[] };
type AccountsResponse = { data?: TenantProviderAccount[] };
type AccessPrincipalResponse = { data?: AccessPrincipalCandidate[] };
type ProviderCatalogStatsResponse = {
  data?: {
    cache_ttl_sec?: number;
    counters?: Record<string, number>;
  };
};

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

type Readiness = "ready" | "needs_credentials" | "needs_access" | "pricing_incomplete";

export default function ProviderAccountWizard() {
  const t = useTranslations("providerAccounts");
  const ux = useTranslations("providerAccounts.ux");

  const [workspaceMode, setWorkspaceMode] = useState<WorkspaceMode>("overview");
  const [setupStep, setSetupStep] = useState<SetupStep>(1);
  const [activeTab, setActiveTab] = useState<WorkspaceTab>("models");

  const [isLoading, setIsLoading] = useState(false);
  const [isSavingModels, setIsSavingModels] = useState(false);
  const [isSavingAccess, setIsSavingAccess] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [isLoadingEffective, setIsLoadingEffective] = useState(false);
  const [isRefreshingCatalog, setIsRefreshingCatalog] = useState(false);
  const [isActivating, setIsActivating] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [catalogStats, setCatalogStats] = useState<{ cacheTTL: number; counters: Record<string, number> }>({
    cacheTTL: 0,
    counters: {},
  });

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);

  const [selectedProviderID, setSelectedProviderID] = useState("");
  const [selectedAccountID, setSelectedAccountID] = useState("");
  const selectedProviderRef = useRef("");
  const selectedAccountRef = useRef("");

  const [account, setAccount] = useState<TenantProviderAccount | null>(null);

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

  const [modelSearchQuery, setModelSearchQuery] = useState("");
  const [modelDrafts, setModelDrafts] = useState<Record<string, ModelDraft>>({});

  const [managers, setManagers] = useState<AccessPrincipal[]>([]);
  const [users, setUsers] = useState<AccessPrincipal[]>([{ type: "everyone", display_name: t("principal.everyone") }]);

  const [managerKind, setManagerKind] = useState<"user" | "team">("user");
  const [userKind, setUserKind] = useState<"user" | "team">("user");

  const [effectiveKeyType, setEffectiveKeyType] = useState<"pat" | "vat">("pat");
  const [effectiveUserID, setEffectiveUserID] = useState("");
  const [effectiveTeamID, setEffectiveTeamID] = useState("");
  const [effectiveModels, setEffectiveModels] = useState<EffectiveModelView[]>([]);
  const [lastCredentialTestMessage, setLastCredentialTestMessage] = useState("");

  const handleEffectiveUserIDChange = useCallback(
    (value: string) => {
      setEffectiveUserID(value);
      if (!value.trim() && !effectiveTeamID.trim()) setEffectiveModels([]);
    },
    [effectiveTeamID],
  );

  const handleEffectiveTeamIDChange = useCallback(
    (value: string) => {
      setEffectiveTeamID(value);
      if (!value.trim() && !effectiveUserID.trim()) setEffectiveModels([]);
    },
    [effectiveUserID],
  );

  const visibleProviders = useMemo(
    () => providers.filter((item) => item.enabled && stringsEqualFold(item.provider, "azure") === false),
    [providers],
  );

  const selectedProvider = useMemo(
    () => visibleProviders.find((item) => stringsEqualFold(item.provider, selectedProviderID)) ?? null,
    [visibleProviders, selectedProviderID],
  );

  const selectedProviderAccounts = useMemo(
    () => accounts.filter((item) => stringsEqualFold(item.provider, selectedProviderID)),
    [accounts, selectedProviderID],
  );

  useEffect(() => {
    selectedProviderRef.current = selectedProviderID;
  }, [selectedProviderID]);

  useEffect(() => {
    selectedAccountRef.current = selectedAccountID;
  }, [selectedAccountID]);

  const filteredProviderModels = useMemo(() => {
    const query = modelSearchQuery.trim().toLowerCase();
    if (!query) return providerModels;
    return providerModels.filter((model) => {
      const display = (model.display_name || "").toLowerCase();
      const key = model.model.toLowerCase();
      return display.includes(query) || key.includes(query);
    });
  }, [providerModels, modelSearchQuery]);

  const selectedModelCount = useMemo(
    () => Object.values(modelDrafts).filter((draft) => draft.selected).length,
    [modelDrafts],
  );

  const unresolvedPricingCount = useMemo(() => {
    let unresolved = 0;
    for (const model of providerModels) {
      const draft = modelDrafts[model.model];
      if (!draft?.selected) continue;
      if (draft.pricingMode === "private") {
        const input = toNumberOrUndefined(draft.inputPrice);
        const output = toNumberOrUndefined(draft.outputPrice);
        if (typeof input !== "number" && typeof output !== "number") unresolved += 1;
        continue;
      }
      const hasCatalog =
        typeof model.public_pricing?.input_usd_per_1m === "number" ||
        typeof model.public_pricing?.output_usd_per_1m === "number";
      if (!hasCatalog) unresolved += 1;
    }
    return unresolved;
  }, [modelDrafts, providerModels]);

  const readiness = useMemo<Readiness>(() => {
    const needsCredential =
      !!account &&
      !account.has_credential &&
      !account.allows_virtual_keys_only &&
      account.auth_mode !== "no_credential";
    if (needsCredential) return "needs_credentials";
    if (managers.length === 0) return "needs_access";
    if (selectedModelCount > 0 && unresolvedPricingCount > 0) return "pricing_incomplete";
    return "ready";
  }, [account, managers.length, selectedModelCount, unresolvedPricingCount]);

  const loadProviderModels = useCallback(
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
        const items = unwrapData<ProviderCatalogModel[]>(response) ?? [];
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

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setBanner(null);
    try {
      const currentProviderID = selectedProviderRef.current;
      const currentAccountID = selectedAccountRef.current;
      const [providerData, accountData] = await Promise.all([
        requestJSON<ProvidersResponse>("/api/provider-catalog/providers"),
        requestJSON<AccountsResponse>("/api/tenant-provider-accounts"),
      ]);

      const providerItems = unwrapData<ProviderCatalogProvider[]>(providerData) ?? [];
      const accountItems = unwrapData<TenantProviderAccount[]>(accountData) ?? [];
      setProviders(providerItems);
      setAccounts(accountItems);

      const initialProvider =
        currentProviderID || accountItems[0]?.provider || providerItems[0]?.provider || "";
      if (initialProvider) setSelectedProviderID(initialProvider);

      if (currentAccountID && accountItems.some((item) => item.id === currentAccountID)) return;

      const firstProviderAccount = accountItems.find((item) => stringsEqualFold(item.provider, initialProvider));
      if (firstProviderAccount) {
        setSelectedAccountID(firstProviderAccount.id);
        setWorkspaceMode("workspace");
      } else {
        setSelectedAccountID("");
        setWorkspaceMode("overview");
      }
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.initialLoadFailed")) });
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  const loadCatalogStats = useCallback(async () => {
    try {
      const response = await requestJSON<ProviderCatalogStatsResponse>("/api/provider-catalog/stats");
      const payload = unwrapData<ProviderCatalogStatsResponse["data"]>(response) ?? response.data;
      setCatalogStats({
        cacheTTL: Number(payload?.cache_ttl_sec ?? 0),
        counters: payload?.counters ?? {},
      });
    } catch {
      // Best-effort observability call; do not block the main flow.
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadInitial();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadInitial]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadCatalogStats();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadCatalogStats]);

  const handleLoadAccount = useCallback(
    async (accountID: string) => {
      if (!accountID) return;
      setIsLoading(true);
      setBanner(null);
      try {
        const details = await requestJSON<AccountDetailsResponse>(
          `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}`,
        );
        const payload = unwrapData<AccountDetailsResponse>(details) ?? details;
        setAccount(payload.account);
        setProvider(payload.account.provider);
        setName(payload.account.name);
        setAuthMode(payload.account.auth_mode);
        setAllowVirtualOnly(payload.account.allows_virtual_keys_only);
        setManagers(payload.access?.managers ?? []);
        setUsers(payload.access?.users ?? []);
        setSelectedProviderID(payload.account.provider);
        setSelectedAccountID(payload.account.id);
        await loadProviderModels(payload.account.provider);
        setModelDrafts((prev) => hydrateDraftsFromAccount(payload.models ?? [], prev));
      } catch (error) {
        setBanner({ kind: "error", message: normalizeError(error, t("messages.accountLoadFailed")) });
      } finally {
        setIsLoading(false);
      }
    },
    [loadProviderModels, t],
  );

  useEffect(() => {
    if (!selectedAccountID || workspaceMode !== "workspace") return;
    const timer = setTimeout(() => {
      void handleLoadAccount(selectedAccountID);
    }, 0);
    return () => clearTimeout(timer);
  }, [selectedAccountID, workspaceMode, handleLoadAccount]);

  function resetCredentialInputs() {
    setAPIKey("");
    setSecretRef("");
    setAzureTenantID("");
    setAzureClientID("");
    setAWSAccessKeyID("");
    setAWSSecretAccessKey("");
    setAWSRegion("us-east-1");
    setGoogleServiceAccountJSON("");
  }

  function startSetup(providerID?: string) {
    const nextProvider = providerID || selectedProviderID || visibleProviders[0]?.provider || "";
    setWorkspaceMode("setup");
    setSetupStep(1);
    setSelectedAccountID("");
    setAccount(null);
    setManagers([]);
    setUsers([{ type: "everyone", display_name: t("principal.everyone") }]);
    setEffectiveModels([]);
    setLastCredentialTestMessage("");
    setModelSearchQuery("");
    resetCredentialInputs();

    if (nextProvider) {
      const providerItem = visibleProviders.find((item) => stringsEqualFold(item.provider, nextProvider));
      setSelectedProviderID(nextProvider);
      setProvider(nextProvider);
      setName(nextProvider);
      setAuthMode(providerItem?.auth_modes?.[0] ?? "api_key");
      setAllowVirtualOnly(false);
      void loadProviderModels(nextProvider);
    }
  }

  async function handleRefreshCatalog() {
    setIsRefreshingCatalog(true);
    setBanner(null);
    try {
      await requestJSON("/api/provider-catalog/refresh", { method: "POST" });
      await loadInitial();
      await loadCatalogStats();
      if (selectedProviderID) {
        await loadProviderModels(selectedProviderID);
      }
      setBanner({ kind: "success", message: ux("messages.catalogRefreshSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, ux("messages.catalogRefreshFailed")) });
    } finally {
      setIsRefreshingCatalog(false);
    }
  }

  function handleSelectProvider(item: ProviderCatalogProvider) {
    setSelectedProviderID(item.provider);
    if (workspaceMode === "setup") {
      setProvider(item.provider);
      setAuthMode(item.auth_modes?.[0] ?? "api_key");
      if (!name.trim()) setName(item.provider);
      void loadProviderModels(item.provider);
    }
  }

  function handleSelectAccount(item: TenantProviderAccount) {
    setSelectedProviderID(item.provider);
    setSelectedAccountID(item.id);
    setWorkspaceMode("workspace");
    setActiveTab("models");
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
      const newAccount = unwrapData<TenantProviderAccount>(created) ?? created;
      setSelectedAccountID(newAccount.id);
      setSelectedProviderID(newAccount.provider);
      setAccount(newAccount);
      await loadInitial();
      await handleLoadAccount(newAccount.id);
      setSetupStep(3);
      setBanner({ kind: "success", message: t("messages.accountCreateSuccess") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accountCreateFailed")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleSaveModels(accountID = selectedAccountID): Promise<boolean> {
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }

    const validationErrors: string[] = [];
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

    for (const item of payloadModels) {
      if (!Array.isArray(item.types) || item.types.length === 0) {
        validationErrors.push(`${item.model}: ${ux("messages.modelTypeRequired")}`);
      }
      if (item.pricing_mode === "private") {
        const input = item.private_pricing?.input_usd_per_1m;
        const output = item.private_pricing?.output_usd_per_1m;
        if (typeof input !== "number" && typeof output !== "number") {
          validationErrors.push(`${item.model}: ${ux("messages.privatePricingRequired")}`);
        }
      }
    }

    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: t("messages.selectAtLeastOneModel") });
      return false;
    }
    if (validationErrors.length > 0) {
      setBanner({ kind: "error", message: validationErrors[0] });
      return false;
    }

    setIsSavingModels(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/models`,
        { method: "PUT", body: JSON.stringify({ models: payloadModels }) },
      );
      setBanner({ kind: "success", message: t("messages.modelsSaveSuccess") });
      await handleLoadAccount(accountID);
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.modelsSaveFailed")) });
      return false;
    } finally {
      setIsSavingModels(false);
    }
  }

  async function handleSaveAccess(accountID = selectedAccountID): Promise<boolean> {
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }
    if (managers.length === 0) {
      setBanner({ kind: "error", message: t("messages.managerRequired") });
      return false;
    }

    setIsSavingAccess(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/access`,
        { method: "PUT", body: JSON.stringify({ access: { managers, users } }) },
      );
      setBanner({ kind: "success", message: t("messages.accessSaveSuccess") });
      await handleLoadAccount(accountID);
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accessSaveFailed")) });
      return false;
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
      const payload = unwrapData<{ success: boolean; message?: string }>(result) ?? result;
      const nextMessage =
        payload.message ?? (payload.success ? t("messages.credentialValid") : t("messages.credentialInvalid"));
      setLastCredentialTestMessage(nextMessage);
      setBanner({ kind: payload.success ? "success" : "error", message: nextMessage });
      await loadInitial();
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.credentialTestFailed")) });
    } finally {
      setIsTesting(false);
    }
  }

  async function handleActivateAccount() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }
    setIsActivating(true);
    try {
      const modelsOK = await handleSaveModels(selectedAccountID);
      if (!modelsOK) return;
      const accessOK = await handleSaveAccess(selectedAccountID);
      if (!accessOK) return;
      setSetupStep(5);
      setWorkspaceMode("workspace");
      setActiveTab("models");
      setBanner({ kind: "success", message: ux("messages.accountActivated") });
    } finally {
      setIsActivating(false);
    }
  }

  const searchPrincipalCandidates = useCallback(async (kind: "user" | "team", query: string) => {
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query.trim())}&limit=20`,
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

  async function loadEffectiveModelsPreview(accountID = selectedAccountID) {
    if (!accountID) {
      setEffectiveModels([]);
      return;
    }
    if (!effectiveUserID.trim() && !effectiveTeamID.trim()) {
      setEffectiveModels([]);
      return;
    }
    setIsLoadingEffective(true);
    try {
      const query = new URLSearchParams();
      query.set("key_type", effectiveKeyType);
      if (effectiveUserID.trim()) query.set("user_id", effectiveUserID.trim());
      if (effectiveTeamID.trim()) query.set("team_id", effectiveTeamID.trim());
      const response = await requestJSON<EffectiveModelsResponse>(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/effective-models?${query.toString()}`,
      );
      const payload = unwrapData<EffectiveModelView[]>(response) ?? [];
      setEffectiveModels(
        payload.map((item) => ({
          ...item,
          types: item.types ?? item.effective_capabilities ?? [],
        })),
      );
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "Failed to load effective models.") });
    } finally {
      setIsLoadingEffective(false);
    }
  }

  const panelTitle =
    workspaceMode === "setup"
      ? ux("panel.setupTitle")
      : workspaceMode === "workspace"
        ? ux("panel.workspaceTitle")
        : ux("panel.overviewTitle");

  const setupSteps: Array<{ id: SetupStep; title: string }> = [
    { id: 1, title: ux("setupSteps.provider") },
    { id: 2, title: ux("setupSteps.accountCredentials") },
    { id: 3, title: ux("setupSteps.models") },
    { id: 4, title: ux("setupSteps.access") },
    { id: 5, title: ux("setupSteps.review") },
  ];

  const workspaceTabs: Array<{ id: WorkspaceTab; label: string; icon: typeof Layers3 }> = [
    { id: "models", label: ux("tabs.models"), icon: Layers3 },
    { id: "access", label: ux("tabs.access"), icon: ShieldCheck },
    { id: "credentials", label: ux("tabs.credentials"), icon: KeyRound },
    { id: "observability", label: ux("tabs.observability"), icon: Eye },
  ];

  return (
    <Card id="tenant-provider-accounts" className="border-border/70 shadow-sm">
      <CardHeader className="rounded-t-xl bg-gradient-to-l from-background to-muted/30">
        <CardTitle className="text-2xl">{t("title")}</CardTitle>
        <CardDescription className="text-base">{t("description")}</CardDescription>
      </CardHeader>

      <CardContent className="space-y-4 pt-6">
        {banner ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

        <div className="grid gap-4 lg:grid-cols-12">
          <aside className="space-y-3 lg:col-span-4 xl:col-span-3">
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{ux("sidebar.connectedProviders")}</CardTitle>
                <CardDescription>{ux("sidebar.connectedProvidersHint")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                {visibleProviders.map((item) => {
                  const providerCount = accounts.filter((accountItem) => stringsEqualFold(accountItem.provider, item.provider)).length;
                  const isSelected = stringsEqualFold(selectedProviderID, item.provider);
                  return (
                    <button
                      key={item.provider}
                      type="button"
                      className={`w-full rounded-lg border px-3 py-2 text-start transition ${
                        isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"
                      }`}
                      onClick={() => handleSelectProvider(item)}
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <Image src={getProviderLogoPath(item.provider)} alt={item.display_name} width={18} height={18} />
                          <span className="font-medium">{item.display_name}</span>
                        </div>
                        <Badge variant="secondary">{providerCount}</Badge>
                      </div>
                    </button>
                  );
                })}
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{ux("sidebar.providerAccounts")}</CardTitle>
                <CardDescription>{selectedProvider?.display_name ?? ux("sidebar.selectProviderPrompt")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                {selectedProviderAccounts.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{ux("sidebar.noAccountYet")}</p>
                ) : (
                  selectedProviderAccounts.map((item) => {
                    const isSelected = item.id === selectedAccountID && workspaceMode === "workspace";
                    return (
                      <button
                        key={item.id}
                        type="button"
                        className={`w-full rounded-lg border px-3 py-2 text-start ${
                          isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"
                        }`}
                        onClick={() => handleSelectAccount(item)}
                      >
                        <div className="flex items-center justify-between gap-2">
                          <div>
                            <div className="font-medium">{item.name}</div>
                            <div className="text-xs text-muted-foreground">{item.auth_mode}</div>
                          </div>
                          <ReadinessBadge
                            readiness={deriveAccountListReadiness(item)}
                            labels={{
                              ready: ux("readiness.ready"),
                              needs_credentials: ux("readiness.needsCredentials"),
                              needs_access: ux("readiness.needsAccess"),
                              pricing_incomplete: ux("readiness.pricingIncomplete"),
                            }}
                          />
                        </div>
                      </button>
                    );
                  })
                )}

                <div className="flex gap-2 pt-2">
                  <Button type="button" size="sm" className="flex-1" onClick={() => startSetup(selectedProviderID)}>
                    {ux("actions.addProviderAccount")}
                  </Button>
                  <Button type="button" size="sm" variant="outline" onClick={() => void handleRefreshCatalog()} disabled={isRefreshingCatalog}>
                    {isRefreshingCatalog ? ux("actions.refreshingCatalog") : ux("actions.refreshCatalog")}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </aside>

          <section className="space-y-4 lg:col-span-8 xl:col-span-9">
            <Card className="border-border/70">
              <CardHeader className="pb-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <CardTitle className="text-xl">{panelTitle}</CardTitle>
                    <CardDescription>
                      {workspaceMode === "setup"
                        ? ux("panel.setupHint")
                        : workspaceMode === "workspace"
                          ? ux("panel.workspaceHint")
                          : ux("panel.overviewHint")}
                    </CardDescription>
                  </div>
                  {workspaceMode === "workspace" && account ? (
                    <div className="flex items-center gap-2">
                      <ReadinessBadge
                        readiness={readiness}
                        labels={{
                          ready: ux("readiness.ready"),
                          needs_credentials: ux("readiness.needsCredentials"),
                          needs_access: ux("readiness.needsAccess"),
                          pricing_incomplete: ux("readiness.pricingIncomplete"),
                        }}
                      />
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          void navigator.clipboard.writeText(`${account.provider}/${account.name}`);
                          setBanner({ kind: "success", message: ux("messages.fqnCopied") });
                        }}
                      >
                        <Copy className="me-2 h-4 w-4" />
                        {ux("actions.copyFqn")}
                      </Button>
                    </div>
                  ) : null}
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {workspaceMode === "overview" ? (
                  <EmptyOverview
                    title={ux("overview.title")}
                    description={ux("overview.description")}
                    setupLabel={ux("actions.setupNewAccount")}
                    openExistingLabel={ux("actions.openExisting")}
                    onStart={() => startSetup(selectedProviderID)}
                    onOpenFirst={() => {
                      const first = selectedProviderAccounts[0];
                      if (first) handleSelectAccount(first);
                    }}
                    hasAccount={selectedProviderAccounts.length > 0}
                  />
                ) : null}

                {workspaceMode === "setup" ? (
                  <>
                    <div className="grid gap-2 md:grid-cols-5">
                      {setupSteps.map((step) => (
                        <button
                          key={step.id}
                          type="button"
                          onClick={() => {
                            if (step.id <= setupStep) setSetupStep(step.id);
                          }}
                          className={`rounded-lg border px-2 py-2 text-start ${
                            setupStep === step.id
                              ? "border-primary bg-primary/10"
                              : setupStep > step.id
                                ? "border-primary/40 bg-primary/5"
                                : "border-border"
                          }`}
                        >
                          <div className="text-xs text-muted-foreground">{ux("setup.stepLabel", { number: step.id })}</div>
                          <div className="text-sm font-medium">{step.title}</div>
                        </button>
                      ))}
                    </div>

                    {setupStep === 1 ? (
                      <section className="space-y-3">
                        <h3 className="text-base font-semibold">{ux("setup.selectProviderTitle")}</h3>
                        <p className="text-sm text-muted-foreground">{ux("setup.selectProviderDesc")}</p>
                        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
                          {visibleProviders.map((item) => {
                            const isSelected = stringsEqualFold(item.provider, provider);
                            return (
                              <button
                                key={item.provider}
                                type="button"
                                className={`rounded-lg border p-3 text-start ${
                                  isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"
                                }`}
                                onClick={() => {
                                  setProvider(item.provider);
                                  setAuthMode(item.auth_modes?.[0] ?? "api_key");
                                  if (!name.trim()) setName(item.provider);
                                  setSelectedProviderID(item.provider);
                                  void loadProviderModels(item.provider);
                                }}
                              >
                                <div className="flex items-center gap-2">
                                  <Image src={getProviderLogoPath(item.provider)} alt={item.display_name} width={18} height={18} />
                                  <span className="font-medium">{item.display_name}</span>
                                </div>
                                <div className="mt-2 text-xs text-muted-foreground">{ux("setup.authModes")}: {item.auth_modes.join(", ")}</div>
                              </button>
                            );
                          })}
                        </div>
                        <div className="flex justify-end">
                          <Button type="button" onClick={() => setSetupStep(2)} disabled={!provider}>{ux("actions.continue")}</Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 2 ? (
                      <section className="space-y-4">
                        <h3 className="text-base font-semibold">{ux("setup.accountBasicsTitle")}</h3>
                        <p className="text-sm text-muted-foreground">
                          {ux("microcopy.virtualOnly")}
                        </p>

                        <div className="grid gap-3 md:grid-cols-2">
                          <Field label={t("fields.provider")} id="provider">
                            <Input id="provider" value={provider} onChange={(event) => setProvider(event.target.value)} />
                          </Field>
                          <Field label={t("fields.accountName")} id="accountName">
                            <Input id="accountName" value={name} onChange={(event) => setName(event.target.value)} placeholder={ux("placeholders.accountName")} />
                          </Field>
                        </div>

                        <div className="grid gap-3 md:grid-cols-2">
                          <Field label={t("fields.authMode")} id="authMode">
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
                          </Field>
                          <Field label={t("fields.virtualKeysMode")} id="virtualOnly">
                            <label className="inline-flex h-10 items-center gap-2 rounded-md border border-input px-3 text-sm">
                              <input
                                id="virtualOnly"
                                type="checkbox"
                                checked={allowVirtualOnly}
                                onChange={(event) => setAllowVirtualOnly(event.target.checked)}
                              />
                              {t("fields.allowVirtualOnly")}
                            </label>
                          </Field>
                        </div>

                        {!allowVirtualOnly ? (
                          <CredentialFields
                            authMode={authMode}
                            apiKey={apiKey}
                            setAPIKey={setAPIKey}
                            secretRef={secretRef}
                            setSecretRef={setSecretRef}
                            azureTenantID={azureTenantID}
                            setAzureTenantID={setAzureTenantID}
                            azureClientID={azureClientID}
                            setAzureClientID={setAzureClientID}
                            awsAccessKeyID={awsAccessKeyID}
                            setAWSAccessKeyID={setAWSAccessKeyID}
                            awsSecretAccessKey={awsSecretAccessKey}
                            setAWSSecretAccessKey={setAWSSecretAccessKey}
                            awsRegion={awsRegion}
                            setAWSRegion={setAWSRegion}
                            googleServiceAccountJSON={googleServiceAccountJSON}
                            setGoogleServiceAccountJSON={setGoogleServiceAccountJSON}
                            t={t}
                          />
                        ) : null}

                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(1)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => void handleCreateAccount()} disabled={isLoading || !provider || !name.trim()}>
                            {isLoading ? t("creating") : t("createProviderAccount")}
                          </Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 3 ? (
                      <section className="space-y-4">
                        <ModelsEditor
                          t={t}
                          filteredProviderModels={filteredProviderModels}
                          modelDrafts={modelDrafts}
                          setModelDrafts={setModelDrafts}
                          modelSearchQuery={modelSearchQuery}
                          setModelSearchQuery={setModelSearchQuery}
                          isSaving={isSavingModels}
                          onSave={() => void handleSaveModels()}
                        />
                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(2)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => setSetupStep(4)} disabled={selectedModelCount === 0}>{ux("actions.continue")}</Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 4 ? (
                      <section className="space-y-4">
                        <AccessEditor
                          t={t}
                          managerKind={managerKind}
                          setManagerKind={setManagerKind}
                          managers={managers}
                          setManagers={setManagers}
                          userKind={userKind}
                          setUserKind={setUserKind}
                          users={users}
                          setUsers={setUsers}
                          onSearchPrincipal={searchPrincipalCandidates}
                          onSave={() => void handleSaveAccess()}
                          isSaving={isSavingAccess}
                          effectiveKeyType={effectiveKeyType}
                          setEffectiveKeyType={setEffectiveKeyType}
                          effectiveUserID={effectiveUserID}
                          setEffectiveUserID={handleEffectiveUserIDChange}
                          effectiveTeamID={effectiveTeamID}
                          setEffectiveTeamID={handleEffectiveTeamIDChange}
                          effectiveModels={effectiveModels}
                          isLoadingEffective={isLoadingEffective}
                          onLoadEffective={() => void loadEffectiveModelsPreview()}
                        />
                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(3)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => setSetupStep(5)} disabled={managers.length === 0}>{ux("actions.continue")}</Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 5 ? (
                      <section className="space-y-4">
                        <h3 className="text-base font-semibold">{ux("review.title")}</h3>
                        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                          <SummaryKPI label={ux("kpi.credentialStatus")} value={account?.has_credential || allowVirtualOnly ? ux("readiness.ready") : ux("readiness.needsCredentials")} />
                          <SummaryKPI label={ux("kpi.enabledModels")} value={selectedModelCount.toString()} />
                          <SummaryKPI label={ux("kpi.unresolvedPricing")} value={unresolvedPricingCount.toString()} />
                          <SummaryKPI label={ux("kpi.managersUsers")} value={`${managers.length} / ${users.length}`} />
                        </div>
                        <Alert>
                          <AlertTitle>{ux("review.activationCheckTitle")}</AlertTitle>
                          <AlertDescription>
                            {ux("microcopy.pricingPolicy")}
                          </AlertDescription>
                        </Alert>
                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(4)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => void handleActivateAccount()} disabled={isActivating}>
                            {isActivating ? ux("actions.activating") : ux("actions.activateAccount")}
                          </Button>
                        </div>
                      </section>
                    ) : null}
                  </>
                ) : null}

                {workspaceMode === "workspace" && account ? (
                  <>
                    <div className="grid gap-3 md:grid-cols-3 xl:grid-cols-4">
                      <SummaryKPI label={ux("kpi.provider")} value={account.provider} />
                      <SummaryKPI label={ux("kpi.enabledModels")} value={selectedModelCount.toString()} />
                      <SummaryKPI label={ux("kpi.unresolvedPricing")} value={unresolvedPricingCount.toString()} />
                      <SummaryKPI label={ux("kpi.managersUsers")} value={`${managers.length} / ${users.length}`} />
                    </div>

                    <div className="flex flex-wrap items-center gap-2 rounded-lg border bg-muted/20 p-2">
                      {workspaceTabs.map((tab) => {
                        const Icon = tab.icon;
                        return (
                          <Button
                            key={tab.id}
                            type="button"
                            variant={activeTab === tab.id ? "default" : "outline"}
                            size="sm"
                            onClick={() => setActiveTab(tab.id)}
                          >
                            <Icon className="me-2 h-4 w-4" />
                            {tab.label}
                          </Button>
                        );
                      })}
                    </div>

                    {activeTab === "models" ? (
                      <ModelsEditor
                        t={t}
                        filteredProviderModels={filteredProviderModels}
                        modelDrafts={modelDrafts}
                        setModelDrafts={setModelDrafts}
                        modelSearchQuery={modelSearchQuery}
                        setModelSearchQuery={setModelSearchQuery}
                        isSaving={isSavingModels}
                        onSave={() => void handleSaveModels()}
                      />
                    ) : null}

                    {activeTab === "access" ? (
                      <AccessEditor
                        t={t}
                        managerKind={managerKind}
                        setManagerKind={setManagerKind}
                        managers={managers}
                        setManagers={setManagers}
                        userKind={userKind}
                        setUserKind={setUserKind}
                        users={users}
                        setUsers={setUsers}
                        onSearchPrincipal={searchPrincipalCandidates}
                        onSave={() => void handleSaveAccess()}
                        isSaving={isSavingAccess}
                        effectiveKeyType={effectiveKeyType}
                        setEffectiveKeyType={setEffectiveKeyType}
                        effectiveUserID={effectiveUserID}
                        setEffectiveUserID={handleEffectiveUserIDChange}
                        effectiveTeamID={effectiveTeamID}
                        setEffectiveTeamID={handleEffectiveTeamIDChange}
                        effectiveModels={effectiveModels}
                        isLoadingEffective={isLoadingEffective}
                        onLoadEffective={() => void loadEffectiveModelsPreview()}
                      />
                    ) : null}

                    {activeTab === "credentials" ? (
                      <section className="space-y-4 rounded-lg border p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h3 className="text-base font-semibold">{ux("tabs.credentials")}</h3>
                            <p className="text-sm text-muted-foreground">{ux("credentials.desc")}</p>
                          </div>
                          <Badge variant="secondary">{account.auth_mode}</Badge>
                        </div>
                        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                          <SummaryKPI label={ux("credentials.hasCredential")} value={account.has_credential ? ux("common.yes") : ux("common.no")} />
                          <SummaryKPI label={ux("credentials.virtualOnly")} value={account.allows_virtual_keys_only ? ux("common.enabled") : ux("common.disabled")} />
                          <SummaryKPI label={ux("credentials.lastTest")} value={lastCredentialTestMessage || ux("credentials.notTested")} />
                        </div>
                        <div>
                          <Button type="button" variant="outline" onClick={() => void handleTestCredentials()} disabled={isTesting}>
                            {isTesting ? t("testing") : t("testCredentials")}
                          </Button>
                        </div>
                      </section>
                    ) : null}

                    {activeTab === "observability" ? (
                      <section className="space-y-4 rounded-lg border p-4">
                        <div>
                          <h3 className="text-base font-semibold">{ux("tabs.observability")}</h3>
                          <p className="text-sm text-muted-foreground">
                            {ux("observability.desc")}
                          </p>
                        </div>
                        <div className="rounded-lg border bg-muted/20 p-3 text-sm">
                          {ux("observability.deniedReasons")}
                        </div>
                        <div className="grid gap-3 md:grid-cols-3">
                          <SummaryKPI
                            label={ux("observability.cacheTTL")}
                            value={`${catalogStats.cacheTTL}`}
                          />
                          <SummaryKPI
                            label={ux("observability.providersCache")}
                            value={`${catalogStats.counters.providers_cache_hit ?? 0} / ${catalogStats.counters.providers_cache_miss ?? 0}`}
                          />
                          <SummaryKPI
                            label={ux("observability.modelsCache")}
                            value={`${catalogStats.counters.models_cache_hit ?? 0} / ${catalogStats.counters.models_cache_miss ?? 0}`}
                          />
                        </div>
                        <div className="grid gap-3 md:grid-cols-2">
                          <SummaryKPI
                            label={ux("observability.discoveryErrors")}
                            value={`${catalogStats.counters.discovery_error ?? 0}`}
                          />
                          <SummaryKPI
                            label={ux("observability.discoveryFallback")}
                            value={`${catalogStats.counters.discovery_fallback ?? 0}`}
                          />
                        </div>
                        <div className="flex flex-wrap items-end gap-2">
                          <div className="grid gap-1">
                            <Label htmlFor="obsKeyType">{ux("access.keyTypeLabel")}</Label>
                            <select
                              id="obsKeyType"
                              className="h-9 rounded-md border border-input bg-background px-3 text-sm"
                              value={effectiveKeyType}
                              onChange={(event) => setEffectiveKeyType(event.target.value as "pat" | "vat")}
                            >
                              <option value="pat">{ux("access.pat")}</option>
                              <option value="vat">{ux("access.vat")}</option>
                            </select>
                          </div>
                          <Field label={ux("access.userIdLabel")} id="obsUserID">
                            <PrincipalSearchInput
                              id="obsUserID"
                              kind="user"
                              selectedID={effectiveUserID}
                              onSelectedIDChange={setEffectiveUserID}
                              onSearch={searchPrincipalCandidates}
                              loadingLabel={ux("actions.loading")}
                              emptyLabel={ux("effectiveModels.empty")}
                              placeholder={`${t("principal.searchPlaceholder")} (${t("principal.user")})`}
                            />
                          </Field>
                          <Field label={ux("access.teamIdLabel")} id="obsTeamID">
                            <PrincipalSearchInput
                              id="obsTeamID"
                              kind="team"
                              selectedID={effectiveTeamID}
                              onSelectedIDChange={setEffectiveTeamID}
                              onSearch={searchPrincipalCandidates}
                              loadingLabel={ux("actions.loading")}
                              emptyLabel={ux("effectiveModels.empty")}
                              placeholder={`${t("principal.searchPlaceholder")} (${t("principal.team")})`}
                            />
                          </Field>
                          <Button type="button" variant="outline" onClick={() => void loadEffectiveModelsPreview()} disabled={isLoadingEffective}>
                            {isLoadingEffective ? ux("actions.loading") : ux("actions.loadEffectiveModels")}
                          </Button>
                        </div>
                        <EffectiveModelsTable models={effectiveModels} />
                      </section>
                    ) : null}
                  </>
                ) : null}
              </CardContent>
            </Card>
          </section>
        </div>
      </CardContent>
    </Card>
  );
}

function EmptyOverview(props: {
  title: string;
  description: string;
  setupLabel: string;
  openExistingLabel: string;
  onStart: () => void;
  onOpenFirst: () => void;
  hasAccount: boolean;
}) {
  return (
    <section className="space-y-3 rounded-lg border border-dashed p-6 text-center">
      <h3 className="text-lg font-semibold">{props.title}</h3>
      <p className="text-sm text-muted-foreground">{props.description}</p>
      <div className="flex flex-wrap items-center justify-center gap-2">
        <Button type="button" onClick={props.onStart}>{props.setupLabel}</Button>
        {props.hasAccount ? (
          <Button type="button" variant="outline" onClick={props.onOpenFirst}>{props.openExistingLabel}</Button>
        ) : null}
      </div>
    </section>
  );
}

function SummaryKPI(props: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-muted/20 p-3">
      <div className="text-xs text-muted-foreground">{props.label}</div>
      <div className="truncate text-sm font-semibold">{props.value}</div>
    </div>
  );
}

function ReadinessBadge(props: {
  readiness: Readiness;
  labels: Record<Readiness, string>;
}) {
  const mapping: Record<Readiness, { label: string; className: string }> = {
    ready: { label: props.labels.ready, className: "bg-emerald-100 text-emerald-700" },
    needs_credentials: { label: props.labels.needs_credentials, className: "bg-amber-100 text-amber-700" },
    needs_access: { label: props.labels.needs_access, className: "bg-amber-100 text-amber-700" },
    pricing_incomplete: { label: props.labels.pricing_incomplete, className: "bg-sky-100 text-sky-700" },
  };
  const view = mapping[props.readiness];
  return <span className={`rounded-full px-2 py-1 text-[10px] font-semibold ${view.className}`}>{view.label}</span>;
}

function deriveAccountListReadiness(account: TenantProviderAccount): Readiness {
  if (!account.has_credential && !account.allows_virtual_keys_only && account.auth_mode !== "no_credential") {
    return "needs_credentials";
  }
  return "ready";
}

function EffectiveModelsTable(props: { models: EffectiveModelView[] }) {
  const ux = useTranslations("providerAccounts.ux");
  if (props.models.length === 0) {
    return <p className="text-sm text-muted-foreground">{ux("effectiveModels.empty")}</p>;
  }
  return (
    <div className="overflow-hidden rounded-lg border">
      <div className="grid grid-cols-[1.2fr_1fr_120px_120px] border-b bg-muted/30 px-3 py-2 text-xs font-semibold text-muted-foreground">
        <div>{ux("effectiveModels.columns.model")}</div>
        <div>{ux("effectiveModels.columns.capabilities")}</div>
        <div>{ux("effectiveModels.columns.pricing")}</div>
        <div>{ux("effectiveModels.columns.status")}</div>
      </div>
      {props.models.map((item) => (
        <div key={item.model} className="grid grid-cols-[1.2fr_1fr_120px_120px] items-center border-b px-3 py-2 text-sm last:border-b-0">
          <div className="truncate font-medium">{item.model}</div>
          <div className="flex flex-wrap gap-1">
            {(item.types ?? []).map((type) => (
              <Badge key={`${item.model}-${type}`} variant="secondary">{type}</Badge>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <span>{item.pricing_mode}</span>
            <Badge variant={item.pricing_resolved ? "default" : "outline"}>
              {item.pricing_resolved ? ux("effectiveModels.resolved") : ux("effectiveModels.unresolved")}
            </Badge>
          </div>
          <div>
            <Badge variant={item.allowed ? "default" : "destructive"}>
              {item.allowed ? ux("effectiveModels.allowed") : ux("effectiveModels.denied")}
            </Badge>
            {!item.allowed && item.denied_reason ? (
              <p className="mt-1 truncate text-xs text-muted-foreground">{item.denied_reason}</p>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}

function ModelsEditor(props: {
  t: ReturnType<typeof useTranslations>;
  filteredProviderModels: ProviderCatalogModel[];
  modelDrafts: Record<string, ModelDraft>;
  setModelDrafts: React.Dispatch<React.SetStateAction<Record<string, ModelDraft>>>;
  modelSearchQuery: string;
  setModelSearchQuery: (value: string) => void;
  isSaving: boolean;
  onSave: () => void;
}) {
  const ux = useTranslations("providerAccounts.ux");
  return (
    <section className="space-y-3 rounded-lg border p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">{ux("tabs.models")}</h3>
          <p className="text-sm text-muted-foreground">{ux("models.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={props.isSaving}>
          {props.isSaving ? props.t("saving") : props.t("saveSelectedModels")}
        </Button>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Input
          value={props.modelSearchQuery}
          onChange={(event) => props.setModelSearchQuery(event.target.value)}
          placeholder={props.t("step3.searchPlaceholder")}
          className="max-w-sm"
        />
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            const allSelected = props.filteredProviderModels.length > 0 && props.filteredProviderModels.every((model) => {
              const draft = props.modelDrafts[model.model];
              return Boolean(draft?.selected);
            });
            props.setModelDrafts((prev) => {
              const next = { ...prev };
              for (const model of props.filteredProviderModels) {
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
          disabled={props.filteredProviderModels.length === 0}
        >
          {props.t("step3.toggleSelectAll")}
        </Button>
      </div>

      {props.filteredProviderModels.length === 0 ? (
        <p className="text-sm text-muted-foreground">{props.t("step3.empty")}</p>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <div className="grid grid-cols-[44px_1fr_1fr_160px_160px] border-b bg-muted/30 px-3 py-2 text-xs font-semibold text-muted-foreground">
            <div />
            <div>{ux("models.columns.model")}</div>
            <div>{ux("models.columns.capabilities")}</div>
            <div>{ux("models.columns.pricingMode")}</div>
            <div>{ux("models.columns.costConfig")}</div>
          </div>
          {props.filteredProviderModels.map((model) => {
            const draft = props.modelDrafts[model.model] ?? {
              selected: false,
              types: model.types ?? [],
              pricingMode: "public" as const,
              inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
              outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
            };

            return (
              <div key={model.model} className="grid grid-cols-[44px_1fr_1fr_160px_160px] items-start gap-2 border-b px-3 py-3 last:border-b-0">
                <div>
                  <input
                    aria-label={`select-${model.model}`}
                    type="checkbox"
                    checked={draft.selected}
                    onChange={(event) =>
                      props.setModelDrafts((prev) => ({
                        ...prev,
                        [model.model]: { ...draft, selected: event.target.checked },
                      }))
                    }
                  />
                </div>
                <div className="min-w-0">
                  <div className="truncate font-medium">{model.display_name || model.model}</div>
                  <div className="text-xs text-muted-foreground">{model.model}</div>
                </div>
                <div className="flex flex-wrap gap-1">
                  {(model.types ?? []).map((capability) => {
                    const checked = draft.types.includes(capability);
                    return (
                      <button
                        key={`${model.model}-${capability}`}
                        type="button"
                        className={`rounded-full border px-2 py-1 text-xs ${checked ? "border-primary bg-primary/10" : "border-border"}`}
                        onClick={() => {
                          props.setModelDrafts((prev) => {
                            const existing = prev[model.model] ?? draft;
                            const exists = existing.types.includes(capability);
                            const nextTypes = exists
                              ? existing.types.filter((item) => item !== capability)
                              : [...existing.types, capability];
                            return {
                              ...prev,
                              [model.model]: {
                                ...existing,
                                types: nextTypes,
                                selected: true,
                              },
                            };
                          });
                        }}
                      >
                        {capability}
                      </button>
                    );
                  })}
                </div>
                <div className="space-y-2">
                  <select
                    className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm"
                    value={draft.pricingMode}
                    onChange={(event) =>
                      props.setModelDrafts((prev) => ({
                        ...prev,
                        [model.model]: {
                          ...draft,
                          selected: true,
                          pricingMode: event.target.value as "public" | "private",
                        },
                      }))
                    }
                  >
                    <option value="public">{ux("models.pricing.public")}</option>
                    <option value="private">{ux("models.pricing.private")}</option>
                  </select>
                </div>
                <div className="space-y-2">
                  {draft.pricingMode === "private" ? (
                    <>
                      <Input
                        value={draft.inputPrice}
                        onChange={(event) =>
                          props.setModelDrafts((prev) => ({
                            ...prev,
                            [model.model]: { ...draft, selected: true, inputPrice: event.target.value },
                          }))
                        }
                        placeholder={ux("models.placeholders.input")}
                      />
                      <Input
                        value={draft.outputPrice}
                        onChange={(event) =>
                          props.setModelDrafts((prev) => ({
                            ...prev,
                            [model.model]: { ...draft, selected: true, outputPrice: event.target.value },
                          }))
                        }
                        placeholder={ux("models.placeholders.output")}
                      />
                    </>
                  ) : (
                    <div className="space-y-1 text-xs text-muted-foreground">
                      <div>{ux("models.catalogPricing")}</div>
                      <div>
                        {ux("models.labels.input")}: {formatCostPerMillion(model.public_pricing?.input_usd_per_1m, ux("models.na"))}
                      </div>
                      <div>
                        {ux("models.labels.output")}: {formatCostPerMillion(model.public_pricing?.output_usd_per_1m, ux("models.na"))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}

function AccessEditor(props: {
  t: ReturnType<typeof useTranslations>;
  managerKind: "user" | "team";
  setManagerKind: (value: "user" | "team") => void;
  managers: AccessPrincipal[];
  setManagers: (value: AccessPrincipal[]) => void;
  userKind: "user" | "team";
  setUserKind: (value: "user" | "team") => void;
  users: AccessPrincipal[];
  setUsers: (value: AccessPrincipal[]) => void;
  onSearchPrincipal: (kind: "user" | "team", query: string) => Promise<AccessPrincipalCandidate[]>;
  onSave: () => void;
  isSaving: boolean;
  effectiveKeyType: "pat" | "vat";
  setEffectiveKeyType: (value: "pat" | "vat") => void;
  effectiveUserID: string;
  setEffectiveUserID: (value: string) => void;
  effectiveTeamID: string;
  setEffectiveTeamID: (value: string) => void;
  effectiveModels: EffectiveModelView[];
  isLoadingEffective: boolean;
  onLoadEffective: () => void;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const hasEffectivePrincipal = Boolean(props.effectiveUserID.trim() || props.effectiveTeamID.trim());

  return (
    <section className="space-y-4 rounded-lg border p-4">
      <div className="flex items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">{ux("tabs.access")}</h3>
          <p className="text-sm text-muted-foreground">{ux("access.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={props.isSaving}>
          {props.isSaving ? props.t("saving") : props.t("saveAccessControl")}
        </Button>
      </div>

      <div className="grid gap-4 xl:grid-cols-2">
        <PrincipalPicker
          title={props.t("step3.managersTitle")}
          kind={props.managerKind}
          setKind={props.setManagerKind}
          onSearchPrincipal={props.onSearchPrincipal}
          selected={props.managers}
          setSelected={props.setManagers}
          allowEveryone={false}
          t={props.t}
        />

        <PrincipalPicker
          title={props.t("step3.usersTitle")}
          kind={props.userKind}
          setKind={props.setUserKind}
          onSearchPrincipal={props.onSearchPrincipal}
          selected={props.users}
          setSelected={props.setUsers}
          allowEveryone
          t={props.t}
        />
      </div>

      <div className="space-y-3 rounded-lg border bg-muted/10 p-3">
        <h4 className="font-medium">{ux("access.effectivePreviewTitle")}</h4>
        <div className="grid gap-3 xl:grid-cols-4">
          <div className="grid gap-1 xl:col-span-1">
            <Label htmlFor="previewKeyType">{ux("access.keyTypeLabel")}</Label>
            <select
              id="previewKeyType"
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              value={props.effectiveKeyType}
              onChange={(event) => props.setEffectiveKeyType(event.target.value as "pat" | "vat")}
            >
              <option value="pat">{ux("access.pat")}</option>
              <option value="vat">{ux("access.vat")}</option>
            </select>
          </div>
          <Field label={ux("access.userIdLabel")} id="previewUserID">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
              <PrincipalSearchInput
                id="previewUserID"
                kind="user"
                selectedID={props.effectiveUserID}
                onSelectedIDChange={props.setEffectiveUserID}
                onSearch={props.onSearchPrincipal}
                loadingLabel={ux("actions.loading")}
                emptyLabel={ux("effectiveModels.empty")}
                placeholder={`${props.t("principal.searchPlaceholder")} (${props.t("principal.user")})`}
                className="flex-1"
              />
              <Button
                type="button"
                variant="ghost"
                onClick={() => props.setEffectiveUserID("")}
                disabled={!props.effectiveUserID.trim()}
                className="w-full sm:w-auto"
              >
                {ux("actions.clear")}
              </Button>
            </div>
          </Field>
          <Field label={ux("access.teamIdLabel")} id="previewTeamID">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
              <PrincipalSearchInput
                id="previewTeamID"
                kind="team"
                selectedID={props.effectiveTeamID}
                onSelectedIDChange={props.setEffectiveTeamID}
                onSearch={props.onSearchPrincipal}
                loadingLabel={ux("actions.loading")}
                emptyLabel={ux("effectiveModels.empty")}
                placeholder={`${props.t("principal.searchPlaceholder")} (${props.t("principal.team")})`}
                className="flex-1"
              />
              <Button
                type="button"
                variant="ghost"
                onClick={() => props.setEffectiveTeamID("")}
                disabled={!props.effectiveTeamID.trim()}
                className="w-full sm:w-auto"
              >
                {ux("actions.clear")}
              </Button>
            </div>
          </Field>
          <div className="flex items-end">
            <Button
              type="button"
              variant="outline"
              onClick={props.onLoadEffective}
              disabled={props.isLoadingEffective || !hasEffectivePrincipal}
              className="w-full xl:w-auto"
            >
              {props.isLoadingEffective ? ux("actions.loading") : ux("access.previewAction")}
            </Button>
          </div>
        </div>
        <EffectiveModelsTable models={props.effectiveModels} />
      </div>
    </section>
  );
}

function CredentialFields(props: {
  authMode: ProviderAuthMode;
  apiKey: string;
  setAPIKey: (value: string) => void;
  secretRef: string;
  setSecretRef: (value: string) => void;
  azureTenantID: string;
  setAzureTenantID: (value: string) => void;
  azureClientID: string;
  setAzureClientID: (value: string) => void;
  awsAccessKeyID: string;
  setAWSAccessKeyID: (value: string) => void;
  awsSecretAccessKey: string;
  setAWSSecretAccessKey: (value: string) => void;
  awsRegion: string;
  setAWSRegion: (value: string) => void;
  googleServiceAccountJSON: string;
  setGoogleServiceAccountJSON: (value: string) => void;
  t: ReturnType<typeof useTranslations>;
}) {
  const ux = useTranslations("providerAccounts.ux");
  return (
    <Card className="border-border/70">
      <CardHeader>
        <CardTitle className="text-base">{ux("setupSteps.accountCredentials")}</CardTitle>
        <CardDescription>{props.t("step2.connectionDescription")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {props.authMode === "api_key" ? (
          <div className="grid gap-3 md:grid-cols-2">
            <Field label={props.t("fields.apiKeyOptional")} id="providerApiKey">
              <Input id="providerApiKey" value={props.apiKey} onChange={(event) => props.setAPIKey(event.target.value)} placeholder={ux("placeholders.apiKey")} />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRef">
              <Input id="providerSecretRef" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefProvider")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "azure_entra_cert" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field label={props.t("fields.azureTenantID")} id="azureTenantID">
              <Input id="azureTenantID" value={props.azureTenantID} onChange={(event) => props.setAzureTenantID(event.target.value)} placeholder={ux("placeholders.tenantId")} />
            </Field>
            <Field label={props.t("fields.azureClientID")} id="azureClientID">
              <Input id="azureClientID" value={props.azureClientID} onChange={(event) => props.setAzureClientID(event.target.value)} placeholder={ux("placeholders.clientId")} />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRefAzure">
              <Input id="providerSecretRefAzure" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefAzure")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "aws_sigv4" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field label={props.t("fields.awsAccessKeyID")} id="awsAccessKeyID">
              <Input id="awsAccessKeyID" value={props.awsAccessKeyID} onChange={(event) => props.setAWSAccessKeyID(event.target.value)} placeholder={ux("placeholders.awsAccessKey")} />
            </Field>
            <Field label={props.t("fields.awsSecretAccessKey")} id="awsSecretAccessKey">
              <Input id="awsSecretAccessKey" value={props.awsSecretAccessKey} onChange={(event) => props.setAWSSecretAccessKey(event.target.value)} placeholder={ux("placeholders.secret")} />
            </Field>
            <Field label={props.t("fields.awsRegion")} id="awsRegion">
              <Input id="awsRegion" value={props.awsRegion} onChange={(event) => props.setAWSRegion(event.target.value)} placeholder={ux("placeholders.awsRegion")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "google_service_account" ? (
          <div className="grid gap-3">
            <Field label={props.t("fields.googleServiceAccountJSON")} id="googleServiceAccountJSON">
              <textarea
                id="googleServiceAccountJSON"
                value={props.googleServiceAccountJSON}
                onChange={(event) => props.setGoogleServiceAccountJSON(event.target.value)}
                placeholder={ux("placeholders.googleServiceJson")}
                className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRefGoogle">
              <Input id="providerSecretRefGoogle" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefGoogle")} />
            </Field>
          </div>
        ) : null}
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
  onSearchPrincipal: (kind: "user" | "team", query: string) => Promise<AccessPrincipalCandidate[]>;
  selected: AccessPrincipal[];
  setSelected: (value: AccessPrincipal[]) => void;
  allowEveryone: boolean;
  t: ReturnType<typeof useTranslations>;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const [candidateID, setCandidateID] = useState("");
  const [candidateOption, setCandidateOption] = useState<PrincipalSearchOption | null>(null);

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

  function addSelectedCandidate() {
    if (!candidateOption?.principal_id) return;
    addPrincipal({
      type: props.kind,
      principal_id: candidateOption.principal_id,
      display_name: candidateOption.display_name,
      secondary: candidateOption.secondary,
    });
    setCandidateID("");
    setCandidateOption(null);
  }

  return (
    <div className="space-y-3 rounded-lg border p-3">
      <h4 className="font-medium">{props.title}</h4>
      <div className="grid gap-2 lg:grid-cols-[120px_minmax(0,1fr)_auto]">
        <select
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          value={props.kind}
          onChange={(event) => {
            props.setKind(event.target.value as "user" | "team");
            setCandidateID("");
            setCandidateOption(null);
          }}
        >
          <option value="user">{props.t("principal.user")}</option>
          <option value="team">{props.t("principal.team")}</option>
        </select>
        <PrincipalSearchInput
          id={`${props.kind}-principal-picker`}
          kind={props.kind}
          selectedID={candidateID}
          onSelectedIDChange={setCandidateID}
          onSelectedOptionChange={setCandidateOption}
          onSearch={props.onSearchPrincipal}
          loadingLabel={ux("actions.loading")}
          emptyLabel={ux("effectiveModels.empty")}
          placeholder={`${props.t("principal.searchPlaceholder")} (${props.t(`principal.${props.kind}`)})`}
          className="w-full"
        />
        <Button
          type="button"
          variant="outline"
          onClick={addSelectedCandidate}
          disabled={!candidateOption?.principal_id}
          className="w-full lg:w-auto"
        >
          {props.t("principal.add")}
        </Button>
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

      <div className="flex flex-wrap gap-2">
        {props.selected.map((item) => (
          <button key={`${item.type}-${item.principal_id ?? "everyone"}`} type="button" className="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs" onClick={() => removePrincipal(item)} title={props.t("principal.remove")}>
            <span>{item.display_name || item.principal_id || item.type}</span>
            <span>{props.t("principal.remove")}</span>
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
      types: normalizeModelTypes(item.types),
      pricingMode: item.pricing_mode ?? "public",
      inputPrice: item.private_pricing?.input_usd_per_1m?.toString() ?? "",
      outputPrice: item.private_pricing?.output_usd_per_1m?.toString() ?? "",
    };
  }
  return next;
}

function normalizeModelTypes(types: string[] | ModelType[] | undefined): ModelType[] {
  if (!Array.isArray(types)) return [];
  const output: ModelType[] = [];
  const seen = new Set<string>();
  for (const item of types) {
    const normalized = String(item ?? "").trim().toLowerCase();
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    output.push(normalized);
  }
  return output;
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

function unwrapData<T>(payload: unknown): T | null {
  if (!payload || typeof payload !== "object") return null;
  const direct = payload as T;
  const candidate = payload as { data?: T };
  if (candidate.data !== undefined) return candidate.data;
  return direct;
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

function formatCostPerMillion(value: number | undefined, naLabel: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return naLabel;
  return `$${value} / 1M`;
}
