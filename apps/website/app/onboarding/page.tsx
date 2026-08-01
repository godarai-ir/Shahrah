"use client";

import { useEffect, useMemo, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import Link from "next/link";
import {
  AlertCircle,
  ArrowRight,
  Building2,
  Check,
  CheckCircle,
  Clock,
  Github,
  Loader2,
  LockKeyhole,
  Mail,
  Send,
  ShieldCheck,
  Smartphone,
} from "lucide-react";

import { ThemeProvider } from "@/app/components/ThemeProvider";
import { GodarAILogo } from "@/app/components/GodarAILogo";
import { mobileAuthUIEnabled } from "@/app/lib/mobile-auth-ui";
import { onboardingMessages as t } from "./messages";

type Step = "details" | "otp" | "done";

type OnboardingPayload = {
  available?: boolean;
  reason?: string;
  request_id?: string;
  tenant_slug?: string;
  channel?: "email" | "mobile";
  destination_masked?: string;
  expires_in_seconds?: number;
  already_pending?: boolean;
  debug_otp?: string;
  login_url?: string;
  error?: string;
};

type SocialUser = { id: string; name: string; email: string; email_verified: boolean };

export default function OnboardingPage() {
  const [step, setStep] = useState<Step>("details");
  const [tenantSlug, setTenantSlug] = useState("");
  const [email, setEmail] = useState("");
  const [mobile, setMobile] = useState("");
  const [otpCode, setOtpCode] = useState("");
  const [requestID, setRequestID] = useState("");
  const [loginURL, setLoginURL] = useState("");
  const [destination, setDestination] = useState("");
  const [debugOTP, setDebugOTP] = useState("");
  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [socialUser, setSocialUser] = useState<SocialUser | null>(null);

  const isExpired = step === "otp" && remainingSeconds <= 0;
  const formattedRemaining = useMemo(() => formatRemaining(remainingSeconds), [remainingSeconds]);

  useEffect(() => {
    if (step !== "otp" || remainingSeconds <= 0) {
      return;
    }
    const timer = window.setInterval(() => {
      setRemainingSeconds((value) => Math.max(0, value - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [remainingSeconds, step]);

  useEffect(() => {
    if (new URLSearchParams(window.location.search).get("social") !== "1") return;
    void fetch("/api/social-auth/session", { cache: "no-store" })
      .then((response) => response.json())
      .then((payload: { user?: SocialUser | null }) => {
        if (payload.user?.email_verified) {
          setSocialUser(payload.user);
          setEmail(payload.user.email);
        } else {
          setError("ایمیل حساب اجتماعی تایید نشده است.");
        }
      })
      .catch(() => setError("اطلاعات حساب اجتماعی دریافت نشد."));
  }, []);

  async function provisionSocial(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      const result = await postOnboarding({
        action: "provision_social",
        tenant_slug: tenantSlug,
        plan_slug: requestedPlanSlug(),
      });
      if (!result.ok) {
        setError(mapOnboardingError(result.status, result.payload, t.errors.requestFailed));
        return;
      }
      setLoginURL(String(result.payload.login_url ?? ""));
      setStep("done");
    } finally {
      setLoading(false);
    }
  }

  async function requestOTP(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    setError("");
    setNotice("");
    setLoading(true);
    try {
      const availability = await postOnboarding({
        action: "check_slug",
        tenant_slug: tenantSlug,
      });
      if (!availability.ok) {
        setError(normalizeError(availability.payload, t.errors.requestFailed));
        return;
      }
      if (!availability.payload.available && availability.payload.reason === "invalid_or_reserved") {
        setError(t.errors.slugInvalid);
        return;
      }

      const otp = await postOnboarding({
        action: "request_otp",
        tenant_slug: tenantSlug,
        plan_slug: requestedPlanSlug(),
        email,
        mobile: mobileAuthUIEnabled ? mobile : undefined,
        metadata: { source: "website-onboarding" },
      });
      if (!otp.ok) {
        setError(mapOnboardingError(otp.status, otp.payload, t.errors.sendFailed));
        return;
      }

      setRequestID(String(otp.payload.request_id ?? ""));
      setDestination(String(otp.payload.destination_masked ?? ""));
      setRemainingSeconds(Number(otp.payload.expires_in_seconds ?? 300));
      setDebugOTP(String(otp.payload.debug_otp ?? ""));
      setOtpCode("");
      setNotice(
        otp.payload.already_pending
          ? t.alreadyPendingDescription
          : t.pendingDescription,
      );
      setStep("otp");
    } finally {
      setLoading(false);
    }
  }

  async function verifyOTP(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isExpired) {
      setError(t.errors.expired);
      return;
    }

    setError("");
    setLoading(true);
    try {
      const result = await postOnboarding({
        action: "verify_otp",
        request_id: requestID,
        otp_code: otpCode,
      });
      if (!result.ok) {
        setError(mapOnboardingError(result.status, result.payload, t.errors.verifyFailed));
        return;
      }
      setLoginURL(String(result.payload.login_url ?? ""));
      setStep("done");
      setNotice("");
    } finally {
      setLoading(false);
    }
  }

  return (
    <ThemeProvider defaultTheme="dark" storageKey="shahrah-theme">
    <main className="relative min-h-screen overflow-hidden bg-slate-50 text-slate-950 transition-colors dark:bg-[#0B0F19] dark:text-white">
      <div className="pointer-events-none absolute inset-0" aria-hidden="true">
        <div className="absolute -right-32 -top-32 size-96 rounded-full bg-indigo-500/10 blur-3xl dark:bg-indigo-500/15" />
        <div className="absolute -bottom-40 -left-32 size-[32rem] rounded-full bg-purple-500/10 blur-3xl dark:bg-purple-500/10" />
        <div className="absolute inset-0 opacity-[0.035] dark:opacity-[0.06] [background-image:radial-gradient(circle_at_1px_1px,currentColor_1px,transparent_0)] [background-size:28px_28px]" />
      </div>

      <div className="relative mx-auto flex min-h-screen w-full max-w-7xl flex-col px-4 py-5 sm:px-6 sm:py-7 lg:px-8">
        <header className="flex items-center justify-between gap-4">
          <Link href="/" className="flex items-center gap-3 transition-opacity hover:opacity-80">
            <GodarAILogo
              decorative
              priority
              className="h-10 w-auto max-w-14 object-contain"
              sizes="3.5rem"
            />
            <span className="text-lg font-bold tracking-tight sm:text-xl">گدارAI</span>
          </Link>
          <Link
            href="/pricing"
            className="inline-flex min-h-11 items-center gap-2 rounded-xl border border-slate-200 bg-white/80 px-4 text-sm font-semibold text-slate-700 shadow-sm backdrop-blur transition hover:border-indigo-300 hover:text-indigo-600 dark:border-slate-800 dark:bg-slate-900/70 dark:text-slate-200 dark:hover:border-indigo-500/50 dark:hover:text-indigo-300"
          >
            <ArrowRight aria-hidden="true" />
            <span className="hidden sm:inline">{t.backToPlans}</span>
            <span className="sm:hidden">بازگشت</span>
          </Link>
        </header>

        <div className="flex flex-1 items-center py-8 sm:py-10">
          <section className="grid w-full gap-6 lg:grid-cols-[18rem_minmax(0,1fr)] lg:gap-8">
            <aside className="order-2 overflow-hidden rounded-3xl border border-slate-200 bg-white/80 p-5 shadow-xl shadow-slate-900/5 backdrop-blur dark:border-slate-800 dark:bg-slate-900/70 dark:shadow-black/20 lg:order-1 lg:p-6">
              <div className="mb-6 hidden lg:block">
                <span className="mb-4 flex size-11 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-300">
                  <ShieldCheck aria-hidden="true" />
                </span>
                <h2 className="font-bold">راه‌اندازی فضای کاری</h2>
                <p className="mt-2 text-sm leading-6 text-slate-500 dark:text-slate-400">
                  تنها سه گام تا ورود به پنل مدیریت باقی مانده است.
                </p>
              </div>

              <ol className="grid grid-cols-3 gap-2 lg:flex lg:flex-col lg:gap-1">
                <StepItem active={step === "details"} done={step !== "details"} label={t.detailsTitle} number="۱" />
                <StepItem active={step === "otp"} done={step === "done"} label={t.otpTitle} number="۲" />
                <StepItem active={step === "done"} done={step === "done"} label={t.doneTitle} number="۳" />
              </ol>

              <div className="mt-6 hidden rounded-2xl border border-indigo-100 bg-indigo-50/70 p-4 text-sm leading-6 text-slate-600 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-slate-300 lg:block">
                <div className="mb-2 flex items-center gap-2 font-semibold text-indigo-700 dark:text-indigo-300">
                  <LockKeyhole aria-hidden="true" />
                  فعال‌سازی امن
                </div>
                اطلاعات تماس شما فقط برای تایید و مدیریت فضای کاری استفاده می‌شود.
              </div>
            </aside>

            <div className="order-1 rounded-3xl border border-slate-200 bg-white p-5 shadow-xl shadow-slate-900/5 dark:border-slate-800 dark:bg-slate-900 dark:shadow-black/20 sm:p-8 lg:order-2 lg:p-10">
              <div className="mb-8 flex items-start justify-between gap-4">
                <div className="max-w-2xl">
                  <p className="mb-2 text-sm font-semibold text-indigo-600 dark:text-indigo-400">شروع کار با گدارAI</p>
                  <h1 className="text-2xl font-bold leading-10 tracking-tight sm:text-3xl">{t.title}</h1>
                  <p className="mt-2 text-sm leading-7 text-slate-500 dark:text-slate-400 sm:text-base">{t.subtitle}</p>
                </div>
                <span className="hidden size-12 shrink-0 items-center justify-center rounded-2xl bg-indigo-600 text-white shadow-lg shadow-indigo-500/25 sm:flex">
                  <Building2 aria-hidden="true" />
                </span>
              </div>

              {step === "otp" && (
                <div className="mb-5 flex items-center justify-between gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm dark:border-slate-800 dark:bg-[#0B0F19]">
                  <span className="text-slate-500 dark:text-slate-400">{t.expiresIn}</span>
                  <span className="font-mono font-semibold text-indigo-600 dark:text-indigo-300" dir="ltr">{formattedRemaining}</span>
                </div>
              )}

              {notice && !isExpired && <StatusMessage variant="info" icon={<Clock />}>{notice}</StatusMessage>}
              {isExpired && (
                <StatusMessage variant="warning" icon={<AlertCircle />}>
                  <strong className="mb-1 block">{t.expiredTitle}</strong>
                  {t.expiredDescription}
                </StatusMessage>
              )}
              {error && <StatusMessage variant="error" icon={<AlertCircle />}>{error}</StatusMessage>}

              {step === "details" && (
                <form className="flex flex-col gap-6" onSubmit={socialUser ? provisionSocial : requestOTP}>
                  {!socialUser ? <SocialAuthButtons /> : (
                    <StatusMessage variant="info" icon={<ShieldCheck />}>
                      ورود امن با حساب <strong dir="ltr">{socialUser.email}</strong> انجام شد. شناسه فضای کاری را انتخاب کنید.
                    </StatusMessage>
                  )}
                  <Field label={t.tenantSlug} hint={t.tenantSlugHint} icon={<Building2 />}>
                    <input
                      id="tenantSlug"
                      value={tenantSlug}
                      onChange={(event) => setTenantSlug(event.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))}
                      placeholder={t.tenantSlugPlaceholder}
                      className={inputClass}
                      dir="ltr"
                      autoComplete="organization"
                      required
                    />
                  </Field>

                  {!socialUser ? <div className={mobileAuthUIEnabled ? "grid gap-5 sm:grid-cols-2" : "grid gap-5"}>
                    <Field label={t.email} icon={<Mail />}>
                      <input id="email" value={email} onChange={(event) => setEmail(event.target.value.trim().toLowerCase())} placeholder={t.emailPlaceholder} className={inputClass} dir="ltr" type="email" autoComplete="email" required={!mobileAuthUIEnabled} />
                    </Field>
                    {mobileAuthUIEnabled ? (
                      <Field label={t.mobile} icon={<Smartphone />}>
                        <input id="mobile" value={mobile} onChange={(event) => setMobile(event.target.value.replace(/[^\d+]/g, ""))} placeholder={t.mobilePlaceholder} className={inputClass} dir="ltr" inputMode="tel" autoComplete="tel" />
                      </Field>
                    ) : null}
                  </div> : null}
                  {!socialUser ? <p className="-mt-3 text-xs leading-5 text-slate-500 dark:text-slate-400">
                    {mobileAuthUIEnabled ? t.contactHint : t.emailHint}
                  </p> : null}

                  <button type="submit" disabled={loading || !tenantSlug || (!socialUser && (mobileAuthUIEnabled ? !email && !mobile : !email))} className={primaryButtonClass}>
                    {loading ? <Loader2 className="animate-spin" aria-hidden="true" /> : <Send aria-hidden="true" />}
                    {loading ? (socialUser ? "در حال ساخت فضای کاری…" : t.sendingCode) : (socialUser ? "ساخت فضای کاری" : t.sendCode)}
                  </button>
                </form>
              )}

              {step === "otp" && (
                <form className="flex flex-col gap-6" onSubmit={verifyOTP}>
                  <div className="rounded-2xl border border-indigo-100 bg-indigo-50/60 p-4 dark:border-indigo-500/20 dark:bg-indigo-500/10">
                    <h2 className="text-base font-semibold">{t.pendingTitle}</h2>
                    <p className="mt-2 text-sm text-slate-600 dark:text-slate-300" dir="ltr">{destination}</p>
                  </div>

                  {debugOTP && <StatusMessage variant="warning">{t.devCode}: <code className="rounded bg-white/70 px-2 py-1 font-mono text-xs dark:bg-black/20" dir="ltr">{debugOTP}</code></StatusMessage>}

                  <Field label={t.code} icon={<LockKeyhole />}>
                    <input id="otpCode" value={otpCode} onChange={(event) => setOtpCode(event.target.value.replace(/\D/g, ""))} placeholder={t.codePlaceholder} maxLength={6} className={`${inputClass} text-center font-mono text-lg tracking-[0.45em]`} dir="ltr" inputMode="numeric" autoComplete="one-time-code" required autoFocus />
                  </Field>

                  <div className="flex flex-col-reverse gap-3 sm:flex-row">
                    <button type="button" onClick={() => void requestOTP()} disabled={loading || !isExpired} className={secondaryButtonClass}>{t.requestAgain}</button>
                    <button type="submit" disabled={loading || otpCode.length !== 6 || isExpired} className={`${primaryButtonClass} flex-1`}>
                      {loading && <Loader2 className="animate-spin" aria-hidden="true" />}
                      {loading ? t.verifying : t.verify}
                    </button>
                  </div>
                </form>
              )}

              {step === "done" && (
                <section className="rounded-2xl border border-emerald-200 bg-emerald-50 p-6 text-emerald-950 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-100">
                  <span className="mb-5 flex size-12 items-center justify-center rounded-full bg-emerald-600 text-white shadow-lg shadow-emerald-500/20"><CheckCircle aria-hidden="true" /></span>
                  <h2 className="text-xl font-bold">{t.doneTitle}</h2>
                  <p className="mt-2 text-sm leading-7">{t.successDescription}</p>
                  {loginURL && (
                    <>
                      <p className="mt-5 text-xs font-semibold">{t.loginUrl}</p>
                      <p className="mt-1 break-all font-mono text-sm" dir="ltr">{loginURL}</p>
                      <a href={loginURL} className="mt-6 inline-flex min-h-12 w-full items-center justify-center rounded-xl bg-emerald-700 px-5 text-sm font-semibold text-white transition hover:bg-emerald-800 sm:w-auto">{t.openPanel}</a>
                    </>
                  )}
                </section>
              )}
            </div>
          </section>
        </div>
      </div>
    </main>
    </ThemeProvider>
  );
}

const inputClass = "min-h-12 w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-start text-sm outline-none transition placeholder:text-slate-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-slate-800 dark:bg-[#0B0F19] dark:placeholder:text-slate-600";
const primaryButtonClass = "inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-indigo-600 px-5 text-sm font-semibold text-white shadow-lg shadow-indigo-500/20 transition hover:bg-indigo-700 disabled:pointer-events-none disabled:opacity-50";
const secondaryButtonClass = "inline-flex min-h-12 items-center justify-center rounded-xl border border-slate-200 bg-white px-5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 disabled:pointer-events-none disabled:opacity-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800";

function SocialAuthButtons() {
  const plan = typeof window === "undefined" ? "developer" : requestedPlanSlug();
  const providers = [
    { id: "google", label: "Google", icon: <span className="text-base font-bold">G</span> },
    { id: "github", label: "GitHub", icon: <Github className="size-5" aria-hidden="true" /> },
    { id: "microsoft", label: "Microsoft", icon: <span className="grid grid-cols-2 gap-0.5" aria-hidden="true">{[0,1,2,3].map((item) => <i key={item} className="size-1.5 bg-current" />)}</span> },
  ];
  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-3 sm:grid-cols-3">
        {providers.map((provider) => {
          const query = new URLSearchParams({
            flow: "onboarding",
            provider: provider.id,
            plan,
          });
          return <a key={provider.id} href={`/auth/social/start?${query.toString()}`} className={`${secondaryButtonClass} gap-2`} dir="ltr">{provider.icon}{provider.label}</a>;
        })}
      </div>
      <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400" aria-hidden="true">
        <span className="h-px flex-1 bg-slate-200 dark:bg-slate-800" /><span>یا ادامه با ایمیل</span><span className="h-px flex-1 bg-slate-200 dark:bg-slate-800" />
      </div>
    </div>
  );
}

function Field({ children, hint, icon, label }: { children: ReactNode; hint?: string; icon: ReactNode; label: string }) {
  const inputId = label === t.tenantSlug ? "tenantSlug" : label === t.email ? "email" : label === t.mobile ? "mobile" : "otpCode";
  return (
    <div className="flex flex-col gap-2">
      <label className="flex items-center gap-2 text-sm font-semibold" htmlFor={inputId}>
        <span className="text-indigo-500 [&>svg]:size-4">{icon}</span>
        {label}
      </label>
      {children}
      {hint && <p className="text-xs leading-5 text-slate-500 dark:text-slate-400">{hint}</p>}
    </div>
  );
}

function StepItem({ active, done, label, number }: { active: boolean; done: boolean; label: string; number: string }) {
  return (
    <li className={`flex min-w-0 items-center gap-2 rounded-xl p-2.5 transition lg:gap-3 lg:p-3 ${active ? "bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-300" : "text-slate-500 dark:text-slate-400"}`}>
      <span
        className={`flex size-8 shrink-0 items-center justify-center rounded-lg text-xs font-bold ${done ? "bg-emerald-500 text-white" : active ? "bg-indigo-600 text-white" : "bg-slate-100 dark:bg-slate-800"}`}
      >
        {done ? <Check aria-hidden="true" /> : number}
      </span>
      <span className={`truncate text-xs lg:text-sm ${active ? "font-semibold" : ""}`}>{label}</span>
    </li>
  );
}

function StatusMessage({
  children,
  icon,
  variant,
}: {
  children: ReactNode;
  icon?: ReactNode;
  variant: "info" | "warning" | "error";
}) {
  const styles = {
    info: "border-indigo-200 bg-indigo-50 text-indigo-950 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-indigo-100",
    warning: "border-amber-200 bg-amber-50 text-amber-950 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-100",
    error: "border-red-200 bg-red-50 text-red-950 dark:border-red-500/20 dark:bg-red-500/10 dark:text-red-100",
  };

  return (
    <div className={`mb-5 flex items-start gap-3 rounded-xl border p-4 text-sm leading-6 ${styles[variant]}`} role={variant === "error" ? "alert" : "status"}>
      {icon && <span className="mt-1 shrink-0 [&>svg]:size-4">{icon}</span>}
      <div>{children}</div>
    </div>
  );
}

async function postOnboarding(body: Record<string, unknown>) {
  const response = await fetch("/api/onboarding", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  return {
    ok: response.ok,
    status: response.status,
    payload: (await response.json()) as OnboardingPayload,
  };
}

function normalizeError(payload: OnboardingPayload, fallback: string) {
  if (!mobileAuthUIEnabled && payload.error && /mobile|phone|موبایل|همراه/i.test(payload.error)) {
    return t.errors.emailRequired;
  }
  return payload.error || fallback;
}

function mapOnboardingError(status: number, payload: OnboardingPayload, fallback: string) {
  if (status === 409) {
    return t.errors.slugTaken;
  }
  if (status === 410) {
    return t.errors.expired;
  }
  return normalizeError(payload, fallback);
}

function requestedPlanSlug() {
  const requested = new URLSearchParams(window.location.search).get("plan")?.trim().toLowerCase();
  return requested && ["developer", "team", "growth", "enterprise"].includes(requested) ? requested : "developer";
}

function formatRemaining(totalSeconds: number) {
  const safeSeconds = Math.max(0, totalSeconds);
  const minutes = Math.floor(safeSeconds / 60);
  const seconds = safeSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}
