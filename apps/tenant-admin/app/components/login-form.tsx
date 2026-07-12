"use client";

import Image from "next/image";
import { type FormEvent, type ReactNode, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { REGEXP_ONLY_DIGITS } from "input-otp";
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle,
  KeyRound,
  LockKeyhole,
  RefreshCw,
  ShieldCheck,
} from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

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

type LoginFieldErrors = {
  identifier?: string;
  captchaAnswer?: string;
};

type LoginTranslator = (key: string, values?: Record<string, string | number>) => string;

const ANIMATED_HERO_INTERVAL_MS = 2200;
const ANIMATED_HERO_TRANSITION_MS = 220;

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
  const [fieldErrors, setFieldErrors] = useState<LoginFieldErrors>({});
  const [loading, setLoading] = useState(false);
  const activeStep = otpRequest ? "verify" : "request";
  const animatedHeroWords = [
    t("hero.animated.words.stable"),
    t("hero.animated.words.saving"),
    t("hero.animated.words.observability"),
    t("hero.animated.words.control"),
  ];

  async function refreshCaptcha(options: { clearError?: boolean } = {}) {
    if (options.clearError ?? true) {
      setError(null);
    }
    try {
      const challenge = await requestJSON<CaptchaChallenge>("/api/auth/captcha");
      setCaptchaID(challenge.captcha_id);
      setCaptchaImage(challenge.image_data_url);
      setCaptchaAnswer("");
      setFieldErrors((current) => ({ ...current, captchaAnswer: undefined }));
    } catch (err) {
      setError(normalizeLoginError(err, common("requestError"), t));
    }
  }

  async function handleRequestOTP(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const validationErrors = validateLoginInput(identifier, captchaAnswer, t);
    setFieldErrors(validationErrors);

    if (validationErrors.identifier || validationErrors.captchaAnswer) {
      setNotice(null);
      setError(t("errors.fixFields"));
      return;
    }

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
      setFieldErrors({});
    } catch (err) {
      setError(normalizeLoginError(err, common("requestError"), t));
      await refreshCaptcha({ clearError: false });
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
      setError(normalizeLoginError(err, common("requestError"), t));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="relative flex min-h-screen overflow-hidden bg-background text-foreground">
      <div
        className="pointer-events-none absolute inset-0"
        aria-hidden="true"
        style={{
          backgroundImage:
            "radial-gradient(circle at center, color-mix(in oklab, var(--foreground) 13%, transparent) 1px, transparent 1.4px), radial-gradient(circle at center, color-mix(in oklab, var(--primary) 16%, transparent) 1px, transparent 1.5px)",
          backgroundPosition: "0 0, 18px 18px",
          backgroundSize: "28px 28px, 56px 56px",
        }}
      />
      <div
        className="pointer-events-none absolute inset-0 opacity-80"
        aria-hidden="true"
        style={{
          backgroundImage:
            "linear-gradient(135deg, color-mix(in oklab, var(--primary) 10%, transparent), transparent 36%), linear-gradient(315deg, color-mix(in oklab, var(--chart-3) 9%, transparent), transparent 42%)",
        }}
      />
      <div
        className="pointer-events-none absolute inset-0"
        aria-hidden="true"
        style={{
          background:
            "radial-gradient(circle at 50% 44%, transparent 0%, transparent 28%, color-mix(in oklab, var(--background) 72%, transparent) 74%)",
        }}
      />

      <div className="relative mx-auto flex w-full max-w-4xl flex-col gap-8 px-4 py-5 sm:px-6 lg:px-8">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="relative flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-border bg-card p-1.5 shadow-sm">
              <Image
                src="/brand/mini_logo.svg"
                alt={t("logoAlt")}
                width={40}
                height={40}
                className="h-full w-full object-contain"
                priority
              />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-foreground">{t("brand")}</p>
              <p className="truncate text-xs text-muted-foreground">{t("eyebrow")}</p>
            </div>
          </div>
          <ThemeToggle />
        </header>

        <div className="flex flex-1 items-center justify-center py-4">
          <section className="mx-auto flex w-full max-w-md flex-col gap-4">
            <div className="flex flex-col items-start gap-4">
              <Badge variant="info" className="w-fit">
                <ShieldCheck className="me-1 size-3.5" aria-hidden="true" />
                {t("hero.badge")}
              </Badge>
              <div className="flex flex-col gap-3">
                <h1 className="text-3xl font-semibold leading-10 tracking-normal text-foreground sm:text-4xl sm:leading-tight">
                  {t("hero.title")}
                </h1>
                <AnimatedHeroText
                  prefix={t("hero.animated.prefix")}
                  words={animatedHeroWords}
                  suffix={t("hero.animated.suffix")}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-2 rounded-xl border border-border bg-card/80 p-1 text-xs font-medium shadow-sm backdrop-blur">
              <StepPill
                active={activeStep === "request"}
                icon={<KeyRound className="size-3.5" aria-hidden="true" />}
                label={t("otp.request")}
              />
              <StepPill
                active={activeStep === "verify"}
                icon={<LockKeyhole className="size-3.5" aria-hidden="true" />}
                label={t("otp.verifyTitle")}
              />
            </div>

            <Card className="rounded-2xl border-border/80 bg-card/95 shadow-2xl shadow-primary/10 backdrop-blur">
              <CardContent className="flex flex-col gap-5 p-4 md:p-5">
                {notice && (
                  <Alert variant="success" icon={<CheckCircle aria-hidden="true" />}>
                    <AlertDescription>{notice}</AlertDescription>
                  </Alert>
                )}
                {error && (
                  <Alert variant="destructive" icon={<AlertCircle aria-hidden="true" />}>
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
                  <form className="flex flex-col gap-5" onSubmit={handleRequestOTP} noValidate>
                    <div className="flex flex-col gap-2">
                      <Label htmlFor="identifier">{t("identifier")}</Label>
                      <Input
                        id="identifier"
                        value={identifier}
                        onChange={(event) => {
                          setIdentifier(event.target.value);
                          setFieldErrors((current) => ({ ...current, identifier: undefined }));
                        }}
                        placeholder={t("identifierPlaceholder")}
                        autoComplete="username"
                        autoFocus
                        required
                        dir="ltr"
                        aria-invalid={Boolean(fieldErrors.identifier)}
                        aria-describedby={fieldErrors.identifier ? "identifier-error" : undefined}
                        className="h-11 text-left"
                      />
                      {fieldErrors.identifier && (
                        <p id="identifier-error" className="text-sm leading-6 text-destructive">
                          {fieldErrors.identifier}
                        </p>
                      )}
                    </div>

                    <div className="flex flex-col gap-3">
                      <Label htmlFor="captchaAnswer">{t("captcha.label")}</Label>
                      <div className="flex flex-col gap-3 rounded-xl border border-border bg-muted/35 p-3">
                        <div className="flex items-center gap-3">
                          <Input
                            id="captchaAnswer"
                            dir="ltr"
                            value={captchaAnswer}
                            onChange={(event) => {
                              setCaptchaAnswer(event.target.value);
                              setFieldErrors((current) => ({
                                ...current,
                                captchaAnswer: undefined,
                              }));
                            }}
                            placeholder={t("captcha.answerPlaceholder")}
                            autoComplete="off"
                            required
                            aria-invalid={Boolean(fieldErrors.captchaAnswer)}
                            aria-describedby={
                              fieldErrors.captchaAnswer ? "captcha-answer-error" : undefined
                            }
                            className="h-11 min-w-0 flex-1 text-left"
                          />
                          {captchaImage ? (
                            <Image
                              src={captchaImage}
                              alt={t("captcha.label")}
                              width={170}
                              height={56}
                              unoptimized
                              className="h-12 w-36 shrink-0 rounded-lg bg-card object-contain sm:w-[170px]"
                            />
                          ) : (
                            <div className="h-12 w-36 shrink-0 rounded-lg bg-card sm:w-[170px]" />
                          )}
                          <Button
                            type="button"
                            variant="outline"
                            size="icon"
                            onClick={() => void refreshCaptcha({ clearError: true })}
                            aria-label={t("captcha.refresh")}
                          >
                            <RefreshCw data-icon="inline-start" aria-hidden="true" />
                          </Button>
                        </div>
                        {fieldErrors.captchaAnswer && (
                          <p
                            id="captcha-answer-error"
                            className="text-sm leading-6 text-destructive"
                          >
                            {fieldErrors.captchaAnswer}
                          </p>
                        )}
                      </div>
                    </div>

                    <Button type="submit" size="lg" disabled={loading} aria-busy={loading} className="w-full">
                      {loading ? <Spinner data-icon="inline-start" /> : null}
                      {loading ? t("otp.requesting") : t("otp.request")}
                    </Button>
                  </form>
                ) : (
                  <form className="flex flex-col gap-5" onSubmit={handleVerifyOTP} noValidate>
                    <div className="flex flex-col gap-2">
                      <Label htmlFor="otpCode">{t("otp.code")}</Label>
                      <InputOTP
                        id="otpCode"
                        dir="ltr"
                        maxLength={6}
                        pattern={REGEXP_ONLY_DIGITS}
                        value={otpCode}
                        onChange={setOTPCode}
                        aria-label={t("otp.codePlaceholder")}
                        autoComplete="one-time-code"
                        required
                        disabled={loading}
                        containerClassName="justify-center"
                      >
                        <InputOTPGroup dir="ltr" className="gap-2">
                          {Array.from({ length: 6 }).map((_, index) => (
                            <InputOTPSlot
                              key={index}
                              index={index}
                              className="size-11 rounded-xl border bg-background text-base font-semibold shadow-xs"
                            />
                          ))}
                        </InputOTPGroup>
                      </InputOTP>
                    </div>

                    <div className="flex flex-col-reverse gap-2 sm:flex-row">
                      <Button
                        type="submit"
                        size="lg"
                        disabled={loading || otpCode.length !== 6}
                        aria-busy={loading}
                        className="flex-1"
                      >
                        {loading ? <Spinner data-icon="inline-start" /> : null}
                        {loading ? t("otp.verifying") : t("otp.verify")}
                        {!loading && <ArrowLeft data-icon="inline-end" aria-hidden="true" />}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="lg"
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
          </section>
        </div>
      </div>
    </main>
  );
}

function AnimatedHeroText({
  prefix,
  words,
  suffix,
}: {
  prefix: string;
  words: string[];
  suffix: string;
}) {
  const [wordIndex, setWordIndex] = useState(0);
  const [isVisible, setIsVisible] = useState(true);
  const currentWord = words[wordIndex] ?? "";

  useEffect(() => {
    if (words.length <= 1) {
      return;
    }

    let timeoutID: number | undefined;
    const interval = window.setInterval(() => {
      setIsVisible(false);

      timeoutID = window.setTimeout(() => {
        setWordIndex((current) => (current + 1) % words.length);
        setIsVisible(true);
      }, ANIMATED_HERO_TRANSITION_MS);
    }, ANIMATED_HERO_INTERVAL_MS);

    return () => {
      window.clearInterval(interval);
      if (timeoutID) {
        window.clearTimeout(timeoutID);
      }
    };
  }, [words.length]);

  return (
    <p className="flex max-w-full items-center justify-center gap-1.5 overflow-hidden whitespace-nowrap text-[clamp(0.875rem,3.2vw,1.25rem)] font-semibold leading-8 text-foreground sm:gap-2">
      <span>{prefix}</span>
      <span
        className={cn(
          "inline-flex min-w-[6.75rem] justify-center rounded-lg bg-primary/10 px-2 py-1 text-primary ring-1 ring-primary/15 transition-all duration-300 motion-reduce:transition-none sm:min-w-[7.75rem] sm:px-2.5",
          isVisible
            ? "translate-y-0 opacity-100 blur-0"
            : "translate-y-1 opacity-0 blur-sm",
        )}
      >
        {currentWord}
      </span>
      <span>{suffix}</span>
    </p>
  );
}

function StepPill({
  active,
  icon,
  label,
}: {
  active: boolean;
  icon: ReactNode;
  label: string;
}) {
  return (
    <div
      className={cn(
        "flex items-center justify-center gap-2 rounded-lg px-3 py-2.5",
        active
          ? "bg-primary text-primary-foreground shadow-sm"
          : "text-muted-foreground",
      )}
    >
      {icon}
      <span className="truncate">{label}</span>
    </div>
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

function validateLoginInput(
  identifier: string,
  captchaAnswer: string,
  t: LoginTranslator,
): LoginFieldErrors {
  const errors: LoginFieldErrors = {};
  const identifierValue = identifier.trim();
  const captchaValue = captchaAnswer.trim();

  if (!identifierValue) {
    errors.identifier = t("errors.identifierRequired");
  } else if (!normalizeIdentifier(identifierValue)) {
    errors.identifier = t("errors.identifierInvalid");
  }

  if (!captchaValue) {
    errors.captchaAnswer = t("errors.captchaRequired");
  }

  return errors;
}

function normalizeIdentifier(raw: string) {
  const value = raw.trim();
  if (!value) {
    return null;
  }

  const emailCandidate = value.toLowerCase();
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailCandidate)) {
    return { channel: "email", normalized: emailCandidate };
  }

  const compact = value.replace(/[\s\-()]/g, "");
  if (/^\+?\d{10,15}$/.test(compact)) {
    return { channel: "mobile", normalized: compact };
  }

  return null;
}

function normalizeLoginError(
  error: unknown,
  fallback: string,
  t: LoginTranslator,
): string {
  if (error instanceof Error) {
    const message = error.message;
    const normalized = message.toLowerCase();

    if (
      normalized.includes("captcha") ||
      message.includes("کپچا") ||
      message.includes("captcha_id") ||
      message.includes("captcha_answer")
    ) {
      return t("errors.captchaInvalidOrExpired");
    }

    if (
      normalized.includes("tenant member not found") ||
      normalized.includes("not found") ||
      normalized.includes("do not have access") ||
      message.includes("You do not have access to this tenant")
    ) {
      return t("errors.identifierNotRegistered");
    }

    if (message.includes("ایمیل یا شماره همراه")) {
      return t("errors.identifierInvalid");
    }

    return message;
  }
  return fallback;
}
