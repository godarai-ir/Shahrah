"use client";

import { useEffect, useMemo, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

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
  const [isLoading, setIsLoading] = useState(false);
  const [isSavingModels, setIsSavingModels] = useState(false);
  const [isSavingAccess, setIsSavingAccess] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(null);

  const [providers, setProviders] = useState<ProviderCatalogProvider[]>([]);
  const [accounts, setAccounts] = useState<TenantProviderAccount[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);

  const [selectedAccountID, setSelectedAccountID] = useState("");

  const [provider, setProvider] = useState("");
  const [name, setName] = useState("");
  const [authMode, setAuthMode] = useState<ProviderAuthMode>("api_key");
  const [allowVirtualOnly, setAllowVirtualOnly] = useState(false);
  const [apiKey, setAPIKey] = useState("");
  const [secretRef, setSecretRef] = useState("");

  const [modelDrafts, setModelDrafts] = useState<Record<string, ModelDraft>>({});
  const [managers, setManagers] = useState<AccessPrincipal[]>([]);
  const [users, setUsers] = useState<AccessPrincipal[]>([{ type: "everyone", display_name: "everyone" }]);

  const [managerKind, setManagerKind] = useState<"user" | "team">("user");
  const [managerQuery, setManagerQuery] = useState("");
  const [managerResults, setManagerResults] = useState<AccessPrincipalCandidate[]>([]);

  const [userKind, setUserKind] = useState<"user" | "team">("user");
  const [userQuery, setUserQuery] = useState("");
  const [userResults, setUserResults] = useState<AccessPrincipalCandidate[]>([]);

  useEffect(() => {
    void loadInitial();
  }, []);

  const selectedProvider = useMemo(
    () => providers.find((item) => item.provider === provider) ?? null,
    [providers, provider],
  );

  async function loadInitial() {
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

      if (!provider && providerItems.length > 0) {
        setProvider(providerItems[0].provider);
        setName(providerItems[0].provider);
        setAuthMode(providerItems[0].auth_modes?.[0] ?? "api_key");
      }
      if (!selectedAccountID && accountItems.length > 0) {
        setSelectedAccountID(accountItems[0].id);
      }
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "بارگذاری اولیه انجام نشد.") });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleLoadProviderModels(nextProvider: string) {
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
        message: normalizeError(error, "خواندن مدل‌های provider از discovery انجام نشد."),
      });
    }
  }

  async function handleCreateAccount() {
    setBanner(null);
    if (!provider || !name.trim()) {
      setBanner({ kind: "error", message: "انتخاب provider و نام اکانت الزامی است." });
      return;
    }

    const payload: Record<string, unknown> = {
      provider,
      name: name.trim(),
      auth_mode: authMode,
      allow_virtual_keys_only: allowVirtualOnly,
      provider_config: {},
    };

    if (apiKey.trim() || secretRef.trim()) {
      payload.credentials = {
        api_key: apiKey.trim() || undefined,
        secret_ref: secretRef.trim() || undefined,
      };
    }

    setIsLoading(true);
    try {
      const created = await requestJSON<TenantProviderAccount>("/api/tenant-provider-accounts", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      setSelectedAccountID(created.id);
      await loadInitial();
      await handleLoadProviderModels(created.provider);
      setBanner({ kind: "success", message: "Provider account با موفقیت ساخته شد." });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "ساخت provider account انجام نشد.") });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleLoadAccount(accountID: string) {
    if (!accountID) {
      return;
    }
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
      setBanner({ kind: "success", message: "جزئیات account بارگذاری شد." });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "بارگذاری account انجام نشد.") });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleSaveModels() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: "ابتدا یک account انتخاب یا ایجاد کنید." });
      return;
    }
    const payloadModels = Object.entries(modelDrafts)
      .filter(([, draft]) => draft.selected)
      .map(([model, draft]) => {
        const privatePricing =
          draft.pricingMode === "private"
            ? {
                input_usd_per_1m: toNumberOrUndefined(draft.inputPrice),
                output_usd_per_1m: toNumberOrUndefined(draft.outputPrice),
              }
            : undefined;

        return {
          model,
          enabled: true,
          types: draft.types,
          pricing_mode: draft.pricingMode,
          private_pricing: privatePricing,
          settings: {},
        };
      });

    if (payloadModels.length === 0) {
      setBanner({ kind: "error", message: "حداقل یک مدل انتخاب کنید." });
      return;
    }

    setIsSavingModels(true);
    setBanner(null);
    try {
      await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/models`, {
        method: "PUT",
        body: JSON.stringify({ models: payloadModels }),
      });
      setBanner({ kind: "success", message: "تنظیمات مدل‌ها ذخیره شد." });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "ذخیره مدل‌ها انجام نشد.") });
    } finally {
      setIsSavingModels(false);
    }
  }

  async function handleSaveAccess() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: "ابتدا یک account انتخاب یا ایجاد کنید." });
      return;
    }
    if (managers.length === 0) {
      setBanner({ kind: "error", message: "حداقل یک manager لازم است." });
      return;
    }

    setIsSavingAccess(true);
    setBanner(null);
    try {
      await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/access`, {
        method: "PUT",
        body: JSON.stringify({
          access: {
            managers,
            users,
          },
        }),
      });
      setBanner({ kind: "success", message: "Access control ذخیره شد." });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "ذخیره access انجام نشد.") });
    } finally {
      setIsSavingAccess(false);
    }
  }

  async function handleTestCredentials() {
    if (!selectedAccountID) {
      setBanner({ kind: "error", message: "ابتدا یک account انتخاب یا ایجاد کنید." });
      return;
    }
    setIsTesting(true);
    setBanner(null);
    try {
      const result = await requestJSON<{ success: boolean; message?: string }>(
        `/api/tenant-provider-accounts/${encodeURIComponent(selectedAccountID)}/test-credentials`,
        {
          method: "POST",
          body: JSON.stringify({}),
        },
      );
      setBanner({
        kind: result.success ? "success" : "error",
        message: result.message ?? (result.success ? "اعتبارسنجی موفق بود." : "اعتبارسنجی ناموفق بود."),
      });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "تست credential انجام نشد.") });
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
      setBanner({ kind: "error", message: normalizeError(error, "جستجوی manager انجام نشد.") });
    }
  }

  async function searchUserPrincipals() {
    try {
      const response = await requestJSON<AccessPrincipalResponse>(
        `/api/access-principals?kind=${userKind}&q=${encodeURIComponent(userQuery.trim())}&limit=20`,
      );
      setUserResults(response.data ?? []);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, "جستجوی user/team انجام نشد.") });
    }
  }

  return (
    <Card id="tenant-provider-accounts">
      <CardHeader>
        <CardTitle>Tenant Provider Accounts</CardTitle>
        <CardDescription>
          انتخاب provider از discovery، مدیریت مدل‌های tenant و تنظیم access با users/teams اصلی پلتفرم
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {banner ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? "خطا" : "موفق"}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

        <section className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
          <div className="grid gap-2">
            <Label htmlFor="providerAccountSelect">Account موجود</Label>
            <select
              id="providerAccountSelect"
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              value={selectedAccountID}
              onChange={(event) => setSelectedAccountID(event.target.value)}
            >
              <option value="">انتخاب کنید</option>
              {accounts.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name} ({item.provider})
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <Button type="button" variant="outline" onClick={() => void handleLoadAccount(selectedAccountID)}>
              بارگذاری Account
            </Button>
          </div>
          <div className="flex items-end">
            <Button type="button" variant="outline" onClick={() => void loadInitial()} disabled={isLoading}>
              {isLoading ? "در حال بارگذاری..." : "Refresh"}
            </Button>
          </div>
          <div className="flex items-end">
            <Button type="button" variant="outline" onClick={() => void handleTestCredentials()} disabled={isTesting}>
              {isTesting ? "در حال تست..." : "Test Credentials"}
            </Button>
          </div>
        </section>

        <Separator />

        <section className="space-y-4">
          <h3 className="text-lg font-semibold">Step 1: Configure Provider Account</h3>
          <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
            <div className="grid gap-2">
              <Label htmlFor="providerSelect">Provider</Label>
              <select
                id="providerSelect"
                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                value={provider}
                onChange={(event) => {
                  const nextProvider = event.target.value;
                  setProvider(nextProvider);
                  const item = providers.find((providerItem) => providerItem.provider === nextProvider);
                  if (item) {
                    setAuthMode(item.auth_modes?.[0] ?? "api_key");
                    if (!name.trim()) {
                      setName(item.provider);
                    }
                  }
                  void handleLoadProviderModels(nextProvider);
                }}
              >
                <option value="">انتخاب Provider</option>
                {providers
                  .filter((item) => item.enabled)
                  .map((item) => (
                    <option key={item.provider} value={item.provider}>
                      {item.display_name}
                    </option>
                  ))}
              </select>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="providerAccountName">Account Name</Label>
              <Input
                id="providerAccountName"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="azure-openai-main"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="authMode">Auth Mode</Label>
              <select
                id="authMode"
                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                value={authMode}
                onChange={(event) => setAuthMode(event.target.value as ProviderAuthMode)}
              >
                {(selectedProvider?.auth_modes ?? ["api_key"]).map((mode) => (
                  <option key={mode} value={mode}>
                    {mode}
                  </option>
                ))}
              </select>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="allowVirtual">Virtual Keys Mode</Label>
              <label className="flex h-10 items-center gap-2 rounded-md border border-input px-3 text-sm">
                <input
                  id="allowVirtual"
                  type="checkbox"
                  checked={allowVirtualOnly}
                  onChange={(event) => setAllowVirtualOnly(event.target.checked)}
                />
                allow_virtual_keys_only
              </label>
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="providerAPIKey">API Key (اختیاری)</Label>
              <Input
                id="providerAPIKey"
                value={apiKey}
                onChange={(event) => setAPIKey(event.target.value)}
                placeholder="sk-..."
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="providerSecretRef">Secret Ref (اختیاری)</Label>
              <Input
                id="providerSecretRef"
                value={secretRef}
                onChange={(event) => setSecretRef(event.target.value)}
                placeholder="vault://..."
              />
            </div>
          </div>

          <Button type="button" onClick={() => void handleCreateAccount()} disabled={isLoading}>
            {isLoading ? "در حال ایجاد..." : "Create Provider Account"}
          </Button>
        </section>

        <Separator />

        <section className="space-y-4">
          <h3 className="text-lg font-semibold">Step 2: Model Selection</h3>
          <p className="text-sm text-muted-foreground">
            مدل‌ها از discovery برای provider انتخاب‌شده خوانده می‌شوند.
          </p>
          <div className="space-y-3">
            {providerModels.length === 0 ? (
              <p className="text-sm text-muted-foreground">مدلی برای این provider دریافت نشد.</p>
            ) : (
              providerModels.map((model) => {
                const draft = modelDrafts[model.model] ?? {
                  selected: false,
                  types: model.types ?? [],
                  pricingMode: "public",
                  inputPrice: "",
                  outputPrice: "",
                };
                return (
                  <div key={model.model} className="rounded-lg border p-3">
                    <div className="flex flex-wrap items-center gap-3">
                      <label className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={draft.selected}
                          onChange={(event) =>
                            setModelDrafts((prev) => ({
                              ...prev,
                              [model.model]: {
                                ...draft,
                                selected: event.target.checked,
                              },
                            }))
                          }
                        />
                        <span className="font-medium">{model.display_name || model.model}</span>
                      </label>
                      <Badge variant="outline">{model.model}</Badge>
                      {(model.types ?? []).map((type) => (
                        <Badge key={type} variant="secondary">
                          {type}
                        </Badge>
                      ))}
                    </div>

                    {draft.selected ? (
                      <div className="mt-3 grid gap-3 md:grid-cols-3">
                        <div className="grid gap-1">
                          <Label>Pricing Mode</Label>
                          <select
                            className="h-9 rounded-md border border-input bg-background px-3 text-sm"
                            value={draft.pricingMode}
                            onChange={(event) =>
                              setModelDrafts((prev) => ({
                                ...prev,
                                [model.model]: {
                                  ...draft,
                                  pricingMode: event.target.value as "public" | "private",
                                },
                              }))
                            }
                          >
                            <option value="public">public</option>
                            <option value="private">private</option>
                          </select>
                        </div>

                        {draft.pricingMode === "private" ? (
                          <>
                            <div className="grid gap-1">
                              <Label>Input $ / 1M</Label>
                              <Input
                                value={draft.inputPrice}
                                onChange={(event) =>
                                  setModelDrafts((prev) => ({
                                    ...prev,
                                    [model.model]: {
                                      ...draft,
                                      inputPrice: event.target.value,
                                    },
                                  }))
                                }
                                placeholder="0.5"
                              />
                            </div>
                            <div className="grid gap-1">
                              <Label>Output $ / 1M</Label>
                              <Input
                                value={draft.outputPrice}
                                onChange={(event) =>
                                  setModelDrafts((prev) => ({
                                    ...prev,
                                    [model.model]: {
                                      ...draft,
                                      outputPrice: event.target.value,
                                    },
                                  }))
                                }
                                placeholder="1.5"
                              />
                            </div>
                          </>
                        ) : null}
                      </div>
                    ) : null}
                  </div>
                );
              })
            )}
          </div>
          <Button type="button" onClick={() => void handleSaveModels()} disabled={isSavingModels}>
            {isSavingModels ? "در حال ذخیره..." : "Save Selected Models"}
          </Button>
        </section>

        <Separator />

        <section className="space-y-4">
          <h3 className="text-lg font-semibold">Step 3: Access Control</h3>
          <div className="grid gap-6 md:grid-cols-2">
            <PrincipalPicker
              title="Managers (only user/team)"
              kind={managerKind}
              setKind={setManagerKind}
              query={managerQuery}
              setQuery={setManagerQuery}
              results={managerResults}
              onSearch={searchManagerPrincipals}
              selected={managers}
              setSelected={setManagers}
              allowEveryone={false}
            />
            <PrincipalPicker
              title="Users (view/use access)"
              kind={userKind}
              setKind={setUserKind}
              query={userQuery}
              setQuery={setUserQuery}
              results={userResults}
              onSearch={searchUserPrincipals}
              selected={users}
              setSelected={setUsers}
              allowEveryone
            />
          </div>
          <Button type="button" onClick={() => void handleSaveAccess()} disabled={isSavingAccess}>
            {isSavingAccess ? "در حال ذخیره..." : "Save Access Control"}
          </Button>
        </section>
      </CardContent>
    </Card>
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
}) {
  function addPrincipal(item: AccessPrincipalCandidate) {
    const exists = props.selected.some(
      (current) =>
        current.type === item.type &&
        (current.principal_id ?? "") === (item.principal_id ?? ""),
    );
    if (exists) {
      return;
    }
    props.setSelected([
      ...props.selected,
      {
        type: item.type,
        principal_id: item.principal_id,
        display_name: item.display_name,
      },
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
        <select
          className="h-9 rounded-md border border-input bg-background px-3 text-sm"
          value={props.kind}
          onChange={(event) => props.setKind(event.target.value as "user" | "team")}
        >
          <option value="user">user</option>
          <option value="team">team</option>
        </select>
        <Input
          value={props.query}
          onChange={(event) => props.setQuery(event.target.value)}
          placeholder="search..."
        />
        <Button type="button" variant="outline" onClick={() => void props.onSearch()}>
          Search
        </Button>
      </div>

      {props.allowEveryone ? (
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            const exists = props.selected.some((item) => item.type === "everyone");
            if (!exists) {
              props.setSelected([
                ...props.selected,
                { type: "everyone", display_name: "everyone" },
              ]);
            }
          }}
        >
          Add everyone
        </Button>
      ) : null}

      <div className="space-y-2">
        {props.results.map((item) => (
          <div key={`${item.type}-${item.principal_id}`} className="flex items-center justify-between rounded border px-2 py-1">
            <div>
              <div className="text-sm font-medium">{item.display_name}</div>
              <div className="text-xs text-muted-foreground">
                {item.type} {item.secondary ? `• ${item.secondary}` : ""}
              </div>
            </div>
            <Button type="button" size="sm" variant="outline" onClick={() => addPrincipal(item)}>
              Add
            </Button>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        {props.selected.map((item) => (
          <button
            key={`${item.type}-${item.principal_id ?? "everyone"}`}
            type="button"
            className="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs"
            onClick={() => removePrincipal(item)}
            title="remove"
          >
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
    const existing = current[model.model];
    next[model.model] =
      existing ??
      {
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

async function requestJSON<T = unknown>(
  path: string,
  options?: RequestInit,
): Promise<T> {
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
    throw new Error(
      extractMessageFromPayload(data) ?? `خطای درخواست (${response.status})`,
    );
  }
  return data as T;
}

function parseResponsePayload(text: string): unknown {
  if (!text.trim()) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessageFromPayload(payload: unknown): string | null {
  if (typeof payload === "string") {
    return payload;
  }
  if (!payload || typeof payload !== "object") {
    return null;
  }

  const candidate = payload as {
    error?: {
      message?: string;
    };
    message?: string;
  };

  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

function toNumberOrUndefined(value: string): number | undefined {
  const raw = value.trim();
  if (!raw) {
    return undefined;
  }
  const parsed = Number(raw);
  if (!Number.isFinite(parsed)) {
    return undefined;
  }
  return parsed;
}
