"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useLocale, useTranslations } from "next-intl";
import { ChevronRight, CircleDollarSign, ExternalLink, KeyRound, Layers3, LayoutDashboard, SearchX, ShieldCheck, Users } from "lucide-react";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { EntityIdentityHeader } from "@/components/app-ui/EntityIdentityHeader";
import { DebouncedTextFilter } from "@/components/app-ui/FilterField";
import { FormSection } from "@/components/app-ui/FormSection";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { SectionHeader } from "@/components/app-ui/SectionHeader";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { UnsavedChangesDialog } from "@/components/app-ui/UnsavedChangesDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
  FieldTitle,
} from "@/components/ui/field";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import {
  buildDefaultDrafts,
  buildDraftsForAccount,
  buildSelectedModelPayload,
  canUseCustomModelPricing,
} from "@/app/components/provider-account-wizard.models";
import { AccessEditor } from "@/app/components/provider-account-wizard/access-editor";
import {
  mergeAccountIntoList,
} from "@/app/components/provider-account-wizard/account-readiness";
import { CredentialFields } from "@/app/components/provider-account-wizard/credential-fields";
import { EffectiveAccessTester } from "@/app/components/provider-account-wizard/effective-access-tester";
import {
  buildProviderCredentialPayload,
  defaultProviderCredentialValues,
  missingRequiredProviderFields,
  providerCredentialValuesForSchema,
  type ProviderCredentialValues,
} from "@/app/components/provider-account-wizard/credential-schema";
import { getCreateStepState, shouldPersistCreateAccount } from "@/app/components/provider-account-wizard/create-flow";
import { ModelsEditor } from "@/app/components/provider-account-wizard/models-editor";
import { EmptyOverview, SummaryKPI } from "@/app/components/provider-account-wizard/overview";
import { buildPrincipalSelectorLabels } from "@/app/components/provider-account-wizard/principal-labels";
import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import { ProviderCredentialSchemaFields } from "@/app/components/provider-account-wizard/provider-credential-schema-fields";
import type {
  AccessPrincipal,
  AccessPrincipalCandidate,
  AccessPrincipalResponse,
  AccountDetailsResponse,
  AccountsResponse,
  Banner,
  CredentialTestResult,
  ModelDraft,
  ProviderAccountWizardMode,
  ProviderAuthMode,
  ProviderCatalogModel,
  ProviderCatalogProvider,
  ProviderConnectionSchema,
  ProviderConnectionSchemasResponse,
  ProviderModelsResponse,
  ProvidersResponse,
  Readiness,
  SetupStep,
  TenantProviderAccount,
  WorkspaceMode,
  WorkspaceTab,
} from "@/app/components/provider-account-wizard/types";
import { WIFCredentialStatus } from "@/app/components/provider-account-wizard/wif-credential-status";
import {
  CreateFlowStepIndicator,
  type CreateWizardSectionID,
} from "@/app/components/provider-account-wizard/wizard-layout";
import {
  buildModelDiscoveryProviderHref,
  extractMessageFromPayload,
  formatDisplayPrice,
  formatProviderDisplayName,
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
import { cn } from "@/lib/utils";

export default function ProviderAccountWizard(props?: {
  mode?: ProviderAccountWizardMode;
  initialProvider?: string;
  initialAccountID?: string;
  docsBaseHref?: string;
  initialSection?: WorkspaceTab;
  currentUser?: {
    id: string;
    email?: string;
    mobile?: string;
  };
}) {
  const mode = props?.mode ?? "manage";
  const isCreateMode = mode === "create";
  const initialProvider = normalizeProviderSlug(props?.initialProvider ?? "");
  const initialAccountID = props?.initialAccountID?.trim() ?? "";
  const docsBaseHref = props?.docsBaseHref?.replace(/\/$/, "") || "/docs";
  const t = useTranslations("providerAccounts");
  const ux = useTranslations("providerAccounts.ux");
  const catalogT = useTranslations("modelCatalog");
  const locale = useLocale();
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
  const [activeTab, setActiveTab] = useState<WorkspaceTab>(props?.initialSection ?? "overview");

  const [isLoading, setIsLoading] = useState(Boolean(initialAccountID));
  const [isSavingModels, setIsSavingModels] = useState(false);
  const [isSavingAccess, setIsSavingAccess] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [isSavingCredentials, setIsSavingCredentials] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteTargetAccount, setDeleteTargetAccount] = useState<TenantProviderAccount | null>(null);
  const [isLoadingProviderModels, setIsLoadingProviderModels] = useState(false);
  const [isActivating, setIsActivating] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);

  const [selectedProviderID, setSelectedProviderID] = useState(initialProvider);
  const [selectedAccountID, setSelectedAccountID] = useState(initialAccountID);
  const selectedProviderRef = useRef(initialProvider);
  const selectedAccountRef = useRef(initialAccountID);
  const accountLoadRequestRef = useRef(0);
  const modelDraftBaselineRef = useRef("");
  const credentialDraftBaselineRef = useRef<{
    authMode: ProviderAuthMode;
    usePlatformManagedCredentials: boolean;
    values: ProviderCredentialValues;
  } | null>(null);

  const [account, setAccount] = useState<TenantProviderAccount | null>(null);

  const [provider, setProvider] = useState(initialProvider);
  const [name, setName] = useState("");
  const [isNameAutoGenerated, setIsNameAutoGenerated] = useState(true);
  const [authMode, setAuthMode] = useState<ProviderAuthMode>("api_key");
  const [usePlatformManagedCredentials, setUsePlatformManagedCredentials] = useState(isCreateMode);

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
  const [connectionSchemas, setConnectionSchemas] = useState<ProviderConnectionSchema[]>([]);
  const [providerCredentialValues, setProviderCredentialValues] = useState<ProviderCredentialValues>({});
  const [isLoadingConnectionSchemas, setIsLoadingConnectionSchemas] = useState(false);
  const [connectionSchemaError, setConnectionSchemaError] = useState("");

  const [modelDrafts, setModelDrafts] = useState<Record<string, ModelDraft>>({});

  const currentUserManager = useMemo<AccessPrincipal | null>(
    () => isCreateMode && props?.currentUser?.id
      ? {
          type: "user",
          principal_id: props.currentUser.id,
          display_name: props.currentUser.email || props.currentUser.mobile || props.currentUser.id,
        }
      : null,
    [isCreateMode, props?.currentUser?.email, props?.currentUser?.id, props?.currentUser?.mobile],
  );
  const [managers, setManagers] = useState<AccessPrincipal[]>(currentUserManager ? [currentUserManager] : []);
  const [users, setUsers] = useState<AccessPrincipal[]>([]);
  const accessAutosaveQueueRef = useRef<Promise<void>>(Promise.resolve());

  const [createCredentialTestResult, setCreateCredentialTestResult] = useState<
    (Banner & Partial<CredentialTestResult>) | null
  >(null);
  const [activeCreateSection, setActiveCreateSection] = useState<CreateWizardSectionID>("details");
  const [furthestCreateStepIndex, setFurthestCreateStepIndex] = useState(0);
  const [providerSearch, setProviderSearch] = useState("");
  const [unsavedDialogOpen, setUnsavedDialogOpen] = useState(false);
  const [pendingWorkspaceNavigation, setPendingWorkspaceNavigation] = useState<
    { kind: "tab"; tab: WorkspaceTab } | { kind: "href"; href: string } | null
  >(null);

  const visibleProviders = useMemo(
    () => {
      const query = providerSearch.trim().toLocaleLowerCase();
      return providers.filter((item) => {
        if (!item.enabled || stringsEqualFold(item.provider, "azure")) return false;
        if (!query) return true;
        return `${item.display_name} ${item.provider} ${item.description ?? ""}`.toLocaleLowerCase().includes(query);
      });
    },
    [providerSearch, providers],
  );

  const selectedProvider = useMemo(
    () => providers.find((item) => stringsEqualFold(item.provider, selectedProviderID)) ?? null,
    [providers, selectedProviderID],
  );
  const activeProvider = useMemo(
    () => providers.find((item) => stringsEqualFold(item.provider, provider)) ?? selectedProvider,
    [providers, provider, selectedProvider],
  );
  const selectedConnectionSchema = useMemo(
    () => connectionSchemas.find((item) => stringsEqualFold(item.auth_mode, authMode)) ?? null,
    [authMode, connectionSchemas],
  );
  const missingRequiredCredentialFields = useMemo(
    () => usePlatformManagedCredentials ? [] : missingRequiredProviderFields(selectedConnectionSchema, providerCredentialValues),
    [usePlatformManagedCredentials, providerCredentialValues, selectedConnectionSchema],
  );
  const canContinueCreateDetails = usePlatformManagedCredentials || (
    !isLoadingConnectionSchemas
    && !connectionSchemaError
    && !!selectedConnectionSchema
    && missingRequiredCredentialFields.length === 0
  );
  const requiresSuccessfulCreatePreflight =
    !usePlatformManagedCredentials && stringsEqualFold(authMode, "api_key");
  const missingRequiredCredentialFieldsForSave = useMemo(() => {
    if (usePlatformManagedCredentials) return [];
    const mustReplaceSecret = !account?.has_credential || !stringsEqualFold(authMode, account.auth_mode);

    return (selectedConnectionSchema?.fields ?? []).filter((field) => {
      if (!field.required) return false;
      if (field.secret && !mustReplaceSecret) return false;
      return !providerCredentialValues[field.key]?.trim();
    });
  }, [account, authMode, providerCredentialValues, selectedConnectionSchema, usePlatformManagedCredentials]);
  const canEditCredentialSettings = isCreateMode || (!!account && !account.use_platform_managed_credentials);
  const canSaveCredentialSettings = canEditCredentialSettings && (
    usePlatformManagedCredentials
    || (
      !isLoadingConnectionSchemas
      && !connectionSchemaError
      && !!selectedConnectionSchema
      && missingRequiredCredentialFieldsForSave.length === 0
    )
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
    let cancelled = false;
    async function loadConnectionSchemas() {
      await Promise.resolve();
      if (cancelled) return;
      if (!provider) {
        setConnectionSchemas([]);
        setProviderCredentialValues({});
        setConnectionSchemaError("");
        return;
      }

      setIsLoadingConnectionSchemas(true);
      setConnectionSchemas([]);
      setProviderCredentialValues({});
      setConnectionSchemaError("");
      try {
        const items = await queryClient.fetchQuery({
          queryKey: ["tenant-admin", "provider-connection-schemas", provider],
          queryFn: async () => {
            const response = await requestJSON<ProviderConnectionSchemasResponse>(
              `/api/provider-catalog/providers/${encodeURIComponent(provider)}/connection-schemas`,
            );
            return unwrapData<ProviderConnectionSchema[]>(response) ?? [];
          },
          staleTime: 1000 * 60 * 30,
          gcTime: 1000 * 60 * 60 * 2,
        });
        if (cancelled) return;
        const preferredAuthMode = isCreateMode ? undefined : account?.auth_mode;
        const nextSchema = preferredAuthMode
          ? items.find((item) => stringsEqualFold(item.auth_mode, preferredAuthMode)) ?? items.find((item) => item.runtime_ready) ?? items[0] ?? null
          : items.find((item) => item.runtime_ready) ?? items[0] ?? null;
        const nextAuthMode = nextSchema?.auth_mode ?? "api_key";
        const nextValues = isCreateMode
          ? defaultProviderCredentialValues(nextSchema)
          : providerCredentialValuesForSchema(nextSchema, account?.provider_config);
        setConnectionSchemas(items);
        setAuthMode(nextAuthMode);
        setProviderCredentialValues(nextValues);
        if (!isCreateMode) {
          credentialDraftBaselineRef.current = {
            authMode: nextAuthMode,
            usePlatformManagedCredentials: account?.use_platform_managed_credentials ?? false,
            values: nextValues,
          };
        }
        if (!nextSchema) setConnectionSchemaError(ux("integrations.wizard.connectionSchemaUnavailable"));
      } catch (error) {
        if (cancelled) return;
        setConnectionSchemas([]);
        setProviderCredentialValues({});
        setConnectionSchemaError(normalizeError(error, ux("integrations.wizard.connectionSchemaUnavailable")));
      } finally {
        if (!cancelled) setIsLoadingConnectionSchemas(false);
      }
    }
    void loadConnectionSchemas();

    return () => {
      cancelled = true;
    };
  }, [account?.auth_mode, account?.provider_config, account?.use_platform_managed_credentials, isCreateMode, provider, queryClient, ux]);

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
  const selectedModels = useMemo(
    () => providerModels.filter((model) => modelDrafts[model.model]?.enabled),
    [modelDrafts, providerModels],
  );
  const modelDraftFingerprint = useMemo(() => JSON.stringify(modelDrafts), [modelDrafts]);
  const hasUnsavedModelChanges = Boolean(
    !isCreateMode
    && modelDraftBaselineRef.current
    && modelDraftFingerprint !== modelDraftBaselineRef.current,
  );
  const credentialDraftFingerprint = useMemo(
    () => credentialFingerprint({
      authMode,
      usePlatformManagedCredentials,
      values: providerCredentialValues,
    }),
    [authMode, providerCredentialValues, usePlatformManagedCredentials],
  );
  const hasUnsavedCredentialChanges = Boolean(
    !isCreateMode
    && credentialDraftBaselineRef.current
    && credentialDraftFingerprint !== credentialFingerprint(credentialDraftBaselineRef.current),
  );
  const hasUnsavedChanges = hasUnsavedModelChanges || hasUnsavedCredentialChanges;

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
      !account.use_platform_managed_credentials &&
      account.auth_mode !== "no_credential";
    if (needsCredential) return "needs_credentials";
    if (createCredentialTestResult?.kind === "error") return "connection_failed";
    if (managers.length === 0) return "needs_access";
    if (selectedModelCount > 0 && unresolvedPricingCount > 0) return "pricing_incomplete";
    return "ready";
  }, [account, createCredentialTestResult?.kind, managers.length, selectedModelCount, unresolvedPricingCount]);
  const accountNameConflict = useMemo(
    () => existingAccountNames.some((existingName) => stringsEqualFold(existingName, name.trim())),
    [existingAccountNames, name],
  );
  const createPreflightReady =
    usePlatformManagedCredentials
    || !requiresSuccessfulCreatePreflight
    || apiKeyTestSucceeded;
  const createReady =
    Boolean(provider && name.trim())
    && !accountNameConflict
    && canContinueCreateDetails
    && createPreflightReady
    && selectedModelCount > 0
    && unresolvedPricingCount === 0
    && managers.length > 0;

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
        const initialProvider = currentProviderID;
        setSelectedProviderID(initialProvider);
        setSelectedAccountID("");
        setAccount(null);
        setWorkspaceMode("setup");
        setSetupStep(currentProviderID ? 2 : 1);
        setManagers(currentUserManager ? [currentUserManager] : []);
        setUsers([]);
        if (initialProvider) {
          const providerItem = createProviders.find((item) => stringsEqualFold(item.provider, initialProvider));
          setProvider(initialProvider);
          setName(generateProviderAccountName(initialProvider, fetchedAccountNames));
          setIsNameAutoGenerated(true);
          setAuthMode(providerItem?.auth_modes?.[0] ?? "api_key");
          setUsePlatformManagedCredentials(true);
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
        setUsePlatformManagedCredentials(initialAccount.use_platform_managed_credentials);
        setWorkspaceMode("workspace");
        setActiveTab(props?.initialSection ?? "overview");
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
  }, [currentUserManager, fetchProviders, isCreateMode, loadProviderModels, props?.initialSection, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadInitial();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadInitial]);

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
        setUsePlatformManagedCredentials(payload.account.use_platform_managed_credentials);
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
        setCreateCredentialTestResult(null);
        setSecretRef("");
        setManagers(payload.access?.managers ?? []);
        setUsers(payload.access?.users ?? []);
        setSelectedProviderID(normalizeProviderSlug(payload.account.provider));
        setSelectedAccountID(payload.account.id);
        const nextDrafts = buildDraftsForAccount(payload.models ?? [], catalogModels);
        modelDraftBaselineRef.current = JSON.stringify(nextDrafts);
        setModelDrafts(nextDrafts);
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

  const handleDeleteAccount = useCallback(
    async (target: TenantProviderAccount) => {
      if (!canDeleteProviderAccount) {
        setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
        return;
      }
      if (!target?.id) return;
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
        showFeedbackToast({ kind: "success", title: t("messages.accountDeleteSuccess") });
        setDeleteDialogOpen(false);
        setDeleteTargetAccount(null);
      } catch (error) {
        setBanner({ kind: "error", message: normalizeError(error, t("messages.accountDeleteFailed")) });
        throw error;
      }
    },
    [accounts, canDeleteProviderAccount, normalizeDeleteAccountError, selectedAccountID, t],
  );

  async function handleSaveCredentials(
    accountID = selectedAccountID,
    options: { reload?: boolean; announce?: boolean } = {},
  ) {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }
    if (!isCreateMode && account?.use_platform_managed_credentials) {
      setBanner({ kind: "error", message: ux("credentials.managedCredentialLockedDescription") });
      return false;
    }
    if (account?.use_platform_managed_credentials && !usePlatformManagedCredentials) {
      setBanner({ kind: "error", message: ux("credentials.managedCredentialLockedDescription") });
      return false;
    }
    if (account && !account.use_platform_managed_credentials && usePlatformManagedCredentials) {
      setBanner({ kind: "error", message: ux("credentials.privateCredentialLockedTitle") });
      return false;
    }

    const savingPlatformManagedCredentials = account?.use_platform_managed_credentials ?? usePlatformManagedCredentials;

    const providerConfig: Record<string, unknown> = {};
    const credentials: Record<string, unknown> = {};

    if (!savingPlatformManagedCredentials && (!selectedConnectionSchema || missingRequiredCredentialFieldsForSave.length > 0)) {
      setBanner({ kind: "error", message: connectionSchemaError || ux("integrations.wizard.requiredCredentialFields") });
      return false;
    }

    if (!savingPlatformManagedCredentials && selectedConnectionSchema) {
      const schemaPayload = buildProviderCredentialPayload(selectedConnectionSchema, providerCredentialValues);
      Object.assign(providerConfig, schemaPayload.providerConfig);
      Object.assign(credentials, schemaPayload.credentials);
    } else if (!savingPlatformManagedCredentials && authMode === "api_key") {
      if (apiKey.trim()) credentials.api_key = apiKey.trim();
      if (!isCreateMode && secretRef.trim()) credentials.secret_ref = secretRef.trim();
    } else if (!savingPlatformManagedCredentials && authMode === "azure_entra_cert") {
      providerConfig.azure_tenant_id = azureTenantID.trim() || undefined;
      providerConfig.azure_client_id = azureClientID.trim() || undefined;
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    } else if (!savingPlatformManagedCredentials && authMode === "aws_sigv4") {
      providerConfig.aws_region = awsRegion.trim() || undefined;
      providerConfig.aws_access_key_id = awsAccessKeyID.trim() || undefined;
      providerConfig.aws_secret_access_key = awsSecretAccessKey.trim() || undefined;
    } else if (!savingPlatformManagedCredentials && authMode === "google_service_account") {
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
            use_platform_managed_credentials: savingPlatformManagedCredentials,
            provider_config: providerConfig,
            credentials: Object.keys(credentials).length > 0 ? credentials : undefined,
          }),
        },
      );
      if (options.announce ?? true) {
        showFeedbackToast({ kind: "success", title: t("messages.accountLoadSuccess") });
      }
      if (options.reload ?? true) {
        await handleLoadAccount(accountID);
      }
      setAPIKeyTouched(false);
      setAPIKeyTestSucceeded(false);
      credentialDraftBaselineRef.current = {
        authMode,
        usePlatformManagedCredentials: savingPlatformManagedCredentials,
        values: { ...providerCredentialValues },
      };
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

  useEffect(() => {
    if (isCreateMode) return;
    const preventUnload = (event: BeforeUnloadEvent) => {
      if (!hasUnsavedChanges) return;
      event.preventDefault();
    };
    window.addEventListener("beforeunload", preventUnload);
    return () => window.removeEventListener("beforeunload", preventUnload);
  }, [hasUnsavedChanges, isCreateMode]);

  function resetCredentialInputs() {
    setAPIKey("");
    setAPIKeyTouched(false);
    setAPIKeyTestSucceeded(false);
    setCreateCredentialTestResult(null);
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
    setUsePlatformManagedCredentials(item.use_platform_managed_credentials);
    resetCredentialInputs();
    setManagers([]);
    setUsers([]);
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
    if (isCreateMode && !canContinueCreateDetails) {
      setBanner({ kind: "error", message: connectionSchemaError || ux("integrations.wizard.requiredCredentialFields") });
      return null;
    }

    const providerConfig: Record<string, unknown> = {};
    const credentials: Record<string, unknown> = {};

    if (!usePlatformManagedCredentials && selectedConnectionSchema) {
      const schemaPayload = buildProviderCredentialPayload(selectedConnectionSchema, providerCredentialValues);
      Object.assign(providerConfig, schemaPayload.providerConfig);
      Object.assign(credentials, schemaPayload.credentials);
    } else if (!usePlatformManagedCredentials && authMode === "api_key") {
      if (apiKey.trim()) credentials.api_key = apiKey.trim();
      if (!isCreateMode && secretRef.trim()) credentials.secret_ref = secretRef.trim();
    } else if (!usePlatformManagedCredentials && authMode === "azure_entra_cert") {
      providerConfig.azure_tenant_id = azureTenantID.trim() || undefined;
      providerConfig.azure_client_id = azureClientID.trim() || undefined;
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    } else if (!usePlatformManagedCredentials && authMode === "aws_sigv4") {
      providerConfig.aws_region = awsRegion.trim() || undefined;
      providerConfig.aws_access_key_id = awsAccessKeyID.trim() || undefined;
      providerConfig.aws_secret_access_key = awsSecretAccessKey.trim() || undefined;
    } else if (!usePlatformManagedCredentials && authMode === "google_service_account") {
      providerConfig.google_service_account_json = googleServiceAccountJSON.trim() || undefined;
      if (secretRef.trim()) credentials.secret_ref = secretRef.trim();
    }

    const payload: Record<string, unknown> = {
      provider,
      name: name.trim(),
      auth_mode: authMode,
      use_platform_managed_credentials: usePlatformManagedCredentials,
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
      setAccounts((current) => mergeAccountIntoList(current, newAccount));
      if (!isCreateMode) {
        await loadInitial();
        await handleLoadAccount(newAccount.id);
      }
      if (advanceToModels) {
        setSetupStep(3);
        setActiveCreateSection("models");
      }
      showFeedbackToast({ kind: "success", title: options.successMessage ?? t("messages.accountCreateSuccess") });
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
    if (usePlatformManagedCredentials) return;
    if (!selectedConnectionSchema || missingRequiredCredentialFields.length > 0) {
      setCreateCredentialTestResult({
        kind: "error",
        message: ux("integrations.wizard.credentialsRequiredForTest"),
      });
      return;
    }

    const schemaPayload = buildProviderCredentialPayload(selectedConnectionSchema, providerCredentialValues);

    setIsTesting(true);
    try {
      const result = await requestJSON<CredentialTestResult>(
        "/api/provider-connections/test",
        {
          method: "POST",
          body: JSON.stringify({
            provider,
            auth_mode: authMode,
            provider_config: schemaPayload.providerConfig,
            credentials: Object.keys(schemaPayload.credentials).length > 0
              ? schemaPayload.credentials
              : undefined,
            use_platform_managed_credentials: false,
            model_hint: providerModels[0]?.model,
          }),
        },
      );
      const payload = unwrapData<CredentialTestResult>(result) ?? result;
      const rawMessage =
        payload.message ?? (payload.success ? t("messages.credentialValid") : t("messages.credentialInvalid"));
      const message = payload.unsupported
        ? rawMessage
        : payload.success
          ? ux("integrations.wizard.connectionTestSuccess", { message: rawMessage })
        : ux("integrations.wizard.connectionTestFailed", { message: rawMessage });
      setCreateCredentialTestResult({
        ...payload,
        kind: payload.unsupported ? "warning" : payload.success ? "success" : "error",
        message,
      });
      setAPIKeyTouched(!payload.success);
      setAPIKeyTestSucceeded(payload.success && !payload.unsupported);
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
    if (!provider || !name.trim()) {
      setBanner({ kind: "error", message: t("messages.validationProviderName") });
      return;
    }
    if (!canContinueCreateDetails) {
      setBanner({ kind: "error", message: connectionSchemaError || ux("integrations.wizard.requiredCredentialFields") });
      return;
    }
    if (requiresSuccessfulCreatePreflight && !apiKeyTestSucceeded) {
      setBanner({ kind: "error", message: ux("integrations.wizard.credentialsRequiredForTest") });
      return;
    }
    if (shouldPersistCreateAccount("continue-details")) return;
    setSetupStep(3);
    setFurthestCreateStepIndex((current) => Math.max(current, 1));
    setActiveCreateSection("models");
  }

  const handleSaveModels = useCallback(async (
    accountID = selectedAccountID,
    options: { reload?: boolean; announce?: boolean; allowEmpty?: boolean } = {},
    draftOverride?: Record<string, ModelDraft>,
  ): Promise<boolean> => {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }

    const { models: payloadModels, validationErrors } = buildSelectedModelPayload(draftOverride ?? modelDrafts, {
      canUseCustomPricing: canUseCustomModelPricing(account),
    });

    if (payloadModels.length === 0 && !options.allowEmpty) {
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
      modelDraftBaselineRef.current = JSON.stringify(draftOverride ?? modelDrafts);
      if (options.announce ?? true) {
        showFeedbackToast({ kind: "success", title: t("messages.modelsSaveSuccess") });
      }
      if (options.reload ?? true) {
        await handleLoadAccount(accountID);
      }
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.modelsSaveFailed")) });
      return false;
    } finally {
      setIsSavingModels(false);
    }
  }, [account, canManageProviderAccount, handleLoadAccount, modelDrafts, selectedAccountID, t, ux]);

  async function handleFinishCreateFlow() {
    if (!provider || !name.trim()) {
      setBanner({ kind: "error", message: t("messages.validationProviderName") });
      return;
    }
    if (accountNameConflict) {
      setBanner({ kind: "error", message: t("messages.accountNameExists") });
      return;
    }
    if (!canContinueCreateDetails) {
      setBanner({ kind: "error", message: connectionSchemaError || ux("integrations.wizard.requiredCredentialFields") });
      return;
    }
    if (!createPreflightReady) {
      setBanner({ kind: "error", message: ux("integrations.wizard.credentialsRequiredForTest") });
      return;
    }
    const { models: payloadModels, validationErrors } = buildSelectedModelPayload(modelDrafts, {
      canUseCustomPricing: canUseCustomModelPricing(account),
    });
    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: t("messages.selectAtLeastOneModel") });
      return;
    }
    if (validationErrors.length > 0 || unresolvedPricingCount > 0) {
      setBanner({ kind: "error", message: t("messages.validPricingRequired") });
      return;
    }
    if (managers.length === 0) {
      setBanner({ kind: "error", message: t("messages.managerRequired") });
      return;
    }
    await handleActivateAccount();
  }

  async function handleContinueFromCreateModels() {
    const { models: payloadModels, validationErrors } = buildSelectedModelPayload(modelDrafts, {
      canUseCustomPricing: canUseCustomModelPricing(account),
    });
    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: t("messages.selectAtLeastOneModel") });
      return;
    }
    if (validationErrors.length > 0) {
      const firstError = validationErrors[0];
      const reasonMessage = firstError.reason === "effective_capabilities_required"
        ? ux("messages.modelTypeRequired")
        : ux("messages.privatePricingRequired");
      setBanner({ kind: "error", message: `${firstError.model}: ${reasonMessage}` });
      return;
    }
    if (shouldPersistCreateAccount("continue-models")) return;
    setFurthestCreateStepIndex((current) => Math.max(current, 2));
    setActiveCreateSection("access");
  }

  async function handleContinueFromCreateAccess() {
    if (managers.length === 0) {
      setBanner({ kind: "error", message: t("messages.managerRequired") });
      return;
    }
    if (shouldPersistCreateAccount("continue-access")) return;
    setFurthestCreateStepIndex((current) => Math.max(current, 3));
    setActiveCreateSection("review");
  }

  const handleSaveAccess = useCallback(async (
    accountID = selectedAccountID,
    options: { reload?: boolean; announce?: boolean } = {},
    accessOverride?: { managers: AccessPrincipal[]; users: AccessPrincipal[] },
  ): Promise<boolean> => {
    const nextManagers = accessOverride?.managers ?? managers;
    const nextUsers = accessOverride?.users ?? users;

    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return false;
    }
    if (!accountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return false;
    }
    if (nextManagers.length === 0) {
      setBanner({ kind: "error", message: t("messages.managerRequired") });
      return false;
    }

    setIsSavingAccess(true);
    try {
      await requestJSON(
        `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/access`,
        { method: "PUT", body: JSON.stringify({ access: { managers: nextManagers, users: nextUsers } }) },
      );
      if (options.announce ?? true) {
        showFeedbackToast({ kind: "success", title: t("messages.accessSaveSuccess") });
      }
      if (options.reload ?? true) {
        await handleLoadAccount(accountID);
      }
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.accessSaveFailed")) });
      return false;
    } finally {
      setIsSavingAccess(false);
    }
  }, [canManageProviderAccount, handleLoadAccount, managers, selectedAccountID, t, users]);

  async function handleTestCredentials() {
    setCreateCredentialTestResult(null);
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: t("messages.selectAccountFirst") });
      return;
    }
    if (!usePlatformManagedCredentials) {
      const saved = await handleSaveCredentials(selectedAccountID, { reload: false, announce: false });
      if (!saved) return;
    }
    setIsTesting(true);
    try {
      const result = await requestJSON<CredentialTestResult>(
        `/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/test-credentials`,
        { method: "POST", body: JSON.stringify({}) },
      );
      const payload = unwrapData<CredentialTestResult>(result) ?? result;
      const nextMessage =
        payload.message ?? (payload.success ? t("messages.credentialValid") : t("messages.credentialInvalid"));
      setCreateCredentialTestResult({
        ...payload,
        kind: payload.success ? "success" : "error",
        message: payload.success
          ? ux("integrations.wizard.connectionTestSuccess", { message: nextMessage })
          : ux("integrations.wizard.connectionTestFailed", { message: nextMessage }),
      });
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

  async function handleActivateAccount() {
    if (!canManageProviderAccount) {
      setBanner({ kind: "error", message: t("messages.insufficientPermissions") });
      return;
    }
    setIsActivating(true);
    try {
      if (!shouldPersistCreateAccount("activate")) return;
      let accountID: string | null = selectedAccountID;
      if (!accountID) {
        accountID = await handleCreateAccount({ advanceToModels: false });
        if (!accountID) return;
      } else {
        const credentialsOK = await handleSaveCredentials(accountID, { reload: false, announce: false });
        if (!credentialsOK) return;
      }
      const modelsOK = await handleSaveModels(accountID, { reload: false, announce: false });
      if (!modelsOK) {
        if (isCreateMode) {
          router.replace(`/provider-accounts?${new URLSearchParams({ account: accountID, section: "models" }).toString()}`);
        }
        return;
      }
      const accessOK = await handleSaveAccess(accountID, { reload: false, announce: false });
      if (!accessOK) {
        if (isCreateMode) {
          router.replace(`/provider-accounts?${new URLSearchParams({ account: accountID, section: "access" }).toString()}`);
        }
        return;
      }
      setSetupStep(5);
      if (isCreateMode) {
        router.replace(`/provider-accounts?${new URLSearchParams({ account: accountID, section: "overview" }).toString()}`);
        return;
      }
      setWorkspaceMode("workspace");
      setActiveTab("models");
      showFeedbackToast({ kind: "success", title: ux("messages.accountActivated") });
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

  const handleSaveModelsAction = useCallback((draftOverride?: Record<string, ModelDraft>) => {
    void handleSaveModels(undefined, { reload: false, announce: false, allowEmpty: true }, draftOverride);
  }, [handleSaveModels]);

  const handleSaveAccessAction = useCallback(() => {
    void handleSaveAccess();
  }, [handleSaveAccess]);

  const handleAccessAutosave = useCallback((nextAccess: { managers: AccessPrincipal[]; users: AccessPrincipal[] }) => {
    if (!selectedAccountID) return;

    accessAutosaveQueueRef.current = accessAutosaveQueueRef.current
      .catch(() => undefined)
      .then(async () => {
        await handleSaveAccess(selectedAccountID, { reload: false, announce: false }, nextAccess);
      });
  }, [handleSaveAccess, selectedAccountID]);

  const handleManagersChange = useCallback((nextManagers: AccessPrincipal[]) => {
    const mustKeepCurrentUser = isCreateMode && Boolean(currentUserManager);
    const hasCurrentUser = !currentUserManager || nextManagers.some(
      (manager) => manager.type === "user" && manager.principal_id === currentUserManager.principal_id,
    );
    const normalizedManagers = mustKeepCurrentUser && currentUserManager && !hasCurrentUser
      ? [currentUserManager, ...nextManagers]
      : nextManagers;
    setManagers(normalizedManagers);
    handleAccessAutosave({ managers: normalizedManagers, users });
  }, [currentUserManager, handleAccessAutosave, isCreateMode, users]);

  const handleUsersChange = useCallback((nextUsers: AccessPrincipal[]) => {
    setUsers(nextUsers);
    handleAccessAutosave({ managers, users: nextUsers });
  }, [handleAccessAutosave, managers]);

  function navigateWorkspaceTab(tab: WorkspaceTab) {
    if (tab === visibleActiveTab) return;
    if (hasUnsavedChanges) {
      setPendingWorkspaceNavigation({ kind: "tab", tab });
      setUnsavedDialogOpen(true);
      return;
    }
    applyWorkspaceTab(tab);
  }

  function applyWorkspaceTab(tab: WorkspaceTab) {
    setActiveTab(tab);
    if (account) {
      router.push(
        `/provider-accounts?${new URLSearchParams({
          account: account.id,
          section: tab,
        }).toString()}`,
      );
    }
  }

  function navigateWorkspaceHref(href: string) {
    if (hasUnsavedChanges) {
      setPendingWorkspaceNavigation({ kind: "href", href });
      setUnsavedDialogOpen(true);
      return;
    }
    router.push(href);
  }

  function discardWorkspaceChanges() {
    if (modelDraftBaselineRef.current) {
      setModelDrafts(JSON.parse(modelDraftBaselineRef.current) as Record<string, ModelDraft>);
    }
    const credentialBaseline = credentialDraftBaselineRef.current;
    if (credentialBaseline) {
      setAuthMode(credentialBaseline.authMode);
      setUsePlatformManagedCredentials(credentialBaseline.usePlatformManagedCredentials);
      setProviderCredentialValues({ ...credentialBaseline.values });
    }
    const pending = pendingWorkspaceNavigation;
    setPendingWorkspaceNavigation(null);
    if (pending?.kind === "tab") applyWorkspaceTab(pending.tab);
    if (pending?.kind === "href") router.push(pending.href);
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
    { id: "overview", label: catalogT("details.sections.overview"), icon: LayoutDashboard },
    { id: "models", label: ux("tabs.models"), icon: Layers3 },
    { id: "access", label: ux("tabs.access"), icon: ShieldCheck },
    { id: "credentials", label: ux("tabs.credentials"), icon: KeyRound },
  ];
  const visibleActiveTab = activeTab;
  const principalSelectorLabels = useMemo(() => buildPrincipalSelectorLabels(t, ux), [t, ux]);
  const createHeaderProviderName = formatProviderDisplayName(
    activeProvider?.provider ?? provider,
    activeProvider?.display_name,
  );
  const createWizardSteps: Array<{ id: CreateWizardSectionID; number: number; title: string }> = [
    { id: "details", number: 1, title: ux("integrations.wizard.integrationDetails") },
    { id: "models", number: 2, title: ux("integrations.wizard.modelProvisioning") },
    { id: "access", number: 3, title: ux("integrations.wizard.accessPermissions") },
    { id: "review", number: 4, title: ux("integrations.wizard.reviewConfirm") },
  ];
  const activeCreateStepIndex = createWizardSteps.findIndex((step) => step.id === activeCreateSection);
  const createStepState = (stepID: CreateWizardSectionID) => {
    return getCreateStepState({
      activeStep: activeCreateSection,
      step: stepID,
      completedThrough: furthestCreateStepIndex,
    });
  };
  const canOpenCreateStep = (stepID: CreateWizardSectionID) => {
    const stepIndex = createWizardSteps.findIndex((step) => step.id === stepID);
    return stepIndex >= 0 && stepIndex <= Math.max(activeCreateStepIndex, furthestCreateStepIndex);
  };
  const CreatePanel = CatalogCreateSection;
  const catalogStepCount = createWizardSteps.length + 1;
  const catalogActiveStepNumber = provider ? Math.max(activeCreateStepIndex, 0) + 2 : 1;
  const catalogActiveStepTitle = provider
    ? createWizardSteps[activeCreateStepIndex]?.title ?? catalogT("create.configureProvider")
    : catalogT("create.chooseProvider");
  const catalogStepProgressLabel = catalogT("create.stepProgress", {
    current: catalogActiveStepNumber,
    total: catalogStepCount,
  });
  const resetCreateProviderSelection = () => {
    setProvider("");
    setSelectedProviderID("");
    setFurthestCreateStepIndex(0);
    setActiveCreateSection("details");
  };

  if (isCreateMode) {
    return (
      <PageContainer size="full" spacing="compact" padding="none">
        <header className="flex min-h-14 items-center rounded-xl border border-card-border bg-card px-4 py-3 shadow-xs sm:px-6">
          <div className="flex min-w-0 items-center gap-2 text-base sm:text-lg">
            <span className="shrink-0 text-muted-foreground">{ux("integrations.pageTitle")}</span>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground rtl:rotate-180" />
            {provider ? (
              <span className="flex min-w-0 items-center gap-2 font-semibold">
                <ProviderLogo
                  provider={activeProvider?.provider ?? provider}
                  displayName={createHeaderProviderName}
                  logoURL={activeProvider?.logo_url}
                  size={20}
                  decorative
                />
                <span className="truncate" dir="ltr">{createHeaderProviderName}</span>
              </span>
            ) : (
              <span className="truncate font-semibold">{catalogT("create.providerPlaceholder")}</span>
            )}
          </div>
        </header>

        {banner && banner.kind !== "success" ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

        {selectedAccountID ? (
          <Alert variant="info">
            <AlertTitle>{ux("integrations.wizard.draftSavedTitle")}</AlertTitle>
            <AlertDescription>{ux("integrations.wizard.draftSavedDescription")}</AlertDescription>
          </Alert>
        ) : null}

        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <nav
            aria-label={ux("integrations.pageTitle")}
            className="min-w-0"
          >
            <div className="rounded-xl border border-card-border bg-card p-4 shadow-xs xl:hidden">
              <div className="flex items-center justify-between gap-3">
                <span className="text-xs font-medium text-muted-foreground">
                  {catalogStepProgressLabel}
                </span>
                <span className="text-sm font-semibold">{catalogActiveStepTitle}</span>
              </div>
              <Progress
                value={(catalogActiveStepNumber / catalogStepCount) * 100}
                className="mt-3 h-1.5"
                dir="ltr"
                aria-label={catalogStepProgressLabel}
              />
            </div>
            <div className="hidden grid-cols-5 gap-2 xl:grid">
              <CreateFlowStepIndicator
                number={1}
                title={catalogT("create.chooseProvider")}
                state={provider ? "complete" : "active"}
                variant="compact"
                onClick={resetCreateProviderSelection}
              />
              {createWizardSteps.map((step) => {
                const stepState = createStepState(step.id);
                const isStepOpenable = Boolean(provider) && canOpenCreateStep(step.id);
                return (
                  <CreateFlowStepIndicator
                    key={step.id}
                    number={step.number + 1}
                    title={step.title}
                    state={provider ? stepState : "idle"}
                    variant="compact"
                    disabled={!isStepOpenable}
                    onClick={() => {
                      if (isStepOpenable) setActiveCreateSection(step.id);
                    }}
                  />
                );
              })}
            </div>
          </nav>

          <div className="min-w-0">
            <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 pb-2">
                {!provider ? (
                  <CreatePanel
                    title=""
                    footer={
                      <Button type="button" variant="outline" asChild>
                        <Link href="/provider-accounts">{ux("actions.cancelDelete")}</Link>
                      </Button>
                    }
                  >
                    <div className="flex flex-col gap-4 sm:gap-5">
                      <DebouncedTextFilter
                        label={catalogT("filters.search")}
                        value={providerSearch}
                        onValueChange={setProviderSearch}
                        placeholder={catalogT("filters.searchProviderPicker")}
                      />
                      <div
                        className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
                      >
                        {isLoading ? Array.from({ length: 6 }).map((_, index) => (
                          <div key={`provider-picker-skeleton-${index}`} className="flex min-h-20 items-center gap-3 rounded-lg border border-card-border bg-background p-3">
                            <Skeleton className="size-10 shrink-0 rounded-lg" />
                            <Skeleton className="h-4 w-32 max-w-full" />
                          </div>
                        )) : visibleProviders.length === 0 ? (
                          <EmptyState
                            icon={SearchX}
                            title={catalogT("empty.noResultsTitle")}
                            description={catalogT("empty.noResultsDescription")}
                            variant="inline"
                            density="compact"
                            className="sm:col-span-2 xl:col-span-3"
                          />
                        ) : visibleProviders.map((item) => {
                          const isSelected = stringsEqualFold(item.provider, provider);
                          return (
                            <button
                              key={item.provider}
                              type="button"
                              className={cn(
                                "group flex min-h-20 min-w-0 items-center gap-3 rounded-lg border border-card-border bg-background p-3 text-start shadow-xs transition hover:border-primary/40 hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring active:translate-y-px",
                                isSelected && "border-primary bg-primary/10",
                              )}
                              onClick={() => {
                                const nextProvider = normalizeProviderSlug(item.provider);
                                setProvider(nextProvider);
                                setAuthMode(item.auth_modes?.[0] ?? "api_key");
                                setName(generateProviderAccountName(nextProvider, existingAccountNames));
                                setIsNameAutoGenerated(true);
                                setSelectedProviderID(nextProvider);
                                setUsePlatformManagedCredentials(true);
                                setAPIKey("");
                                setAPIKeyTestSucceeded(false);
                                 setCreateCredentialTestResult(null);
                                 setFurthestCreateStepIndex(0);
                                 setActiveCreateSection("details");
                                 void loadProviderModels(item.provider);
                              }}
                            >
                              <ProviderLogo
                                provider={item.provider}
                                displayName={formatProviderDisplayName(item.provider, item.display_name)}
                                logoURL={item.logo_url}
                                size={28}
                                decorative
                              />
                              <p className="min-w-0 flex-1 truncate text-sm font-semibold" dir="ltr">
                                {formatProviderDisplayName(item.provider, item.display_name)}
                              </p>
                              <ChevronRight
                                aria-hidden="true"
                                className="size-4 shrink-0 text-muted-foreground transition-colors group-hover:text-primary rtl:rotate-180"
                              />
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  </CreatePanel>
                ) : (
                  <>
                    {activeCreateSection === "details" ? (
                      <CreatePanel
                        id="connection-details"
                        title=""
                        fillHeight={false}
                        footer={(
                          <div className="flex w-full flex-wrap items-center justify-between gap-2">
                            <Button type="button" variant="outline" onClick={resetCreateProviderSelection}>
                              {ux("actions.back")}
                            </Button>
                            <Button
                              type="button"
                              onClick={() => void handleContinueFromCreateDetails()}
                              disabled={!canCreateProviderAccount || !canContinueCreateDetails || (requiresSuccessfulCreatePreflight && !apiKeyTestSucceeded) || isLoading || isSavingCredentials || !provider || !name.trim()}
                            >
                              {isLoading || isSavingCredentials ? t("saving") : ux("integrations.wizard.nextStep")}
                            </Button>
                          </div>
                        )}
                      >
                        <div className="flex flex-col gap-6">
                          <EntityIdentityHeader
                            title={<span dir="ltr">{createHeaderProviderName}</span>}
                            leading={(
                              <div className="flex size-12 items-center justify-center rounded-lg border border-border bg-muted/30">
                                <ProviderLogo
                                  provider={activeProvider?.provider ?? provider}
                                  displayName={createHeaderProviderName}
                                  logoURL={activeProvider?.logo_url}
                                  size={28}
                                  decorative
                                />
                              </div>
                            )}
                            className="border-b border-border pb-5"
                            titleClassName="flex h-12 items-center"
                          />

                          <IdentifierField
                            id="accountName"
                            label={t("fields.accountName")}
                            value={name}
                            onChange={(event) => {
                              setName(event.target.value);
                              setIsNameAutoGenerated(false);
                            }}
                            placeholder={ux("placeholders.accountName")}
                            disabled={Boolean(selectedAccountID)}
                            error={accountNameConflict ? t("messages.accountNameExists") : undefined}
                            containerClassName="max-w-2xl"
                          />

                      <div className="text-start">
                        <FieldSet className="gap-0">
                          <FieldLegend className="mb-0">
                            {ux("integrations.wizard.useOwnApiKeyQuestion", { providerName: createHeaderProviderName })}
                          </FieldLegend>
                          <FieldDescription className="mt-1 text-sm leading-6">
                            {ux.rich("integrations.wizard.useOwnApiKeyDescription", {
                              guideLink: (chunks) => (
                                <a
                                  href={`${docsBaseHref}/quick-start`}
                                  target="_blank"
                                  rel="noreferrer"
                                  className="font-medium text-primary underline-offset-4 hover:underline"
                                >
                                  {chunks}
                                </a>
                              ),
                            })}
                          </FieldDescription>

                          <RadioGroup
                            name="create-own-api-key"
                            value={usePlatformManagedCredentials ? "platform-managed" : "byok"}
                            onValueChange={(nextValue) => {
                               if (account) return;
                               setUsePlatformManagedCredentials(nextValue === "platform-managed");
                               setAPIKeyTestSucceeded(false);
                               setCreateCredentialTestResult(null);
                            }}
                            disabled={!!account}
                            className="mt-4 grid gap-3 md:grid-cols-2"
                          >
                            <Field
                              orientation="horizontal"
                              className="relative order-1 cursor-pointer items-start rounded-md border border-border p-4 transition-colors hover:bg-muted/30 focus-within:ring-2 focus-within:ring-ring/50 has-[[data-state=checked]]:border-primary has-[[data-state=checked]]:bg-primary/5"
                            >
                              <FieldLabel
                                htmlFor="create-own-api-key-byok"
                                className="absolute inset-0 size-full cursor-pointer rounded-md"
                              >
                                <span className="sr-only">{ux("integrations.wizard.useOwnApiKeyYes")}</span>
                              </FieldLabel>
                              <RadioGroupItem id="create-own-api-key-byok" value="byok" className="pointer-events-none relative z-10 mt-1" />
                              <FieldContent className="pointer-events-none relative z-10 min-w-0">
                                <FieldTitle>
                                  {ux("integrations.wizard.useOwnApiKeyYes")}
                                </FieldTitle>
                                <FieldDescription className="text-sm leading-6">
                                  {ux.rich("integrations.wizard.useOwnApiKeyYesDescription", {
                                    byokLink: (chunks) => (
                                      <a
                                        href={`${docsBaseHref}/guides/byok`}
                                        target="_blank"
                                        rel="noreferrer"
                                        className="pointer-events-auto relative z-20 font-medium text-primary underline-offset-4 hover:underline"
                                      >
                                        {chunks}
                                      </a>
                                    ),
                                  })}
                                </FieldDescription>
                              </FieldContent>
                            </Field>

                            <Field
                              orientation="horizontal"
                              className="relative order-3 cursor-pointer items-start rounded-md border border-border p-4 transition-colors hover:bg-muted/30 focus-within:ring-2 focus-within:ring-ring/50 has-[[data-state=checked]]:border-primary has-[[data-state=checked]]:bg-primary/5 md:order-2"
                            >
                              <FieldLabel
                                htmlFor="create-own-api-key-platform"
                                className="absolute inset-0 size-full cursor-pointer rounded-md"
                              >
                                <span className="sr-only">{ux("integrations.wizard.useOwnApiKeyNo")}</span>
                              </FieldLabel>
                              <RadioGroupItem id="create-own-api-key-platform" value="platform-managed" className="pointer-events-none relative z-10 mt-1" />
                              <FieldContent className="pointer-events-none relative z-10 min-w-0">
                                <FieldTitle>
                                  {ux("integrations.wizard.useOwnApiKeyNo")}
                                </FieldTitle>
                                <FieldDescription className="text-sm leading-6">
                                  {ux("integrations.wizard.useOwnApiKeyNoDescription", { providerName: createHeaderProviderName })}
                                </FieldDescription>
                              </FieldContent>
                            </Field>

                            {!usePlatformManagedCredentials ? (
                              <FieldGroup className="order-2 rounded-e-lg border-s-2 border-primary/40 bg-muted/20 p-4 md:order-3 md:col-span-2 md:rounded-none md:border-s-0 md:bg-transparent md:p-0">
                                {isLoadingConnectionSchemas ? (
                                  <p className="text-sm text-muted-foreground" role="status">
                                    {ux("integrations.wizard.loadingConnectionSchema")}
                                  </p>
                                ) : null}
                                {connectionSchemaError ? (
                                  <Alert variant="destructive">
                                    <AlertTitle>{ux("integrations.wizard.connectionSchemaErrorTitle")}</AlertTitle>
                                    <AlertDescription>{connectionSchemaError}</AlertDescription>
                                  </Alert>
                                ) : null}
                                {connectionSchemas.length > 1 ? (
                                  <AppSelectField
                                    id="createProviderAuthMode"
                                    label={t("fields.authMode")}
                                    value={authMode}
                                    optionTextDir="ltr"
                                    options={connectionSchemas.map((schema) => ({
                                      value: schema.auth_mode,
                                      label: schema.auth_mode,
                                    }))}
                                    onValueChange={(value) => {
                                      const nextSchema = connectionSchemas.find((schema) => schema.auth_mode === value) ?? null;
                                     setAuthMode(value as ProviderAuthMode);
                                     setProviderCredentialValues(defaultProviderCredentialValues(nextSchema));
                                     setAPIKeyTestSucceeded(false);
                                     setCreateCredentialTestResult(null);
                                    }}
                                  />
                                ) : null}
                                {selectedConnectionSchema ? (
                                  <ProviderCredentialSchemaFields
                                    schema={selectedConnectionSchema}
                                    values={providerCredentialValues}
                                    onValueChange={(key, value) => {
                                      setProviderCredentialValues((current) => ({ ...current, [key]: value }));
                                      setAPIKeyTestSucceeded(false);
                                      setCreateCredentialTestResult(null);
                                    }}
                                    disabled={isLoading || isSavingCredentials || isTesting}
                                    connectionFieldsTitle={ux("integrations.wizard.connectionFieldsTitle")}
                                    credentialFieldsTitle={ux("integrations.wizard.credentialFieldsTitle")}
                                    writeOnlyLabel={ux("integrations.wizard.writeOnlyCredential")}
                                    documentationLabel={ux("integrations.wizard.providerDocumentation")}
                                  />
                                ) : null}
                                <div className="flex flex-wrap items-center justify-between gap-3">
                                  <p className="text-sm leading-6 text-muted-foreground">
                                    {ux("integrations.wizard.testConnectionDescription")}
                                  </p>
                                  <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => void handleCreateCredentialTest()}
                                    disabled={!canCreateProviderAccount || !canContinueCreateDetails || isLoading || isSavingCredentials || isTesting}
                                  >
                                    {isTesting || isSavingCredentials || isLoading
                                      ? ux("integrations.wizard.testingConnection")
                                      : ux("integrations.wizard.testConnection")}
                                  </Button>
                                </div>
                                {createCredentialTestResult ? (
                                  <Alert variant={createCredentialTestResult.kind === "error" ? "destructive" : createCredentialTestResult.kind === "warning" ? "warning" : "default"}>
                                    <AlertTitle>
                                      {createCredentialTestResult.kind === "warning"
                                        ? ux("connectionTest.warningTitle")
                                        : createCredentialTestResult.kind === "error"
                                        ? ux("integrations.wizard.connectionTestFailedTitle")
                                        : ux("integrations.wizard.connectionTestSuccessTitle")}
                                    </AlertTitle>
                                    <AlertDescription>{createCredentialTestResult.message}</AlertDescription>
                                  </Alert>
                                ) : null}
                                <WIFCredentialStatus result={createCredentialTestResult} />
                              </FieldGroup>
                            ) : null}
                          </RadioGroup>
                        </FieldSet>
                      </div>
                    </div>
                  </CreatePanel>
                    ) : null}

                  {activeCreateSection === "models" ? (
                    <CreatePanel
                      id="model-provisioning"
                      title=""
                      fillHeight={false}
                      contentClassName="flex min-h-0 flex-1 flex-col"
                      footer={(
                        <div className="flex w-full flex-wrap items-center justify-between gap-3">
                          <Button type="button" variant="outline" onClick={() => setActiveCreateSection("details")}>
                            {ux("actions.back")}
                          </Button>
                          <div className="flex flex-wrap items-center justify-end gap-3">
                            {selectedModelCount === 0 ? (
                              <span className="text-sm text-destructive">
                                {ux("integrations.wizard.modelSelectionRequired")}
                              </span>
                            ) : null}
                            <Button
                              type="button"
                              onClick={() => void handleContinueFromCreateModels()}
                              disabled={!canManageProviderAccount || selectedModelCount === 0 || isSavingModels}
                            >
                              {isSavingModels ? t("saving") : ux("actions.continue")}
                            </Button>
                          </div>
                        </div>
                      )}
                    >
                    <ModelsEditor
                      t={t}
                      account={account}
                      pendingAccountName={name}
                      pendingProvider={provider}
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
                  </CreatePanel>
                  ) : null}

                  {activeCreateSection === "access" ? (
                    <CreatePanel
                      id="access-permissions"
                      title={ux("integrations.wizard.accessPermissions")}
                      description={ux("integrations.wizard.accessPermissionsDescription")}
                      fillHeight={false}
                      footer={(
                        <div className="flex w-full flex-wrap items-center justify-between gap-3">
                          <Button type="button" variant="outline" onClick={() => setActiveCreateSection("models")}>
                            {ux("actions.back")}
                          </Button>
                          <div className="flex flex-wrap items-center justify-end gap-3">
                            {managers.length === 0 ? <span className="text-sm text-destructive">{ux("access.managerRequiredHint")}</span> : null}
                            <Button
                              type="button"
                              onClick={() => void handleContinueFromCreateAccess()}
                              disabled={!canManageProviderAccount || managers.length === 0 || isSavingAccess}
                            >
                              {isSavingAccess ? t("saving") : ux("actions.continue")}
                            </Button>
                          </div>
                        </div>
                      )}
                    >
                    <AccessEditor
                      t={t}
                      managers={managers}
                      setManagers={handleManagersChange}
                      users={users}
                      setUsers={handleUsersChange}
                      onSearchPrincipal={searchPrincipalCandidates}
                      principalSelectorLabels={principalSelectorLabels}
                      onSave={handleSaveAccessAction}
                      isSaving={isSavingAccess}
                       canManage={canManageProviderAccount}
                       showSaveAction={false}
                       embedded
                       isManagerRemovable={currentUserManager
                         ? (manager) => !(
                             manager.type === "user"
                             && manager.principal_id === currentUserManager.principal_id
                           )
                         : undefined}
                       getLockedManagerLabel={(manager) => ux("access.currentUserLocked", {
                         identity: manager.display_name || manager.principal_id || manager.type,
                       })}
                     />
                  </CreatePanel>
                  ) : null}

                  {activeCreateSection === "review" ? (
                    <CreatePanel
                      id="review-confirm"
                      title={ux("integrations.wizard.reviewConfirm")}
                      description={ux("integrations.wizard.reviewConfirmDescription")}
                      fillHeight={false}
                      footer={(
                        <div className="flex w-full flex-wrap items-center justify-between gap-3">
                          <Button type="button" variant="outline" onClick={() => setActiveCreateSection("access")}>
                            {ux("actions.back")}
                          </Button>
                          <div className="flex flex-wrap items-center justify-end gap-3">
                            {!createReady ? (
                              <span className="text-sm text-muted-foreground">{catalogT("create.completeRequired")}</span>
                            ) : null}
                            <Button
                              type="button"
                              onClick={() => void handleFinishCreateFlow()}
                              disabled={!canCreateProviderAccount || !createReady || isActivating || isTesting || isLoading}
                              aria-busy={isActivating}
                            >
                              {isActivating ? ux("actions.activating") : ux("actions.activateAccount")}
                            </Button>
                          </div>
                        </div>
                      )}
                    >
                    <div className="space-y-4">
                      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                        <SummaryKPI label={ux("kpi.credentialStatus")} value={account?.has_credential || usePlatformManagedCredentials || apiKey.trim() ? ux("readiness.ready") : ux("readiness.needsCredentials")} />
                        <SummaryKPI label={ux("kpi.enabledModels")} value={selectedModelCount.toString()} />
                        <SummaryKPI label={ux("kpi.unresolvedPricing")} value={unresolvedPricingCount.toString()} />
                        <SummaryKPI label={ux("kpi.managersUsers")} value={`${managers.length} / ${users.length}`} />
                      </div>
                      <section className="space-y-3 rounded-lg border border-border bg-background p-4">
                        <div className="flex items-center justify-between gap-3">
                          <h3 className="font-semibold">{ux("review.connectionTitle")}</h3>
                          <Button type="button" variant="link" size="sm" onClick={() => setActiveCreateSection("details")}>
                            {ux("actions.edit")}
                          </Button>
                        </div>
                        <DefinitionList columns={2}>
                          <DefinitionListItem label={ux("review.provider")} value={createHeaderProviderName} valueDir="ltr" surface />
                          <DefinitionListItem label={ux("review.accountName")} value={name} valueDir="ltr" mono surface />
                          <DefinitionListItem label={ux("review.authMode")} value={authMode.replaceAll("_", " ")} valueDir="ltr" surface />
                          <DefinitionListItem
                            label={ux("review.credentialSource")}
                            value={usePlatformManagedCredentials ? ux("review.godarKey") : ux("review.providerKey")}
                            surface
                          />
                        </DefinitionList>
                      </section>

                      <section className="space-y-3 rounded-lg border border-border bg-background p-4">
                        <div className="flex items-center justify-between gap-3">
                          <h3 className="font-semibold">{ux("review.modelsTitle")}</h3>
                          <Button type="button" variant="link" size="sm" onClick={() => setActiveCreateSection("models")}>
                            {ux("actions.edit")}
                          </Button>
                        </div>
                        <div className="grid gap-2 sm:grid-cols-2">
                          {selectedModels.map((model) => (
                            <div key={model.model} className="min-w-0 rounded-lg border border-border bg-muted/20 px-3 py-2">
                              <p className="truncate text-sm font-medium" dir="ltr" title={model.display_name || model.model}>
                                {model.display_name || model.model}
                              </p>
                              <p className="mt-1 text-xs text-muted-foreground">
                                {ux("models.card.context")}: <bdi>{model.context_window?.toLocaleString() ?? ux("models.na")}</bdi>
                                {" · "}{ux("models.labels.input")}: <bdi dir="ltr">{formatDisplayPrice(model.public_pricing?.input_usd_per_1m, {
                                  currency: "USD",
                                  scale: "PER_MILLION_TOKENS",
                                  usdToIrr: 0,
                                  naLabel: ux("models.na"),
                                })}</bdi>
                                {" · "}{ux("models.labels.output")}: <bdi dir="ltr">{formatDisplayPrice(model.public_pricing?.output_usd_per_1m, {
                                  currency: "USD",
                                  scale: "PER_MILLION_TOKENS",
                                  usdToIrr: 0,
                                  naLabel: ux("models.na"),
                                })}</bdi>
                              </p>
                            </div>
                          ))}
                        </div>
                      </section>

                      <section className="space-y-3 rounded-lg border border-border bg-background p-4">
                        <div className="flex items-center justify-between gap-3">
                          <h3 className="font-semibold">{ux("review.accessTitle")}</h3>
                          <Button type="button" variant="link" size="sm" onClick={() => setActiveCreateSection("access")}>
                            {ux("actions.edit")}
                          </Button>
                        </div>
                        <DefinitionList columns={2}>
                          <DefinitionListItem
                            label={ux("review.managers")}
                            value={managers.map((item) => item.display_name || item.principal_id || item.type).join("، ")}
                            wrap
                            surface
                          />
                          <DefinitionListItem
                            label={ux("review.users")}
                            value={users.length > 0
                              ? users.map((item) => item.display_name || item.principal_id || item.type).join("، ")
                              : ux("review.noUsers")}
                            wrap
                            surface
                          />
                        </DefinitionList>
                      </section>
                      <Alert>
                        <AlertTitle>{ux("review.activationCheckTitle")}</AlertTitle>
                        <AlertDescription>
                          {ux("microcopy.pricingPolicy")}
                        </AlertDescription>
                      </Alert>
                    </div>
                  </CreatePanel>
                  ) : null}
                  </>
                )}
              </div>
            </div>
          </div>
        </PageContainer>
    );
  }

  if (initialAccountID && !account) {
    return (
      <div id="tenant-provider-accounts" className="min-w-0">
        {isLoading ? (
          <ProviderDetailsSkeleton label={catalogT("details.loading")} />
        ) : (
          <div className="flex flex-col gap-4">
            <OperationNotice
              notice={{ kind: "error", message: banner?.message ?? catalogT("errors.loadDetails") }}
              labels={{
                success: catalogT("notices.success"),
                error: catalogT("notices.error"),
                warning: catalogT("notices.warning"),
                info: catalogT("notices.info"),
              }}
            />
            <div>
              <Button type="button" variant="outline" onClick={() => void loadInitial()}>
                {catalogT("actions.retry")}
              </Button>
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div id="tenant-provider-accounts" className="min-w-0">
      <div className="flex min-w-0 flex-col gap-4">
        {banner ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? t("errorTitle") : t("successTitle")}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

        <section className="space-y-5">
          {workspaceMode === "workspace" ? null : (
            <CardHeader className="px-0 pb-4 pt-0">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <CardTitle className="text-xl">{panelTitle}</CardTitle>
                  <CardDescription>
                    {workspaceMode === "setup" ? ux("panel.setupHint") : ux("panel.overviewHint")}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
          )}
          <div className="space-y-5">
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
                            const selectProvider = () => {
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
                            };
                            const providerProfileHref = buildModelDiscoveryProviderHref(
                              item.provider,
                              process.env.NEXT_PUBLIC_MODEL_DISCOVERY_APP_URL,
                            );
                            return (
                              <div
                                key={item.provider}
                                className={`rounded-lg border text-start transition ${
                                  isSelected ? "border-primary bg-primary/10" : "border-border hover:border-primary/40"
                                }`}
                              >
                                <div className="flex items-center gap-2 p-3 pb-0">
                                  <button
                                    type="button"
                                    className="flex min-w-0 flex-1 items-center gap-2 text-start"
                                    onClick={selectProvider}
                                  >
                                    <ProviderLogo
                                      key={`${item.provider}:${item.logo_url ?? ""}`}
                                      provider={item.provider}
                                      displayName={item.display_name}
                                      logoURL={item.logo_url}
                                    />
                                    <span className="truncate font-medium">{item.display_name}</span>
                                  </button>
                                  <Button type="button" variant="ghost" size="icon" className="h-8 w-8 shrink-0" asChild>
                                    <a
                                      href={providerProfileHref}
                                      target="_blank"
                                      rel="noreferrer"
                                      aria-label={ux("setup.providerProfileLink", { provider: item.display_name })}
                                      title={ux("setup.providerProfileLink", { provider: item.display_name })}
                                    >
                                      <ExternalLink className="h-4 w-4" />
                                    </a>
                                  </Button>
                                </div>
                                <button type="button" className="block w-full p-3 pt-2 text-start" onClick={selectProvider}>
                                  <span className="text-xs text-muted-foreground">{ux("setup.authModes")}: {item.auth_modes.join(", ")}</span>
                                </button>
                              </div>
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
                          {ux("microcopy.platformManagedCredentials")}
                        </p>

                        <div className="grid gap-3 md:grid-cols-2">
                          <Field>
                            <FieldLabel htmlFor="provider">{t("fields.provider")}</FieldLabel>
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
                          <IdentifierField
                            id="accountName"
                            label={t("fields.accountName")}
                            value={name}
                            onChange={(event) => {
                              setName(event.target.value);
                              setIsNameAutoGenerated(false);
                            }}
                            placeholder={ux("placeholders.accountName")}
                          />
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
                          <Field>
                            <FieldLabel htmlFor="platformManagedCredentials">{t("fields.credentialSourceMode")}</FieldLabel>
                            <label className="inline-flex h-10 items-center gap-2 rounded-md border border-input px-3 text-sm">
                              <input
                                id="platformManagedCredentials"
                                type="checkbox"
                                checked={usePlatformManagedCredentials}
                                onChange={(event) => setUsePlatformManagedCredentials(event.target.checked)}
                              />
                              {t("fields.usePlatformManagedCredentials")}
                            </label>
                          </Field>
                        </div>

                        {!usePlatformManagedCredentials ? (
                          selectedConnectionSchema ? (
                            <ProviderCredentialSchemaFields
                              schema={selectedConnectionSchema}
                              values={providerCredentialValues}
                              onValueChange={(key, value) => {
                                setProviderCredentialValues((current) => ({ ...current, [key]: value }));
                                setAPIKeyTestSucceeded(false);
                                setCreateCredentialTestResult(null);
                              }}
                              disabled={isLoading || isSavingCredentials || isTesting}
                              connectionFieldsTitle={ux("integrations.wizard.connectionFieldsTitle")}
                              credentialFieldsTitle={ux("integrations.wizard.credentialFieldsTitle")}
                              writeOnlyLabel={ux("integrations.wizard.writeOnlyCredential")}
                              documentationLabel={ux("integrations.wizard.providerDocumentation")}
                            />
                          ) : (
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
                          )
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
                          setManagers={handleManagersChange}
                          users={users}
                          setUsers={handleUsersChange}
                          onSearchPrincipal={searchPrincipalCandidates}
                          principalSelectorLabels={principalSelectorLabels}
                          onSave={handleSaveAccessAction}
                          isSaving={isSavingAccess}
                          canManage={canManageProviderAccount}
                          showSaveAction={false}
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
                          <SummaryKPI label={ux("kpi.credentialStatus")} value={account?.has_credential || usePlatformManagedCredentials ? ux("readiness.ready") : ux("readiness.needsCredentials")} />
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
                    <div className="flex min-w-0 flex-col gap-3 rounded-xl border border-card-border bg-card p-2 shadow-xs sm:flex-row sm:items-center sm:justify-between">
                      <Tabs
                        value={visibleActiveTab}
                        onValueChange={(value) => navigateWorkspaceTab(value as WorkspaceTab)}
                        className="min-w-0 flex-1"
                      >
                        <TabsList variant="line" aria-label={catalogT("details.sectionsLabel")} className="scrollbar-hidden w-full max-w-full justify-start overflow-x-auto">
                          {workspaceTabs.map((tab) => {
                            const Icon = tab.icon;
                            return (
                              <TabsTrigger key={tab.id} value={tab.id}>
                                <Icon data-icon="inline-start" aria-hidden="true" />
                                {tab.label}
                              </TabsTrigger>
                            );
                          })}
                        </TabsList>
                      </Tabs>
                    </div>
                    <EntityIdentityHeader
                      title={<span dir="ltr">{account.name}</span>}
                      description={formatProviderDisplayName(account.provider, activeProvider?.display_name)}
                      leading={
                        <ProviderLogo
                          provider={account.provider}
                          displayName={formatProviderDisplayName(account.provider, activeProvider?.display_name)}
                          logoURL={activeProvider?.logo_url}
                          size={32}
                        />
                      }
                      status={
                        <>
                          <CopyAction
                            value={`${account.provider}:${account.name}`}
                            copyLabel={catalogT("actions.copyConnectionIdentifier")}
                            copyingLabel={catalogT("actions.copying")}
                            copiedLabel={catalogT("actions.copied")}
                          />
                          <StatusBadge
                            status={readiness === "ready"
                              ? "success"
                              : readiness === "connection_failed" || readiness === "needs_credentials"
                                ? "error"
                                : "warning"}
                          >
                            {readiness === "ready"
                              ? ux("readiness.ready")
                              : readiness === "connection_failed"
                                ? ux("readiness.connectionFailed")
                                : readiness === "needs_credentials"
                                  ? ux("readiness.needsCredentials")
                                  : readiness === "needs_access"
                                    ? ux("readiness.needsAccess")
                                    : ux("readiness.pricingIncomplete")}
                          </StatusBadge>
                        </>
                      }
                      actions={
                        <>
                          {!account.use_platform_managed_credentials ? (
                            <Button type="button" variant="outline" onClick={() => void handleTestCredentials()}>
                              {catalogT("actions.connectionTest")}
                            </Button>
                          ) : null}
                          <Button type="button" variant="outline" onClick={() => navigateWorkspaceHref("/provider-accounts")}>
                            {catalogT("actions.backToCatalog")}
                          </Button>
                        </>
                      }
                      className="rounded-xl border border-card-border bg-card p-4 shadow-xs sm:p-5"
                      titleClassName="text-lg sm:text-xl"
                    />

                    {visibleActiveTab === "overview" ? (
                      <div className="flex flex-col gap-4 sm:gap-6">
                        <div className="grid gap-3 sm:grid-cols-2 md:grid-cols-4">
                          <MetricCard
                            label={catalogT("details.modelCount")}
                            value={new Intl.NumberFormat(locale).format(selectedModelCount)}
                            icon={Layers3}
                            density="compact"
                          />
                          <MetricCard
                            label={catalogT("details.managerCount")}
                            value={new Intl.NumberFormat(locale).format(managers.length)}
                            icon={ShieldCheck}
                            density="compact"
                            tone="muted"
                          />
                          <MetricCard
                            label={catalogT("details.userCount")}
                            value={new Intl.NumberFormat(locale).format(users.length)}
                            icon={Users}
                            density="compact"
                            tone="muted"
                          />
                          <MetricCard
                            label={catalogT("details.unresolvedPricing")}
                            value={unresolvedPricingCount === 0
                              ? catalogT("details.none")
                              : new Intl.NumberFormat(locale).format(unresolvedPricingCount)}
                            icon={CircleDollarSign}
                            density="compact"
                            tone={unresolvedPricingCount > 0 ? "warning" : "success"}
                          />
                        </div>
                        <div className="grid min-w-0 gap-4 xl:grid-cols-3">
                          <div className="min-w-0 rounded-xl border border-card-border bg-card p-4 shadow-xs sm:p-5 xl:col-span-3">
                            <DefinitionList columns={2}>
                              <DefinitionListItem label={catalogT("details.authMode")} value={account.auth_mode} valueDir="ltr" mono />
                              <DefinitionListItem
                                label={catalogT("drawer.credentialSource")}
                                value={account.use_platform_managed_credentials ? catalogT("credential.managed") : catalogT("credential.byok")}
                              />
                            </DefinitionList>
                          </div>
                          {!selectedModels[0] ? (
                            <div className="xl:col-span-3">
                              <OperationNotice
                                notice={{ kind: "warning", message: catalogT("details.noModelRecovery") }}
                                labels={{
                                  success: catalogT("notices.success"),
                                  error: catalogT("notices.error"),
                                  warning: catalogT("notices.warning"),
                                  info: catalogT("notices.info"),
                                }}
                                action={
                                  <Button type="button" variant="outline" size="sm" onClick={() => setActiveTab("models")}>
                                    {catalogT("actions.manageModel")}
                                  </Button>
                                }
                              />
                            </div>
                          ) : null}
                        </div>
                      </div>
                    ) : null}

                    {visibleActiveTab === "models" ? (
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

                    {visibleActiveTab === "access" ? (
                      <div className="flex flex-col gap-6">
                        <AccessEditor
                          t={t}
                          managers={managers}
                          setManagers={handleManagersChange}
                          users={users}
                          setUsers={handleUsersChange}
                          onSearchPrincipal={searchPrincipalCandidates}
                          principalSelectorLabels={principalSelectorLabels}
                          onSave={handleSaveAccessAction}
                          isSaving={isSavingAccess}
                          canManage={canManageProviderAccount}
                          showSaveAction={false}
                        />
                        <FormSection
                          title={ux("accessTester.title")}
                          description={ux("accessTester.description", {
                            name: account.name,
                            provider: account.provider,
                          })}
                        >
                          <EffectiveAccessTester
                            account={account}
                            onSearchPrincipal={searchPrincipalCandidates}
                            principalSelectorLabels={principalSelectorLabels}
                          />
                        </FormSection>
                      </div>
                    ) : null}

                    {visibleActiveTab === "credentials" ? (
                      <section className="overflow-hidden rounded-xl border border-border bg-card text-card-foreground">
                        <div className="flex flex-col gap-4 p-4 sm:p-5">
                          {account.credential_rotation_required ? (
                            <Alert variant="warning">
                              <AlertTitle>{ux("credentials.rotationRequiredTitle")}</AlertTitle>
                              <AlertDescription>{ux("credentials.rotationRequiredDescription")}</AlertDescription>
                            </Alert>
                          ) : null}
                          {account.use_platform_managed_credentials ? (
                            <Alert>
                              <AlertTitle>{ux("credentials.managedCredentialLockedTitle")}</AlertTitle>
                              <AlertDescription>{ux("credentials.managedCredentialLockedDescription")}</AlertDescription>
                            </Alert>
                          ) : (
                            <>
                              <div className="rounded-lg border border-info/30 bg-info/5 px-4 py-3 text-sm leading-6 text-foreground">
                                <p>
                                  {ux.rich("credentials.privateCredentialLockedDescription", {
                                    byokLink: (chunks) => (
                                      <a
                                        href={`${docsBaseHref}/guides/byok`}
                                        target="_blank"
                                        rel="noreferrer"
                                        className="font-medium text-primary underline-offset-4 hover:underline"
                                      >
                                        {chunks}
                                      </a>
                                    ),
                                  })}
                                </p>
                              </div>

                              <div className="text-start">
                                <FieldGroup className="rounded-lg border border-border/70 bg-background p-4">
                                  {isLoadingConnectionSchemas ? (
                                    <p className="text-sm text-muted-foreground" role="status">
                                      {ux("integrations.wizard.loadingConnectionSchema")}
                                    </p>
                                  ) : null}
                                  {connectionSchemaError ? (
                                    <Alert variant="destructive">
                                      <AlertTitle>{ux("integrations.wizard.connectionSchemaErrorTitle")}</AlertTitle>
                                      <AlertDescription>{connectionSchemaError}</AlertDescription>
                                    </Alert>
                                  ) : null}
                                  {connectionSchemas.length > 1 ? (
                                    <AppSelectField
                                      id="credentialAuthMode"
                                      label={t("fields.authMode")}
                                      value={authMode}
                                      disabled={!canManageProviderAccount || !canEditCredentialSettings || isSavingCredentials || isTesting}
                                      optionTextDir="ltr"
                                      options={connectionSchemas.map((schema) => ({
                                        value: schema.auth_mode,
                                        label: schema.auth_mode,
                                      }))}
                                      onValueChange={(value) => {
                                        const nextSchema = connectionSchemas.find((schema) => schema.auth_mode === value) ?? null;
                                        setAuthMode(value as ProviderAuthMode);
                                        setProviderCredentialValues(providerCredentialValuesForSchema(nextSchema, account.provider_config));
                                        setAPIKeyTestSucceeded(false);
                                        setCreateCredentialTestResult(null);
                                      }}
                                    />
                                  ) : null}
                                  {selectedConnectionSchema ? (
                                    <ProviderCredentialSchemaFields
                                      schema={selectedConnectionSchema}
                                      values={providerCredentialValues}
                                      onValueChange={(key, value) => {
                                        setProviderCredentialValues((current) => ({ ...current, [key]: value }));
                                        setAPIKeyTestSucceeded(false);
                                        setCreateCredentialTestResult(null);
                                      }}
                                      disabled={!canManageProviderAccount || !canEditCredentialSettings || isSavingCredentials || isTesting}
                                      connectionFieldsTitle={ux("integrations.wizard.connectionFieldsTitle")}
                                      credentialFieldsTitle={ux("integrations.wizard.credentialFieldsTitle")}
                                      writeOnlyLabel={ux("integrations.wizard.writeOnlyCredential")}
                                      documentationLabel={ux("integrations.wizard.providerDocumentation")}
                                    />
                                  ) : null}
                                  <div className="flex flex-wrap items-center justify-between gap-3">
                                    <p className="text-sm leading-6 text-muted-foreground">
                                      {ux("integrations.wizard.testConnectionDescription")}
                                    </p>
                                    <Button
                                      type="button"
                                      variant="outline"
                                      onClick={() => void handleTestCredentials()}
                                      disabled={!canManageProviderAccount || !canSaveCredentialSettings || isSavingCredentials || isTesting}
                                    >
                                      {isTesting || isSavingCredentials
                                        ? ux("integrations.wizard.testingConnection")
                                        : ux("integrations.wizard.testConnection")}
                                    </Button>
                                  </div>
                                  {createCredentialTestResult ? (
                                    <Alert variant={createCredentialTestResult.kind === "error" ? "destructive" : createCredentialTestResult.kind === "warning" ? "warning" : "default"}>
                                      <AlertTitle>
                                        {createCredentialTestResult.kind === "warning"
                                          ? ux("connectionTest.warningTitle")
                                          : createCredentialTestResult.kind === "error"
                                          ? ux("integrations.wizard.connectionTestFailedTitle")
                                          : ux("integrations.wizard.connectionTestSuccessTitle")}
                                      </AlertTitle>
                                      <AlertDescription>{createCredentialTestResult.message}</AlertDescription>
                                    </Alert>
                                  ) : null}
                                  <WIFCredentialStatus result={createCredentialTestResult} />
                                </FieldGroup>
                              </div>

                              <div className="flex flex-wrap justify-end gap-2 pt-1">
                                <Button
                                  type="button"
                                  onClick={() => void handleSaveCredentials()}
                                  disabled={!canManageProviderAccount || !canSaveCredentialSettings || isSavingCredentials || isTesting}
                                >
                                  {isSavingCredentials ? t("saving") : t("save")}
                                </Button>
                              </div>
                            </>
                          )}
                        </div>
                      </section>
                    ) : null}
                  </>
                ) : null}
          </div>
        </section>
        <TypedConfirmDialog
          open={deleteDialogOpen}
          onOpenChange={(open) => {
            setDeleteDialogOpen(open);
            if (!open) setDeleteTargetAccount(null);
          }}
          title={t("messages.accountDeleteConfirmTitle")}
          description={t("messages.accountDeleteConfirmBody", { name: deleteTargetAccount?.name ?? "" })}
          expectedValue={deleteTargetAccount?.name ?? ""}
          inputLabel={t("messages.accountDeleteConfirmLabel")}
          cancelLabel={ux("actions.cancelDelete")}
          confirmLabel={ux("actions.confirmDelete")}
          pendingLabel={ux("actions.loading")}
          onConfirm={() => deleteTargetAccount ? handleDeleteAccount(deleteTargetAccount) : Promise.resolve()}
          disabled={!deleteTargetAccount}
          inputDir="auto"
        />
        <UnsavedChangesDialog
          open={unsavedDialogOpen}
          onOpenChange={(open) => {
            setUnsavedDialogOpen(open);
            if (!open) setPendingWorkspaceNavigation(null);
          }}
          title={catalogT("details.unsavedTitle")}
          description={catalogT("details.unsavedDescription")}
          stayLabel={catalogT("details.stay")}
          discardLabel={catalogT("details.discard")}
          onDiscard={discardWorkspaceChanges}
        />
      </div>
    </div>
  );
}

function CatalogCreateSection({
  title,
  description,
  children,
  footer,
  actions,
  id,
  setSectionNode,
  fillHeight = false,
  scrollable = false,
  contentClassName,
}: {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  actions?: ReactNode;
  id?: string;
  setSectionNode?: (node: HTMLElement | null) => void;
  fillHeight?: boolean;
  scrollable?: boolean;
  contentClassName?: string;
}) {
  return (
    <section
      id={id}
      ref={setSectionNode}
      className={cn("min-w-0", fillHeight && "flex min-h-0 flex-1 flex-col")}
    >
      <FormSection
        footer={footer}
        className={cn(fillHeight && "min-h-0 flex-1", scrollable && "overflow-y-auto")}
      >
        {title || description || actions ? (
          <SectionHeader
            title={title}
            description={description}
            action={actions}
          />
        ) : null}
        <div className={cn("min-w-0", contentClassName)}>{children}</div>
      </FormSection>
    </section>
  );
}

function ProviderDetailsSkeleton({ label }: { label: string }) {
  return (
    <div className="flex flex-col gap-5" role="status" aria-label={label}>
      <div className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <Skeleton className="size-10 shrink-0 rounded-lg" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-5 w-48 max-w-full" />
            <Skeleton className="h-4 w-64 max-w-full" />
          </div>
        </div>
        <Skeleton className="h-9 w-32" />
      </div>
      <Skeleton className="h-12 w-full rounded-lg" />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={`provider-detail-metric-${index}`} className="h-24 rounded-xl" />
        ))}
      </div>
      <Skeleton className="h-64 w-full rounded-xl" />
    </div>
  );
}

function credentialFingerprint(input: {
  authMode: ProviderAuthMode;
  usePlatformManagedCredentials: boolean;
  values: ProviderCredentialValues;
}) {
  return JSON.stringify({
    authMode: input.authMode,
    usePlatformManagedCredentials: input.usePlatformManagedCredentials,
    values: Object.fromEntries(
      Object.entries(input.values).sort(([left], [right]) => left.localeCompare(right)),
    ),
  });
}
