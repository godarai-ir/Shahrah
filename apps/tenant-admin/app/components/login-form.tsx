"use client";

import Image from "next/image";
import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";
import {
  AlertCircle,
  CheckCircle,
  KeyRound,
  RefreshCw,
  ShieldCheck,
} from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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
  redirectTo: string;
  initialIdentifier?: string;
};

export default function LoginForm({ initialCaptcha, redirectTo, initialIdentifier = "" }: LoginFormProps) {
  const t = useTranslations("login");
  const common = useTranslations("common");

  const [identifier, setIdentifier] = useState(initialIdentifier);
  const [captchaID, setCaptchaID] = useState(initialCaptcha.captcha_id);
  const [captchaImage, setCaptchaImage] = useState(initialCaptcha.image_data_url);
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
      setError(normalizeError(err, common("requestError")));
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
      setNotice(t("otp.sent", { destination: result.destination_masked }));
      setCaptchaAnswer("");
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
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
      window.location.href = redirectTo;
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-screen bg-background px-4 py-6 text-foreground sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="relative flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-card p-1.5 shadow-sm">
              <Image
                src="/brand/shahrah-logo.png"
                alt={t("logoAlt")}
                width={40}
                height={40}
                className="h-full w-full object-contain"
                priority
              />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-foreground">Shahrah</p>
              <p className="truncate text-xs text-muted-foreground">{t("subtitle")}</p>
            </div>
          </div>
          <ThemeToggle />
        </header>

        <div className="grid flex-1 items-center gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(24rem,28rem)]">
          <section className="hidden min-w-0 flex-col gap-5 rounded-lg border border-border bg-muted/35 p-6 lg:order-2 lg:flex">
            <div className="flex items-start gap-3">
              <span className="flex size-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card text-primary">
                <ShieldCheck className="size-5" aria-hidden="true" />
              </span>
              <div className="min-w-0">
                <h1 className="text-2xl font-semibold tracking-normal">{t("title")}</h1>
                <p className="mt-2 max-w-xl text-sm leading-6 text-muted-foreground">
                  {t("subtitle")}
                </p>
              </div>
            </div>

            <div className="grid gap-3 border-t border-border pt-5">
              <div className="flex items-center gap-3 rounded-lg border border-border bg-card p-3">
                <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-secondary text-secondary-foreground">
                  <KeyRound className="size-4" aria-hidden="true" />
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-medium">{t("identifier")}</p>
                  <p className="text-xs leading-5 text-muted-foreground">{t("otp.request")}</p>
                </div>
              </div>
              <div className="flex items-center gap-3 rounded-lg border border-border bg-card p-3">
                <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-secondary text-secondary-foreground">
                  <ShieldCheck className="size-4" aria-hidden="true" />
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-medium">{t("otp.verifyTitle")}</p>
                  <p className="text-xs leading-5 text-muted-foreground">{t("otp.verify")}</p>
                </div>
              </div>
            </div>
          </section>

          <div className="mx-auto w-full max-w-md animate-fade-in lg:order-1">
            <div className="mb-4 flex flex-col gap-3 lg:hidden">
              <div className="flex items-center gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card text-primary">
                  <ShieldCheck className="size-5" aria-hidden="true" />
                </span>
                <div className="min-w-0">
                  <h1 className="text-xl font-semibold tracking-normal">{t("title")}</h1>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">{t("subtitle")}</p>
                </div>
              </div>
            </div>

            <div className="mb-3 grid grid-cols-2 overflow-hidden rounded-lg border border-border bg-card text-xs font-medium">
              <div
                className={
                  !otpRequest
                    ? "bg-primary px-3 py-2 text-center text-primary-foreground"
                    : "px-3 py-2 text-center text-muted-foreground"
                }
              >
                {t("otp.request")}
              </div>
              <div
                className={
                  otpRequest
                    ? "bg-primary px-3 py-2 text-center text-primary-foreground"
                    : "px-3 py-2 text-center text-muted-foreground"
                }
              >
                {t("otp.verifyTitle")}
              </div>
            </div>

            <Card className="rounded-lg border-border shadow-sm">
              <CardHeader className="border-b border-border pb-4">
                <CardTitle className="text-base">
                  {otpRequest ? t("otp.verifyTitle") : t("title")}
                </CardTitle>
                <CardDescription>
                  {otpRequest
                    ? t("otp.enterCode", { destination: otpRequest.destination_masked })
                    : t("subtitle")}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-5 pt-5 md:pt-5">
                {notice && (
                  <Alert variant="success" icon={<CheckCircle className="size-4" />}>
                    <AlertDescription>{notice}</AlertDescription>
                  </Alert>
                )}
                {error && (
                  <Alert variant="destructive" icon={<AlertCircle className="size-4" />}>
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                {otpRequest?.debug_otp && (
                  <Alert variant="warning">
                    <AlertDescription>
                      {t("otp.devCode")}:{" "}
                      <code
                        dir="ltr"
                        className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs"
                      >
                        {otpRequest.debug_otp}
                      </code>
                    </AlertDescription>
                  </Alert>
                )}

                {!otpRequest ? (
                  <form className="grid gap-5" onSubmit={handleRequestOTP}>
                    <div className="grid gap-2">
                      <Label htmlFor="identifier">{t("identifier")}</Label>
                      <Input
                        id="identifier"
                        value={identifier}
                        onChange={(event) => setIdentifier(event.target.value)}
                        placeholder={t("identifierPlaceholder")}
                        autoComplete="username"
                        required
                        className="text-start"
                      />
                    </div>

                    <div className="grid gap-3">
                      <Label htmlFor="captchaAnswer">{t("captcha.label")}</Label>
                      <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/35 p-3">
                        <div className="flex items-center justify-between gap-3">
                          {captchaImage ? (
                            <Image
                              src={captchaImage}
                              alt={t("captcha.label")}
                              width={170}
                              height={56}
                              unoptimized
                              className="h-14 w-[170px] rounded-md border border-border bg-card object-contain"
                            />
                          ) : (
                            <div className="h-14 w-[170px] rounded-md border border-border bg-card" />
                          )}
                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            onClick={() => void refreshCaptcha()}
                            aria-label={t("captcha.refresh")}
                          >
                            <RefreshCw className="size-4" aria-hidden="true" />
                          </Button>
                        </div>
                        <Input
                          id="captchaAnswer"
                          dir="ltr"
                          value={captchaAnswer}
                          onChange={(event) => setCaptchaAnswer(event.target.value)}
                          placeholder={t("captcha.answerPlaceholder")}
                          autoComplete="off"
                          required
                          className="text-start"
                        />
                      </div>
                    </div>

                    <Button type="submit" loading={loading} className="w-full">
                      {loading ? t("otp.requesting") : t("otp.request")}
                    </Button>
                  </form>
                ) : (
                  <form className="grid gap-5" onSubmit={handleVerifyOTP}>
                    <div className="grid gap-2">
                      <Label htmlFor="otpCode">{t("otp.code")}</Label>
                      <Input
                        id="otpCode"
                        dir="ltr"
                        value={otpCode}
                        onChange={(event) => setOTPCode(event.target.value)}
                        placeholder={t("otp.codePlaceholder")}
                        autoComplete="one-time-code"
                        required
                        maxLength={6}
                        inputMode="numeric"
                        className="font-mono text-start tracking-normal"
                      />
                    </div>

                    <div className="flex flex-col-reverse gap-2 sm:flex-row">
                      <Button type="submit" loading={loading} className="flex-1">
                        {loading ? t("otp.verifying") : t("otp.verify")}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                          setOTPRequest(null);
                          setOTPCode("");
                          setNotice(null);
                          void refreshCaptcha();
                        }}
                      >
                        {t("otp.back")}
                      </Button>
                    </div>
                  </form>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </main>
  );
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

function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}
