"use client";

import Image from "next/image";
import { FormEvent, useState } from "react";

type CaptchaChallenge = {
  captcha_id: string;
  image_data_url: string;
  expires_in_seconds: number;
};

type OTPRequestResponse = {
  request_id: string;
  channel: "email" | "mobile";
  destination_masked: string;
  expires_in_seconds: number;
  debug_otp?: string;
};

type LoginFormProps = {
  initialCaptcha: CaptchaChallenge;
};

export default function LoginForm(props: LoginFormProps) {
  const [identifier, setIdentifier] = useState("");
  const [captchaID, setCaptchaID] = useState(props.initialCaptcha.captcha_id);
  const [captchaImage, setCaptchaImage] = useState(
    props.initialCaptcha.image_data_url,
  );
  const [captchaAnswer, setCaptchaAnswer] = useState("");

  const [otpRequest, setOTPRequest] = useState<OTPRequestResponse | null>(null);
  const [otpCode, setOTPCode] = useState("");
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function refreshCaptcha() {
    setError(null);
    try {
      const challenge = await requestJSON<CaptchaChallenge>("/api/auth/captcha");
      setCaptchaID(challenge.captcha_id);
      setCaptchaImage(challenge.image_data_url);
      setCaptchaAnswer("");
    } catch (err) {
      setError(normalizeError(err));
    }
  }

  async function handleRequestOTP(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setNotice(null);

    try {
      const result = await requestJSON<OTPRequestResponse>("/api/auth/request-otp", {
        method: "POST",
        body: JSON.stringify({
          identifier,
          captcha_id: captchaID,
          captcha_answer: captchaAnswer,
        }),
      });
      setOTPRequest(result);
      setNotice(`کد OTP به ${result.destination_masked} ارسال شد.`);
      setCaptchaAnswer("");
    } catch (err) {
      setError(normalizeError(err));
      await refreshCaptcha();
    } finally {
      setLoading(false);
    }
  }

  async function handleVerifyOTP(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!otpRequest) {
      return;
    }

    setLoading(true);
    setError(null);
    setNotice(null);

    try {
      await requestJSON("/api/auth/verify-otp", {
        method: "POST",
        body: JSON.stringify({
          request_id: otpRequest.request_id,
          otp_code: otpCode,
        }),
      });
      window.location.href = "/";
    } catch (err) {
      setError(normalizeError(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen px-4 py-8 md:px-8 md:py-14">
      <div className="mx-auto w-full max-w-lg rounded-3xl border border-[var(--border)] bg-[var(--surface)]/95 p-6 shadow-[0_28px_70px_rgba(2,8,23,0.5)]">
        <p className="text-xs tracking-[0.35em] text-[var(--muted)]">
          پنل مدیریت تننت شاهراه
        </p>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight">ورود</h1>
        <p className="mt-2 text-sm text-[var(--muted)]">
          با ایمیل یا شماره همراه وارد شوید. کد OTP بر اساس نوع ورودی ارسال
          می‌شود.
        </p>

        {notice ? (
          <div className="mt-4 rounded-xl border border-[var(--good)]/40 bg-[var(--good)]/10 px-3 py-2 text-sm text-[var(--good)]">
            {notice}
          </div>
        ) : null}
        {error ? (
          <div className="mt-4 rounded-xl border border-[var(--danger)]/45 bg-[var(--danger)]/10 px-3 py-2 text-sm text-[var(--danger)]">
            {error}
          </div>
        ) : null}

        {!otpRequest ? (
          <form className="mt-5 grid gap-4" onSubmit={handleRequestOTP}>
            <label className="text-sm text-[var(--muted)]">
              ایمیل یا شماره همراه
              <input
                value={identifier}
                onChange={(event) => setIdentifier(event.target.value)}
                className={inputClassName}
                placeholder="name@example.com یا +98912..."
                autoComplete="username"
              />
            </label>

            <div className="grid gap-3">
              <label className="text-sm text-[var(--muted)]">کپچای تصویری</label>
              <Image
                src={captchaImage}
                alt="captcha"
                width={170}
                height={56}
                unoptimized
                className="h-14 w-[170px] rounded-lg border border-[var(--border)] bg-slate-100"
              />
              <div className="flex items-end gap-2">
                <label className="flex-1 text-sm text-[var(--muted)]">
                  پاسخ کپچا
                  <input
                    value={captchaAnswer}
                    onChange={(event) => setCaptchaAnswer(event.target.value)}
                    className={inputClassName}
                    placeholder="متن کپچا را وارد کنید"
                    autoComplete="off"
                  />
                </label>
                <button
                  type="button"
                  onClick={() => void refreshCaptcha()}
                  className="rounded-xl border border-[var(--border)] px-3 py-2 text-sm text-[var(--muted)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
                >
                  کپچای جدید
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? "در حال ارسال OTP..." : "دریافت OTP"}
            </button>
          </form>
        ) : (
          <form className="mt-5 grid gap-4" onSubmit={handleVerifyOTP}>
            <p className="text-sm text-[var(--muted)]">
              کد OTP ارسال‌شده به <b>{otpRequest.destination_masked}</b> را وارد
              کنید.
            </p>
            {otpRequest.debug_otp ? (
              <p className="rounded-lg border border-amber-300/45 bg-amber-100/10 px-3 py-2 text-xs text-amber-200">
                کد OTP توسعه: <code>{otpRequest.debug_otp}</code>
              </p>
            ) : null}
            <label className="text-sm text-[var(--muted)]">
              کد OTP
              <input
                value={otpCode}
                onChange={(event) => setOTPCode(event.target.value)}
                className={inputClassName}
                placeholder="کد ۶ رقمی"
                autoComplete="one-time-code"
              />
            </label>

            <div className="flex flex-wrap gap-2">
              <button
                type="submit"
                disabled={loading}
                className="rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {loading ? "در حال تایید..." : "ورود به پنل"}
              </button>
              <button
                type="button"
                onClick={() => {
                  setOTPRequest(null);
                  setOTPCode("");
                  setNotice(null);
                  void refreshCaptcha();
                }}
                className="rounded-xl border border-[var(--border)] px-4 py-2 text-sm text-[var(--muted)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
              >
                بازگشت
              </button>
            </div>
          </form>
        )}
      </div>
    </main>
  );
}

const inputClassName =
  "mt-1 w-full rounded-xl border border-[var(--border)] bg-[var(--surface-2)] px-3 py-2 text-sm outline-none ring-0 focus:border-[var(--accent)]";

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
  const payload = parsePayload(text);
  if (!response.ok) {
    throw new Error(getErrorMessage(payload) ?? `خطای درخواست (${response.status})`);
  }
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function getErrorMessage(payload: unknown): string | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  const candidate = payload as {
    error?: { message?: string };
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
