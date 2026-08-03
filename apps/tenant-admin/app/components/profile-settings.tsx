"use client";

import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";
import { Link2, Mail, Send, Smartphone } from "lucide-react";

import { PageHeader, SettingsCard, StatusBadge } from "@/components/app-ui";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";

type UserProfile = {
  id: string;
  email?: string;
  mobile?: string;
  socialProviders?: string[];
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
  const t = useTranslations("settings");
  const common = useTranslations("common");

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
  const [fieldError, setFieldError] = useState<{
    field: "email" | "mobile";
    message: string;
  } | null>(null);
  const [socialLoading, setSocialLoading] = useState<string | null>(null);
  const connectedSocialProviders = new Set(user.socialProviders ?? []);

  async function connectSocialAccount(provider: "google" | "github" | "microsoft") {
    setSocialLoading(provider);
    try {
      const result = await requestJSON<{ redirect_url: string }>("/api/auth/social/link-ticket", {
        method: "POST",
        body: JSON.stringify({ provider }),
      });
      window.location.assign(result.redirect_url);
    } catch (err) {
      showFeedbackToast({
        kind: "error",
        title: common("requestError"),
        description: normalizeError(err, common("requestError")),
      });
      setSocialLoading(null);
    }
  }

  async function handleRequestChange(field: "email" | "mobile") {
    setLoadingField(field);
    setFieldError(null);

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

    } catch (err) {
      if (
        !mobileAuthUIEnabled &&
        field === "email" &&
        err instanceof Error &&
        /mobile|phone|موبایل|همراه/i.test(err.message)
      ) {
        setFieldError({ field, message: t("errors.emailInvalid") });
      } else {
        setFieldError({ field, message: normalizeError(err, common("requestError")) });
      }
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
    setFieldError(null);

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
      setUser((current) => ({
        ...result.user,
        socialProviders: result.user.socialProviders ?? current.socialProviders,
      }));

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
      showFeedbackToast({
        kind: "success",
        title: field === "email" ? t("messages.emailSuccess") : t("messages.mobileSuccess"),
      });
    } catch (err) {
      setFieldError({ field, message: normalizeError(err, common("requestError")) });
    } finally {
      setLoadingField(null);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={<>{t("subtitle")} - {t("role")}: <b>{user.role}</b></>}
      />

      <section className={cn("grid gap-6", mobileAuthUIEnabled && "lg:grid-cols-2")}>
        <ContactChangeCard
          title={t("sections.email")}
          description={t("email.description")}
          icon={Mail}
          currentLabel={t("currentValue")}
          currentValue={user.email ?? common("notSet")}
          fieldLabel={t("newEmail")}
          fieldDescription={t("email.fieldDescription")}
          inputID="newEmail"
          inputType="email"
          inputMode="email"
          autoComplete="email"
          value={emailState.value}
          placeholder="name@example.com"
          isLoading={loadingField === "email"}
          onValueChange={(value) =>
            setEmailState((current) => ({
              ...current,
              value,
            }))
          }
          onRequestCode={() => void handleRequestChange("email")}
          sendCodeLabel={t("sendCode")}
          requestID={emailState.requestID}
          destinationMasked={emailState.destinationMasked}
          debugOTP={emailState.debugOTP}
          otpValue={emailState.otpCode}
          onOTPChange={(otpCode) =>
            setEmailState((current) => ({
              ...current,
              otpCode,
            }))
          }
          onVerify={(event) => void handleVerifyChange(event, "email")}
          otpSentLabel={t("otpSentTo", { destination: emailState.destinationMasked })}
          otpCodeLabel={t("otpCode")}
          otpCodePlaceholder={t("otpCodePlaceholder")}
          devCodeLabel={t("devCode")}
          confirmLabel={t("confirmEmail")}
          error={fieldError?.field === "email" ? fieldError.message : undefined}
        />

        {mobileAuthUIEnabled ? (
          <ContactChangeCard
            title={t("sections.mobile")}
            description={t("mobile.description")}
            icon={Smartphone}
            currentLabel={t("currentValue")}
            currentValue={user.mobile ?? common("notSet")}
            fieldLabel={t("newMobile")}
            fieldDescription={t("mobile.fieldDescription")}
            inputID="newMobile"
            inputType="tel"
            inputMode="tel"
            autoComplete="tel"
            value={mobileState.value}
            placeholder="+98912..."
            isLoading={loadingField === "mobile"}
            onValueChange={(value) =>
              setMobileState((current) => ({
                ...current,
                value,
              }))
            }
            onRequestCode={() => void handleRequestChange("mobile")}
            sendCodeLabel={t("sendCode")}
            requestID={mobileState.requestID}
            destinationMasked={mobileState.destinationMasked}
            debugOTP={mobileState.debugOTP}
            otpValue={mobileState.otpCode}
            onOTPChange={(otpCode) =>
              setMobileState((current) => ({
                ...current,
                otpCode,
              }))
            }
            onVerify={(event) => void handleVerifyChange(event, "mobile")}
            otpSentLabel={t("otpSentTo", { destination: mobileState.destinationMasked })}
            otpCodeLabel={t("otpCode")}
            otpCodePlaceholder={t("otpCodePlaceholder")}
            devCodeLabel={t("devCode")}
            confirmLabel={t("confirmMobile")}
            error={fieldError?.field === "mobile" ? fieldError.message : undefined}
          />
        ) : null}
      </section>

      <SettingsCard
        title={<span className="flex items-center gap-2"><span className="flex size-9 items-center justify-center rounded-md bg-muted text-muted-foreground"><Link2 aria-hidden="true" /></span><span>{t("social.title")}</span></span>}
        description={t("social.description")}
        className="max-w-none"
      >
        <div className="grid gap-2 sm:grid-cols-3">
          {([
            ["google", "Google", <GoogleBrandIcon key="g" />],
            ["github", "GitHub", <GithubBrandIcon key="gh" />],
            ["microsoft", "Microsoft", <MicrosoftBrandIcon key="ms" />],
          ] as const).map(([provider, label, icon]) => {
            const isConnected = connectedSocialProviders.has(provider);

            return (
              <div
                key={provider}
                className={cn(
                  "flex min-h-14 items-center justify-between gap-3 rounded-md border border-border bg-card p-3",
                  isConnected && "border-success-soft-border bg-success-soft",
                )}
              >
                <span className="flex min-w-0 items-center gap-2" dir="ltr">
                  {icon}
                  <span className="truncate text-sm font-medium">{label}</span>
                </span>
                {isConnected ? (
                  <StatusBadge status="success">{t("social.connected")}</StatusBadge>
                ) : (
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => void connectSocialAccount(provider)}
                    disabled={socialLoading !== null}
                    aria-busy={socialLoading === provider}
                  >
                    {socialLoading === provider ? <Spinner data-icon="inline-start" /> : null}
                    {t("social.connect")}
                  </Button>
                )}
              </div>
            );
          })}
        </div>
      </SettingsCard>
    </div>
  );
}

function GoogleBrandIcon() {
  return (
    <svg className="size-5 shrink-0" viewBox="0 0 48 48" aria-hidden="true">
      <path
        fill="#FFC107"
        d="M43.611 20H24v8h11.303C33.654 32.657 29.223 36 24 36c-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-8.955 20-20c0-1.341-.138-2.65-.389-4Z"
      />
      <path
        fill="#FF3D00"
        d="m6.306 14.691 6.571 4.819C14.655 15.108 18.961 12 24 12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4c-7.682 0-14.35 4.337-17.694 10.691Z"
      />
      <path
        fill="#4CAF50"
        d="M24 44c5.166 0 9.86-1.977 13.409-5.192l-6.19-5.238A11.91 11.91 0 0 1 24 36c-5.202 0-9.619-3.317-11.283-7.946l-6.522 5.025C9.505 39.556 16.227 44 24 44Z"
      />
      <path
        fill="#1976D2"
        d="M43.611 20H24v8h11.303a12.04 12.04 0 0 1-4.087 5.571l.003-.002 6.19 5.238C36.971 39.205 44 34 44 24c0-1.341-.138-2.65-.389-4Z"
      />
    </svg>
  );
}

function GithubBrandIcon() {
  return (
    <svg className="size-5 shrink-0" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 .7a12 12 0 0 0-3.79 23.39c.6.11.82-.26.82-.58v-2.23c-3.34.73-4.04-1.42-4.04-1.42-.55-1.39-1.33-1.76-1.33-1.76-1.09-.74.08-.73.08-.73 1.2.09 1.84 1.24 1.84 1.24 1.07 1.83 2.81 1.3 3.5.99.11-.78.42-1.3.76-1.6-2.67-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23A11.5 11.5 0 0 1 12 6.5c1.02 0 2.04.14 3 .4 2.29-1.55 3.3-1.23 3.3-1.23.66 1.66.24 2.88.12 3.18.77.84 1.23 1.91 1.23 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22v3.3c0 .32.22.7.82.58A12 12 0 0 0 12 .7Z"
      />
    </svg>
  );
}

function MicrosoftBrandIcon() {
  return (
    <svg className="size-5 shrink-0" viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#F25022" d="M1 1h10v10H1z" />
      <path fill="#7FBA00" d="M13 1h10v10H13z" />
      <path fill="#00A4EF" d="M1 13h10v10H1z" />
      <path fill="#FFB900" d="M13 13h10v10H13z" />
    </svg>
  );
}

type ContactChangeCardProps = {
  title: string;
  description: string;
  icon: typeof Mail;
  currentLabel: string;
  currentValue: string;
  fieldLabel: string;
  fieldDescription: string;
  inputID: string;
  inputType: "email" | "tel";
  inputMode: "email" | "tel";
  autoComplete: string;
  value: string;
  placeholder: string;
  isLoading: boolean;
  onValueChange: (value: string) => void;
  onRequestCode: () => void;
  sendCodeLabel: string;
  requestID: string;
  destinationMasked: string;
  debugOTP?: string;
  otpValue: string;
  onOTPChange: (value: string) => void;
  onVerify: (event: FormEvent<HTMLFormElement>) => void;
  otpSentLabel: string;
  otpCodeLabel: string;
  otpCodePlaceholder: string;
  devCodeLabel: string;
  confirmLabel: string;
  error?: string;
};

function ContactChangeCard(props: ContactChangeCardProps) {
  const Icon = props.icon;

  return (
    <SettingsCard
      title={
        <span className="flex items-center gap-2">
          <span className="flex size-9 items-center justify-center rounded-md bg-muted text-muted-foreground">
            <Icon aria-hidden="true" />
          </span>
          <span>{props.title}</span>
        </span>
      }
      description={props.description}
      className="max-w-none"
    >
      <div className="rounded-md border border-border bg-muted/30 p-3">
        <p className="text-xs font-medium text-muted-foreground">{props.currentLabel}</p>
        <p dir="ltr" className="mt-1 truncate text-sm font-semibold text-foreground">
          {props.currentValue}
        </p>
      </div>

      <FieldGroup>
        <Field>
          <FieldLabel htmlFor={props.inputID}>{props.fieldLabel}</FieldLabel>
          <Input
            id={props.inputID}
            type={props.inputType}
            inputMode={props.inputMode}
            autoComplete={props.autoComplete}
            dir="ltr"
            value={props.value}
            onChange={(event) => props.onValueChange(event.target.value)}
            placeholder={props.placeholder}
          />
          <FieldDescription>{props.fieldDescription}</FieldDescription>
        </Field>
      </FieldGroup>

      <Button
        type="button"
        className="w-full sm:w-fit"
        onClick={props.onRequestCode}
        disabled={props.isLoading || !props.value.trim()}
        aria-busy={props.isLoading}
      >
        {props.isLoading ? <Spinner data-icon="inline-start" /> : <Send data-icon="inline-start" aria-hidden="true" />}
        {props.sendCodeLabel}
      </Button>

      {props.error ? (
        <Alert variant="destructive">
          <AlertDescription>{props.error}</AlertDescription>
        </Alert>
      ) : null}

      {props.requestID ? (
        <form className="flex flex-col gap-4 rounded-md border border-border bg-background p-4" onSubmit={props.onVerify}>
          <Alert>
            <AlertDescription>{props.otpSentLabel}</AlertDescription>
          </Alert>
          {props.debugOTP ? (
            <Alert variant="warning">
              <AlertTitle>{props.devCodeLabel}</AlertTitle>
              <AlertDescription>
                <code dir="ltr" className="font-mono text-sm">
                  {props.debugOTP}
                </code>
              </AlertDescription>
            </Alert>
          ) : null}
          <FieldGroup className="gap-3">
            <Field>
              <FieldLabel htmlFor={`${props.inputID}OTP`}>{props.otpCodeLabel}</FieldLabel>
              <Input
                id={`${props.inputID}OTP`}
                dir="ltr"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={props.otpValue}
                onChange={(event) => props.onOTPChange(event.target.value)}
                placeholder={props.otpCodePlaceholder}
              />
            </Field>
          </FieldGroup>
          <Button type="submit" variant="secondary" disabled={props.isLoading || !props.otpValue.trim()} aria-busy={props.isLoading}>
            {props.isLoading ? <Spinner data-icon="inline-start" /> : null}
            {props.confirmLabel}
          </Button>
        </form>
      ) : null}
    </SettingsCard>
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
