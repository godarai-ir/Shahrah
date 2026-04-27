"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

type UserProfile = {
  id: string;
  email?: string;
  mobile?: string;
  role: string;
};

type OTPRequestResponse = {
  request_id: string;
  destination_masked: string;
  debug_otp?: string;
};

type FieldState = {
  value: string;
  requestID: string;
  otpCode: string;
  destinationMasked: string;
  debugOTP?: string;
};

type ProfileSettingsProps = {
  initialUser: UserProfile;
};

export default function ProfileSettings(props: ProfileSettingsProps) {
  const [user, setUser] = useState<UserProfile>(props.initialUser);
  const [emailState, setEmailState] = useState<FieldState>({
    value: props.initialUser.email ?? "",
    requestID: "",
    otpCode: "",
    destinationMasked: "",
  });
  const [mobileState, setMobileState] = useState<FieldState>({
    value: props.initialUser.mobile ?? "",
    requestID: "",
    otpCode: "",
    destinationMasked: "",
  });
  const [loadingField, setLoadingField] = useState<"email" | "mobile" | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleRequestChange(field: "email" | "mobile") {
    setLoadingField(field);
    setError(null);
    setMessage(null);

    const value = field === "email" ? emailState.value : mobileState.value;
    try {
      const result = await requestJSON<OTPRequestResponse>(
        "/api/auth/profile/request-contact-change",
        {
          method: "POST",
          body: JSON.stringify({
            field,
            new_value: value,
          }),
        },
      );

      if (field === "email") {
        setEmailState((current) => ({
          ...current,
          requestID: result.request_id,
          destinationMasked: result.destination_masked,
          debugOTP: result.debug_otp,
          otpCode: "",
        }));
      } else {
        setMobileState((current) => ({
          ...current,
          requestID: result.request_id,
          destinationMasked: result.destination_masked,
          debugOTP: result.debug_otp,
          otpCode: "",
        }));
      }

      setMessage(`کد تایید به ${result.destination_masked} ارسال شد.`);
    } catch (err) {
      setError(normalizeError(err));
    } finally {
      setLoadingField(null);
    }
  }

  async function handleVerifyChange(
    event: FormEvent<HTMLFormElement>,
    field: "email" | "mobile",
  ) {
    event.preventDefault();
    setLoadingField(field);
    setError(null);
    setMessage(null);

    const state = field === "email" ? emailState : mobileState;
    try {
      const result = await requestJSON<{ user: UserProfile }>(
        "/api/auth/profile/verify-contact-change",
        {
          method: "POST",
          body: JSON.stringify({
            field,
            request_id: state.requestID,
            otp_code: state.otpCode,
          }),
        },
      );
      setUser(result.user);

      if (field === "email") {
        setEmailState((current) => ({
          ...current,
          requestID: "",
          otpCode: "",
          destinationMasked: "",
          debugOTP: undefined,
        }));
      } else {
        setMobileState((current) => ({
          ...current,
          requestID: "",
          otpCode: "",
          destinationMasked: "",
          debugOTP: undefined,
        }));
      }
      setMessage(field === "email" ? "ایمیل با موفقیت تغییر کرد." : "شماره همراه با موفقیت تغییر کرد.");
    } catch (err) {
      setError(normalizeError(err));
    } finally {
      setLoadingField(null);
    }
  }

  return (
    <main className="min-h-screen px-4 py-8 md:px-8 md:py-12">
      <div className="mx-auto w-full max-w-3xl space-y-6">
        <header className="rounded-3xl border border-[var(--border)] bg-[var(--surface)]/92 p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="text-xs tracking-[0.35em] text-[var(--muted)]">
                تنظیمات کاربری
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight">
                حساب کاربری
              </h1>
              <p className="mt-2 text-sm text-[var(--muted)]">
                نقش فعلی: <b>{user.role}</b>. در فاز فعلی همه کاربران دسترسی یکسان
                به پنل دارند.
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <Link
                href="/"
                className="rounded-xl border border-[var(--border)] px-3 py-2 text-sm text-[var(--muted)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
              >
                بازگشت به داشبورد
              </Link>
              <form action="/logout" method="post">
                <button
                  type="submit"
                  className="rounded-xl border border-[var(--danger)]/55 bg-[var(--danger)]/10 px-3 py-2 text-sm text-[var(--danger)] transition hover:bg-[var(--danger)]/20"
                >
                  خروج
                </button>
              </form>
            </div>
          </div>
        </header>

        {message ? (
          <div className="rounded-xl border border-[var(--good)]/40 bg-[var(--good)]/10 px-3 py-2 text-sm text-[var(--good)]">
            {message}
          </div>
        ) : null}
        {error ? (
          <div className="rounded-xl border border-[var(--danger)]/45 bg-[var(--danger)]/10 px-3 py-2 text-sm text-[var(--danger)]">
            {error}
          </div>
        ) : null}

        <section className="grid gap-6 md:grid-cols-2">
          <article className="rounded-2xl border border-[var(--border)] bg-[var(--surface)]/92 p-5">
            <h2 className="text-lg font-semibold">ایمیل</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              مقدار فعلی: {user.email ?? "ثبت نشده"}
            </p>

            <label className="mt-4 block text-sm text-[var(--muted)]">
              ایمیل جدید
              <input
                value={emailState.value}
                onChange={(event) =>
                  setEmailState((current) => ({
                    ...current,
                    value: event.target.value,
                  }))
                }
                className={inputClassName}
                placeholder="name@example.com"
              />
            </label>
            <button
              type="button"
              onClick={() => void handleRequestChange("email")}
              disabled={loadingField === "email"}
              className="mt-3 rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
            >
              ارسال کد تایید
            </button>

            {emailState.requestID ? (
              <form
                className="mt-4 grid gap-3"
                onSubmit={(event) => void handleVerifyChange(event, "email")}
              >
                <p className="text-xs text-[var(--muted)]">
                  ارسال شده به {emailState.destinationMasked}
                </p>
                {emailState.debugOTP ? (
                  <p className="rounded-lg border border-amber-300/45 bg-amber-100/10 px-3 py-2 text-xs text-amber-200">
                    کد توسعه: <code>{emailState.debugOTP}</code>
                  </p>
                ) : null}
                <label className="text-sm text-[var(--muted)]">
                  کد OTP
                  <input
                    value={emailState.otpCode}
                    onChange={(event) =>
                      setEmailState((current) => ({
                        ...current,
                        otpCode: event.target.value,
                      }))
                    }
                    className={inputClassName}
                    placeholder="کد ۶ رقمی"
                  />
                </label>
                <button
                  type="submit"
                  disabled={loadingField === "email"}
                  className="rounded-xl border border-[var(--good)]/55 bg-[var(--good)]/20 px-4 py-2 text-sm font-medium text-[var(--good)] transition hover:bg-[var(--good)]/30 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  تایید تغییر ایمیل
                </button>
              </form>
            ) : null}
          </article>

          <article className="rounded-2xl border border-[var(--border)] bg-[var(--surface)]/92 p-5">
            <h2 className="text-lg font-semibold">شماره همراه</h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              مقدار فعلی: {user.mobile ?? "ثبت نشده"}
            </p>

            <label className="mt-4 block text-sm text-[var(--muted)]">
              شماره همراه جدید
              <input
                value={mobileState.value}
                onChange={(event) =>
                  setMobileState((current) => ({
                    ...current,
                    value: event.target.value,
                  }))
                }
                className={inputClassName}
                placeholder="+98912..."
              />
            </label>
            <button
              type="button"
              onClick={() => void handleRequestChange("mobile")}
              disabled={loadingField === "mobile"}
              className="mt-3 rounded-xl border border-[var(--accent)]/55 bg-[var(--accent)] px-4 py-2 text-sm font-medium text-slate-950 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
            >
              ارسال کد تایید
            </button>

            {mobileState.requestID ? (
              <form
                className="mt-4 grid gap-3"
                onSubmit={(event) => void handleVerifyChange(event, "mobile")}
              >
                <p className="text-xs text-[var(--muted)]">
                  ارسال شده به {mobileState.destinationMasked}
                </p>
                {mobileState.debugOTP ? (
                  <p className="rounded-lg border border-amber-300/45 bg-amber-100/10 px-3 py-2 text-xs text-amber-200">
                    کد توسعه: <code>{mobileState.debugOTP}</code>
                  </p>
                ) : null}
                <label className="text-sm text-[var(--muted)]">
                  کد OTP
                  <input
                    value={mobileState.otpCode}
                    onChange={(event) =>
                      setMobileState((current) => ({
                        ...current,
                        otpCode: event.target.value,
                      }))
                    }
                    className={inputClassName}
                    placeholder="کد ۶ رقمی"
                  />
                </label>
                <button
                  type="submit"
                  disabled={loadingField === "mobile"}
                  className="rounded-xl border border-[var(--good)]/55 bg-[var(--good)]/20 px-4 py-2 text-sm font-medium text-[var(--good)] transition hover:bg-[var(--good)]/30 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  تایید تغییر شماره همراه
                </button>
              </form>
            ) : null}
          </article>
        </section>
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
