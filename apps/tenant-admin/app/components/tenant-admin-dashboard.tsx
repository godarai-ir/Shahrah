"use client";

import Link from "next/link";
import { FormEvent, ReactNode, useCallback, useMemo, useState } from "react";

import { extractTenantSlugFromHost } from "@/lib/tenant";

type TenantKey = {
  id: string;
  key?: string;
  key_alias: string;
  key_name?: string;
  role: string;
  models: string[];
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

type KeyListResponse = {
  data?: TenantKey[];
};

type ModelListResponse = {
  data?: TenantModel[];
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
  currentUser: {
    id: string;
    email?: string;
    mobile?: string;
    role: string;
  };
};

const numberFormatter = new Intl.NumberFormat("fa-IR");

export default function TenantAdminDashboard(props: TenantAdminDashboardProps) {
  const [keys, setKeys] = useState<TenantKey[]>(props.initialKeys);
  const [models, setModels] = useState<TenantModel[]>(props.initialModels);
  const [wallet, setWallet] = useState<Wallet | null>(props.initialWallet);
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
  const [isTopUpLoading, setIsTopUpLoading] = useState(false);

  const [keyAlias, setKeyAlias] = useState("tenant-dashboard-key");
  const [keyModelsInput, setKeyModelsInput] = useState("gpt-4o-mini");

  const [modelName, setModelName] = useState("gpt-4o-mini");
  const [provider, setProvider] = useState("openai");
  const [providerModel, setProviderModel] = useState("gpt-4o-mini");
  const [apiBase, setApiBase] = useState("http://provider.example/v1");
  const [apiKey, setApiKey] = useState("sk-provider");
  const [rpm, setRPM] = useState("250");
  const [tpm, setTPM] = useState("150000");
  const [weight, setWeight] = useState("1");

  const [topUpAmount, setTopUpAmount] = useState("1500000");

  const tenantSlug = useMemo(() => {
    const slug = extractTenantSlugFromHost(props.initialHost);
    return slug || "default";
  }, [props.initialHost]);

  const loadDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const [walletData, keysData, modelsData] = await Promise.all([
        requestJSON<Wallet>("/api/wallet"),
        requestJSON<KeyListResponse>("/api/keys"),
        requestJSON<ModelListResponse>("/api/models"),
      ]);

      setWallet(walletData);
      setKeys(keysData.data ?? []);
      setModels(modelsData.data ?? []);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error),
      });
    } finally {
      setIsLoading(false);
    }
  }, []);

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
        message: "وارد کردن نام کلید و حداقل یک مدل الزامی است.",
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
          role: "virtual_key",
        }),
      });

      setLastCreatedKey(created.key ?? null);
      await loadDashboard();
      setBanner({
        kind: "success",
        message: "کلید API با موفقیت ساخته شد.",
      });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error),
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
        message: "تکمیل همه فیلدهای ضروری مدل الزامی است.",
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
        message: "مدل جدید برای Tenant ثبت شد.",
      });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error),
      });
    } finally {
      setIsCreatingModel(false);
    }
  }

  async function handleTopUp(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);

    const amount = Number(topUpAmount);
    if (!Number.isFinite(amount) || amount <= 0) {
      setBanner({
        kind: "error",
        message: "مبلغ شارژ باید عددی بیشتر از صفر باشد.",
      });
      return;
    }

    setIsTopUpLoading(true);
    try {
      await requestJSON("/api/wallet/topup", {
        method: "POST",
        body: JSON.stringify({
          amount_irr: amount,
          metadata: {
            source: "tenant-admin-ui",
          },
        }),
      });

      await loadDashboard();
      setBanner({
        kind: "success",
        message: "کیف پول با موفقیت شارژ شد.",
      });
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error),
      });
    } finally {
      setIsTopUpLoading(false);
    }
  }

  return (
    <main className="min-h-screen px-4 py-6 md:px-8 md:py-10">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
        <header className="rounded-3xl border border-[var(--border)] bg-[var(--surface)]/90 p-6 shadow-[0_30px_80px_rgba(3,10,25,0.55)] backdrop-blur">
          <p className="text-xs uppercase tracking-[0.35em] text-[var(--muted)]">
            پنل مدیریت تننت شاهراه
          </p>
          <div className="mt-3 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">
                {tenantSlug}.localhost
              </h1>
              <p className="mt-2 text-sm text-[var(--muted)]">
                محدوده Tenant از روی hostname تشخیص داده می‌شود.
              </p>
              <p className="mt-2 text-xs text-[var(--muted)]">
                کاربر جاری:{" "}
                <span className="font-mono">
                  {props.currentUser.email ??
                    props.currentUser.mobile ??
                    props.currentUser.id}
                </span>{" "}
                ({props.currentUser.role})
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => void loadDashboard()}
                disabled={isLoading}
                className="rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)]/10 px-4 py-2 text-sm font-medium text-[var(--accent)] transition hover:bg-[var(--accent)]/20 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isLoading ? "در حال بارگذاری..." : "به‌روزرسانی"}
              </button>
              <Link
                href="/settings"
                className="rounded-xl border border-[var(--border)] px-4 py-2 text-sm text-[var(--muted)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
              >
                تنظیمات
              </Link>
              <form action="/logout" method="post">
                <button
                  type="submit"
                  className="rounded-xl border border-[var(--danger)]/55 bg-[var(--danger)]/10 px-4 py-2 text-sm text-[var(--danger)] transition hover:bg-[var(--danger)]/20"
                >
                  خروج
                </button>
              </form>
            </div>
          </div>
        </header>

        {banner ? (
          <div
            className={`rounded-2xl border px-4 py-3 text-sm ${
              banner.kind === "success"
                ? "border-[var(--good)]/40 bg-[var(--good)]/10 text-[var(--good)]"
                : "border-[var(--danger)]/45 bg-[var(--danger)]/10 text-[var(--danger)]"
            }`}
          >
            {banner.message}
          </div>
        ) : null}

        {lastCreatedKey ? (
          <section className="rounded-2xl border border-amber-300/40 bg-amber-100/10 p-4 text-sm text-amber-200">
            <p className="font-medium">
              کلید ساخته‌شده فقط یک‌بار نمایش داده می‌شود:
            </p>
            <code className="mt-2 block overflow-x-auto rounded-lg bg-black/35 px-3 py-2 font-mono text-xs">
              {lastCreatedKey}
            </code>
          </section>
        ) : null}

        <section className="grid gap-4 md:grid-cols-3">
          <SummaryCard
            title="موجودی کیف پول"
            value={wallet ? formatIRR(wallet.balance_irr) : "-"}
            subtitle="اعتبار قابل استفاده"
          />
          <SummaryCard
            title="مبلغ معلق"
            value={wallet ? formatIRR(wallet.pending_irr) : "-"}
            subtitle="تراکنش‌های در انتظار"
          />
          <SummaryCard
            title="کلیدهای API"
            value={`${numberFormatter.format(keys.length)}`}
            subtitle="کلیدهای فعال Tenant"
          />
        </section>

        <section className="grid gap-6 xl:grid-cols-2">
          <article className="rounded-3xl border border-[var(--border)] bg-[var(--surface)]/85 p-5">
            <h2 className="text-xl font-semibold">مدیریت کلید API</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              مبتنی بر `POST /api/v1/keys` و `GET /api/v1/keys`
            </p>

            <form className="mt-4 grid gap-3" onSubmit={handleCreateKey}>
              <label className="text-sm text-[var(--muted)]">
                نام کلید
                <input
                  value={keyAlias}
                  onChange={(event) => setKeyAlias(event.target.value)}
                  className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--surface-2)] px-3 py-2 text-sm outline-none ring-0 focus:border-[var(--accent)]"
                  placeholder="tenant-dashboard-key"
                />
              </label>
              <label className="text-sm text-[var(--muted)]">
                مدل‌های مجاز (با ویرگول جدا کنید)
                <input
                  value={keyModelsInput}
                  onChange={(event) => setKeyModelsInput(event.target.value)}
                  className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--surface-2)] px-3 py-2 text-sm outline-none ring-0 focus:border-[var(--accent)]"
                  placeholder="gpt-4o-mini,gpt-4.1-mini"
                />
              </label>
              <button
                type="submit"
                disabled={isCreatingKey}
                className="mt-1 w-fit rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isCreatingKey ? "در حال ساخت..." : "ساخت کلید"}
              </button>
            </form>

            <div className="mt-5 overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)] text-right text-[var(--muted)]">
                    <th className="px-2 py-2 font-medium">نام</th>
                    <th className="px-2 py-2 font-medium">نقش</th>
                    <th className="px-2 py-2 font-medium">مدل‌ها</th>
                    <th className="px-2 py-2 font-medium">کلید</th>
                  </tr>
                </thead>
                <tbody>
                  {keys.length === 0 ? (
                    <tr>
                      <td className="px-2 py-4 text-[var(--muted)]" colSpan={4}>
                        هنوز کلیدی برای این Tenant ثبت نشده است.
                      </td>
                    </tr>
                  ) : (
                    keys.map((item) => (
                      <tr key={item.id} className="border-b border-[var(--border)]/60">
                        <td className="px-2 py-2">{item.key_alias}</td>
                        <td className="px-2 py-2">{item.role}</td>
                        <td className="px-2 py-2">
                          {(item.models ?? []).length > 0
                            ? item.models.join("، ")
                            : "-"}
                        </td>
                        <td className="px-2 py-2 font-mono text-xs">
                          {item.key_name ?? "پنهان"}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </article>

          <article className="rounded-3xl border border-[var(--border)] bg-[var(--surface)]/85 p-5">
            <h2 className="text-xl font-semibold">مدیریت مدل‌ها</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              مبتنی بر `POST /api/v1/models` و `GET /api/v1/models`
            </p>

            <form className="mt-4 grid gap-3 sm:grid-cols-2" onSubmit={handleCreateModel}>
              <Field label="نام نمایشی مدل">
                <input
                  value={modelName}
                  onChange={(event) => setModelName(event.target.value)}
                  className={inputClassName}
                  placeholder="gpt-4o-mini"
                />
              </Field>
              <Field label="پرووایدر">
                <input
                  value={provider}
                  onChange={(event) => setProvider(event.target.value)}
                  className={inputClassName}
                  placeholder="openai"
                />
              </Field>
              <Field label="نام مدل در پرووایدر">
                <input
                  value={providerModel}
                  onChange={(event) => setProviderModel(event.target.value)}
                  className={inputClassName}
                  placeholder="gpt-4o-mini"
                />
              </Field>
              <Field label="API Base">
                <input
                  value={apiBase}
                  onChange={(event) => setApiBase(event.target.value)}
                  className={inputClassName}
                  placeholder="http://provider.example/v1"
                />
              </Field>
              <Field label="API Key">
                <input
                  value={apiKey}
                  onChange={(event) => setApiKey(event.target.value)}
                  className={inputClassName}
                  placeholder="sk-provider"
                />
              </Field>
              <Field label="RPM">
                <input
                  value={rpm}
                  onChange={(event) => setRPM(event.target.value)}
                  className={inputClassName}
                  placeholder="250"
                />
              </Field>
              <Field label="TPM">
                <input
                  value={tpm}
                  onChange={(event) => setTPM(event.target.value)}
                  className={inputClassName}
                  placeholder="150000"
                />
              </Field>
              <Field label="وزن">
                <input
                  value={weight}
                  onChange={(event) => setWeight(event.target.value)}
                  className={inputClassName}
                  placeholder="1"
                />
              </Field>

              <div className="sm:col-span-2">
                <button
                  type="submit"
                  disabled={isCreatingModel}
                  className="rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isCreatingModel ? "در حال افزودن..." : "افزودن مدل"}
                </button>
              </div>
            </form>

            <div className="mt-5 overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)] text-right text-[var(--muted)]">
                    <th className="px-2 py-2 font-medium">مدل</th>
                    <th className="px-2 py-2 font-medium">پرووایدر</th>
                    <th className="px-2 py-2 font-medium">نام‌های جایگزین</th>
                  </tr>
                </thead>
                <tbody>
                  {models.length === 0 ? (
                    <tr>
                      <td className="px-2 py-4 text-[var(--muted)]" colSpan={3}>
                        هنوز مدلی برای این Tenant ثبت نشده است.
                      </td>
                    </tr>
                  ) : (
                    models.map((item) => (
                      <tr key={item.id} className="border-b border-[var(--border)]/60">
                        <td className="px-2 py-2">{item.id}</td>
                        <td className="px-2 py-2">{item.provider ?? "-"}</td>
                        <td className="px-2 py-2">
                          {(item.aliases ?? []).length > 0
                            ? (item.aliases ?? []).join("، ")
                            : "-"}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </article>
        </section>

        <section className="rounded-3xl border border-[var(--border)] bg-[var(--surface)]/85 p-5">
          <h2 className="text-xl font-semibold">شارژ کیف پول</h2>
          <p className="mt-1 text-sm text-[var(--muted)]">
            مبتنی بر `GET /api/v1/wallet` و `POST /api/v1/wallet/topup`
          </p>

          <form
            onSubmit={handleTopUp}
            className="mt-4 flex flex-col gap-3 md:flex-row md:items-end"
          >
            <label className="w-full max-w-xs text-sm text-[var(--muted)]">
              مبلغ (ریال)
              <input
                value={topUpAmount}
                onChange={(event) => setTopUpAmount(event.target.value)}
                className="mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--surface-2)] px-3 py-2 text-sm outline-none ring-0 focus:border-[var(--accent)]"
                placeholder="1500000"
              />
            </label>
            <button
              type="submit"
              disabled={isTopUpLoading}
              className="w-fit rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isTopUpLoading ? "در حال پردازش..." : "شارژ کیف پول"}
            </button>
          </form>
        </section>
      </div>
    </main>
  );
}

function SummaryCard(props: {
  title: string;
  value: string;
  subtitle: string;
}) {
  return (
    <article className="rounded-2xl border border-[var(--border)] bg-[var(--surface)]/90 p-4">
      <p className="text-xs tracking-[0.28em] text-[var(--muted)]">{props.title}</p>
      <p className="mt-3 text-2xl font-semibold">{props.value}</p>
      <p className="mt-2 text-sm text-[var(--muted)]">{props.subtitle}</p>
    </article>
  );
}

function Field(props: { label: string; children: ReactNode }) {
  return (
    <label className="text-sm text-[var(--muted)]">
      {props.label}
      <div className="mt-1">{props.children}</div>
    </label>
  );
}

const inputClassName =
  "w-full rounded-xl border border-[var(--border)] bg-[var(--surface-2)] px-3 py-2 text-sm outline-none ring-0 focus:border-[var(--accent)]";

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

function normalizeError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "درخواست با خطا مواجه شد.";
}

function toNumberOrUndefined(value: string): number | undefined {
  const raw = value.trim();
  if (!raw) {
    return undefined;
  }
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function formatIRR(value: number): string {
  return `${numberFormatter.format(Math.round(value))} ریال`;
}
