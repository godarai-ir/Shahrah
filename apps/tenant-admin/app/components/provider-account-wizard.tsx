"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { ChevronRight, Copy, ExternalLink, Eye, KeyRound, Layers3, Pencil, ShieldCheck, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import {
  buildDefaultDrafts,
  buildDraftsForAccount,
  buildSelectedModelPayload,
} from "@/app/components/provider-account-wizard.models";
import { AccessEditor } from "@/app/components/provider-account-wizard/access-editor";
import {
  mergeAccountIntoList,
} from "@/app/components/provider-account-wizard/account-readiness";
import { CredentialFields } from "@/app/components/provider-account-wizard/credential-fields";
import { EffectiveModelsTable } from "@/app/components/provider-account-wizard/effective-models-table";
import { Field } from "@/app/components/provider-account-wizard/field";
import { ModelsEditor } from "@/app/components/provider-account-wizard/models-editor";
import { EmptyOverview, ReadinessBadge, SummaryKPI } from "@/app/components/provider-account-wizard/overview";
import { buildPrincipalSelectorLabels } from "@/app/components/provider-account-wizard/principal-labels";
import { PROVIDER_DOCS_URLS } from "@/app/components/provider-account-wizard/provider-docs";
import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import type {
  AccessPrincipal,
  AccessPrincipalCandidate,
  AccessPrincipalResponse,
  AccountDetailsResponse,
  AccountsResponse,
  Banner,
  EffectiveModelsResponse,
  EffectiveModelView,
  ModelDraft,
  ProviderAccountWizardMode,
  ProviderAuthMode,
  ProviderCatalogModel,
  ProviderCatalogProvider,
  ProviderCatalogStatsResponse,
  ProviderModelsResponse,
  ProvidersResponse,
  Readiness,
  SetupStep,
  TenantProviderAccount,
  WorkspaceMode,
  WorkspaceTab,
} from "@/app/components/provider-account-wizard/types";
import {
  CreateFlowStepIndicator,
  type CreateWizardSectionID,
  WizardPanel,
} from "@/app/components/provider-account-wizard/wizard-layout";
import {
  extractMessageFromPayload,
  generateProviderAccountName,
  normalizeError,
  normalizeProviderSlug,
  parseResponsePayload,
  requestJSON,
  shouldGenerateProviderAccountName,
  stringsEqualFold,
  toNumberOrUndefined,
  unwrapData,
} from "@/app/components/provider-account-wizard/utils";
import { usePermissions } from "@/app/components/permissions-context";
import { canPermission } from "@/lib/authz";

export default function ProviderAccountWizard(props?: {
  mode?: ProviderAccountWizardMode;
  initialProvider?: string;
  initialAccountID?: string;
}) {
  const mode = props?.mode ?? "manage";
  const isCreateMode = mode === "create";
  const initialProvider = normalizeProviderSlug(props?.initialProvider ?? "");
  const initialAccountID = props?.initialAccountID?.trim() ?? "";
  const t = useTranslations("providerAccounts");
  const ux = useTranslations("providerAccounts.ux");
  const router = useRouter();
  const queryClient = useQueryClient();
  const permissions = usePermissions();
  const canCreateProviderAccount = permissions ? canPermission(permissions, "provider_account:create") : false;
  const canDeleteProviderAccount = permissions ? canPermission(permissions, "provider_account:delete") : false;
  const canManageProviderAccount = permissions
    ? canPermission(permissions, "provider_account:manage") || canPermission(permissions, "provider_account:create")
    : false;

  const [workspaceMode, setWorkspaceMode] = useState<WorkspaceMode>(isCreateMode ? "setup" : "overview");
  const [setupStep, setSetupStep] = useState<SetupStep>(1);
  const [activeTab, setActiveTab] = useState<WorkspaceTab>("models");

  const [isLoading, setIsLoading] = useState(false);
  const [isSavingModels, setIsSavingModels] = useState(false);
  const [isSavingAccess, setIsSavingAccess] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [isSavingCredentials, setIsSavingCredentials] = useState(false);
  const [deletingAccountID, setDeletingAccountID] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteTargetAccount, setDeleteTargetAccount] = useState<TenantProviderAccount | null>(null);
  const [deleteAccountNameInput, setDeleteAccountNameInput] = useState("");
  const [isLoadingEffective, setIsLoadingEffective] = useState(false);
  const [isLoadingProviderModels, setIsLoadingProviderModels] = useState(false);
  const [isActivating, setIsActivating] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [catalogStats, setCatalogStats] = useState<{ cacheTTL: number; counters: Record<string, number> }>({
    cacheTTL: 0,
    counters: {},
  });

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);

  const [selectedProviderID, setSelectedProviderID] = useState(initialProvider);
  const [selectedAccountID, setSelectedAccountID] = useState(initialAccountID);
  const selectedProviderRef = useRef(initialProvider);
  const selectedAccountRef = useRef(initialAccountID);
  const accountLoadRequestRef = useRef(0);

  const [account, setAccount] = useState<TenantProviderAccount | null>(null);

  const [provider, setProvider] = useState(initialProvider);
  const [name, setName] = useState("");
  const [isNameAutoGenerated, setIsNameAutoGenerated] = useState(true);
  const [authMode, setAuthMode] = useState<ProviderAuthMode>("api_key");
  const [allowVirtualOnly, setAllowVirtualOnly] = useState(isCreateMode);

  const [apiKey, setAPIKey] = useState("");
  const [apiKeyTouched, setAPIKeyTouched] = useState(false);
  const [apiKeyTestSucceeded, setAPIKeyTestSucceeded] = useState(false);
  const [secretRef, setSecretRef] = useState("");
  const [azureTenantID, setAzureTenantID] = useState("");
  const [azureClientID, setAzureClientID] = useState("");
  const [awsAccessKeyID, setAWSAccessKeyID] = useState("");
  const [awsSecretAccessKey, setAWSSecretAccessKey] = useState("");
  const [awsRegion, setAWSRegion] = useState("us-east-1");
  const [googleServiceAccountJSON, setGoogleServiceAccountJSON] = useState("");

  const [modelDrafts, setModelDrafts] = useState<Record<string, ModelDraft>>({});

  const [managers, setManagers] = useState<AccessPrincipal[]>([]);
  const [users, setUsers] = useState<AccessPrincipal[]>([{ type: "everyone", display_name: t("principal.everyone") }]);

  const [effectiveKeyType, setEffectiveKeyType] = useState<"pat" | "vat">("pat");
  const [effectiveUserID, setEffectiveUserID] = useState("");
  const [effectiveTeamID, setEffectiveTeamID] = useState("");
  const [effectivePrincipalSelection, setEffectivePrincipalSelection] = useState<PrincipalEntitySelection | null>(null);
  const [effectiveModels, setEffectiveModels] = useState<EffectiveModelView[]>([]);
  const [lastCredentialTestMessage, setLastCredentialTestMessage] = useState("");
  const [createCredentialTestResult, setCreateCredentialTestResult] = useState<Banner | null>(null);
  const [activeCreateSection, setActiveCreateSection] = useState<CreateWizardSectionID>("details");

  useEffect(() => {
    if (!isCreateMode || banner?.kind !== "success") return;
    toast.success(banner.message, { duration: 3500 });
    const timeout = window.setTimeout(() => setBanner(null), 0);
    return () => window.clearTimeout(timeout);
  }, [banner, isCreateMode]);

  const handleEffectivePrincipalChange = useCallback((selection: PrincipalEntitySelection | null) => {
    setEffectivePrincipalSelection(selection);
    if (selection?.type === "user") {
      setEffectiveUserID((selection.principal_id ?? "").trim());
      setEffectiveTeamID("");
      return;
    }
    if (selection?.type === "team") {
      setEffectiveUserID("");
      setEffectiveTeamID((selection.principal_id ?? "").trim());
      return;
    }
    setEffectiveUserID("");
    setEffectiveTeamID("");
    setEffectiveModels([]);
  }, []);

  const visibleProviders = useMemo(
    () => providers.filter((item) => item.enabled && stringsEqualFold(item.provider, "azure") === false),
    [providers],
  );

  const selectedProvider = useMemo(
    () => providers.find((item) => stringsEqualFold(item.provider, selectedProviderID)) ?? null,
    [providers, selectedProviderID],
  );
  const activeProvider = useMemo(
    () => providers.find((item) => stringsEqualFold(item.provider, provider)) ?? selectedProvider,
    [providers, provider, selectedProvider],
  );

  const selectedProviderAccounts = useMemo(
    () => accounts.filter((item) => stringsEqualFold(item.provider, selectedProviderID)),
    [accounts, selectedProviderID],
  );
  const existingAccountNames = useMemo(() => accounts.map((item) => item.name), [accounts]);

  const fetchProviders = useCallback(async () => {
    return queryClient.fetchQuery({
      queryKey: ["tenant-admin", "provider-catalog", "providers"],
      queryFn: async () => {
        const providerData = await requestJSON<ProvidersResponse>("/api/provider-catalog/providers");
        return unwrapData<ProviderCatalogProvider[]>(providerData) ?? [];
      },
      staleTime: 1000 * 60 * 30,
      gcTime: 1000 * 60 * 60 * 2,
    });
  }, [queryClient]);

  useEffect(() => {
    selectedProviderRef.current = selectedProviderID;
  }, [selectedProviderID]);

  useEffect(() => {
    selectedAccountRef.current = selectedAccountID;
  }, [selectedAccountID]);

  const selectedModelCount = useMemo(
    () => Object.values(modelDrafts).filter((draft) => draft.enabled).length,
    [modelDrafts],
  );

  const unresolvedPricingCount = useMemo(() => {
    let unresolved = 0;
    for (const model of providerModels) {
      const draft = modelDrafts[model.model];
      if (!draft?.enabled) continue;
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
  const storedAPIKeyHint = account?.credential_hint ?? "";
  const shouldShowStoredAPIKey = authMode === "api_key" && apiKeyTestSucceeded && !apiKeyTouched && !!storedAPIKeyHint;

  const fetchProviderModels = useCallback(async (nextProvider: string) => {
    const normalized = nextProvider.trim().toLowerCase();
    if (!normalized) return [];
    return queryClient.fetchQuery({
      queryKey: ["tenant-admin", "provider-catalog", "models", normalized],
      queryFn: async () => {
        const response = await requestJSON<ProviderModelsResponse>(
          `/api/provider-catalog/providers/${encodeURIComponent(normalized)}/models`,
        );
        return unwrapData<ProviderCatalogModel[]>(response) ?? [];
      },
      staleTime: 1000 * 60 * 30,
      gcTime: 1000 * 60 * 60 * 2,
    });
  }, [queryClient]);

  const loadProviderModels = useCallback(
    async (nextProvider: string) => {
      const normalized = nextProvider.trim().toLowerCase();
      if (!normalized) {
        setProviderModels([]);
        setModelDrafts({});
        return [];
      }
      setIsLoadingProviderModels(true);
      try {
        const items = await fetchProviderModels(normalized);
        setProviderModels(items);
        setModelDrafts((prev) => buildDefaultDrafts(items, prev));
        return items;
      } catch (error) {
        setProviderModels([]);
        setModelDrafts({});
        setBanner({
          kind: "error",
          message: normalizeError(error, t("messages.providerModelsLoadFailed")),
        });
        return [];
      } finally {
        setIsLoadingProviderModels(false);
      }
    },
    [fetchProviderModels, t],
  );

  const loadInitial = useCallback(async () => {
    setIsLoading(true);
    setBanner(null);
    try {
      const currentProviderID = selectedProviderRef.current;
      const currentAccountID = selectedAccountRef.current;
      const providerItems = await fetchProviders();
      const accountData = await requestJSON<AccountsResponse>("/api/tenant-provider-accounts");
      const accountItems = unwrapData<TenantProviderAccount[]>(accountData) ?? [];
      const fetchedAccountNames = accountItems.map((item) => item.name);
      setProviders(providerItems);
      setAccounts(accountItems);

      if (isCreateMode) {
        const createProviders = providerItems.filter(
          (item) => item.enabled && stringsEqualFold(item.provider, "azure") === false,
        );
        const initialProvider = currentProviderID || createProviders[0]?.provider || "";
        setSelectedProviderID(initialProvider);
        setSelectedAccountID("");
        setAccount(null);
        setWorkspaceMode("setup");
        setSetupStep(currentProviderID ? 2 : 1);
        setManagers([]);
        setUsers([{ type: "everyone", display_name: t("principal.everyone") }]);
        if (initialProvider) {
          const providerItem = createProviders.find((item) => stringsEqualFold(item.provider, initialProvider));
          setProvider(initialProvider);
          setName(generateProviderAccountName(initialProvider, fetchedAccountNames));
          setIsNameAutoGenerated(true);
          setAuthMode(providerItem?.auth_modes?.[0] ?? "api_key");
          setAllowVirtualOnly(true);
          void loadProviderModels(initialProvider);
        }
        return;
      }

      const initialAccount = currentAccountID
        ? accountItems.find((item) => item.id === currentAccountID) ?? null
        : null;
      const initialProvider = currentProviderID || initialAccount?.provider || accountItems[0]?.provider || "";
      setSelectedProviderID(initialProvider);

      if (initialAccount) {
        setSelectedAccountID(initialAccount.id);
        setAccount(initialAccount);
        setProvider(normalizeProviderSlug(initialAccount.provider));
        setName(initialAccount.name);
        setIsNameAutoGenerated(false);
        setAuthMode(initialAccount.auth_mode);
        setAllowVirtualOnly(initialAccount.allows_virtual_keys_only);
        setWorkspaceMode("workspace");
        setActiveTab("models");
        return;
      }

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
  }, [fetchProviders, isCreateMode, loadProviderModels, t]);

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
      const requestID = ++accountLoadRequestRef.current;
      setIsLoading(true);
      setIsLoadingProviderModels(true);
      setBanner(null);
      try {
        const details = await requestJSON<AccountDetailsResponse>(
          `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}`,
        );
        const payload = unwrapData<AccountDetailsResponse>(details) ?? details;

        if (requestID !== accountLoadRequestRef.current) return;

        let catalogModels: ProviderCatalogModel[] = [];
        try {
          catalogModels = await fetchProviderModels(payload.account.provider);
        } catch (error) {
          if (requestID !== accountLoadRequestRef.current) return;
          setBanner({
            kind: "error",
            message: normalizeError(error, t("messages.providerModelsLoadFailed")),
          });
        }

        if (requestID !== accountLoadRequestRef.current) return;

        setAccount(payload.account);
        setAccounts((prev) => mergeAccountIntoList(prev, payload.account));
        setProviderModels(catalogModels);
        setProvider(normalizeProviderSlug(payload.account.provider));
        setName(payload.account.name);
        setIsNameAutoGenerated(false);
        setAuthMode(payload.account.auth_mode);
        setAllowVirtualOnly(payload.account.allows_virtual_keys_only);
        const cfg = payload.account.provider_config ?? {};
        setAzureTenantID(typeof cfg.azure_tenant_id === "string" ? cfg.azure_tenant_id : "");
        setAzureClientID(typeof cfg.azure_client_id === "string" ? cfg.azure_client_id : "");
        setAWSRegion(typeof cfg.aws_region === "string" ? cfg.aws_region : "us-east-1");
        setAWSAccessKeyID(typeof cfg.aws_access_key_id === "string" ? cfg.aws_access_key_id : "");
        setAWSSecretAccessKey(typeof cfg.aws_secret_access_key === "string" ? cfg.aws_secret_access_key : "");
        setGoogleServiceAccountJSON(typeof cfg.google_service_account_json === "string" ? cfg.google_service_account_json : "");
        setAPIKey("");
        setAPIKeyTouched(false);
        setAPIKeyTestSucceeded(false);
        setSecretRef("");
        setManagers(payload.access?.managers ?? []);
        setUsers(payload.access?.users ?? []);
        setSelectedProviderID(normalizeProviderSlug(payload.account.provider));
        setSelectedAccountID(payload.account.id);
        setModelDrafts(buildDraftsForAccount(payload.models ?? [], catalogModels));
      } catch (error) {
        if (requestID === accountLoadRequestRef.current) {
          setBanner({ kind: "error", message: normalizeError(error, t("messages.accountLoadFailed")) });
        }
      } finally {
        if (requestID === accountLoadRequestRef.current) {
          setIsLoading(false);
          setIsLoadingProviderModels(false);
        }
      }
    },
    [fetchProviderModels, t],
  );

  const normalizeDeleteAccountError = useCallback(
    (message: string, status: number): string => {
      const normalized = message.trim().toLowerCase();
      if (status === 403 || normalized.includes("insufficient permissions")) return t("messages.accountDeleteForbidden");
      if (status === 404 || normalized.includes("not found")) return t("messages.accountDeleteNotFound");
      if (status === 409 || normalized.includes("active keys") || normalized.includes("referenced")) {
        return t("messages.accountDeleteConflict");
      }
      return message || t("messages.accountDeleteFailed");
    },
    [t],
  );

  const openDeleteAccountDialog = useCallback((target: TenantProviderAccount) => {
    if (!canDeleteProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return;
    }
    setDeleteTargetAccount(target);
    setDeleteAccountNameInput("");
    setDeleteDialogOpen(true);
  }, [canDeleteProviderAccount, t]);

  const handleDeleteAccount = useCallback(
    async (target: TenantProviderAccount) => {
      if (!canDeleteProviderAccount) {
        setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
        return;
      }
      if (!target?.id) return;
      if (deleteAccountNameInput !== target.name) {
        setBanner({ kind: "error", message: t("messages.accountDeleteNameMismatch") });
        return;
      }

      setDeletingAccountID(target.id);
      try {
        const response = await fetch(`/api/tenant-provider-accounts/${encodeURIComponent(target.id)}`, {
          method: "DELETE",
          cache: "no-store",
        });
        const text = await response.text();
        const payload = parseResponsePayload(text);
        if (!response.ok) {
          const upstreamMessage = extractMessageFromPayload(payload) ?? `Request failed (${response.status})`;
          throw new Error(normalizeDeleteAccountError(upstreamMessage, response.status));
        }

        const remainingAccounts = accounts.filter((item) => item.id !== target.id);
        setAccounts(remainingAccounts);

        if (selectedAccountID === target.id) {
          const nextAccount = remainingAccounts.find((item) => stringsEqualFold(item.provider, target.provider)) ?? remainingAccounts[0];
          if (nextAccount) {
            setSelectedAccountID(nextAccount.id);
            setSelectedProviderID(normalizeProviderSlug(nextAccount.provider));
            setWorkspaceMode("workspace");
          } else {
            setSelectedAccountID("");
            setSelectedProviderID("");
            setAccount(null);
            setProviderModels([]);
            setModelDrafts({});
            setManagers([]);
            setUsers([{ type: "everyone", display_name: t("principal.everyone") }]);
            setWorkspaceMode("overview");
          }
        }
        setBanner({ kind: "success", message: t("messages.accountDeleteSuccess") });
        setDeleteDialogOpen(false);
        setDeleteTargetAccount(null);
        setDeleteAccountNameInput("");
      } catch (error) {
        setBanner({ kind: "error", message: normalizeError(error, t("messages.accountDeleteFailed")) });
      } finally {
        setDeletingAccountID("");
      }
    },
    [accounts, canDeleteProviderAccount, deleteAccountNameInput, normalizeDeleteAccountError, selectedAccountID, t],
  );

  async function handleSaveCredentials(accountID = selectedAccountID) {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }

    const providerConfig: Record<string, unknown> = {};
    const credentials: Record<string, unknown> = {};

    if (authMode === "api_key") {
      if (!allowVirtualOnly && apiKey.trim()) credentials.api_key = apiKey.trim();
      if (!isCreateMode && secretRef.trim()) credentials.secret_ref = secretRef.trim();
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

    setIsSavingCredentials(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/credentials`,
        {
          method: "PUT",
          body: JSON.stringify({
            auth_mode: authMode,
            allow_virtual_keys_only: allowVirtualOnly,
            provider_config: providerConfig,
            credentials: Object.keys(credentials).length > 0 ? credentials : undefined,
          }),
        },
      );
      setBanner({ kind: "success", message: t("messages.accountLoadSuccess") });
      await handleLoadAccount(accountID);
      setAPIKeyTouched(false);
      setAPIKeyTestSucceeded(false);
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accountLoadFailed")) });
      return false;
    } finally {
      setIsSavingCredentials(false);
    }
  }

  useEffect(() => {
    if (!selectedAccountID || workspaceMode !== "workspace") return;
    const timer = setTimeout(() => {
      void handleLoadAccount(selectedAccountID);
    }, 0);
    return () => clearTimeout(timer);
  }, [selectedAccountID, workspaceMode, handleLoadAccount]);

  function resetCredentialInputs() {
    setAPIKey("");
    setAPIKeyTouched(false);
    setAPIKeyTestSucceeded(false);
    setSecretRef("");
    setAzureTenantID("");
    setAzureClientID("");
    setAWSAccessKeyID("");
    setAWSSecretAccessKey("");
    setAWSRegion("us-east-1");
    setGoogleServiceAccountJSON("");
  }

  function handleSelectAccount(item: TenantProviderAccount) {
    const shouldReloadCurrentAccount = item.id === selectedAccountID && workspaceMode === "workspace";
    accountLoadRequestRef.current += 1;
    setSelectedProviderID(normalizeProviderSlug(item.provider));
    setSelectedAccountID(item.id);
    setAccount(item);
    setProvider(normalizeProviderSlug(item.provider));
    setName(item.name);
    setIsNameAutoGenerated(false);
    setAuthMode(item.auth_mode);
    setAllowVirtualOnly(item.allows_virtual_keys_only);
    resetCredentialInputs();
    setManagers([]);
    setUsers([]);
    setEffectiveModels([]);
    setLastCredentialTestMessage("");
    setModelDrafts({});
    setWorkspaceMode("workspace");
    setActiveTab("models");
    if (shouldReloadCurrentAccount) void handleLoadAccount(item.id);
  }

  async function handleCreateAccount(options: { advanceToModels?: boolean; successMessage?: string } = {}) {
    const advanceToModels = options.advanceToModels ?? true;
    setBanner(null);
    if (!canCreateProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return null;
    }
    if (!provider || !name.trim()) {
      setBanner({ kind: "error", message: t("messages.validationProviderName") });
      return null;
    }

    const providerConfig: Record<string, unknown> = {};
    const credentials: Record<string, unknown> = {};

    if (authMode === "api_key") {
      if (!allowVirtualOnly && apiKey.trim()) credentials.api_key = apiKey.trim();
      if (!isCreateMode && secretRef.trim()) credentials.secret_ref = secretRef.trim();
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
      if (advanceToModels) {
        setSetupStep(3);
        setActiveCreateSection("models");
      }
      setBanner({ kind: "success", message: options.successMessage ?? t("messages.accountCreateSuccess") });
      return newAccount.id;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accountCreateFailed")) });
      return null;
    } finally {
      setIsLoading(false);
    }
  }

  async function handleCreateCredentialTest() {
    setCreateCredentialTestResult(null);
    setAPIKeyTestSucceeded(false);
    if (allowVirtualOnly) return;
    if (!apiKey.trim()) {
      setCreateCredentialTestResult({
        kind: "error",
        message: ux("integrations.wizard.apiKeyRequiredForTest"),
      });
      return;
    }

    let accountID = selectedAccountID;
    if (accountID) {
      const saved = await handleSaveCredentials(accountID);
      if (!saved) return;
    } else {
      const createdID = await handleCreateAccount({
        advanceToModels: false,
        successMessage: ux("integrations.wizard.accountSavedBeforeTest"),
      });
      if (!createdID) return;
      accountID = createdID;
    }

    setIsTesting(true);
    try {
      const result = await requestJSON<{ success: boolean; message?: string }>(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/test-credentials`,
        { method: "POST", body: JSON.stringify({}) },
      );
      const payload = unwrapData<{ success: boolean; message?: string }>(result) ?? result;
      const rawMessage =
        payload.message ?? (payload.success ? t("messages.credentialValid") : t("messages.credentialInvalid"));
      const message = payload.success
        ? ux("integrations.wizard.connectionTestSuccess", { message: rawMessage })
        : ux("integrations.wizard.connectionTestFailed", { message: rawMessage });
      setLastCredentialTestMessage(rawMessage);
      setCreateCredentialTestResult({ kind: payload.success ? "success" : "error", message });
      setAPIKeyTouched(!payload.success);
      setAPIKeyTestSucceeded(payload.success);
    } catch (error) {
      setAPIKeyTestSucceeded(false);
      const message = normalizeError(error, t("messages.credentialTestFailed"));
      setCreateCredentialTestResult({
        kind: "error",
        message: ux("integrations.wizard.connectionTestFailed", { message }),
      });
    } finally {
      setIsTesting(false);
    }
  }

  async function handleContinueFromCreateDetails() {
    if (!selectedAccountID) {
      await handleCreateAccount();
      return;
    }

    const saved = await handleSaveCredentials(selectedAccountID);
    if (!saved) return;
    setSetupStep(3);
    setActiveCreateSection("models");
  }

  const handleSaveModels = useCallback(async (accountID = selectedAccountID): Promise<boolean> => {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }

    const { models: payloadModels, validationErrors } = buildSelectedModelPayload(modelDrafts);

    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: t("messages.selectAtLeastOneModel") });
      return false;
    }
    if (validationErrors.length > 0) {
      const firstError = validationErrors[0];
      const reasonMessage =
        firstError.reason === "effective_capabilities_required"
          ? ux("messages.modelTypeRequired")
          : ux("messages.privatePricingRequired");
      setBanner({ kind: "error", message: `${firstError.model}: ${reasonMessage}` });
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
  }, [canManageProviderAccount, handleLoadAccount, modelDrafts, selectedAccountID, t, ux]);

  async function handleFinishCreateFlow() {
    await handleActivateAccount();
  }

  async function handleContinueFromCreateModels() {
    const saved = await handleSaveModels(selectedAccountID);
    if (!saved) return;
    setActiveCreateSection("access");
  }

  async function handleContinueFromCreateAccess() {
    const saved = await handleSaveAccess(selectedAccountID);
    if (!saved) return;
    setActiveCreateSection("review");
  }

  const handleSaveAccess = useCallback(async (accountID = selectedAccountID): Promise<boolean> => {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
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
  }, [canManageProviderAccount, handleLoadAccount, managers, selectedAccountID, t, users]);

  async function handleTestCredentials() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }
    if (!allowVirtualOnly && apiKeyTouched) {
      if (!apiKey.trim()) {
        setBanner({ kind: "error", message: t("messages.credentialInvalid") });
        return;
      }
      const saved = await handleSaveCredentials(selectedAccountID);
      if (!saved) return;
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
      setAPIKeyTouched(!payload.success);
      setAPIKeyTestSucceeded(payload.success);
      await loadInitial();
    } catch (error) {
      setAPIKeyTestSucceeded(false);
      setBanner({ kind: "error", message: normalizeError(error, t("messages.credentialTestFailed")) });
    } finally {
      setIsTesting(false);
    }
  }

  async function handleActivateAccount() {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return;
    }
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
      if (isCreateMode) {
        router.push("/provider-accounts");
        return;
      }
      setWorkspaceMode("workspace");
      setActiveTab("models");
      setBanner({ kind: "success", message: ux("messages.accountActivated") });
    } finally {
      setIsActivating(false);
    }
  }

  const searchPrincipalCandidates = useCallback(async (kind: PrincipalSearchKind, query: string) => {
    if (kind === "external_identity") return [];
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

  const loadEffectiveModelsPreview = useCallback(async (accountID = selectedAccountID) => {
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
  }, [effectiveKeyType, effectiveTeamID, effectiveUserID, selectedAccountID]);

  const handleSaveModelsAction = useCallback(() => {
    void handleSaveModels();
  }, [handleSaveModels]);

  const handleSaveAccessAction = useCallback(() => {
    void handleSaveAccess();
  }, [handleSaveAccess]);

  const handleLoadEffectiveModelsAction = useCallback(() => {
    void loadEffectiveModelsPreview();
  }, [loadEffectiveModelsPreview]);

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
  const principalSelectorLabels = useMemo(() => buildPrincipalSelectorLabels(t, ux), [t, ux]);
  const providerDocsURL = PROVIDER_DOCS_URLS[normalizeProviderSlug(provider)] ?? "";
  const createProviderDescription =
    activeProvider && "description" in activeProvider && typeof activeProvider.description === "string"
      ? activeProvider.description
      : ux("integrations.providerDescriptionFallback");
  const createHeaderProviderName = activeProvider?.display_name ?? provider;
  const createWizardSteps: Array<{ id: CreateWizardSectionID; number: number; title: string }> = [
    { id: "details", number: 1, title: ux("integrations.wizard.integrationDetails") },
    { id: "models", number: 2, title: ux("integrations.wizard.modelProvisioning") },
    { id: "access", number: 3, title: ux("integrations.wizard.accessPermissions") },
    { id: "review", number: 4, title: ux("integrations.wizard.reviewConfirm") },
  ];
  const activeCreateStepIndex = createWizardSteps.findIndex((step) => step.id === activeCreateSection);
  const isModelsCreateSection = activeCreateSection === "models";
  const createStepState = (stepID: CreateWizardSectionID) => {
    if (activeCreateSection === stepID) return "active";
    if (stepID === "details" && !!selectedAccountID) return "complete";
    if (stepID === "models" && selectedModelCount > 0) return "complete";
    if (stepID === "access" && managers.length > 0) return "complete";
    return "idle";
  };
  const canOpenCreateStep = (stepID: CreateWizardSectionID) => {
    const stepIndex = createWizardSteps.findIndex((step) => step.id === stepID);
    return stepIndex >= 0 && stepIndex <= activeCreateStepIndex;
  };

  if (isCreateMode) {
    return (
      <main className="-m-4 flex min-h-[calc(100dvh-4rem)] md:-m-6 lg:h-[calc(100dvh-5.25rem)] lg:min-h-[560px]">
        <section className="flex w-full flex-col overflow-hidden rounded-xl border border-border bg-background">
          <header className="flex h-14 shrink-0 items-center border-b border-border px-4">
            <div className="flex min-w-0 items-center gap-2 text-xl">
              <span className="text-muted-foreground">{ux("integrations.pageTitle")}</span>
              <ChevronRight className="size-4 text-muted-foreground rtl:rotate-180" />
              {provider ? (
                <span className="flex min-w-0 items-center gap-2 font-semibold" dir="ltr">
                  <ProviderLogo
                    provider={activeProvider?.provider ?? provider}
                    displayName={createHeaderProviderName}
                    logoURL={activeProvider?.logo_url}
                    size={18}
                  />
                  <span className="truncate">{createHeaderProviderName}</span>
                </span>
              ) : (
                <span className="font-semibold">{ux("setup.selectProviderTitle")}</span>
              )}
            </div>
          </header>

          {banner && banner.kind !== "success" ? (
            <div className="shrink-0 px-4 pt-3">
              <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
                <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
                <AlertDescription>{banner.message}</AlertDescription>
              </Alert>
            </div>
          ) : null}

          <div className="grid flex-1 gap-y-6 overflow-hidden px-4 py-6 lg:min-h-0 lg:grid-cols-[200px_minmax(0,1fr)] lg:gap-x-10 xl:gap-x-12">
            <aside className="min-h-0 lg:self-stretch">
              <ScrollArea className="min-h-[5.5rem] lg:h-full [&_[data-slot=scroll-area-viewport]]:overscroll-contain">
                <div className="flex flex-row-reverse gap-3 py-2 lg:flex-col lg:gap-5">
                {createWizardSteps.map((step) => {
                  const stepState = createStepState(step.id);
                  const isStepOpenable = canOpenCreateStep(step.id);
                  return (
                    <CreateFlowStepIndicator
                      key={step.id}
                      number={step.number}
                      title={step.title}
                      state={stepState}
                      disabled={!isStepOpenable}
                      onClick={() => {
                        if (isStepOpenable) setActiveCreateSection(step.id);
                      }}
                    />
                  );
                })}
                </div>
              </ScrollArea>
            </aside>

            <div className={isModelsCreateSection ? "min-h-[640px] min-w-0 lg:ms-2 lg:flex lg:h-full lg:min-h-0 lg:flex-col" : "min-h-[640px] min-w-0 overflow-y-auto lg:ms-2 lg:h-full lg:min-h-0"}>
              <div className={isModelsCreateSection ? "mx-auto flex w-full max-w-5xl min-h-0 flex-1 flex-col gap-4 pb-2" : "mx-auto flex w-full max-w-5xl flex-col gap-8 pb-2"}>
                {!provider ? (
                  <WizardPanel
                    title={ux("setup.selectProviderTitle")}
                    description={ux("setup.selectProviderDesc")}
                    footer={
                      <>
                        <Button type="button" variant="outline" asChild>
                          <Link href="/provider-accounts">{ux("actions.cancelDelete")}</Link>
                        </Button>
                        <Button type="button" onClick={() => setSetupStep(2)} disabled={!provider}>
                          {ux("actions.continue")}
                        </Button>
                      </>
                    }
                  >
                    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                      {visibleProviders.map((item) => {
                        const isSelected = stringsEqualFold(item.provider, provider);
                        return (
                          <button
                            key={item.provider}
                            type="button"
                            className={`rounded-xl border p-4 text-start transition ${isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"}`}
                            onClick={() => {
                              const nextProvider = normalizeProviderSlug(item.provider);
                              setProvider(nextProvider);
                              setAuthMode(item.auth_modes?.[0] ?? "api_key");
                              setName(generateProviderAccountName(nextProvider, existingAccountNames));
                              setIsNameAutoGenerated(true);
                              setSelectedProviderID(nextProvider);
                              setAllowVirtualOnly(true);
                              setAPIKey("");
                              setCreateCredentialTestResult(null);
                              void loadProviderModels(item.provider);
                            }}
                          >
                            <ProviderLogo
                              provider={item.provider}
                              displayName={item.display_name}
                              logoURL={item.logo_url}
                            />
                            <p className="mt-4 font-medium" dir="ltr">{item.display_name}</p>
                          </button>
                        );
                      })}
                    </div>
                  </WizardPanel>
                ) : (
                  <>
                    {activeCreateSection === "details" ? (
                      <WizardPanel
                    id="connection-details"
                    title={ux("integrations.wizard.detailsTitle")}
                    description={ux("integrations.wizard.detailsDescription")}
                    fillHeight={false}
                    footer={
                      <>
                        <Button type="button" variant="outline" asChild>
                          <Link href="/provider-accounts">{ux("actions.cancelDelete")}</Link>
                        </Button>
                        <Button type="button" onClick={() => void handleContinueFromCreateDetails()} disabled={!canCreateProviderAccount || isLoading || isSavingCredentials || !provider || !name.trim()}>
                          {isLoading || isSavingCredentials ? t("saving") : ux("integrations.wizard.nextStep")}
                        </Button>
                      </>
                    }
                  >
                    <div className="flex flex-col gap-8">
                      <div dir="rtl" className="flex items-start gap-3 text-right">
                        <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border bg-card">
                          <ProviderLogo
                            provider={activeProvider?.provider ?? provider}
                            displayName={activeProvider?.display_name ?? provider}
                            logoURL={activeProvider?.logo_url}
                          />
                        </div>
                        <div className="min-w-0">
                          <h3 className="truncate text-base font-semibold" dir="auto">{activeProvider?.display_name ?? provider}</h3>
                          <p className="mt-1 text-sm leading-6 text-muted-foreground">{createProviderDescription}</p>
                        </div>
                      </div>

                      <div className="grid gap-5">
                        <Field label={t("fields.accountName")} id="accountName">
                          <Input
                            id="accountName"
                            value={name}
                            onChange={(event) => {
                              setName(event.target.value);
                              setIsNameAutoGenerated(false);
                            }}
                            placeholder={ux("placeholders.accountName")}
                          />
                        </Field>
                      </div>

                      <div dir="rtl" className="text-right">
                        <div className="space-y-1">
                          <h3 className="text-base font-semibold">{ux("integrations.wizard.useOwnApiKeyQuestion")}</h3>
                          <p className="text-sm leading-6 text-muted-foreground">{ux("integrations.wizard.useOwnApiKeyDescription")}</p>
                        </div>

                        <div className="mt-4 grid gap-3 md:grid-cols-2">
                          <label className={`flex cursor-pointer gap-3 rounded-lg border p-4 text-right transition ${!allowVirtualOnly ? "border-primary bg-primary/10" : "border-border hover:bg-muted/40"}`}>
                            <input
                              type="radio"
                              name="create-own-api-key"
                              className="mt-1"
                              checked={!allowVirtualOnly}
                              onChange={() => {
                                setAllowVirtualOnly(false);
                                setCreateCredentialTestResult(null);
                              }}
                            />
                            <span className="min-w-0">
                              <span className="block text-sm font-medium">{ux("integrations.wizard.useOwnApiKeyYes")}</span>
                            </span>
                          </label>
                          <label className={`flex cursor-pointer gap-3 rounded-lg border p-4 text-right transition ${allowVirtualOnly ? "border-primary bg-primary/10" : "border-border hover:bg-muted/40"}`}>
                            <input
                              type="radio"
                              name="create-own-api-key"
                              className="mt-1"
                              checked={allowVirtualOnly}
                              onChange={() => {
                                setAllowVirtualOnly(true);
                                setCreateCredentialTestResult(null);
                              }}
                            />
                            <span className="min-w-0">
                              <span className="block text-sm font-medium">{ux("integrations.wizard.useOwnApiKeyNo")}</span>
                              <span className="mt-1 block text-sm leading-6 text-muted-foreground">
                                {ux("integrations.wizard.useOwnApiKeyNoDescription")}
                              </span>
                            </span>
                          </label>
                        </div>

                        {!allowVirtualOnly ? (
                          <div className="mt-5 space-y-4">
                            {providerDocsURL ? (
                              <div className="flex justify-end">
                                <Button type="button" variant="link" asChild className="h-auto px-0">
                                  <a href={providerDocsURL} target="_blank" rel="noreferrer">
                                    {ux("integrations.wizard.grabApiKey")}
                                    <ExternalLink data-icon="inline-end" />
                                  </a>
                                </Button>
                              </div>
                            ) : null}
                            <Field label={ux("integrations.wizard.apiKeyLabel")} id="providerApiKey">
                              {shouldShowStoredAPIKey ? (
                                <div className="flex h-10 items-center justify-between gap-2 rounded-md border border-input bg-muted/30 px-3">
                                  <span className="min-w-0 truncate font-mono text-sm text-foreground" dir="ltr">
                                    {storedAPIKeyHint}
                                  </span>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 shrink-0"
                                    onClick={() => {
                                      setAPIKeyTouched(true);
                                      setAPIKeyTestSucceeded(false);
                                      setAPIKey("");
                                      setCreateCredentialTestResult(null);
                                    }}
                                    aria-label={ux("credentials.editStoredApiKey")}
                                  >
                                    <Pencil className="h-4 w-4" />
                                  </Button>
                                </div>
                              ) : (
                                <Input
                                  id="providerApiKey"
                                  value={apiKey}
                                  onChange={(event) => {
                                    setAPIKey(event.target.value);
                                    setAPIKeyTouched(true);
                                    setAPIKeyTestSucceeded(false);
                                    setCreateCredentialTestResult(null);
                                  }}
                                  placeholder={ux("placeholders.apiKey")}
                                />
                              )}
                            </Field>
                            <div className="flex flex-wrap items-center justify-between gap-3">
                              <p className="text-sm leading-6 text-muted-foreground">
                                {ux("integrations.wizard.testConnectionDescription")}
                              </p>
                              <Button
                                type="button"
                                variant="outline"
                                onClick={() => void handleCreateCredentialTest()}
                                disabled={!canCreateProviderAccount || isLoading || isSavingCredentials || isTesting}
                              >
                                {isTesting || isSavingCredentials || isLoading
                                  ? ux("integrations.wizard.testingConnection")
                                  : ux("integrations.wizard.testConnection")}
                              </Button>
                            </div>
                            {createCredentialTestResult ? (
                              <Alert variant={createCredentialTestResult.kind === "error" ? "destructive" : "default"}>
                                <AlertTitle>
                                  {createCredentialTestResult.kind === "error"
                                    ? ux("integrations.wizard.connectionTestFailedTitle")
                                    : ux("integrations.wizard.connectionTestSuccessTitle")}
                                </AlertTitle>
                                <AlertDescription>{createCredentialTestResult.message}</AlertDescription>
                              </Alert>
                            ) : null}
                          </div>
                        ) : null}
                      </div>
                    </div>
                  </WizardPanel>
                    ) : null}

                  {activeCreateSection === "models" ? (
                    <WizardPanel
                    id="model-provisioning"
                    title={ux("integrations.wizard.modelsHeaderTitle")}
                    description={ux("integrations.wizard.modelsHeaderDescription")}
                    fillHeight
                    contentClassName="flex min-h-0 flex-1 flex-col p-0"
                    footer={
                      <>
                        <Button type="button" variant="outline" onClick={() => setActiveCreateSection("details")}>
                          {ux("actions.back")}
                        </Button>
                        <Button
                          type="button"
                          onClick={() => void handleContinueFromCreateModels()}
                          disabled={!canManageProviderAccount || !selectedAccountID || selectedModelCount === 0 || isSavingModels}
                        >
                          {isSavingModels ? t("saving") : ux("actions.continue")}
                        </Button>
                      </>
                    }
                  >
                    <ModelsEditor
                      t={t}
                      account={account}
                      setBanner={setBanner}
                      providerModels={providerModels}
                      modelDrafts={modelDrafts}
                      setModelDrafts={setModelDrafts}
                      isSaving={isSavingModels}
                      isLoadingModels={isLoadingProviderModels}
                      canManage={canManageProviderAccount}
                      onSave={handleSaveModelsAction}
                      variant="compact"
                    />
                  </WizardPanel>
                  ) : null}

                  {activeCreateSection === "access" ? (
                    <WizardPanel
                    id="access-permissions"
                    title={ux("integrations.wizard.accessPermissions")}
                    description={ux("integrations.wizard.accessPermissionsDescription")}
                    fillHeight={false}
                    footer={
                      <>
                        <Button type="button" variant="outline" onClick={() => setActiveCreateSection("models")}>
                          {ux("actions.back")}
                        </Button>
                        <Button
                          type="button"
                          onClick={() => void handleContinueFromCreateAccess()}
                          disabled={!canManageProviderAccount || !selectedAccountID || managers.length === 0 || isSavingAccess}
                        >
                          {isSavingAccess ? t("saving") : ux("actions.continue")}
                        </Button>
                      </>
                    }
                  >
                    <AccessEditor
                      t={t}
                      managers={managers}
                      setManagers={setManagers}
                      users={users}
                      setUsers={setUsers}
                      onSearchPrincipal={searchPrincipalCandidates}
                      principalSelectorLabels={principalSelectorLabels}
                      onSave={handleSaveAccessAction}
                      isSaving={isSavingAccess}
                      canManage={canManageProviderAccount}
                      effectiveKeyType={effectiveKeyType}
                      setEffectiveKeyType={setEffectiveKeyType}
                      effectivePrincipal={effectivePrincipalSelection}
                      setEffectivePrincipal={handleEffectivePrincipalChange}
                      effectiveModels={effectiveModels}
                      isLoadingEffective={isLoadingEffective}
                      onLoadEffective={handleLoadEffectiveModelsAction}
                    />
                  </WizardPanel>
                  ) : null}

                  {activeCreateSection === "review" ? (
                    <WizardPanel
                    id="review-confirm"
                    title={ux("integrations.wizard.reviewConfirm")}
                    description={ux("integrations.wizard.reviewConfirmDescription")}
                    fillHeight={false}
                    footer={
                      <>
                        <Button type="button" variant="outline" onClick={() => setActiveCreateSection("access")}>
                          {ux("actions.back")}
                        </Button>
                        <Button type="button" onClick={() => void handleFinishCreateFlow()} disabled={!canManageProviderAccount || isActivating}>
                          {isActivating ? ux("actions.activating") : ux("actions.activateAccount")}
                        </Button>
                      </>
                    }
                  >
                    <div className="space-y-4">
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
                    </div>
                  </WizardPanel>
                  ) : null}
                  </>
                )}
              </div>
            </div>
          </div>
        </section>
      </main>
    );
  }

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

        <div className="space-y-4">
          <section className="space-y-4">
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
                          void navigator.clipboard.writeText(`${account.provider}:${account.name}`);
                          setBanner({ kind: "success", message: ux("messages.fqnCopied") });
                        }}
                      >
                        <Copy className="me-2 h-4 w-4" />
                        {ux("actions.copyFqn")}
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="destructive"
                        onClick={() => openDeleteAccountDialog(account)}
                        disabled={deletingAccountID === account.id || !canDeleteProviderAccount}
                      >
                        <Trash2 className="me-2 h-4 w-4" />
                        {deletingAccountID === account.id ? ux("actions.loading") : ux("actions.deleteAccount")}
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
                    setupHref="/provider-accounts/new"
                    onOpenFirst={() => {
                      const first = selectedProviderAccounts[0];
                      if (first) handleSelectAccount(first);
                    }}
                    hasAccount={selectedProviderAccounts.length > 0}
                    createDisabled={!canCreateProviderAccount}
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
                                  const nextProvider = normalizeProviderSlug(item.provider);
                                  const shouldRegenerateName = shouldGenerateProviderAccountName({
                                    currentName: name,
                                    currentProvider: provider,
                                    isAutoGenerated: isNameAutoGenerated,
                                  });
                                  setProvider(nextProvider);
                                  setAuthMode(item.auth_modes?.[0] ?? "api_key");
                                  if (shouldRegenerateName) {
                                    setName(generateProviderAccountName(nextProvider, existingAccountNames));
                                    setIsNameAutoGenerated(true);
                                  }
                                  setSelectedProviderID(nextProvider);
                                  void loadProviderModels(item.provider);
                                }}
                              >
                                <div className="flex items-center gap-2">
                                  <ProviderLogo
                                    key={`${item.provider}:${item.logo_url ?? ""}`}
                                    provider={item.provider}
                                    displayName={item.display_name}
                                    logoURL={item.logo_url}
                                  />
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
                            <div
                              id="provider"
                              className="flex min-h-10 items-center gap-3 rounded-md border border-input bg-muted/30 px-3 py-2 text-sm"
                              aria-readonly="true"
                            >
                              <ProviderLogo
                                key={`${activeProvider?.provider ?? provider}:${activeProvider?.logo_url ?? ""}`}
                                provider={activeProvider?.provider ?? provider}
                                displayName={activeProvider?.display_name ?? provider}
                                logoURL={activeProvider?.logo_url}
                                size={20}
                              />
                              <div className="min-w-0">
                                <div className="truncate font-medium">{activeProvider?.display_name ?? provider}</div>
                                {activeProvider?.display_name && !stringsEqualFold(activeProvider.display_name, provider) ? (
                                  <div className="truncate text-xs text-muted-foreground">{provider}</div>
                                ) : null}
                              </div>
                            </div>
                          </Field>
                          <Field label={t("fields.accountName")} id="accountName">
                            <Input
                              id="accountName"
                              value={name}
                              onChange={(event) => {
                                setName(event.target.value);
                                setIsNameAutoGenerated(false);
                              }}
                              placeholder={ux("placeholders.accountName")}
                            />
                          </Field>
                        </div>

                        <div className="grid gap-3 md:grid-cols-2">
                          <AppSelectField
                            id="authMode"
                            label={t("fields.authMode")}
                            value={authMode}
                            onValueChange={(value) => setAuthMode(value as ProviderAuthMode)}
                            optionTextDir="ltr"
                            options={(selectedProvider?.auth_modes ?? ["api_key"]).map((mode) => ({ value: mode, label: mode }))}
                          />
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
                            apiKeyTouched={apiKeyTouched}
                            setAPIKeyTouched={(value) => {
                              setAPIKeyTouched(value);
                              if (value) {
                                setAPIKeyTestSucceeded(false);
                                setCreateCredentialTestResult(null);
                              }
                            }}
                            apiKeyHint={apiKeyTestSucceeded ? account?.credential_hint : undefined}
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
                          <Button type="button" onClick={() => void handleCreateAccount()} disabled={!canCreateProviderAccount || isLoading || !provider || !name.trim()}>
                            {isLoading ? t("creating") : t("createProviderAccount")}
                          </Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 3 ? (
                      <section className="space-y-4">
                        <ModelsEditor
                          t={t}
                          account={account}
                          setBanner={setBanner}
                          providerModels={providerModels}
                          modelDrafts={modelDrafts}
                          setModelDrafts={setModelDrafts}
                          isSaving={isSavingModels}
                          isLoadingModels={isLoadingProviderModels}
                          canManage={canManageProviderAccount}
                          onSave={handleSaveModelsAction}
                        />
                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(2)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => setSetupStep(4)} disabled={!canManageProviderAccount || selectedModelCount === 0}>{ux("actions.continue")}</Button>
                        </div>
                      </section>
                    ) : null}

                    {setupStep === 4 ? (
                      <section className="space-y-4">
                        <AccessEditor
                          t={t}
                          managers={managers}
                          setManagers={setManagers}
                          users={users}
                          setUsers={setUsers}
                          onSearchPrincipal={searchPrincipalCandidates}
                          principalSelectorLabels={principalSelectorLabels}
                          onSave={handleSaveAccessAction}
                          isSaving={isSavingAccess}
                          canManage={canManageProviderAccount}
                          effectiveKeyType={effectiveKeyType}
                          setEffectiveKeyType={setEffectiveKeyType}
                          effectivePrincipal={effectivePrincipalSelection}
                          setEffectivePrincipal={handleEffectivePrincipalChange}
                          effectiveModels={effectiveModels}
                          isLoadingEffective={isLoadingEffective}
                          onLoadEffective={handleLoadEffectiveModelsAction}
                        />
                        <div className="flex justify-between">
                          <Button type="button" variant="outline" onClick={() => setSetupStep(3)}>{ux("actions.back")}</Button>
                          <Button type="button" onClick={() => setSetupStep(5)} disabled={!canManageProviderAccount || managers.length === 0}>{ux("actions.continue")}</Button>
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
                          <Button type="button" onClick={() => void handleActivateAccount()} disabled={!canManageProviderAccount || isActivating}>
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
                        account={account}
                        setBanner={setBanner}
                        providerModels={providerModels}
                        modelDrafts={modelDrafts}
                        setModelDrafts={setModelDrafts}
                        isSaving={isSavingModels}
                        isLoadingModels={isLoadingProviderModels}
                        canManage={canManageProviderAccount}
                        onSave={handleSaveModelsAction}
                      />
                    ) : null}

                    {activeTab === "access" ? (
                      <AccessEditor
                        t={t}
                        managers={managers}
                        setManagers={setManagers}
                        users={users}
                        setUsers={setUsers}
                        onSearchPrincipal={searchPrincipalCandidates}
                        principalSelectorLabels={principalSelectorLabels}
                        onSave={handleSaveAccessAction}
                        isSaving={isSavingAccess}
                        canManage={canManageProviderAccount}
                        effectiveKeyType={effectiveKeyType}
                        setEffectiveKeyType={setEffectiveKeyType}
                        effectivePrincipal={effectivePrincipalSelection}
                        setEffectivePrincipal={handleEffectivePrincipalChange}
                        effectiveModels={effectiveModels}
                        isLoadingEffective={isLoadingEffective}
                        onLoadEffective={handleLoadEffectiveModelsAction}
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
                        <AppSelectField
                          id="credentialAuthMode"
                          label={t("fields.authMode")}
                          value={authMode}
                          disabled={!canManageProviderAccount}
                          onValueChange={(value) => setAuthMode(value as ProviderAuthMode)}
                          optionTextDir="ltr"
                          options={(selectedProvider?.auth_modes ?? ["api_key"]).map((mode) => ({ value: mode, label: mode }))}
                        />
                        <Field label={t("fields.allowVirtualOnly")} id="credentialVirtualOnly">
                          <label className="inline-flex items-center gap-2 text-sm">
                            <input
                              id="credentialVirtualOnly"
                              type="checkbox"
                              checked={allowVirtualOnly}
                              disabled={!canManageProviderAccount}
                              onChange={(event) => setAllowVirtualOnly(event.target.checked)}
                            />
                            {t("fields.allowVirtualOnly")}
                          </label>
                        </Field>
                        {!allowVirtualOnly ? (
                          <CredentialFields
                            authMode={authMode}
                            apiKey={apiKey}
                            setAPIKey={setAPIKey}
                            apiKeyTouched={apiKeyTouched}
                            setAPIKeyTouched={(value) => {
                              setAPIKeyTouched(value);
                              if (value) {
                                setAPIKeyTestSucceeded(false);
                                setLastCredentialTestMessage("");
                                setCreateCredentialTestResult(null);
                              }
                            }}
                            apiKeyHint={apiKeyTestSucceeded ? account.credential_hint : undefined}
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
                        <div className="flex flex-wrap gap-2">
                          <Button type="button" onClick={() => void handleSaveCredentials()} disabled={!canManageProviderAccount || isSavingCredentials}>
                            {isSavingCredentials ? t("saving") : t("save")}
                          </Button>
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
                          <AppSelectField
                            id="obsKeyType"
                            label={ux("access.keyTypeLabel")}
                            value={effectiveKeyType}
                            onValueChange={(value) => setEffectiveKeyType(value as "pat" | "vat")}
                            options={[
                              { value: "pat", label: ux("access.pat") },
                              { value: "vat", label: ux("access.vat") },
                            ]}
                          />
                          <Field label={`${ux("access.userIdLabel")} / ${ux("access.teamIdLabel")}`} id="obsPrincipal">
                            <PrincipalEntitySelect
                              id="obsPrincipal"
                              selected={effectivePrincipalSelection}
                              onSelectedChange={handleEffectivePrincipalChange}
                              onSearch={searchPrincipalCandidates}
                              labels={principalSelectorLabels}
                              kinds={["user", "team"]}
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
        <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("messages.accountDeleteConfirmTitle")}</DialogTitle>
              <DialogDescription>
                {t("messages.accountDeleteConfirmBody", { name: deleteTargetAccount?.name ?? "" })}
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-2">
              <Label htmlFor="delete-account-name-input">{t("messages.accountDeleteConfirmLabel")}</Label>
              <Input
                id="delete-account-name-input"
                value={deleteAccountNameInput}
                onChange={(event) => setDeleteAccountNameInput(event.target.value)}
                placeholder={deleteTargetAccount?.name ?? ""}
                aria-invalid={
                  Boolean(deleteAccountNameInput) &&
                  deleteAccountNameInput !== (deleteTargetAccount?.name ?? "")
                }
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setDeleteDialogOpen(false);
                  setDeleteTargetAccount(null);
                  setDeleteAccountNameInput("");
                }}
              >
                {ux("actions.cancelDelete")}
              </Button>
              <Button
                type="button"
                variant="destructive"
                onClick={() => {
                  if (deleteTargetAccount) {
                    void handleDeleteAccount(deleteTargetAccount);
                  }
                }}
                disabled={
                  !deleteTargetAccount ||
                  deletingAccountID === deleteTargetAccount.id ||
                  deleteAccountNameInput !== (deleteTargetAccount?.name ?? "")
                }
              >
                {deletingAccountID ? ux("actions.loading") : ux("actions.confirmDelete")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}


