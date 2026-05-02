"use client";

import Image from "next/image";
import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";
import { AlertCircle, CheckCircle, RefreshCw } from "lucide-react";

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
};

export default function LoginForm({ initialCaptcha, redirectTo }: LoginFormProps) {
  const t = useTranslations("login");
  const common = useTranslations("common");

  const [identifier, setIdentifier] = useState("");
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
    <main className="relative flex min-h-screen items-center justify-center px-4 py-12 md:px-8 md:py-16">
      {/* Background Gradient */}
      <div className="absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute left-1/2 top-0 h-[500px] w-[800px] -translate-x-1/2 opacity-20 dark:opacity-10">
          <div className="bg-gradient-to-r from-indigo-500 to-purple-500 h-full w-full blur-[120px] rounded-full" />
        </div>
      </div>

      <div className="mx-auto w-full max-w-md space-y-6 animate-fade-in">
        {/* Theme Toggle */}
        <div className="flex justify-start">
          <ThemeToggle />
        </div>

        {/* Logo */}
        <div className="flex flex-col items-center gap-3 text-center">
          <div className="relative flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl border border-border bg-card p-2 shadow-lg">
            <Image
              src="/brand/shahrah-logo.png"
              alt={t("logoAlt")}
              width={56}
              height={56}
              className="h-full w-full object-contain"
              priority
            />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">{t("title")}</h1>
            <p className="mt-1 text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
        </div>

        {/* Auth Card */}
        <Card className="border-border/60 shadow-lg">
          <CardHeader className="pb-4">
            <CardTitle className="text-lg">
              {otpRequest ? t("otp.verifyTitle") : t("title")}
            </CardTitle>
            <CardDescription>
              {otpRequest
                ? t("otp.enterCode", { destination: otpRequest.destination_masked })
                : t("subtitle")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            {notice && (
              <Alert variant="success" icon={<CheckCircle className="h-4 w-4" />}>
                <AlertDescription>{notice}</AlertDescription>
              </Alert>
            )}
            {error && (
              <Alert variant="destructive" icon={<AlertCircle className="h-4 w-4" />}>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {otpRequest?.debug_otp && (
              <Alert variant="warning">
                <AlertDescription>
                  {t("otp.devCode")}:{" "}
                  <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
                    {otpRequest.debug_otp}
                  </code>
                </AlertDescription>
              </Alert>
            )}

            {!otpRequest ? (
              <form className="grid gap-5" onSubmit={handleRequestOTP}>
                {/* Identifier */}
                <div className="grid gap-2">
                  <Label htmlFor="identifier">{t("identifier")}</Label>
                  <Input
                    id="identifier"
                    value={identifier}
                    onChange={(event) => setIdentifier(event.target.value)}
                    placeholder={t("identifierPlaceholder")}
                    autoComplete="username"
                    required
                  />
                </div>

                {/* Captcha */}
                <div className="grid gap-3">
                  <Label>{t("captcha.label")}</Label>
                  <Image
                    src={captchaImage}
                    alt="captcha"
                    width={170}
                    height={56}
                    unoptimized
                    className="h-14 w-[170px] rounded-lg border border-border bg-secondary"
                  />
                  <div className="flex gap-2">
                    <Input
                      value={captchaAnswer}
                      onChange={(event) => setCaptchaAnswer(event.target.value)}
                      placeholder={t("captcha.answerPlaceholder")}
                      autoComplete="off"
                      required
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() => void refreshCaptcha()}
                      aria-label={t("captcha.refresh")}
                    >
                      <RefreshCw className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {/* Submit */}
                <Button type="submit" loading={loading} className="w-full">
                  {loading ? t("otp.requesting") : t("otp.request")}
                </Button>
              </form>
            ) : (
              <form className="grid gap-5" onSubmit={handleVerifyOTP}>
                {/* OTP Code */}
                <div className="grid gap-2">
                  <Label htmlFor="otpCode">{t("otp.code")}</Label>
                  <Input
                    id="otpCode"
                    value={otpCode}
                    onChange={(event) => setOTPCode(event.target.value)}
                    placeholder={t("otp.codePlaceholder")}
                    autoComplete="one-time-code"
                    required
                    maxLength={6}
                    inputMode="numeric"
                  />
                </div>

                {/* Actions */}
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
