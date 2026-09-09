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

import { GodarAILogo } from "@/components/app-ui/GodarAILogo";

import { LocaleToggle } from "@/components/locale-toggle";
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
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";
import {
  refreshLoginCaptcha,
  requestLoginOTP,
  verifyLoginOTP,
  type CaptchaChallenge,
  type OTPRequestResponse,
} from "./api";

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

  const [identifier, setIdentifier] = useState(
    mobileAuthUIEnabled || normalizeIdentifier(initialIdentifier)?.channel === "email"
      ? initialIdentifier
      : "",
  );
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
      const challenge = await refreshLoginCaptcha();
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
      const result = await requestLoginOTP({
        identifier,
        captchaID,
        captchaAnswer,
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
      await verifyLoginOTP(otpRequest.request_id, otpCode);
      window.location.href = redirectTo;
    } catch (err) {
      setError(normalizeLoginError(err, common("requestError"), t));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="theme-brand-grid relative flex min-h-screen overflow-hidden bg-background text-foreground">
      <div className="relative mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-8 px-4 py-5 sm:px-6 lg:px-8">
        <header className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="relative flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-primary-subtle-border bg-card p-1.5 shadow-sm">
              <GodarAILogo
                alt={t("logoAlt")}
                className="h-full w-auto max-w-full object-contain"
                priority
                sizes="2.75rem"
              />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-foreground">{t("brand")}</p>
              <p className="truncate text-xs text-muted-foreground">{t("eyebrow")}</p>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <LocaleToggle />
            <ThemeToggle />
          </div>
        </header>

        <div className="flex flex-1 items-center justify-center py-4">
          <section className="mx-auto flex w-full max-w-md flex-col gap-4">
            <div className="flex flex-col items-center gap-4 text-center">
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

            <div className="grid grid-cols-2 gap-2 rounded-lg border border-card-border bg-card p-1 text-xs font-medium shadow-xs">
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

            <Card className="theme-accent-line overflow-hidden shadow-md">
              <CardContent className="flex flex-col gap-5 p-4 md:p-5">
                {!otpRequest ? <SocialLoginButtons redirectTo={redirectTo} labels={{
                  title: t("social.title"), google: t("social.google"), github: t("social.github"), microsoft: t("social.microsoft"), divider: t("social.divider"),
                }} /> : null}
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
                      <Label htmlFor="identifier">
                        {mobileAuthUIEnabled ? t("identifier") : t("identifierEmailOnly")}
                      </Label>
                      <Input
                        id="identifier"
                        value={identifier}
                        onChange={(event) => {
                          setIdentifier(event.target.value);
                          setFieldErrors((current) => ({ ...current, identifier: undefined }));
                        }}
                        placeholder={
                          mobileAuthUIEnabled
                            ? t("identifierPlaceholder")
                            : t("identifierEmailOnlyPlaceholder")
                        }
                        autoComplete="username"
                        autoFocus
                        required
                        dir="ltr"
                        aria-invalid={Boolean(fieldErrors.identifier)}
                        aria-describedby={fieldErrors.identifier ? "identifier-error" : undefined}
                        className="h-11 text-start"
                      />
                      {fieldErrors.identifier && (
                        <p id="identifier-error" className="text-sm leading-6 text-destructive">
                          {fieldErrors.identifier}
                        </p>
                      )}
                    </div>

                    <div className="flex flex-col gap-3">
                      <Label htmlFor="captchaAnswer">{t("captcha.label")}</Label>
                      <div className="flex flex-col gap-3 rounded-lg border border-border bg-neutral-soft/55 p-3">
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
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
                            className="h-11 min-w-0 flex-1 text-start"
                          />
                          <div className="flex items-center gap-3 sm:contents">
                            {captchaImage ? (
                              <Image
                                src={captchaImage}
                                alt={t("captcha.label")}
                                width={170}
                                height={56}
                                unoptimized
                                className="h-12 min-w-0 flex-1 rounded-md bg-card object-contain sm:w-40 sm:flex-none"
                              />
                            ) : (
                              <div className="h-12 min-w-0 flex-1 rounded-md bg-card sm:w-40 sm:flex-none" />
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
                              className="size-11 rounded-md border bg-background text-base font-semibold shadow-xs first:rounded-md last:rounded-md"
                            />
                          ))}
                        </InputOTPGroup>
                      </InputOTP>
                    </div>

                    <div className="flex flex-col gap-2">
                      <Button
                        type="submit"
                        size="lg"
                        disabled={loading || otpCode.length !== 6}
                        aria-busy={loading}
                        className="w-full"
                      >
                        {loading ? <Spinner data-icon="inline-start" /> : null}
                        {loading ? t("otp.verifying") : t("otp.verify")}
                        {!loading && <ArrowLeft data-icon="inline-end" aria-hidden="true" />}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="lg"
                        className="w-full"
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

function SocialLoginButtons({ redirectTo, labels }: { redirectTo: string; labels: Record<"title" | "google" | "github" | "microsoft" | "divider", string> }) {
  const providers = [
    { id: "google", label: labels.google, icon: <span className="text-base font-bold">G</span> },
    { id: "github", label: labels.github, icon: <span className="text-xs font-bold" aria-hidden="true">GH</span> },
    { id: "microsoft", label: labels.microsoft, icon: <span className="grid grid-cols-2 gap-0.5" aria-hidden="true">{[0, 1, 2, 3].map((item) => <i key={item} className="size-1.5 bg-current" />)}</span> },
  ];
  return (
    <section className="flex flex-col gap-3" aria-label={labels.title}>
      <div className="grid gap-2 sm:grid-cols-3">
        {providers.map((provider) => {
          const query = new URLSearchParams({ provider: provider.id, return_to: redirectTo });
          return <Button key={provider.id} asChild variant="outline" className="w-full"><a href={`/auth/social/start?${query.toString()}`} dir="ltr">{provider.icon}{provider.label}</a></Button>;
        })}
      </div>
      <div className="flex items-center gap-3 text-xs text-muted-foreground" aria-hidden="true">
        <span className="h-px flex-1 bg-border" /><span>{labels.divider}</span><span className="h-px flex-1 bg-border" />
      </div>
    </section>
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
    <p className="flex max-w-full items-center justify-center gap-1.5 overflow-hidden whitespace-nowrap text-sm font-semibold leading-8 text-foreground sm:gap-2 sm:text-xl">
      <span>{prefix}</span>
      <span
        className={cn(
          "inline-flex min-w-[6.75rem] justify-center rounded-md bg-primary-soft px-2 py-1 text-primary-soft-foreground ring-1 ring-primary-subtle-border transition-all duration-300 motion-reduce:transition-none sm:min-w-[7.75rem] sm:px-2.5",
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

function validateLoginInput(
  identifier: string,
  captchaAnswer: string,
  t: LoginTranslator,
): LoginFieldErrors {
  const errors: LoginFieldErrors = {};
  const identifierValue = identifier.trim();
  const captchaValue = captchaAnswer.trim();

  if (!identifierValue) {
    errors.identifier = mobileAuthUIEnabled
      ? t("errors.identifierRequired")
      : t("errors.identifierEmailOnlyRequired");
  } else if (!normalizeIdentifier(identifierValue)) {
    errors.identifier = mobileAuthUIEnabled
      ? t("errors.identifierInvalid")
      : t("errors.identifierEmailOnlyInvalid");
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
      return mobileAuthUIEnabled
        ? t("errors.identifierNotRegistered")
        : t("errors.identifierEmailOnlyNotRegistered");
    }

    if (message.includes("ایمیل یا شماره همراه")) {
      return mobileAuthUIEnabled
        ? t("errors.identifierInvalid")
        : t("errors.identifierEmailOnlyInvalid");
    }

    return message;
  }
  return fallback;
}
