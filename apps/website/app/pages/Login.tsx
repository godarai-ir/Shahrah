"use client";

import Link from "next/link";
import { FormEvent, useMemo, useState } from "react";
import { ArrowLeft, Building2, Loader2, Mail, Search, UserRound } from "lucide-react";

import { mobileAuthUIEnabled } from "@/app/lib/mobile-auth-ui";

type LoginMethod = "identifier" | "tenant";

type Tenant = {
  id: string;
  name?: string;
  slug: string;
  domain?: string;
  loginUrl: string;
};

type IdentifierResponse = {
  status: "none" | "single" | "multiple";
  tenants: Tenant[];
};

type TenantResponse = {
  exists: boolean;
  tenant?: Tenant;
};

const tenantNamePattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function Login({ tenantBaseDomain = "" }: { tenantBaseDomain?: string }) {
  const [method, setMethod] = useState<LoginMethod>("identifier");
  const [identifier, setIdentifier] = useState("");
  const [tenantName, setTenantName] = useState("");
  const [identifierError, setIdentifierError] = useState("");
  const [tenantError, setTenantError] = useState("");
  const [message, setMessage] = useState("");
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(false);

  const domainSuffix = useMemo(() => {
    if (typeof window === "undefined") {
      return ".godarai.ir";
    }
    const host = tenantBaseDomain || window.location.host;
    return `.${normalizeHostForDisplay(host)}`;
  }, [tenantBaseDomain]);

  async function handleIdentifierSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalized = normalizeIdentifier(identifier);
    setIdentifierError("");
    setTenantError("");
    setMessage("");
    setTenants([]);

    if (!normalized) {
      setIdentifierError(
        mobileAuthUIEnabled
          ? "لطفاً یک ایمیل یا شماره موبایل معتبر وارد کنید."
          : "لطفاً یک ایمیل معتبر وارد کنید.",
      );
      return;
    }

    setLoading(true);
    try {
      const result = await requestJSON<IdentifierResponse>("/api/login-discovery", {
        action: "identifier",
        identifier: normalized,
      });
      if (result.status === "single" && result.tenants[0]) {
        redirectToTenant(result.tenants[0], normalized);
        return;
      }
      if (result.status === "multiple") {
        setTenants(result.tenants);
        setMessage("فضایی را که می‌خواهید وارد آن شوید انتخاب کنید.");
        return;
      }
      setMessage(
        mobileAuthUIEnabled
          ? "هیچ فضای کاری مرتبط با این ایمیل یا شماره موبایل پیدا نشد. می‌توانید یک فضای کاری جدید بسازید یا از مدیر تیم خود دعوت‌نامه بگیرید."
          : "هیچ فضای کاری مرتبط با این ایمیل پیدا نشد. می‌توانید یک فضای کاری جدید بسازید یا از مدیر تیم خود دعوت‌نامه بگیرید.",
      );
    } catch (error) {
      setIdentifierError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function handleTenantSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalized = tenantName.trim().toLowerCase();
    setIdentifierError("");
    setTenantError("");
    setMessage("");
    setTenants([]);

    if (!tenantNamePattern.test(normalized)) {
      setTenantError("نام فضای کاری فقط می‌تواند شامل حروف کوچک انگلیسی، عدد و خط تیره باشد.");
      return;
    }

    setLoading(true);
    try {
      const result = await requestJSON<TenantResponse>("/api/login-discovery", {
        action: "tenant",
        tenantName: normalized,
      });
      if (result.exists && result.tenant) {
        redirectToTenant(result.tenant);
        return;
      }
      setTenantError("فضای کاری با این نام پیدا نشد. نام را بررسی کنید و دوباره تلاش کنید.");
    } catch (error) {
      setTenantError(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="bg-slate-50 px-4 py-10 text-slate-950 dark:bg-[#0B0F19] dark:text-white sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-8">
        <header className="max-w-2xl">
          <p className="mb-3 text-sm font-semibold text-indigo-600 dark:text-indigo-400">ورود</p>
          <h1 className="text-3xl font-bold tracking-normal sm:text-4xl">فضای کاری خود را پیدا کنید</h1>
          <p className="mt-3 text-base leading-7 text-slate-600 dark:text-slate-300">
            {mobileAuthUIEnabled
              ? "ورود از طریق فضای کاری سازمان شما انجام می‌شود. اگر نام فضای کاری را نمی‌دانید، ایمیل یا شماره موبایل خود را وارد کنید."
              : "ورود از طریق فضای کاری سازمان شما انجام می‌شود. اگر نام فضای کاری را نمی‌دانید، ایمیل خود را وارد کنید."}
          </p>
        </header>

        <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
          <div className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-gray-800 dark:bg-gray-900 sm:p-6">
            <div className="mb-6 grid grid-cols-2 rounded-xl border border-gray-200 bg-slate-100 p-1 text-sm font-medium dark:border-gray-800 dark:bg-[#0B0F19]">
              <button
                type="button"
                onClick={() => setMethod("identifier")}
                className={tabClass(method === "identifier")}
                aria-pressed={method === "identifier"}
              >
                <Mail className="h-4 w-4" aria-hidden="true" />
                {mobileAuthUIEnabled ? "ایمیل یا موبایل" : "ایمیل"}
              </button>
              <button
                type="button"
                onClick={() => setMethod("tenant")}
                className={tabClass(method === "tenant")}
                aria-pressed={method === "tenant"}
              >
                <Building2 className="h-4 w-4" aria-hidden="true" />
                نام فضای کاری
              </button>
            </div>

            {method === "identifier" ? (
              <form className="flex flex-col gap-5" onSubmit={handleIdentifierSubmit}>
                <div className="flex flex-col gap-2">
                  <label htmlFor="identifier" className="text-sm font-medium">
                    {mobileAuthUIEnabled ? "ایمیل یا شماره موبایل" : "ایمیل"}
                  </label>
                  <input
                    id="identifier"
                    value={identifier}
                    onChange={(event) => setIdentifier(event.target.value)}
                    placeholder={mobileAuthUIEnabled ? "name@example.com یا +98912..." : "name@example.com"}
                    autoComplete="username"
                    inputMode="email"
                    dir="ltr"
                    aria-invalid={Boolean(identifierError)}
                    aria-describedby={identifierError ? "identifier-error" : undefined}
                    className="min-h-12 rounded-lg border border-gray-200 bg-white px-4 py-3 text-start outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-gray-800 dark:bg-[#0B0F19]"
                  />
                  {identifierError ? (
                    <p id="identifier-error" className="text-sm text-red-600 dark:text-red-400">
                      {identifierError}
                    </p>
                  ) : null}
                </div>
                <SubmitButton loading={loading}>ادامه</SubmitButton>
              </form>
            ) : (
              <form className="flex flex-col gap-5" onSubmit={handleTenantSubmit}>
                <div className="flex flex-col gap-2">
                  <label htmlFor="tenantName" className="text-sm font-medium">
                    نام فضای کاری
                  </label>
                  <div
                    className="flex min-h-12 overflow-hidden rounded-lg border border-gray-200 bg-white focus-within:border-indigo-500 focus-within:ring-2 focus-within:ring-indigo-500/20 dark:border-gray-800 dark:bg-[#0B0F19]"
                    dir="ltr"
                  >
                    <input
                      id="tenantName"
                      value={tenantName}
                      onChange={(event) => setTenantName(event.target.value.toLowerCase())}
                      placeholder="acme"
                      autoComplete="organization"
                      aria-invalid={Boolean(tenantError)}
                      aria-describedby={tenantError ? "tenant-error" : "tenant-domain"}
                      className="min-w-0 flex-1 bg-transparent px-4 py-3 text-start outline-none"
                      dir="ltr"
                    />
                    <span
                      id="tenant-domain"
                      className="flex shrink-0 items-center border-s border-gray-200 bg-slate-50 px-3 text-sm text-slate-500 dark:border-gray-800 dark:bg-gray-900 dark:text-slate-400"
                      dir="ltr"
                    >
                      {domainSuffix}
                    </span>
                  </div>
                  {tenantError ? (
                    <p id="tenant-error" className="text-sm text-red-600 dark:text-red-400">
                      {tenantError}
                    </p>
                  ) : null}
                </div>
                <SubmitButton loading={loading}>باز کردن فضای کاری</SubmitButton>
              </form>
            )}

            {message ? (
              <div className="mt-6 rounded-lg border border-gray-200 bg-slate-50 p-4 text-sm leading-6 text-slate-700 dark:border-gray-800 dark:bg-[#0B0F19] dark:text-slate-300">
                <p>{message}</p>
                {tenants.length === 0 ? (
                  <Link
                    href="/onboarding"
                    className="mt-3 inline-flex items-center gap-2 font-medium text-indigo-600 dark:text-indigo-400"
                  >
                    شروع راه‌اندازی
                    <ArrowLeft className="h-4 w-4" aria-hidden="true" />
                  </Link>
                ) : null}
              </div>
            ) : null}

            {tenants.length > 0 ? (
              <div className="mt-5 flex flex-col gap-3">
                {tenants.map((tenant) => (
                  <button
                    key={tenant.id || tenant.slug}
                    type="button"
                    onClick={() => redirectToTenant(tenant, normalizeIdentifier(identifier))}
                    className="flex w-full items-center gap-3 rounded-lg border border-gray-200 bg-white p-4 text-start transition hover:border-indigo-300 hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-500/30 dark:border-gray-800 dark:bg-gray-900 dark:hover:border-indigo-500/60 dark:hover:bg-indigo-500/10"
                  >
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-300">
                      <UserRound className="h-5 w-5" aria-hidden="true" />
                    </span>
                    <span className="min-w-0">
                      <span className="block font-medium">{tenant.name || tenant.slug}</span>
                      <span className="block truncate text-sm text-slate-500 dark:text-slate-400" dir="ltr">
                        {tenant.domain || tenant.slug}
                      </span>
                    </span>
                  </button>
                ))}
              </div>
            ) : null}
          </div>

          <aside className="rounded-2xl border border-gray-200 bg-white p-5 text-sm leading-6 text-slate-600 shadow-sm dark:border-gray-800 dark:bg-gray-900 dark:text-slate-300">
            <Search className="mb-4 h-6 w-6 text-indigo-500" aria-hidden="true" />
            <h2 className="mb-2 text-base font-semibold text-slate-950 dark:text-white">نحوه ورود</h2>
            <p>
              {mobileAuthUIEnabled
                ? "اگر ایمیل یا شماره شما فقط به یک فضای کاری وصل باشد، مستقیم به صفحه ورود همان فضا می‌روید. اگر به چند فضای کاری دسترسی داشته باشید، ابتدا فضای موردنظر را انتخاب می‌کنید."
                : "اگر ایمیل شما فقط به یک فضای کاری وصل باشد، مستقیم به صفحه ورود همان فضا می‌روید. اگر به چند فضای کاری دسترسی داشته باشید، ابتدا فضای موردنظر را انتخاب می‌کنید."}
            </p>
          </aside>
        </section>
      </div>
    </main>
  );
}

function SubmitButton({ loading, children }: { loading: boolean; children: string }) {
  return (
    <button
      type="submit"
      disabled={loading}
      className="inline-flex min-h-12 items-center justify-center gap-2 rounded-lg bg-indigo-500 px-4 py-3 text-sm font-medium text-white shadow-[0_0_15px_rgba(99,102,241,0.3)] transition hover:bg-indigo-600 disabled:cursor-not-allowed disabled:opacity-70"
    >
      {loading ? <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" /> : null}
      {children}
    </button>
  );
}

function tabClass(active: boolean) {
  return [
    "flex min-h-10 items-center justify-center gap-2 rounded-lg px-3 py-2 transition",
    active ? "bg-white text-slate-950 shadow-sm dark:bg-gray-900 dark:text-white" : "text-slate-600 dark:text-slate-400",
  ].join(" ");
}

async function requestJSON<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  const payload = (await response.json().catch(() => ({}))) as T & { error?: string };
  if (!response.ok) {
    throw new Error(translateLoginError(payload.error));
  }
  return payload;
}

function normalizeIdentifier(value: string) {
  const trimmed = value.trim();
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed.toLowerCase())) {
    return trimmed.toLowerCase();
  }
  const mobile = trimmed.replace(/[\s-]/g, "").replace(/^00/, "+");
  if (/^\+?[0-9]{10,15}$/.test(mobile)) {
    return mobile;
  }
  return "";
}

function redirectToTenant(tenant: Tenant, identifier?: string) {
  const url = new URL(tenant.loginUrl);
  if (identifier) {
    url.searchParams.set("identifier", encodeBase64URL(identifier));
  }
  window.location.assign(url.toString());
}

function normalizeHostForDisplay(host: string) {
  return host
    .trim()
    .toLowerCase()
    .replace(/^https?:\/\//, "")
    .split("/")[0]
    ?.replace(/^www\./, "")
    .replace(/\.$/, "") ?? "";
}

function encodeBase64URL(value: string) {
  const bytes = new TextEncoder().encode(value);
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return window
    .btoa(binary)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/g, "");
}

function getErrorMessage(error: unknown) {
  return error instanceof Error ? translateLoginError(error.message) : defaultLoginError;
}

const defaultLoginError = "در حال حاضر امکان بررسی فضاهای کاری وجود ندارد. لطفاً دوباره تلاش کنید.";

function translateLoginError(message?: string) {
  switch (message) {
    case "Invalid login discovery request.":
      return "درخواست ورود معتبر نیست.";
    case "Please enter a valid email address or mobile number.":
      return mobileAuthUIEnabled
        ? "لطفاً یک ایمیل یا شماره موبایل معتبر وارد کنید."
        : "لطفاً یک ایمیل معتبر وارد کنید.";
    case "Tenant name can only include lowercase letters, numbers, and hyphens.":
      return "نام فضای کاری فقط می‌تواند شامل حروف کوچک انگلیسی، عدد و خط تیره باشد.";
    case "We could not check workspaces right now. Please try again.":
    case undefined:
    case "":
      return defaultLoginError;
    default:
      return message;
  }
}
