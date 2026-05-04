"use client";

import { useState } from "react";
import Link from "next/link";

type Step = "details" | "otp" | "done";

export default function OnboardingPage() {
  const [step, setStep] = useState<Step>("details");
  const [tenantSlug, setTenantSlug] = useState("");
  const [email, setEmail] = useState("");
  const [mobile, setMobile] = useState("");
  const [otpCode, setOtpCode] = useState("");
  const [requestID, setRequestID] = useState("");
  const [loginURL, setLoginURL] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function requestOTP() {
    setError("");
    setLoading(true);
    try {
      const availability = await postOnboarding({
        action: "check_slug",
        tenant_slug: tenantSlug,
      });
      if (!availability.ok) {
        setError(availability.payload.error ?? "Could not check tenant slug availability.");
        return;
      }
      if (!availability.payload.available) {
        const reason = availability.payload.reason === "invalid_or_reserved" ? "reserved or invalid" : "already taken";
        setError(`Tenant slug is ${reason}.`);
        return;
      }

      const otp = await postOnboarding({
        action: "request_otp",
        tenant_slug: tenantSlug,
        tenant_domain: `${tenantSlug}.localhost`,
        plan_slug: "free",
        email,
        mobile,
        metadata: { source: "website-onboarding" },
      });
      if (!otp.ok) {
        setError(otp.payload.error ?? "Could not send OTP.");
        return;
      }
      setRequestID(String(otp.payload.request_id ?? ""));
      setStep("otp");
    } finally {
      setLoading(false);
    }
  }

  async function verifyOTP() {
    setError("");
    setLoading(true);
    try {
      const result = await postOnboarding({
        action: "verify_otp",
        request_id: requestID,
        otp_code: otpCode,
      });
      if (!result.ok) {
        setError(result.payload.error ?? "Invalid or expired OTP.");
        return;
      }
      setLoginURL(String(result.payload.login_url ?? ""));
      setStep("done");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-white px-6 py-12 text-slate-950 dark:bg-[#0B0F19] dark:text-white">
      <div className="mx-auto max-w-xl">
        <Link href="/pricing" className="text-sm font-semibold text-indigo-600 dark:text-indigo-400">
          Back to plans
        </Link>
        <h1 className="mt-8 text-3xl font-bold">Create your Shahrah tenant</h1>
        <p className="mt-3 text-sm text-slate-600 dark:text-slate-400">
          Choose a globally unique tenant slug. The workspace is provisioned only after OTP verification.
        </p>

        {step === "details" && (
          <section className="mt-8 space-y-4">
            <input
              value={tenantSlug}
              onChange={(event) => setTenantSlug(event.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))}
              placeholder="tenant slug, for example mytenant"
              className="w-full rounded-lg border border-slate-300 bg-transparent px-4 py-3 text-left"
              dir="ltr"
            />
            <input
              value={email}
              onChange={(event) => setEmail(event.target.value.trim().toLowerCase())}
              placeholder="email"
              className="w-full rounded-lg border border-slate-300 bg-transparent px-4 py-3 text-left"
              dir="ltr"
            />
            <input
              value={mobile}
              onChange={(event) => setMobile(event.target.value.replace(/[^\d+]/g, ""))}
              placeholder="mobile"
              className="w-full rounded-lg border border-slate-300 bg-transparent px-4 py-3 text-left"
              dir="ltr"
            />
            <button
              onClick={requestOTP}
              disabled={loading || !tenantSlug || (!email && !mobile)}
              className="w-full rounded-lg bg-indigo-600 px-4 py-3 font-bold text-white disabled:bg-slate-300"
            >
              Send OTP
            </button>
          </section>
        )}

        {step === "otp" && (
          <section className="mt-8 space-y-4">
            <input
              value={otpCode}
              onChange={(event) => setOtpCode(event.target.value.replace(/\D/g, ""))}
              placeholder="OTP code"
              maxLength={6}
              className="w-full rounded-lg border border-slate-300 bg-transparent px-4 py-3 text-center tracking-[0.35em]"
              dir="ltr"
            />
            <button
              onClick={verifyOTP}
              disabled={loading || otpCode.length !== 6}
              className="w-full rounded-lg bg-indigo-600 px-4 py-3 font-bold text-white disabled:bg-slate-300"
            >
              Verify and provision tenant
            </button>
          </section>
        )}

        {step === "done" && (
          <section className="mt-8 rounded-lg border border-emerald-300 bg-emerald-50 p-5 text-emerald-900">
            <h2 className="font-bold">Tenant provisioned</h2>
            <p className="mt-2 text-sm">Login URL:</p>
            <p className="mt-1 font-mono text-sm" dir="ltr">{loginURL}</p>
          </section>
        )}

        {error && <p className="mt-4 text-sm font-semibold text-red-600">{error}</p>}
      </div>
    </main>
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
    payload: await response.json(),
  };
}
