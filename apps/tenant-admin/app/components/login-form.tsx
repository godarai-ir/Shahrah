"use client";

import Image from "next/image";
import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";

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
    <main className="min-h-screen px-4 py-8 md:px-8 md:py-14">
      <div className="mx-auto w-full max-w-lg space-y-4">
        <div className="flex justify-start">
          <ThemeToggle />
        </div>

        <Card>
          <CardHeader>
            <div className="mb-2 flex items-center justify-center">
              <Image
                src="/brand/shahrah-logo.png"
                alt={t("logoAlt")}
                width={68}
                height={68}
                className="h-16 w-16 rounded-xl border bg-muted/40 p-1 object-contain"
                priority
              />
            </div>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("subtitle")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {notice ? (
              <Alert className="border-primary/30 bg-primary/5">
                <AlertDescription>{notice}</AlertDescription>
              </Alert>
            ) : null}
            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}

            {!otpRequest ? (
              <form className="grid gap-4" onSubmit={handleRequestOTP}>
                <div className="grid gap-2">
                  <Label htmlFor="identifier">{t("identifier")}</Label>
                  <Input
                    id="identifier"
                    value={identifier}
                    onChange={(event) => setIdentifier(event.target.value)}
                    placeholder={t("identifierPlaceholder")}
                    autoComplete="username"
                  />
                </div>

                <div className="grid gap-3">
                  <Label>{t("captcha.label")}</Label>
                  <Image
                    src={captchaImage}
                    alt="captcha"
                    width={170}
                    height={56}
                    unoptimized
                    className="h-14 w-[170px] rounded-md border bg-secondary"
                  />
                  <div className="grid gap-3 sm:grid-cols-[1fr_auto] sm:items-end">
                    <div className="grid gap-2">
                      <Label htmlFor="captchaAnswer">{t("captcha.answer")}</Label>
                      <Input
                        id="captchaAnswer"
                        value={captchaAnswer}
                        onChange={(event) => setCaptchaAnswer(event.target.value)}
                        placeholder={t("captcha.answerPlaceholder")}
                        autoComplete="off"
                      />
                    </div>
                    <Button type="button" variant="outline" onClick={() => void refreshCaptcha()}>
                      {t("captcha.refresh")}
                    </Button>
                  </div>
                </div>

                <Button type="submit" disabled={loading}>
                  {loading ? t("otp.requesting") : t("otp.request")}
                </Button>
              </form>
            ) : (
              <form className="grid gap-4" onSubmit={handleVerifyOTP}>
                <p className="text-sm text-muted-foreground">
                  {t("otp.enterCode", { destination: otpRequest.destination_masked })}
                </p>
                {otpRequest.debug_otp ? (
                  <Alert className="border-amber-400/45 bg-amber-100/40 text-amber-950 dark:bg-amber-500/10 dark:text-amber-200">
                    <AlertDescription>
                      {t("otp.devCode")}: <code>{otpRequest.debug_otp}</code>
                    </AlertDescription>
                  </Alert>
                ) : null}
                <div className="grid gap-2">
                  <Label htmlFor="otpCode">{t("otp.code")}</Label>
                  <Input
                    id="otpCode"
                    value={otpCode}
                    onChange={(event) => setOTPCode(event.target.value)}
                    placeholder={t("otp.codePlaceholder")}
                    autoComplete="one-time-code"
                  />
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button type="submit" disabled={loading}>
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
