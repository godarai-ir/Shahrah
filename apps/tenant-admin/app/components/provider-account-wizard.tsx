"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { Copy, Eye, KeyRound, Layers3, ShieldCheck, Trash2 } from "lucide-react";

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
  deriveAccountDetailsReadiness,
  deriveAccountListReadiness,
  mergeAccountIntoList,
} from "@/app/components/provider-account-wizard/account-readiness";
import { CredentialFields } from "@/app/components/provider-account-wizard/credential-fields";
import { EffectiveModelsTable } from "@/app/components/provider-account-wizard/effective-models-table";
import { Field } from "@/app/components/provider-account-wizard/field";
import { ModelsEditor } from "@/app/components/provider-account-wizard/models-editor";
import { EmptyOverview, ReadinessBadge, SummaryKPI } from "@/app/components/provider-account-wizard/overview";
import { buildPrincipalSelectorLabels } from "@/app/components/provider-account-wizard/principal-labels";
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
export default function ProviderAccountWizard(props?: { mode?: ProviderAccountWizardMode }) {
  const mode = props?.mode ?? "manage";
  const isCreateMode = mode === "create";
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
  const [isRefreshingCatalog, setIsRefreshingCatalog] = useState(false);
  const [isActivating, setIsActivating] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [catalogStats, setCatalogStats] = useState<{ cacheTTL: number; counters: Record<string, number> }>({
    cacheTTL: 0,
    counters: {},
  });

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [accountReadiness, setAccountReadiness] = useState<Record<string, Readiness>>({});
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);

  const [selectedProviderID, setSelectedProviderID] = useState("");
  const [selectedAccountID, setSelectedAccountID] = useState("");
  const selectedProviderRef = useRef("");
  const selectedAccountRef = useRef("");
  const accountLoadRequestRef = useRef(0);

  const [account, setAccount] = useState<TenantProviderAccount | null>(null);

  const [provider, setProvider] = useState("");
  const [name, setName] = useState("");
  const [isNameAutoGenerated, setIsNameAutoGenerated] = useState(true);
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

  const [effectiveKeyType, setEffectiveKeyType] = useState<"pat" | "vat">("pat");
  const [effectiveUserID, setEffectiveUserID] = useState("");
  const [effectiveTeamID, setEffectiveTeamID] = useState("");
  const [effectivePrincipalSelection, setEffectivePrincipalSelection] = useState<PrincipalEntitySelection | null>(null);
  const [effectiveModels, setEffectiveModels] = useState<EffectiveModelView[]>([]);
  const [lastCredentialTestMessage, setLastCredentialTestMessage] = useState("");

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
  const hasVisibleProviders = visibleProviders.length > 0;
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
      setAccountReadiness((prev) => {
        const activeIDs = new Set(accountItems.map((item) => item.id));
        return Object.fromEntries(Object.entries(prev).filter(([id]) => activeIDs.has(id)));
      });

      if (isCreateMode) {
        const createProviders = providerItems.filter(
          (item) => item.enabled && stringsEqualFold(item.provider, "azure") === false,
        );
        const initialProvider = currentProviderID || createProviders[0]?.provider || "";
        setSelectedProviderID(initialProvider);
        setSelectedAccountID("");
        setAccount(null);
        setWorkspaceMode("setup");
        setSetupStep(1);
        setManagers([]);
        setUsers([{ type: "everyone", display_name: t("principal.everyone") }]);
        if (initialProvider) {
          const providerItem = createProviders.find((item) => stringsEqualFold(item.provider, initialProvider));
          setProvider(initialProvider);
          setName(generateProviderAccountName(initialProvider, fetchedAccountNames));
          setIsNameAutoGenerated(true);
          setAuthMode(providerItem?.auth_modes?.[0] ?? "api_key");
          setAllowVirtualOnly(false);
          void loadProviderModels(initialProvider);
        }
        return;
      }

      const initialProvider = currentProviderID || accountItems[0]?.provider || "";
      setSelectedProviderID(initialProvider);

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
        setSecretRef("");
        setManagers(payload.access?.managers ?? []);
        setUsers(payload.access?.users ?? []);
        setSelectedProviderID(normalizeProviderSlug(payload.account.provider));
        setSelectedAccountID(payload.account.id);
        setModelDrafts(buildDraftsForAccount(payload.models ?? [], catalogModels));
        setAccountReadiness((prev) => ({
          ...prev,
          [payload.account.id]: deriveAccountDetailsReadiness(
            payload.account,
            payload.models ?? [],
            payload.access,
            catalogModels,
          ),
        }));
      } catch (error) {
        if (requestID === accountLoadRequestRef.current) {
          setBanner({ kind: "error", message: normalizeError(error, t("messages.accountLoadFailed")) });
        }
      } finally {
        if (requestID === accountLoadRequestRef.current) setIsLoading(false);
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
        setAccountReadiness((prev) =>
          Object.fromEntries(Object.entries(prev).filter(([id]) => id !== target.id)),
        );

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
    setSecretRef("");
    setAzureTenantID("");
    setAzureClientID("");
    setAWSAccessKeyID("");
    setAWSSecretAccessKey("");
    setAWSRegion("us-east-1");
    setGoogleServiceAccountJSON("");
  }

  async function handleRefreshCatalog() {
    setIsRefreshingCatalog(true);
    setBanner(null);
    try {
      await requestJSON("/api/provider-catalog/refresh", { method: "POST" });
      await queryClient.invalidateQueries({ queryKey: ["tenant-admin", "provider-catalog"] });
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
    setSelectedProviderID(normalizeProviderSlug(item.provider));
    if (workspaceMode === "setup") {
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
      void loadProviderModels(item.provider);
      return;
    }

    const firstProviderAccount = accounts.find((accountItem) => stringsEqualFold(accountItem.provider, item.provider));
    if (firstProviderAccount) {
      handleSelectAccount(firstProviderAccount);
    } else {
      accountLoadRequestRef.current += 1;
      setSelectedAccountID("");
      setAccount(null);
      setWorkspaceMode("overview");
      setProvider(item.provider);
      setName(generateProviderAccountName(item.provider, existingAccountNames));
      setIsNameAutoGenerated(true);
      setAuthMode(item.auth_modes?.[0] ?? "api_key");
      setAllowVirtualOnly(false);
      setManagers([]);
      setUsers([{ type: "everyone", display_name: t("principal.everyone") }]);
      setEffectiveModels([]);
      void loadProviderModels(item.provider);
    }
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
    setModelSearchQuery("");
    setModelDrafts({});
    setWorkspaceMode("workspace");
    setActiveTab("models");
    if (shouldReloadCurrentAccount) void handleLoadAccount(item.id);
  }

  async function handleCreateAccount() {
    setBanner(null);
    if (!canCreateProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return;
    }
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
  }

  async function handleSaveAccess(accountID = selectedAccountID): Promise<boolean> {
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
  const principalSelectorLabels = buildPrincipalSelectorLabels(t, ux);

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
          {!isCreateMode ? (
          <aside className="space-y-3 lg:col-span-4 xl:col-span-3">
            {hasVisibleProviders ? (
              <>
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-base">{ux("sidebar.catalogProviders")}</CardTitle>
                    <CardDescription>{ux("sidebar.catalogProvidersHint")}</CardDescription>
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
                            isSelected
                              ? "border-primary bg-primary/10"
                              : providerCount > 0
                                ? "border-primary/30 bg-primary/[0.04] hover:border-primary/50"
                                : "border-border hover:border-primary/40"
                          }`}
                          onClick={() => handleSelectProvider(item)}
                        >
                          <div className="flex items-center justify-between gap-2">
                            <div className="flex items-center gap-2">
                              <ProviderLogo
                                key={`${item.provider}:${item.logo_url ?? ""}`}
                                provider={item.provider}
                                displayName={item.display_name}
                                logoURL={item.logo_url}
                              />
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
                        const isDeleting = deletingAccountID === item.id;
                        return (
                          <div
                            key={item.id}
                            className={`w-full rounded-lg border px-2 py-2 ${
                              isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"
                            }`}
                          >
                            <div className="flex items-center gap-2">
                              <button
                                type="button"
                                className="min-w-0 flex-1 text-start"
                                onClick={() => handleSelectAccount(item)}
                              >
                                <div className="font-medium">{item.name}</div>
                                <div className="text-xs text-muted-foreground">{item.auth_mode}</div>
                              </button>
                              <div className="shrink-0">
                                <ReadinessBadge
                                  readiness={accountReadiness[item.id] ?? deriveAccountListReadiness(item)}
                                  labels={{
                                    ready: ux("readiness.ready"),
                                    needs_credentials: ux("readiness.needsCredentials"),
                                    needs_access: ux("readiness.needsAccess"),
                                    pricing_incomplete: ux("readiness.pricingIncomplete"),
                                  }}
                                />
                              </div>
                            </div>
                            <div className="mt-2 flex justify-end">
                              <Button
                                type="button"
                                size="sm"
                                variant="ghost"
                                className="text-destructive hover:text-destructive"
                                onClick={() => openDeleteAccountDialog(item)}
                                disabled={isDeleting || !canDeleteProviderAccount}
                              >
                                <Trash2 className="me-1 h-4 w-4" />
                                {isDeleting ? ux("actions.loading") : ux("actions.deleteAccount")}
                              </Button>
                            </div>
                          </div>
                        );
                      })
                    )}

                    <div className="flex gap-2 pt-2">
                      <Button type="button" size="sm" className="flex-1" asChild disabled={!canCreateProviderAccount}>
                        <Link href="/provider-accounts/new">{ux("actions.addProviderAccount")}</Link>
                      </Button>
                      <Button type="button" size="sm" variant="outline" onClick={() => void handleRefreshCatalog()} disabled={isRefreshingCatalog}>
                        {isRefreshingCatalog ? ux("actions.refreshingCatalog") : ux("actions.refreshCatalog")}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </>
            ) : (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">{ux("sidebar.providerAccounts")}</CardTitle>
                  <CardDescription>{ux("overview.title")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm text-muted-foreground">{ux("overview.description")}</p>
                  <div className="flex gap-2">
                    <Button type="button" size="sm" className="flex-1" asChild disabled={!canCreateProviderAccount}>
                      <Link href="/provider-accounts/new">{ux("actions.setupNewAccount")}</Link>
                    </Button>
                    <Button type="button" size="sm" variant="outline" onClick={() => void handleRefreshCatalog()} disabled={isRefreshingCatalog}>
                      {isRefreshingCatalog ? ux("actions.refreshingCatalog") : ux("actions.refreshCatalog")}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}
          </aside>
          ) : null}

          <section className={`space-y-4 ${isCreateMode ? "lg:col-span-12" : "lg:col-span-8 xl:col-span-9"}`}>
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
                          filteredProviderModels={filteredProviderModels}
                          modelDrafts={modelDrafts}
                          setModelDrafts={setModelDrafts}
                          modelSearchQuery={modelSearchQuery}
                          setModelSearchQuery={setModelSearchQuery}
                          isSaving={isSavingModels}
                          canManage={canManageProviderAccount}
                          onSave={() => void handleSaveModels()}
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
                          onSave={() => void handleSaveAccess()}
                          isSaving={isSavingAccess}
                          canManage={canManageProviderAccount}
                          effectiveKeyType={effectiveKeyType}
                          setEffectiveKeyType={setEffectiveKeyType}
                          effectivePrincipal={effectivePrincipalSelection}
                          setEffectivePrincipal={handleEffectivePrincipalChange}
                          effectiveModels={effectiveModels}
                          isLoadingEffective={isLoadingEffective}
                          onLoadEffective={() => void loadEffectiveModelsPreview()}
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
                        filteredProviderModels={filteredProviderModels}
                        modelDrafts={modelDrafts}
                        setModelDrafts={setModelDrafts}
                        modelSearchQuery={modelSearchQuery}
                        setModelSearchQuery={setModelSearchQuery}
                        isSaving={isSavingModels}
                        canManage={canManageProviderAccount}
                        onSave={() => void handleSaveModels()}
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
                        onSave={() => void handleSaveAccess()}
                        isSaving={isSavingAccess}
                        canManage={canManageProviderAccount}
                        effectiveKeyType={effectiveKeyType}
                        setEffectiveKeyType={setEffectiveKeyType}
                        effectivePrincipal={effectivePrincipalSelection}
                        setEffectivePrincipal={handleEffectivePrincipalChange}
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
                        <Field label={t("fields.authMode")} id="credentialAuthMode">
                          <select
                            id="credentialAuthMode"
                            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                            value={authMode}
                            disabled={!canManageProviderAccount}
                            onChange={(event) => setAuthMode(event.target.value as ProviderAuthMode)}
                          >
                            {(selectedProvider?.auth_modes ?? ["api_key"]).map((mode) => (
                              <option key={mode} value={mode}>{mode}</option>
                            ))}
                          </select>
                        </Field>
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


