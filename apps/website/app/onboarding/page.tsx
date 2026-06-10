"use client";

import { useEffect, useMemo, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import Link from "next/link";
import { AlertCircle, CheckCircle, Clock, Loader2, Send, ShieldCheck } from "lucide-react";

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
        plan_slug: "free",
        email,
        mobile,
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
    <main className="min-h-screen bg-background px-4 py-8 text-foreground sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <header className="flex items-center justify-between gap-4">
          <Link
            href="/pricing"
            className="rounded-lg border border-border bg-card px-3 py-2 text-sm font-medium text-foreground transition hover:bg-accent"
          >
            {t.backToPlans}
          </Link>
          <div className="rounded-full border border-border bg-card px-3 py-1 text-xs font-medium text-muted-foreground">
            {t.expiresIn}: {t.fiveMinutes}
          </div>
        </header>

        <section className="grid gap-6 lg:grid-cols-[1fr_360px]">
          <div className="rounded-lg border border-border bg-card p-5 shadow-sm sm:p-6">
            <div className="mb-6 flex items-start gap-3">
              <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <ShieldCheck className="h-5 w-5" aria-hidden="true" />
              </span>
              <div>
                <h1 className="text-2xl font-bold leading-9">{t.title}</h1>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">{t.subtitle}</p>
              </div>
            </div>

            {notice && !isExpired && (
              <StatusMessage variant="info" icon={<Clock className="h-4 w-4" />}>
                {notice}
              </StatusMessage>
            )}
            {isExpired && (
              <StatusMessage variant="warning" icon={<AlertCircle className="h-4 w-4" />}>
                <strong className="mb-1 block">{t.expiredTitle}</strong>
                {t.expiredDescription}
              </StatusMessage>
            )}
            {error && (
              <StatusMessage variant="error" icon={<AlertCircle className="h-4 w-4" />}>
                {error}
              </StatusMessage>
            )}

            {step === "details" && (
              <form className="mt-6 grid gap-5" onSubmit={requestOTP}>
                <div className="grid gap-2">
                  <label className="text-sm font-semibold" htmlFor="tenantSlug">
                    {t.tenantSlug}
                  </label>
                  <input
                    id="tenantSlug"
                    value={tenantSlug}
                    onChange={(event) =>
                      setTenantSlug(event.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))
                    }
                    placeholder={t.tenantSlugPlaceholder}
                    className="h-11 w-full rounded-lg border border-border bg-input-background px-3 text-left text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-ring/30"
                    dir="ltr"
                    autoComplete="organization"
                    required
                  />
                  <p className="text-xs leading-5 text-muted-foreground">{t.tenantSlugHint}</p>
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="grid gap-2">
                    <label className="text-sm font-semibold" htmlFor="email">
                      {t.email}
                    </label>
                    <input
                      id="email"
                      value={email}
                      onChange={(event) => setEmail(event.target.value.trim().toLowerCase())}
                      placeholder={t.emailPlaceholder}
                      className="h-11 w-full rounded-lg border border-border bg-input-background px-3 text-left text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-ring/30"
                      dir="ltr"
                      type="email"
                      autoComplete="email"
                    />
                  </div>
                  <div className="grid gap-2">
                    <label className="text-sm font-semibold" htmlFor="mobile">
                      {t.mobile}
                    </label>
                    <input
                      id="mobile"
                      value={mobile}
                      onChange={(event) => setMobile(event.target.value.replace(/[^\d+]/g, ""))}
                      placeholder={t.mobilePlaceholder}
                      className="h-11 w-full rounded-lg border border-border bg-input-background px-3 text-left text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-ring/30"
                      dir="ltr"
                      inputMode="tel"
                      autoComplete="tel"
                    />
                  </div>
                </div>
                <p className="text-xs leading-5 text-muted-foreground">{t.contactHint}</p>

                <button
                  type="submit"
                  disabled={loading || !tenantSlug || (!email && !mobile)}
                  className="inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:opacity-90 disabled:pointer-events-none disabled:opacity-50"
                >
                  {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                  {loading ? t.sendingCode : t.sendCode}
                </button>
              </form>
            )}

            {step === "otp" && (
              <form className="mt-6 grid gap-5" onSubmit={verifyOTP}>
                <div className="rounded-lg border border-border bg-secondary p-4">
                  <h2 className="text-base font-semibold">{t.pendingTitle}</h2>
                  <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    <span>{destination}</span>
                    <span className="rounded-full bg-card px-2 py-1 font-mono text-xs" dir="ltr">
                      {formattedRemaining}
                    </span>
                  </div>
                </div>

                {debugOTP && (
                  <StatusMessage variant="warning">
                    {t.devCode}:{" "}
                    <code className="rounded bg-card px-1.5 py-0.5 font-mono text-xs" dir="ltr">
                      {debugOTP}
                    </code>
                  </StatusMessage>
                )}

                <div className="grid gap-2">
                  <label className="text-sm font-semibold" htmlFor="otpCode">
                    {t.code}
                  </label>
                  <input
                    id="otpCode"
                    value={otpCode}
                    onChange={(event) => setOtpCode(event.target.value.replace(/\D/g, ""))}
                    placeholder={t.codePlaceholder}
                    maxLength={6}
                    className="h-12 w-full rounded-lg border border-border bg-input-background px-3 text-center font-mono text-lg tracking-[0.35em] outline-none transition focus:border-primary focus:ring-2 focus:ring-ring/30"
                    dir="ltr"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    required
                  />
                </div>

                <div className="flex flex-col-reverse gap-2 sm:flex-row">
                  <button
                    type="button"
                    onClick={() => void requestOTP()}
                    disabled={loading || !isExpired}
                    className="inline-flex h-11 items-center justify-center rounded-lg border border-border bg-card px-4 text-sm font-semibold transition hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                  >
                    {t.requestAgain}
                  </button>
                  <button
                    type="submit"
                    disabled={loading || otpCode.length !== 6 || isExpired}
                    className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-lg bg-primary px-4 text-sm font-semibold text-primary-foreground shadow-sm transition hover:opacity-90 disabled:pointer-events-none disabled:opacity-50"
                  >
                    {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                    {loading ? t.verifying : t.verify}
                  </button>
                </div>
              </form>
            )}

            {step === "done" && (
              <section className="mt-6 rounded-lg border border-emerald-200 bg-emerald-50 p-5 text-emerald-950">
                <div className="flex items-start gap-3">
                  <CheckCircle className="mt-1 h-5 w-5 shrink-0" aria-hidden="true" />
                  <div>
                    <h2 className="font-bold">{t.doneTitle}</h2>
                    <p className="mt-2 text-sm leading-6">{t.successDescription}</p>
                    <p className="mt-4 text-xs font-semibold">{t.loginUrl}</p>
                    <p className="mt-1 break-all font-mono text-sm" dir="ltr">
                      {loginURL}
                    </p>
                    {loginURL && (
                      <a
                        href={loginURL}
                        className="mt-4 inline-flex h-10 items-center justify-center rounded-lg bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800"
                      >
                        {t.openPanel}
                      </a>
                    )}
                  </div>
                </div>
              </section>
            )}
          </div>

          <aside className="rounded-lg border border-border bg-card p-5 shadow-sm">
            <h2 className="text-base font-bold">{step === "details" ? t.detailsTitle : t.otpTitle}</h2>
            <ol className="mt-5 grid gap-3 text-sm text-muted-foreground">
              <StepItem active={step === "details"} done={step !== "details"} label={t.detailsTitle} />
              <StepItem active={step === "otp"} done={step === "done"} label={t.otpTitle} />
              <StepItem active={step === "done"} done={step === "done"} label={t.doneTitle} />
            </ol>
          </aside>
        </section>
      </div>
    </main>
  );
}

function StepItem({ active, done, label }: { active: boolean; done: boolean; label: string }) {
  return (
    <li className="flex items-center gap-2">
      <span
        className={[
          "h-2.5 w-2.5 rounded-full",
          done ? "bg-emerald-600" : active ? "bg-primary" : "bg-muted",
        ].join(" ")}
      />
      <span className={active ? "font-semibold text-foreground" : ""}>{label}</span>
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
    info: "border-blue-200 bg-blue-50 text-blue-950",
    warning: "border-amber-200 bg-amber-50 text-amber-950",
    error: "border-red-200 bg-red-50 text-red-950",
  };

  return (
    <div className={`mb-4 flex items-start gap-3 rounded-lg border p-4 text-sm leading-6 ${styles[variant]}`}>
      {icon && <span className="mt-1 shrink-0">{icon}</span>}
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

function formatRemaining(totalSeconds: number) {
  const safeSeconds = Math.max(0, totalSeconds);
  const minutes = Math.floor(safeSeconds / 60);
  const seconds = safeSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}
