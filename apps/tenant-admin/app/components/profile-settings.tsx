"use client";

import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";
import { CheckCircle, Mail, Send, Smartphone } from "lucide-react";

import { PageHeader, SettingsCard } from "@/components/app-ui";
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

      setMessage(t("messages.sent", { destination: result.destination_masked }));
    } catch (err) {
      if (
        !mobileAuthUIEnabled &&
        field === "email" &&
        err instanceof Error &&
        /mobile|phone|موبایل|همراه/i.test(err.message)
      ) {
        setError(t("errors.emailInvalid"));
      } else {
        setError(normalizeError(err, common("requestError")));
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
      setMessage(field === "email" ? t("messages.emailSuccess") : t("messages.mobileSuccess"));
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
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

      {message ? (
        <Alert variant="success" icon={<CheckCircle aria-hidden="true" />}>
          <AlertDescription>{message}</AlertDescription>
        </Alert>
      ) : null}
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

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
          />
        ) : null}
      </section>
    </div>
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
