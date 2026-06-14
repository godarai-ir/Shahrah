"use client";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { Highlight, themes } from "prism-react-renderer";
import { Prism } from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-powershell";
import { ArrowUpDown, CheckSquare, Code2, Copy, Eye, GitCompare, KeyRound, Layers3, RotateCcw, Search, ShieldCheck, Trash2, X } from "lucide-react";

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
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import { PrincipalPicker } from "@/app/components/principal-picker";
import { ModelCard, type ModelCardData } from "@/app/components/model-card";
import {
  buildDefaultDrafts,
  buildDraftsForAccount,
  buildSelectedModelPayload,
  clearVisibleModelSelection,
  toggleVisibleModelSelection,
} from "@/app/components/provider-account-wizard.models";
import { usePermissions } from "@/app/components/permissions-context";
import { canPermission } from "@/lib/authz";
import { cn } from "@/lib/utils";

type ProviderAuthMode =
  | "api_key"
  | "azure_entra_cert"
  | "aws_sigv4"
  | "google_service_account"
  | "no_credential";

type ModelType = string;

type PrincipalType = "user" | "team" | "org_unit" | "everyone";
type SetupStep = 1 | 2 | 3 | 4 | 5;
type WorkspaceTab = "models" | "access" | "credentials" | "observability";
type WorkspaceMode = "overview" | "setup" | "workspace";

type AccessPrincipal = {
  type: PrincipalType;
  principal_id?: string;
  display_name?: string;
};

type AccessPrincipalCandidate = {
  type: PrincipalSearchKind;
  principal_id: string;
  display_name: string;
  secondary?: string;
};

type ProviderCatalogProvider = {
  provider: string;
  display_name: string;
  logo_url?: string;
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
  context_window?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  features?: string[];
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
  provider_config?: Record<string, unknown>;
  has_credential: boolean;
  allows_virtual_keys_only: boolean;
};

type TenantProviderAccountModel = {
  model: string;
  enabled: boolean;
  types: ModelType[];
  input_modalities?: string[];
  output_modalities?: string[];
  pricing_mode: "public" | "private";
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
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
type DisplaySettingsResponse = {
  data?: {
    currencyUnit?: "USD" | "IRR" | "TOMAN";
    priceScale?: "PER_MILLION_TOKENS" | "PER_TOKEN";
    exchangeRate?: {
      usdToIrr?: number;
    };
  };
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type CurlGuideShell = "powershell" | "bash";

type CurlGuide = {
  modelName: string;
  modelFqn: string;
  commands: Record<CurlGuideShell, string>;
};

type ModelDraft = {
  enabled: boolean;
  compared: boolean;
  selectedCapabilities: string[];
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
};

type Readiness = "ready" | "needs_credentials" | "needs_access" | "pricing_incomplete";

type ProviderAccountWizardMode = "manage" | "create";

const CURL_GUIDE_LANGUAGE: Record<CurlGuideShell, "powershell" | "bash"> = {
  powershell: "powershell",
  bash: "bash",
};

function CodeExampleBlock(props: {
  code: string;
  language: "powershell" | "bash";
  shellLabel: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-2xl border border-slate-800 bg-slate-950 shadow-[0_24px_80px_-32px_rgba(15,23,42,0.85)]",
        props.className,
      )}
    >
      <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900/80 px-4 py-3 text-slate-200">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-rose-400" />
            <span className="h-2.5 w-2.5 rounded-full bg-amber-300" />
            <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" />
          </div>
          <Badge variant="secondary" className="border border-slate-700 bg-slate-800 px-2.5 py-0.5 font-mono text-[11px] text-slate-200">
            {props.shellLabel}
          </Badge>
        </div>
      </div>
      <Highlight
        code={props.code}
        language={props.language}
        prism={Prism}
        theme={themes.nightOwl}
      >
        {({ className, style, tokens, getLineProps, getTokenProps }) => (
          <pre
            dir="ltr"
            className={cn("custom-scrollbar overflow-x-auto p-4 text-left text-[13px] leading-7", className)}
            style={{
              ...style,
              margin: 0,
              background: "transparent",
            }}
          >
            {tokens.map((line, lineIndex) => {
              const lineProps = getLineProps({ line });
              return (
                <div
                  key={`${props.language}-${lineIndex}`}
                  {...lineProps}
                  className={cn("table w-full table-fixed", lineProps.className)}
                >
                  <span className="table-cell w-10 select-none pe-4 text-right text-slate-500">
                    {lineIndex + 1}
                  </span>
                  <span className="table-cell w-full">
                    {line.length === 0
                      ? "\n"
                      : line.map((token, tokenIndex) => {
                          const tokenProps = getTokenProps({ token });
                          return <span key={`${lineIndex}-${tokenIndex}`} {...tokenProps} />;
                        })}
                  </span>
                </div>
              );
            })}
          </pre>
        )}
      </Highlight>
    </div>
  );
}

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

function EmptyOverview(props: {
  title: string;
  description: string;
  setupLabel: string;
  setupHref: string;
  openExistingLabel: string;
  onOpenFirst: () => void;
  hasAccount: boolean;
  createDisabled?: boolean;
}) {
  return (
    <section className="space-y-3 rounded-lg border border-dashed p-6 text-center">
      <h3 className="text-lg font-semibold">{props.title}</h3>
      <p className="text-sm text-muted-foreground">{props.description}</p>
      <div className="flex flex-wrap items-center justify-center gap-2">
        <Button type="button" asChild disabled={props.createDisabled}>
          <Link href={props.setupHref}>{props.setupLabel}</Link>
        </Button>
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

function deriveAccountDetailsReadiness(
  account: TenantProviderAccount,
  models: TenantProviderAccountModel[],
  access: AccountDetailsResponse["access"] | undefined,
  catalogModels: ProviderCatalogModel[],
): Readiness {
  if (!account.has_credential && !account.allows_virtual_keys_only && account.auth_mode !== "no_credential") {
    return "needs_credentials";
  }
  if ((access?.managers ?? []).length === 0) {
    return "needs_access";
  }

  const catalogByModel = new Map(catalogModels.map((model) => [model.model, model]));
  const hasUnresolvedPricing = models.some((model) => {
    if (!model.enabled) return false;
    if (model.pricing_mode === "private") {
      return (
        typeof model.private_pricing?.input_usd_per_1m !== "number" &&
        typeof model.private_pricing?.output_usd_per_1m !== "number"
      );
    }
    const catalog = catalogByModel.get(model.model);
    return (
      typeof catalog?.public_pricing?.input_usd_per_1m !== "number" &&
      typeof catalog?.public_pricing?.output_usd_per_1m !== "number"
    );
  });
  return hasUnresolvedPricing ? "pricing_incomplete" : "ready";
}

function mergeAccountIntoList(
  accounts: TenantProviderAccount[],
  nextAccount: TenantProviderAccount,
): TenantProviderAccount[] {
  let found = false;
  const next = accounts.map((item) => {
    if (item.id !== nextAccount.id) return item;
    found = true;
    return nextAccount;
  });
  return found ? next : [nextAccount, ...next];
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
  account: TenantProviderAccount | null;
  setBanner: React.Dispatch<React.SetStateAction<Banner | null>>;
  filteredProviderModels: ProviderCatalogModel[];
  modelDrafts: Record<string, ModelDraft>;
  setModelDrafts: React.Dispatch<React.SetStateAction<Record<string, ModelDraft>>>;
  modelSearchQuery: string;
  setModelSearchQuery: (value: string) => void;
  isSaving: boolean;
  canManage: boolean;
  onSave: () => void;
}) {
  const ux = useTranslations("providerAccounts.ux");
  type ModelFilterState = {
    search: string;
    minPrice: string;
    maxPrice: string;
    minContext: string;
    inputModality: string;
    outputModality: string;
    featureFilter: string;
    sortBy: "price_asc" | "price_desc" | "context_asc" | "context_desc";
  };

  const initialFilters: ModelFilterState = {
    search: props.modelSearchQuery,
    minPrice: "",
    maxPrice: "",
    minContext: "",
    inputModality: "",
    outputModality: "",
    featureFilter: "",
    sortBy: "price_asc",
  };

  const [filters, setFilters] = useState<ModelFilterState>(initialFilters);
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareSortBy, setCompareSortBy] = useState<"input_asc" | "output_asc" | "context_desc" | "name_asc">("input_asc");
  const [compareDiffOnly, setCompareDiffOnly] = useState(false);
  const [compareNormalizedPricing, setCompareNormalizedPricing] = useState(true);
  const [displayCurrency, setDisplayCurrency] = useState<"USD" | "TOMAN">("USD");
  const [displayScale, setDisplayScale] = useState<"PER_MILLION_TOKENS" | "PER_TOKEN">("PER_MILLION_TOKENS");
  const [usdToIrr, setUsdToIrr] = useState(0);
  const [curlGuide, setCurlGuide] = useState<CurlGuide | null>(null);
  const [curlGuideShell, setCurlGuideShell] = useState<CurlGuideShell>("powershell");
  const [pricingEditor, setPricingEditor] = useState<{
    open: boolean;
    modelKey: string;
    pricingMode: "public" | "private";
    inputPrice: string;
    outputPrice: string;
    error: string;
  }>({
    open: false,
    modelKey: "",
    pricingMode: "public",
    inputPrice: "",
    outputPrice: "",
    error: "",
  });
  const tenantBaseUrl = useMemo(() => {
    if (typeof window === "undefined") {
      return "http://tenant1.localhost:4000";
    }
    return `${window.location.protocol}//${window.location.hostname}:4000`;
  }, []);

  const normalizeProviderScopedModelName = useCallback((provider: string, modelName: string) => {
    const normalizedProvider = provider.trim().toLowerCase();
    const normalizedModelName = modelName.trim();
    if (!normalizedProvider || !normalizedModelName) return normalizedModelName;
    const prefix = `${normalizedProvider}/`;
    if (normalizedModelName.toLowerCase().startsWith(prefix)) {
      return normalizedModelName.slice(prefix.length);
    }
    return normalizedModelName;
  }, []);

  const buildModelFqn = useCallback((modelValue: ProviderCatalogModel) => {
    if (!props.account) return modelValue.model;
    const officialModel = normalizeProviderScopedModelName(props.account.provider, modelValue.model);
    return `${props.account.provider}:${props.account.name}:${officialModel}`;
  }, [normalizeProviderScopedModelName, props.account]);

  const openCurlGuide = useCallback((modelValue: ProviderCatalogModel) => {
    const modelFqn = buildModelFqn(modelValue);
    const normalizedProvider = (modelValue.provider || props.account?.provider || "").trim().toLowerCase();
    const useResponsesApi = normalizedProvider === "openai";
    const endpoint = useResponsesApi ? "/v1/responses" : "/v1/chat/completions";
    const bashCommand = useResponsesApi
      ? [
          `curl --location '${tenantBaseUrl}${endpoint}' \\`,
          `  --header 'Authorization: Bearer <PAT_TOKEN>' \\`,
          `  --header 'Content-Type: application/json' \\`,
          "  --data '{",
          `    \"model\": \"${modelFqn}\",`,
          '    "input": "سلام، یک پاسخ کوتاه بده"',
          "  }'",
        ].join("\n")
      : [
          `curl --location '${tenantBaseUrl}${endpoint}' \\`,
          `  --header 'Authorization: Bearer <PAT_TOKEN>' \\`,
          `  --header 'Content-Type: application/json' \\`,
          "  --data '{",
          `    \"model\": \"${modelFqn}\",`,
          '    "messages": [',
          '      {"role":"user","content":"سلام، یک پاسخ کوتاه بده"}',
          "    ]",
          "  }'",
        ].join("\n");
    const powershellCommand = useResponsesApi
      ? [
          "$headers = @{",
          '  Authorization = "Bearer <PAT_TOKEN>"',
          '  "Content-Type" = "application/json"',
          "}",
          "",
          "$body = @{",
          `  model = "${modelFqn}"`,
          '  input = "سلام، یک پاسخ کوتاه بده"',
          "} | ConvertTo-Json -Depth 5",
          "",
          "Invoke-RestMethod -Method Post `",
          `  -Uri "${tenantBaseUrl}${endpoint}" \``,
          "  -Headers $headers `",
          "  -Body $body",
        ].join("\n")
      : [
          "$headers = @{",
          '  Authorization = "Bearer <PAT_TOKEN>"',
          '  "Content-Type" = "application/json"',
          "}",
          "",
          "$body = @{",
          `  model = "${modelFqn}"`,
          "  messages = @(",
          "    @{",
          '      role = "user"',
          '      content = "سلام، یک پاسخ کوتاه بده"',
          "    }",
          "  )",
          "} | ConvertTo-Json -Depth 5",
          "",
          "Invoke-RestMethod -Method Post `",
          `  -Uri "${tenantBaseUrl}${endpoint}" \``,
          "  -Headers $headers `",
          "  -Body $body",
        ].join("\n");
    setCurlGuideShell("powershell");
    setCurlGuide({
      modelName: modelValue.display_name || modelValue.model,
      modelFqn,
      commands: {
        powershell: powershellCommand,
        bash: bashCommand,
      },
    });
  }, [buildModelFqn, props.account?.provider, tenantBaseUrl]);

  const comparedModels = useMemo(
    () => props.filteredProviderModels.filter((model) => props.modelDrafts[model.model]?.compared),
    [props.filteredProviderModels, props.modelDrafts],
  );
  const enabledFilteredModelsCount = useMemo(
    () => props.filteredProviderModels.filter((model) => props.modelDrafts[model.model]?.enabled).length,
    [props.filteredProviderModels, props.modelDrafts],
  );
  const availableFilteredModelsCount = Math.max(props.filteredProviderModels.length - enabledFilteredModelsCount, 0);

  const allInputModalities = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => model.input_modalities ?? []))),
    [props.filteredProviderModels],
  );
  const allOutputModalities = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => model.output_modalities ?? []))),
    [props.filteredProviderModels],
  );
  const allFeatures = useMemo(
    () => Array.from(new Set(props.filteredProviderModels.flatMap((model) => model.features ?? model.types ?? []))),
    [props.filteredProviderModels],
  );

  const visibleModels = useMemo(() => {
    const minPriceValue = toNumberOrUndefined(filters.minPrice);
    const maxPriceValue = toNumberOrUndefined(filters.maxPrice);
    const minContextValue = toNumberOrUndefined(filters.minContext);
    const filtered = props.filteredProviderModels.filter((model) => {
      const inputPrice = model.public_pricing?.input_usd_per_1m;
      const context = model.context_window;
      if (typeof minPriceValue === "number" && (typeof inputPrice !== "number" || inputPrice < minPriceValue)) return false;
      if (typeof maxPriceValue === "number" && (typeof inputPrice !== "number" || inputPrice > maxPriceValue)) return false;
      if (typeof minContextValue === "number" && (typeof context !== "number" || context < minContextValue)) return false;
      if (filters.inputModality && !(model.input_modalities ?? []).includes(filters.inputModality)) return false;
      if (filters.outputModality && !(model.output_modalities ?? []).includes(filters.outputModality)) return false;
      if (filters.featureFilter && !(model.features ?? model.types ?? []).includes(filters.featureFilter)) return false;
      return true;
    });

    const sorted = [...filtered];
    sorted.sort((a, b) => {
      const aEnabled = props.modelDrafts[a.model]?.enabled ? 1 : 0;
      const bEnabled = props.modelDrafts[b.model]?.enabled ? 1 : 0;
      if (aEnabled !== bEnabled) return bEnabled - aEnabled;
      const aPrice = a.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const bPrice = b.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const aContext = a.context_window ?? 0;
      const bContext = b.context_window ?? 0;
      switch (filters.sortBy) {
        case "price_desc":
          return bPrice - aPrice;
        case "context_asc":
          return aContext - bContext;
        case "context_desc":
          return bContext - aContext;
        case "price_asc":
        default:
          return aPrice - bPrice;
      }
    });
    return sorted;
  }, [filters.featureFilter, filters.inputModality, filters.maxPrice, filters.minContext, filters.minPrice, filters.outputModality, filters.sortBy, props.filteredProviderModels, props.modelDrafts]);

  const resetFilters = () => {
    setFilters((prev) => ({
      ...initialFilters,
      search: prev.search,
    }));
  };

  const activeFilters = useMemo(() => {
    const chips: Array<{ key: keyof ModelFilterState; label: string }> = [];
    if (filters.minPrice.trim()) chips.push({ key: "minPrice", label: `${ux("models.filters.minPrice")}: ${filters.minPrice}` });
    if (filters.maxPrice.trim()) chips.push({ key: "maxPrice", label: `${ux("models.filters.maxPrice")}: ${filters.maxPrice}` });
    if (filters.minContext.trim()) chips.push({ key: "minContext", label: `${ux("models.filters.minContext")}: ${filters.minContext}` });
    if (filters.inputModality.trim()) chips.push({ key: "inputModality", label: `${ux("models.filters.allInputModalities")}: ${filters.inputModality}` });
    if (filters.outputModality.trim()) chips.push({ key: "outputModality", label: `${ux("models.filters.allOutputModalities")}: ${filters.outputModality}` });
    if (filters.featureFilter.trim()) chips.push({ key: "featureFilter", label: `${ux("models.filters.allFeatures")}: ${filters.featureFilter}` });
    return chips;
  }, [filters.featureFilter, filters.inputModality, filters.maxPrice, filters.minContext, filters.minPrice, filters.outputModality, ux]);

  const hasInvalidPriceRange = useMemo(() => {
    const min = toNumberOrUndefined(filters.minPrice);
    const max = toNumberOrUndefined(filters.maxPrice);
    if (typeof min !== "number" || typeof max !== "number") return false;
    return min > max;
  }, [filters.maxPrice, filters.minPrice]);

  const compareModels = useMemo(() => {
    const items = [...comparedModels];
    items.sort((a, b) => {
      const aInput = a.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const bInput = b.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const aOutput = a.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const bOutput = b.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY;
      const aContext = a.context_window ?? 0;
      const bContext = b.context_window ?? 0;
      const aName = (a.display_name || a.model).toLowerCase();
      const bName = (b.display_name || b.model).toLowerCase();
      switch (compareSortBy) {
        case "output_asc":
          return aOutput - bOutput;
        case "context_desc":
          return bContext - aContext;
        case "name_asc":
          return aName.localeCompare(bName);
        case "input_asc":
        default:
          return aInput - bInput;
      }
    });
    return items;
  }, [compareSortBy, comparedModels]);

  const compareKpis = useMemo(() => {
    const validInput = compareModels.filter((item) => typeof item.public_pricing?.input_usd_per_1m === "number");
    const validOutput = compareModels.filter((item) => typeof item.public_pricing?.output_usd_per_1m === "number");
    const validContext = compareModels.filter((item) => typeof item.context_window === "number");
    const cheapestInput = validInput.sort((a, b) => (a.public_pricing?.input_usd_per_1m ?? 0) - (b.public_pricing?.input_usd_per_1m ?? 0))[0];
    const cheapestOutput = validOutput.sort((a, b) => (a.public_pricing?.output_usd_per_1m ?? 0) - (b.public_pricing?.output_usd_per_1m ?? 0))[0];
    const largestContext = validContext.sort((a, b) => (b.context_window ?? 0) - (a.context_window ?? 0))[0];
    const maxFeatures = compareModels.reduce((best, item) => {
      const count = (item.features ?? item.types ?? []).length;
      return count > best.count ? { model: item, count } : best;
    }, { model: null as ProviderCatalogModel | null, count: -1 });
    return { cheapestInput, cheapestOutput, largestContext, mostFeatures: maxFeatures.model };
  }, [compareModels]);

  const bestMarkers = useMemo(() => {
    const minInput = Math.min(...compareModels.map((item) => item.public_pricing?.input_usd_per_1m ?? Number.POSITIVE_INFINITY));
    const minOutput = Math.min(...compareModels.map((item) => item.public_pricing?.output_usd_per_1m ?? Number.POSITIVE_INFINITY));
    const maxContext = Math.max(...compareModels.map((item) => item.context_window ?? 0));
    return {
      input: Number.isFinite(minInput) ? minInput : null,
      output: Number.isFinite(minOutput) ? minOutput : null,
      context: maxContext > 0 ? maxContext : null,
    };
  }, [compareModels]);

  const compareRows = useMemo(() => {
    const rows = [
      {
        section: "pricing",
        key: "input",
        label: ux("models.compare.rows.inputPrice"),
        values: compareModels.map((item) => item.public_pricing?.input_usd_per_1m),
        render: (value: unknown) => formatDisplayPrice(typeof value === "number" ? value : undefined, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") }),
      },
      {
        section: "pricing",
        key: "output",
        label: ux("models.compare.rows.outputPrice"),
        values: compareModels.map((item) => item.public_pricing?.output_usd_per_1m),
        render: (value: unknown) => formatDisplayPrice(typeof value === "number" ? value : undefined, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") }),
      },
      {
        section: "capacity",
        key: "context",
        label: ux("models.compare.rows.context"),
        values: compareModels.map((item) => item.context_window),
        render: (value: unknown) => (typeof value === "number" ? value.toLocaleString() : ux("models.na")),
      },
      {
        section: "modalities",
        key: "inputModality",
        label: ux("models.compare.rows.inputModality"),
        values: compareModels.map((item) => (item.input_modalities ?? []).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
      {
        section: "modalities",
        key: "outputModality",
        label: ux("models.compare.rows.outputModality"),
        values: compareModels.map((item) => (item.output_modalities ?? []).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
      {
        section: "capabilities",
        key: "features",
        label: ux("models.compare.rows.features"),
        values: compareModels.map((item) => (item.features ?? item.types ?? []).slice().sort().join(", ")),
        render: (value: unknown) => (typeof value === "string" && value ? value : ux("models.na")),
      },
    ] as const;

    return rows.map((row) => {
      const uniqueCount = new Set(row.values.map((value) => `${value ?? ""}`)).size;
      return { ...row, isDifferent: uniqueCount > 1 };
    });
  }, [compareModels, displayCurrency, displayScale, usdToIrr, ux]);

  const visibleCompareRows = useMemo(
    () => (compareDiffOnly ? compareRows.filter((row) => row.isDifferent) : compareRows),
    [compareDiffOnly, compareRows],
  );

  const removeComparedModel = (modelKey: string) => {
    props.setModelDrafts((prev) => {
      const draft = prev[modelKey];
      if (!draft) return prev;
      return { ...prev, [modelKey]: { ...draft, compared: false } };
    });
  };

  const renderCompareValue = (rowKey: string, value: unknown, render: (value: unknown) => string): string => {
    if (!compareNormalizedPricing && (rowKey === "input" || rowKey === "output")) {
      if (typeof value !== "number") return render(value);
      return render(value / 1_000_000);
    }
    return render(value);
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      props.setModelSearchQuery(filters.search);
    }, 250);
    return () => clearTimeout(timer);
  }, [filters.search, props]);

  const openPricingEditor = (model: ProviderCatalogModel) => {
    const draft = props.modelDrafts[model.model] ?? {
      enabled: false,
      compared: false,
      selectedCapabilities: [],
      pricingMode: "public" as const,
      inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
      outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
    };
    setPricingEditor({
      open: true,
      modelKey: model.model,
      pricingMode: draft.pricingMode,
      inputPrice: draft.inputPrice,
      outputPrice: draft.outputPrice,
      error: "",
    });
  };

  const applyPricingEditor = () => {
    const inputNumber = toNumberOrUndefined(pricingEditor.inputPrice);
    const outputNumber = toNumberOrUndefined(pricingEditor.outputPrice);
    if (pricingEditor.pricingMode === "private") {
      if (pricingEditor.inputPrice.trim() && (typeof inputNumber !== "number" || inputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidInputPrice") }));
        return;
      }
      if (pricingEditor.outputPrice.trim() && (typeof outputNumber !== "number" || outputNumber < 0)) {
        setPricingEditor((prev) => ({ ...prev, error: ux("models.validation.invalidOutputPrice") }));
        return;
      }
    }

    props.setModelDrafts((prev) => {
      const current = prev[pricingEditor.modelKey];
      if (!current) return prev;
      return {
        ...prev,
        [pricingEditor.modelKey]: {
          ...current,
          enabled: true,
          pricingMode: pricingEditor.pricingMode,
          inputPrice: pricingEditor.inputPrice.trim(),
          outputPrice: pricingEditor.outputPrice.trim(),
        },
      };
    });
    setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
  };

  const toModelCardData = (model: ProviderCatalogModel): ModelCardData => {
    const caps = model.features ?? model.types ?? [];
    const categories: ModelCardData["category"] = [];
    if (caps.some((c) => c.toLowerCase().includes("vision") || c.toLowerCase().includes("image"))) categories.push("vision");
    if (caps.some((c) => c.toLowerCase().includes("reason"))) categories.push("reasoning");
    if (caps.some((c) => c.toLowerCase().includes("code"))) categories.push("code");
    if (caps.some((c) => c.toLowerCase().includes("audio"))) categories.push("audio");
    if (categories.length === 0) categories.push("chat");
    const modelFqn = buildModelFqn(model);

    return {
      id: model.model,
      name: model.display_name || model.model,
      provider: {
        id: model.provider,
        name: model.provider,
      },
      description: `${ux("models.columns.modalities")}: ${(model.input_modalities ?? []).join(", ") || ux("models.na")} / ${(model.output_modalities ?? []).join(", ") || ux("models.na")}`,
      category: categories,
      status: "stable",
      contextWindow: model.context_window,
      maxOutputTokens: undefined,
      pricing: {
        inputPerMillion: model.public_pricing?.input_usd_per_1m,
        outputPerMillion: model.public_pricing?.output_usd_per_1m,
      },
      latencyTier: "medium",
      qualityTier: "good",
      capabilities: caps,
      tradeoffs: [],
      metadata: {
        modelId: modelFqn,
        mode: "chat",
      },
    };
  };

  const buildModelDiscoveryDetailsHref = (model: ProviderCatalogModel): string => {
    const base =
      process.env.NEXT_PUBLIC_MODEL_DISCOVERY_APP_URL?.replace(/\/$/, "") ??
      "/catalog";
    const provider = encodeURIComponent(model.provider);
    const canonical = model.model.startsWith(`${model.provider}/`)
      ? model.model.slice(model.provider.length + 1)
      : model.model;
    const path = canonical
      .split("/")
      .filter(Boolean)
      .map((part) => encodeURIComponent(part))
      .join("/");
    return `${base}/models/${provider}/${path || encodeURIComponent(model.model)}`;
  };

  useEffect(() => {
    let canceled = false;
    void (async () => {
      try {
        const payload = await requestJSON<DisplaySettingsResponse>("/api/provider-catalog/display-settings");
        if (canceled) return;
        const data = unwrapData<DisplaySettingsResponse["data"]>(payload) ?? payload.data;
        if (!data) return;
        if (data.currencyUnit === "TOMAN") setDisplayCurrency("TOMAN");
        else setDisplayCurrency("USD");
        if (data.priceScale === "PER_TOKEN" || data.priceScale === "PER_MILLION_TOKENS") {
          setDisplayScale(data.priceScale);
        }
        setUsdToIrr(Number(data.exchangeRate?.usdToIrr ?? 0));
      } catch {
        // Keep defaults when display settings are unavailable.
      }
    })();
    return () => {
      canceled = true;
    };
  }, []);

  return (
    <section className="space-y-3 rounded-lg border p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">{ux("tabs.models")}</h3>
          <p className="text-sm text-muted-foreground">{ux("models.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={!props.canManage || props.isSaving}>
          {props.isSaving ? props.t("saving") : props.t("saveSelectedModels")}
        </Button>
      </div>

      <div className="rounded-lg border border-border/70 bg-gradient-to-l from-background to-emerald-500/[0.05] p-3">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline">{ux("models.summary.catalog", { count: props.filteredProviderModels.length })}</Badge>
          <Badge className="border-emerald-500/30 bg-emerald-500/15 text-emerald-700 hover:bg-emerald-500/20 dark:text-emerald-300" variant="outline">
            {ux("models.summary.enabled", { count: enabledFilteredModelsCount })}
          </Badge>
          <Badge variant="secondary">{ux("models.summary.available", { count: availableFilteredModelsCount })}</Badge>
          <Badge variant="secondary">{ux("models.summary.showing", { visible: visibleModels.length, total: props.filteredProviderModels.length })}</Badge>
        </div>
        <p className="mt-2 text-sm text-muted-foreground">{ux("models.catalogScopeHint")}</p>
      </div>

      <div className="space-y-3 rounded-lg border border-border/70 bg-muted/20 p-3">
        <div className="relative">
          <Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={filters.search}
            onChange={(event) => setFilters((prev) => ({ ...prev, search: event.target.value }))}
            placeholder={props.t("step3.searchPlaceholder")}
            className="ps-9 pe-9"
          />
          {filters.search.trim() ? (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="absolute end-1 top-1/2 h-8 w-8 -translate-y-1/2"
              onClick={() => setFilters((prev) => ({ ...prev, search: "" }))}
              aria-label={ux("actions.clear")}
              title={ux("actions.clear")}
            >
              <X className="h-4 w-4" />
            </Button>
          ) : null}
        </div>

        <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
          <Input
            value={filters.minPrice}
            onChange={(event) => setFilters((prev) => ({ ...prev, minPrice: event.target.value }))}
            placeholder={ux("models.filters.minPrice")}
            inputMode="decimal"
          />
          <Input
            value={filters.maxPrice}
            onChange={(event) => setFilters((prev) => ({ ...prev, maxPrice: event.target.value }))}
            placeholder={ux("models.filters.maxPrice")}
            inputMode="decimal"
          />
          <Input
            value={filters.minContext}
            onChange={(event) => setFilters((prev) => ({ ...prev, minContext: event.target.value }))}
            placeholder={ux("models.filters.minContext")}
            inputMode="numeric"
          />
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.inputModality} onChange={(event) => setFilters((prev) => ({ ...prev, inputModality: event.target.value }))}>
          <option value="">{ux("models.filters.allInputModalities")}</option>
          {allInputModalities.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.outputModality} onChange={(event) => setFilters((prev) => ({ ...prev, outputModality: event.target.value }))}>
          <option value="">{ux("models.filters.allOutputModalities")}</option>
          {allOutputModalities.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.featureFilter} onChange={(event) => setFilters((prev) => ({ ...prev, featureFilter: event.target.value }))}>
          <option value="">{ux("models.filters.allFeatures")}</option>
          {allFeatures.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={filters.sortBy} onChange={(event) => setFilters((prev) => ({ ...prev, sortBy: event.target.value as ModelFilterState["sortBy"] }))}>
          <option value="price_asc">{ux("models.sort.priceAsc")}</option>
          <option value="price_desc">{ux("models.sort.priceDesc")}</option>
          <option value="context_asc">{ux("models.sort.contextAsc")}</option>
          <option value="context_desc">{ux("models.sort.contextDesc")}</option>
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={displayCurrency} onChange={(event) => setDisplayCurrency(event.target.value as "USD" | "TOMAN")}>
          <option value="USD">{ux("models.displayUnit.currencyUsd")}</option>
          <option value="TOMAN">{ux("models.displayUnit.currencyToman")}</option>
          </select>
          <select className="h-10 rounded-md border border-input bg-background px-3 text-sm" value={displayScale} onChange={(event) => setDisplayScale(event.target.value as "PER_MILLION_TOKENS" | "PER_TOKEN")}>
          <option value="PER_MILLION_TOKENS">{ux("models.displayUnit.perMillionTokens")}</option>
          <option value="PER_TOKEN">{ux("models.displayUnit.perToken")}</option>
          </select>
        </div>

        {hasInvalidPriceRange ? (
          <p className="text-xs text-destructive">{ux("models.validation.invalidInputPrice")}</p>
        ) : null}

        {activeFilters.length > 0 ? (
          <div className="flex flex-wrap items-center gap-2">
            {activeFilters.map((item) => (
              <Badge key={item.key} variant="secondary" className="flex items-center gap-1 pe-1">
                <span className="max-w-56 truncate">{item.label}</span>
                <button
                  type="button"
                  className="rounded p-0.5 hover:bg-muted"
                  onClick={() => setFilters((prev) => ({ ...prev, [item.key]: "" }))}
                  aria-label={ux("actions.clear")}
                  title={ux("actions.clear")}
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : null}
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            props.setModelDrafts((prev) => toggleVisibleModelSelection(visibleModels, prev));
          }}
          disabled={!props.canManage || visibleModels.length === 0}
        >
          <CheckSquare className="h-4 w-4" />
          {props.t("step3.toggleSelectAll")}
        </Button>
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            props.setModelDrafts((prev) => clearVisibleModelSelection(visibleModels, prev));
          }}
          disabled={!props.canManage || visibleModels.length === 0}
        >
          <X className="h-4 w-4" />
          {ux("actions.clear")}
        </Button>
        <Button type="button" variant="outline" onClick={resetFilters}>
          <RotateCcw className="h-4 w-4" />
          {ux("models.actions.clearFilters")}
        </Button>
        <Button type="button" variant="outline" disabled={comparedModels.length < 2} onClick={() => setCompareOpen(true)}>
          <GitCompare className="h-4 w-4" />
          {ux("models.actions.compare", { count: comparedModels.length })}
        </Button>
        <div className="ms-auto flex items-center gap-2 text-xs text-muted-foreground">
          <ArrowUpDown className="h-4 w-4" />
          <span>{visibleModels.length}</span>
        </div>
      </div>

      {visibleModels.length === 0 ? (
        <p className="text-sm text-muted-foreground">{props.t("step3.empty")}</p>
      ) : (
        <div className="space-y-3">
          {visibleModels.map((model) => {
            const availableCapabilities = normalizeCapabilityList(model.types);
            const draft = props.modelDrafts[model.model] ?? {
              enabled: false,
              compared: false,
              selectedCapabilities: availableCapabilities.length === 1 ? availableCapabilities : [],
              pricingMode: "public" as const,
              inputPrice: model.public_pricing?.input_usd_per_1m?.toString() ?? "",
              outputPrice: model.public_pricing?.output_usd_per_1m?.toString() ?? "",
            };
            return (
              <Popover
                key={model.model}
                  open={pricingEditor.open && pricingEditor.modelKey === model.model}
                  onOpenChange={(open) => {
                    if (!open) {
                      setPricingEditor((prev) => ({ ...prev, open: false, error: "" }));
                    } else {
                      openPricingEditor(model);
                    }
                  }}
                >
                  <div className="space-y-2">
                    <ModelCard
                      model={toModelCardData(model)}
                      detailsHref={buildModelDiscoveryDetailsHref(model)}
                      pricingDisplay={{
                        currency: displayCurrency,
                        scale: displayScale,
                        usdToIrr,
                      }}
                      isSelected={draft.enabled}
                      stateBadge={
                        draft.enabled
                          ? { label: ux("models.state.enabledOnAccount"), tone: "active" as const }
                          : { label: ux("models.state.availableToAdd"), tone: "inactive" as const }
                      }
                      onSelectedChange={(checked) =>
                        props.setModelDrafts((prev) => ({
                          ...prev,
                          [model.model]: {
                            ...draft,
                            enabled: checked,
                            selectedCapabilities:
                              checked && draft.selectedCapabilities.length === 0 && availableCapabilities.length === 1
                                ? availableCapabilities
                                : draft.selectedCapabilities,
                          },
                        }))
                      }
                      selectedCheckboxLabel={ux("common.enabled")}
                      isCompared={draft.compared}
                      onComparedChange={(checked) =>
                        props.setModelDrafts((prev) => ({
                          ...prev,
                          [model.model]: { ...draft, compared: checked },
                        }))
                      }
                      compareCheckboxLabel={ux("models.actions.compareToggle")}
                      pricingModeBadge={draft.pricingMode === "private" ? ux("models.pricing.private") : ux("models.pricing.public")}
                      className={draft.enabled ? "ring-1 ring-emerald-500/20" : undefined}
                      footerActions={
                        <Button type="button" variant="outline" size="sm" className="gap-1" onClick={() => openCurlGuide(model)}>
                          <Code2 className="h-3.5 w-3.5" />
                          {ux("models.card.curlGuide")}
                        </Button>
                      }
                      pricingAction={
                        <PopoverTrigger asChild>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={!props.canManage}
                            aria-label={`${ux("models.actions.editPricing")} ${model.display_name || model.model}`}
                            onClick={() => openPricingEditor(model)}
                          >
                            {ux("models.actions.editPricing")}
                          </Button>
                        </PopoverTrigger>
                      }
                      labels={{
                        context: ux("models.card.context"),
                        maxOutput: ux("models.card.maxOutput"),
                        input: ux("models.labels.input"),
                        output: ux("models.labels.output"),
                        latency: ux("models.card.latency"),
                        quality: ux("models.card.quality"),
                        pricing: ux("models.card.pricing"),
                        pricingUnavailable: ux("models.card.pricingUnavailable"),
                        technicalMetadata: ux("models.card.technicalMetadata"),
                        modelId: ux("models.card.modelId"),
                        providerId: ux("models.card.providerId"),
                        apiMode: ux("models.card.apiMode"),
                        releaseDate: ux("models.card.releaseDate"),
                        knowledgeCutoff: ux("models.card.knowledgeCutoff"),
                        regions: ux("models.card.regions"),
                        rateLimitNotes: ux("models.card.rateLimitNotes"),
                        license: ux("models.card.license"),
                        viewDetails: ux("models.card.viewDetails"),
                        copyId: ux("models.card.copyFqn"),
                        copied: ux("models.card.copied"),
                        lowerCost: ux("models.card.lowerCost"),
                        recommended: ux("models.card.recommended"),
                        popular: ux("models.card.popular"),
                        tradeOffs: ux("models.card.tradeOffs"),
                        cachedInput: ux("models.card.cachedInput"),
                        image: ux("models.card.image"),
                        audio: ux("models.card.audio"),
                        statusStable: ux("models.card.statusStable"),
                        statusPreview: ux("models.card.statusPreview"),
                        statusDeprecated: ux("models.card.statusDeprecated"),
                        statusNew: ux("models.card.statusNew"),
                        categoryChat: ux("models.card.categoryChat"),
                        categoryReasoning: ux("models.card.categoryReasoning"),
                        categoryEmbedding: ux("models.card.categoryEmbedding"),
                        categoryVision: ux("models.card.categoryVision"),
                        categoryAudio: ux("models.card.categoryAudio"),
                        categoryCode: ux("models.card.categoryCode"),
                        categoryMultimodal: ux("models.card.categoryMultimodal"),
                      }}
                    />
                    {draft.enabled && availableCapabilities.length > 0 ? (
                      <div className="rounded-xl border border-border/70 bg-muted/20 p-3">
                        <div className="flex items-center justify-between gap-2">
                          <p className="text-sm font-medium">{ux("models.columns.capabilities")}</p>
                          <Badge variant="secondary">{draft.selectedCapabilities.length}/{availableCapabilities.length}</Badge>
                        </div>
                        <div className="mt-2 flex flex-wrap gap-2">
                          {availableCapabilities.map((capability) => {
                            const checked = draft.selectedCapabilities.includes(capability);
                            return (
                              <label
                                key={`${model.model}-${capability}`}
                                className={cn(
                                  "inline-flex cursor-pointer items-center gap-2 rounded-full border px-3 py-1.5 text-sm",
                                  checked ? "border-emerald-500 bg-emerald-500/10 text-emerald-700" : "border-border bg-background",
                                )}
                              >
                                <input
                                  type="checkbox"
                                  className="h-4 w-4"
                                  checked={checked}
                                  onChange={(event) =>
                                    props.setModelDrafts((prev) => {
                                      const current = prev[model.model] ?? draft;
                                      const nextCapabilities = event.target.checked
                                        ? [...current.selectedCapabilities, capability]
                                        : current.selectedCapabilities.filter((item) => item !== capability);
                                      return {
                                        ...prev,
                                        [model.model]: {
                                          ...current,
                                          selectedCapabilities: normalizeCapabilityList(nextCapabilities),
                                        },
                                      };
                                    })
                                  }
                                />
                                <span>{capability}</span>
                              </label>
                            );
                          })}
                        </div>
                        {draft.selectedCapabilities.length === 0 ? (
                          <p className="mt-2 text-xs text-destructive">{ux("messages.modelTypeRequired")}</p>
                        ) : null}
                      </div>
                    ) : null}
                  </div>
                  <PopoverContent align="start" className="w-80 space-y-3">
                    <div className="space-y-1">
                      <p className="text-sm font-medium">{ux("models.actions.editPricing")}</p>
                      <p className="text-xs text-muted-foreground">{model.display_name || model.model}</p>
                    </div>
                    <div className="space-y-2">
                      <Label>{ux("models.columns.pricingMode")}</Label>
                      <select
                        className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm"
                        value={pricingEditor.pricingMode}
                        onChange={(event) =>
                          setPricingEditor((prev) => ({
                            ...prev,
                            pricingMode: event.target.value as "public" | "private",
                            error: "",
                          }))
                        }
                      >
                        <option value="public">{ux("models.pricing.public")}</option>
                        <option value="private">{ux("models.pricing.private")}</option>
                      </select>
                    </div>
                    {pricingEditor.pricingMode === "private" ? (
                      <div className="space-y-2">
                        <Input
                          value={pricingEditor.inputPrice}
                          onChange={(event) => setPricingEditor((prev) => ({ ...prev, inputPrice: event.target.value, error: "" }))}
                          placeholder={ux("models.placeholders.input")}
                        />
                        <Input
                          value={pricingEditor.outputPrice}
                          onChange={(event) => setPricingEditor((prev) => ({ ...prev, outputPrice: event.target.value, error: "" }))}
                          placeholder={ux("models.placeholders.output")}
                        />
                      </div>
                    ) : (
                      <div className="space-y-1 text-xs text-muted-foreground">
                        <div>{ux("models.catalogPricing")}</div>
                        <div>{ux("models.labels.input")}: {formatDisplayPrice(model.public_pricing?.input_usd_per_1m, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") })}</div>
                        <div>{ux("models.labels.output")}: {formatDisplayPrice(model.public_pricing?.output_usd_per_1m, { currency: displayCurrency, scale: displayScale, usdToIrr, naLabel: ux("models.na") })}</div>
                      </div>
                    )}
                    {pricingEditor.error ? (
                      <p className="text-xs text-destructive">{pricingEditor.error}</p>
                    ) : null}
                    <div className="flex justify-end gap-2">
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => setPricingEditor((prev) => ({ ...prev, open: false, error: "" }))}
                      >
                        {ux("models.actions.cancel")}
                      </Button>
                      <Button type="button" size="sm" onClick={applyPricingEditor}>
                        {ux("models.actions.apply")}
                      </Button>
                    </div>
                  </PopoverContent>
                </Popover>
            );
          })}
        </div>
      )}
      <Dialog open={compareOpen} onOpenChange={setCompareOpen}>
        <DialogContent className="max-h-[90vh] max-w-7xl overflow-hidden p-0">
          <DialogHeader>
            <div className="space-y-1 px-6 pt-6">
              <DialogTitle>{ux("models.compare.title")}</DialogTitle>
              <DialogDescription>{ux("models.compare.description")}</DialogDescription>
            </div>
          </DialogHeader>
          <div className="space-y-4 overflow-y-auto px-6 pb-6">
            <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
              <Card className="border-emerald-500/30 bg-emerald-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestInput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestInput ? compareKpis.cheapestInput.display_name || compareKpis.cheapestInput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-emerald-500/30 bg-emerald-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.bestOutput")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.cheapestOutput ? compareKpis.cheapestOutput.display_name || compareKpis.cheapestOutput.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-sky-500/30 bg-sky-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.largestContext")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.largestContext ? compareKpis.largestContext.display_name || compareKpis.largestContext.model : ux("models.na")}</CardContent>
              </Card>
              <Card className="border-violet-500/30 bg-violet-500/5">
                <CardHeader className="pb-2">
                  <CardDescription>{ux("models.compare.kpis.richestFeatures")}</CardDescription>
                </CardHeader>
                <CardContent className="text-sm font-medium">{compareKpis.mostFeatures ? compareKpis.mostFeatures.display_name || compareKpis.mostFeatures.model : ux("models.na")}</CardContent>
              </Card>
            </div>

            <div className="flex flex-wrap items-center gap-3 rounded-md border p-3">
              <Label htmlFor="compare-sort">{ux("models.compare.sort.label")}</Label>
              <select
                id="compare-sort"
                className="h-9 rounded-md border border-input bg-background px-2 text-sm"
                value={compareSortBy}
                onChange={(event) => setCompareSortBy(event.target.value as typeof compareSortBy)}
              >
                <option value="input_asc">{ux("models.compare.sort.inputAsc")}</option>
                <option value="output_asc">{ux("models.compare.sort.outputAsc")}</option>
                <option value="context_desc">{ux("models.compare.sort.contextDesc")}</option>
                <option value="name_asc">{ux("models.compare.sort.nameAsc")}</option>
              </select>
              <label className="inline-flex items-center gap-2 text-sm">
                <input type="checkbox" checked={compareDiffOnly} onChange={(event) => setCompareDiffOnly(event.target.checked)} />
                {ux("models.compare.toggles.differencesOnly")}
              </label>
              <label className="inline-flex items-center gap-2 text-sm">
                <input type="checkbox" checked={compareNormalizedPricing} onChange={(event) => setCompareNormalizedPricing(event.target.checked)} />
                {ux("models.compare.toggles.normalizedPricing")}
              </label>
            </div>

            <div className="hidden md:block overflow-x-auto rounded-md border">
              <div className="min-w-[980px]">
                <div
                  className="sticky top-0 z-10 grid border-b bg-background"
                  style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}
                >
                  <div className="sticky left-0 z-20 border-e bg-background p-3 text-sm font-semibold">{ux("models.compare.columns.attribute")}</div>
                  {compareModels.map((model) => (
                    <div key={`compare-head-${model.model}`} className="border-s px-3 py-3">
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold">{model.display_name || model.model}</p>
                          <p className="truncate text-xs text-muted-foreground">{model.provider}</p>
                        </div>
                        <Button type="button" size="icon" variant="ghost" onClick={() => removeComparedModel(model.model)} aria-label={ux("models.compare.actions.removeModel")}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>

                {[
                  ["pricing", ux("models.compare.sections.pricing")],
                  ["capacity", ux("models.compare.sections.capacity")],
                  ["modalities", ux("models.compare.sections.modalities")],
                  ["capabilities", ux("models.compare.sections.capabilities")],
                ].map(([sectionKey, sectionLabel]) => (
                  <div key={sectionKey}>
                    <div
                      className="grid border-b bg-muted/40"
                      style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}
                    >
                      <div className="sticky left-0 z-[1] border-e bg-muted/50 p-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{sectionLabel}</div>
                      {compareModels.map((model) => <div key={`${sectionKey}-${model.model}`} className="border-s p-2" />)}
                    </div>
                    {visibleCompareRows.filter((row) => row.section === sectionKey).map((row) => (
                      <div key={row.key} className="grid border-b" style={{ gridTemplateColumns: `240px repeat(${compareModels.length}, minmax(220px, 1fr))` }}>
                        <div className="sticky left-0 z-[1] border-e bg-background p-3 text-sm font-medium">{row.label}</div>
                        {row.values.map((value, index) => {
                          const model = compareModels[index];
                          const isBestInput = row.key === "input" && typeof value === "number" && value === bestMarkers.input;
                          const isBestOutput = row.key === "output" && typeof value === "number" && value === bestMarkers.output;
                          const isBestContext = row.key === "context" && typeof value === "number" && value === bestMarkers.context;
                          return (
                            <div
                              key={`${row.key}-${model.model}`}
                              className={`border-s p-3 text-sm ${row.isDifferent ? "bg-amber-50/40 dark:bg-amber-950/10" : ""} ${(isBestInput || isBestOutput || isBestContext) ? "bg-emerald-50 dark:bg-emerald-950/20" : ""}`}
                            >
                              {row.key === "features" || row.key === "inputModality" || row.key === "outputModality" ? (
                                <div className="flex flex-wrap gap-1">
                                  {`${value ?? ""}`.split(",").map((item) => item.trim()).filter(Boolean).length > 0 ? (
                                    `${value ?? ""}`.split(",").map((item) => item.trim()).filter(Boolean).map((item) => (
                                      <Badge key={`${row.key}-${model.model}-${item}`} variant="secondary">{item}</Badge>
                                    ))
                                  ) : (
                                    <span className="text-muted-foreground">{ux("models.na")}</span>
                                  )}
                                </div>
                              ) : (
                                <span>
                                  {renderCompareValue(row.key, value, row.render)}
                                </span>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-3 md:hidden">
              {compareModels.map((model) => (
                <Card key={`compare-mobile-${model.model}`}>
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <CardTitle className="truncate text-sm">{model.display_name || model.model}</CardTitle>
                        <CardDescription className="truncate">{model.provider}</CardDescription>
                      </div>
                      <Button type="button" size="icon" variant="ghost" onClick={() => removeComparedModel(model.model)} aria-label={ux("models.compare.actions.removeModel")}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    {visibleCompareRows.map((row) => {
                      const rowIndex = compareModels.findIndex((item) => item.model === model.model);
                      const value = row.values[rowIndex];
                      return (
                        <div key={`mobile-${model.model}-${row.key}`} className="rounded border p-2">
                          <div className="text-xs text-muted-foreground">{row.label}</div>
                          <div className="mt-1">
                            {row.key === "features" || row.key === "inputModality" || row.key === "outputModality"
                              ? (`${value ?? ""}` || ux("models.na"))
                              : row.key === "input" || row.key === "output"
                                ? renderCompareValue(row.key, value, row.render)
                                : row.render(value)}
                          </div>
                        </div>
                      );
                    })}
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </DialogContent>
      </Dialog>
      <Dialog open={curlGuide !== null} onOpenChange={(open) => { if (!open) setCurlGuide(null); }}>
        <DialogContent className="max-w-4xl overflow-hidden p-0">
          <div className="border-b border-border/70 bg-[linear-gradient(135deg,color-mix(in_oklab,var(--primary)_10%,transparent),transparent_58%),linear-gradient(180deg,color-mix(in_oklab,var(--muted)_65%,transparent),transparent)] px-6 py-6">
            <DialogHeader className="space-y-3">
              <div className="flex items-center justify-between gap-4">
                <div className="space-y-2">
                  <DialogTitle>{ux("models.curlGuide.title")}</DialogTitle>
                  <DialogDescription>{ux("models.curlGuide.description", { model: curlGuide?.modelName ?? "" })}</DialogDescription>
                </div>
                <Badge className="border border-primary/20 bg-primary/10 text-primary">
                  <Code2 className="me-1 h-3.5 w-3.5" />
                  {ux("models.curlGuide.readyBadge")}
                </Badge>
              </div>
            </DialogHeader>
            <div className="mt-5 grid gap-3 md:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
              <div className="rounded-2xl border border-border/70 bg-background/80 p-4 shadow-sm backdrop-blur">
                <div className="text-xs font-medium text-muted-foreground">{ux("models.curlGuide.modelFqnLabel")}</div>
                <div dir="ltr" className="mt-2 break-all text-left font-mono text-sm text-foreground">{curlGuide?.modelFqn}</div>
              </div>
              <div className="rounded-2xl border border-border/70 bg-background/80 p-4 shadow-sm backdrop-blur">
                <div className="text-xs font-medium text-muted-foreground">{ux("models.curlGuide.instructionsLabel")}</div>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{ux("models.curlGuide.instructions")}</p>
              </div>
            </div>
          </div>
          <div className="space-y-5 px-6 py-6">
            <Tabs value={curlGuideShell} onValueChange={(value) => setCurlGuideShell(value as CurlGuideShell)} className="space-y-4">
              <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                <TabsList className="h-auto w-full justify-start rounded-xl bg-muted/70 p-1 md:w-auto">
                  <TabsTrigger value="powershell" className="min-w-32">{ux("models.curlGuide.shells.powershell")}</TabsTrigger>
                  <TabsTrigger value="bash" className="min-w-24">{ux("models.curlGuide.shells.bash")}</TabsTrigger>
                </TabsList>
                <Button
                  type="button"
                  variant="outline"
                  className="w-full md:w-auto"
                  onClick={() => {
                    const command = curlGuide?.commands[curlGuideShell];
                    if (!command) return;
                    void navigator.clipboard.writeText(command);
                    props.setBanner({ kind: "success", message: ux("messages.curlCopied") });
                  }}
                >
                  <Copy className="me-2 h-4 w-4" />
                  {ux("models.curlGuide.copyCommand")}
                </Button>
              </div>
              <TabsContent value="powershell">
                <CodeExampleBlock
                  code={curlGuide?.commands.powershell ?? ""}
                  language={CURL_GUIDE_LANGUAGE.powershell}
                  shellLabel={ux("models.curlGuide.shells.powershell")}
                />
              </TabsContent>
              <TabsContent value="bash">
                <CodeExampleBlock
                  code={curlGuide?.commands.bash ?? ""}
                  language={CURL_GUIDE_LANGUAGE.bash}
                  shellLabel={ux("models.curlGuide.shells.bash")}
                />
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter className="border-t border-border/70 bg-muted/20 px-6 py-4">
            <Button type="button" onClick={() => setCurlGuide(null)}>
              {ux("models.curlGuide.close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}

function AccessEditor(props: {
  t: ReturnType<typeof useTranslations>;
  managers: AccessPrincipal[];
  setManagers: (value: AccessPrincipal[]) => void;
  users: AccessPrincipal[];
  setUsers: (value: AccessPrincipal[]) => void;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<AccessPrincipalCandidate[]>;
  principalSelectorLabels: PrincipalEntitySelectLabels;
  onSave: () => void;
  isSaving: boolean;
  canManage: boolean;
  effectiveKeyType: "pat" | "vat";
  setEffectiveKeyType: (value: "pat" | "vat") => void;
  effectivePrincipal: PrincipalEntitySelection | null;
  setEffectivePrincipal: (value: PrincipalEntitySelection | null) => void;
  effectiveModels: EffectiveModelView[];
  isLoadingEffective: boolean;
  onLoadEffective: () => void;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const hasEffectivePrincipal = Boolean((props.effectivePrincipal?.principal_id ?? "").trim());

  return (
    <section className="space-y-5 rounded-lg border border-border/70 bg-background p-4 sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h3 className="text-base font-semibold">{ux("tabs.access")}</h3>
          <p className="mt-1 text-sm leading-6 text-muted-foreground">{ux("access.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={!props.canManage || props.isSaving} className="shrink-0">
          <ShieldCheck className="h-4 w-4" />
          {props.isSaving ? props.t("saving") : props.t("saveAccessControl")}
        </Button>
      </div>

      <div className="grid items-stretch gap-4">
        <PrincipalPicker
          id="managers-principal-picker"
          title={props.t("step3.managersTitle")}
          onSearchPrincipal={props.onSearchPrincipal}
          principalSelectorLabels={props.principalSelectorLabels}
          selected={props.managers}
          setSelected={props.setManagers}
          disabled={!props.canManage}
          allowEveryone={false}
          addLabel={props.t("principal.add")}
          removeLabel={props.t("principal.remove")}
        />

        <PrincipalPicker
          id="users-principal-picker"
          title={props.t("step3.usersTitle")}
          onSearchPrincipal={props.onSearchPrincipal}
          principalSelectorLabels={props.principalSelectorLabels}
          selected={props.users}
          setSelected={props.setUsers}
          disabled={!props.canManage}
          allowEveryone
          addLabel={props.t("principal.add")}
          removeLabel={props.t("principal.remove")}
        />
      </div>

      <div className="space-y-4 rounded-lg border border-border/70 bg-muted/10 p-4">
        <h4 className="text-base font-semibold">{ux("access.effectivePreviewTitle")}</h4>
        <div className="grid gap-3 lg:grid-cols-[150px_minmax(0,1fr)_auto]">
          <div className="grid gap-2">
            <Label htmlFor="previewKeyType">{ux("access.keyTypeLabel")}</Label>
            <select
              id="previewKeyType"
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
              value={props.effectiveKeyType}
              onChange={(event) => props.setEffectiveKeyType(event.target.value as "pat" | "vat")}
            >
              <option value="pat">{ux("access.pat")}</option>
              <option value="vat">{ux("access.vat")}</option>
            </select>
          </div>
          <Field label={`${ux("access.userIdLabel")} / ${ux("access.teamIdLabel")}`} id="previewPrincipal">
            <div className="grid grid-cols-[minmax(0,1fr)_40px] gap-2">
              <PrincipalEntitySelect
                id="previewPrincipal"
                selected={props.effectivePrincipal}
                onSelectedChange={props.setEffectivePrincipal}
                onSearch={props.onSearchPrincipal}
                labels={props.principalSelectorLabels}
                kinds={["user", "team"]}
                className="min-w-0"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => props.setEffectivePrincipal(null)}
                disabled={!hasEffectivePrincipal}
                title={ux("actions.clear")}
                aria-label={ux("actions.clear")}
                className="mt-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </Field>
          <div className="flex items-end">
            <Button
              type="button"
              variant="outline"
              onClick={props.onLoadEffective}
              disabled={props.isLoadingEffective || !hasEffectivePrincipal}
              className="h-10 w-full xl:w-auto"
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

function buildPrincipalSelectorLabels(
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
    empty: t("principal.searchHint"),
  };
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

function normalizeCapabilityList(values: string[] | undefined): string[] {
  if (!Array.isArray(values)) return [];
  const seen = new Set<string>();
  const next: string[] = [];
  for (const item of values) {
    const normalized = item.trim().toLowerCase();
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    next.push(normalized);
  }
  return next;
}

const FRIENDLY_ACCOUNT_ADJECTIVES = [
  "amber",
  "brisk",
  "calm",
  "clear",
  "cobalt",
  "daring",
  "eager",
  "ember",
  "gentle",
  "golden",
  "lively",
  "lucky",
  "minty",
  "nimble",
  "pearl",
  "rapid",
  "silver",
  "sunny",
  "tidy",
  "velvet",
];

const FRIENDLY_ACCOUNT_NOUNS = [
  "anchor",
  "breeze",
  "comet",
  "falcon",
  "harbor",
  "lighthouse",
  "lollipop",
  "lotus",
  "meadow",
  "nebula",
  "oasis",
  "orbit",
  "pioneer",
  "ribbon",
  "sparrow",
  "summit",
  "tidal",
  "voyager",
  "willow",
  "zephyr",
];

function ProviderLogo(props: { provider: string; displayName: string; logoURL?: string; size?: number }) {
  const size = props.size ?? 18;
  const fallbackSrc = getProviderFallbackLogoPath(props.provider);
  const [hasError, setHasError] = useState(false);
  const src = hasError ? fallbackSrc : getProviderLogoSrc(props.provider, props.logoURL);

  return (
    <Image
      src={src}
      alt={props.displayName}
      width={size}
      height={size}
      className="shrink-0 rounded-sm object-contain"
      unoptimized
      onError={() => setHasError(true)}
    />
  );
}

function getProviderLogoSrc(provider: string, logoURL?: string): string {
  const proxiedLogoURL = toTenantProviderAssetPath(logoURL);
  if (proxiedLogoURL) return proxiedLogoURL;
  return getProviderFallbackLogoPath(provider);
}

function toTenantProviderAssetPath(logoURL?: string): string {
  const raw = logoURL?.trim() ?? "";
  if (!raw) return "";
  const candidate = (() => {
    if (raw.startsWith("/")) return raw;
    try {
      return new URL(raw).pathname;
    } catch {
      return raw;
    }
  })();
  const match = candidate.match(/^\/api\/v1\/providers\/assets\/([^/]+)\/([^/]+)$/i);
  if (!match) return raw;
  const [, bucket, fileName] = match;
  return `/api/provider-catalog/assets/${encodeURIComponent(bucket)}/${encodeURIComponent(fileName)}`;
}

function getProviderFallbackLogoPath(provider: string): string {
  const key = normalizeProviderSlug(provider);
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

function shouldGenerateProviderAccountName(input: {
  currentName: string;
  currentProvider: string;
  isAutoGenerated: boolean;
}): boolean {
  const currentName = input.currentName.trim().toLowerCase();
  const currentProvider = normalizeProviderSlug(input.currentProvider);
  if (!currentName) return true;
  if (input.isAutoGenerated) return true;
  if (!currentProvider) return false;
  return currentName === currentProvider || currentName.startsWith(`${currentProvider}-`);
}

function generateProviderAccountName(provider: string, existingNames: string[]): string {
  const normalizedProvider = normalizeProviderSlug(provider) || "provider";
  const usedNames = new Set(existingNames.map((item) => item.trim().toLowerCase()).filter(Boolean));

  for (let attempt = 0; attempt < 24; attempt += 1) {
    const adjective = pickRandom(FRIENDLY_ACCOUNT_ADJECTIVES);
    const noun = pickRandom(FRIENDLY_ACCOUNT_NOUNS);
    const candidate = `${normalizedProvider}-${adjective}-${noun}`;
    if (!usedNames.has(candidate.toLowerCase())) return candidate;
  }

  return `${normalizedProvider}-${pickRandom(FRIENDLY_ACCOUNT_ADJECTIVES)}-${Math.floor(100 + Math.random() * 900)}`;
}

function pickRandom(values: readonly string[]): string {
  return values[Math.floor(Math.random() * values.length)] ?? "account";
}

function normalizeProviderSlug(value: string): string {
  const normalized = value.trim().toLowerCase();
  switch (normalized) {
    case "perplexity":
      return "perplexity-ai";
    case "zai":
      return "xai";
    default:
      return normalized;
  }
}

function stringsEqualFold(a: string, b: string): boolean {
  return normalizeProviderSlug(a) === normalizeProviderSlug(b);
}

function formatDisplayPrice(
  usdPerMillion: number | undefined,
  options: {
    currency: "USD" | "TOMAN";
    scale: "PER_MILLION_TOKENS" | "PER_TOKEN";
    usdToIrr: number;
    naLabel: string;
  },
): string {
  if (typeof usdPerMillion !== "number" || !Number.isFinite(usdPerMillion)) return options.naLabel;

  let value = usdPerMillion;
  if (options.scale === "PER_TOKEN") {
    value = value / 1_000_000;
  }

  let suffix = options.scale === "PER_TOKEN" ? "/ token" : "/ 1M";
  if (options.currency === "TOMAN") {
    const tomanValue = (value * options.usdToIrr) / 10;
    const formatted = new Intl.NumberFormat("en-US", {
      maximumFractionDigits: options.scale === "PER_TOKEN" ? 2 : 0,
    }).format(tomanValue);
    suffix = options.scale === "PER_TOKEN" ? " toman / token" : " toman / 1M";
    return `${formatted} ${suffix}`;
  }

  const formattedUsd = new Intl.NumberFormat("en-US", {
    maximumFractionDigits: options.scale === "PER_TOKEN" ? 8 : 6,
  }).format(value);
  return `$${formattedUsd} ${suffix}`;
}

