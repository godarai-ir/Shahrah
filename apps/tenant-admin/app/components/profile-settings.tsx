"use client";

import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";

import { FormSection, PageHeader, SettingsCard } from "@/components/app-ui";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

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
      setError(normalizeError(err, common("requestError")));
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
        <Alert className="border-primary/30 bg-primary/5">
          <AlertDescription>{message}</AlertDescription>
        </Alert>
      ) : null}
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-6 md:grid-cols-2">
        <SettingsCard
          title={t("sections.email")}
          description={<>{t("currentValue")}: {user.email ?? common("notSet")}</>}
        >
          <FormSection title={t("sections.email")} description={t("newEmail")} className="border-0 p-0 md:grid-cols-1 md:p-0">
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="newEmail">{t("newEmail")}</FieldLabel>
                <Input
                  id="newEmail"
                  value={emailState.value}
                  onChange={(event) =>
                    setEmailState((current) => ({
                      ...current,
                      value: event.target.value,
                    }))
                  }
                  placeholder="name@example.com"
                />
              </Field>
            </FieldGroup>
            <Button
              className="w-fit"
              type="button"
              onClick={() => void handleRequestChange("email")}
              disabled={loadingField === "email"}
            >
              {t("sendCode")}
            </Button>

            {emailState.requestID ? (
              <form className="grid gap-3" onSubmit={(event) => void handleVerifyChange(event, "email")}>
                <p className="text-xs text-muted-foreground">
                  {t("otpSentTo", { destination: emailState.destinationMasked })}
                </p>
                {emailState.debugOTP ? (
                  <Alert className="tone-warning">
                    <AlertDescription>
                      {t("devCode")}: <code>{emailState.debugOTP}</code>
                    </AlertDescription>
                  </Alert>
                ) : null}
                <FieldGroup className="gap-3">
                  <Field>
                    <FieldLabel htmlFor="emailOTP">{t("otpCode")}</FieldLabel>
                    <Input
                      id="emailOTP"
                      value={emailState.otpCode}
                      onChange={(event) =>
                        setEmailState((current) => ({
                          ...current,
                          otpCode: event.target.value,
                        }))
                      }
                      placeholder={t("otpCodePlaceholder")}
                    />
                  </Field>
                </FieldGroup>
                <Button type="submit" variant="secondary" disabled={loadingField === "email"}>
                  {t("confirmEmail")}
                </Button>
              </form>
            ) : null}
          </FormSection>
        </SettingsCard>

        <SettingsCard
          title={t("sections.mobile")}
          description={<>{t("currentValue")}: {user.mobile ?? common("notSet")}</>}
        >
          <FormSection title={t("sections.mobile")} description={t("newMobile")} className="border-0 p-0 md:grid-cols-1 md:p-0">
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="newMobile">{t("newMobile")}</FieldLabel>
                <Input
                  id="newMobile"
                  value={mobileState.value}
                  onChange={(event) =>
                    setMobileState((current) => ({
                      ...current,
                      value: event.target.value,
                    }))
                  }
                  placeholder="+98912..."
                />
              </Field>
            </FieldGroup>
            <Button
              className="w-fit"
              type="button"
              onClick={() => void handleRequestChange("mobile")}
              disabled={loadingField === "mobile"}
            >
              {t("sendCode")}
            </Button>

            {mobileState.requestID ? (
              <form className="grid gap-3" onSubmit={(event) => void handleVerifyChange(event, "mobile")}>
                <p className="text-xs text-muted-foreground">
                  {t("otpSentTo", { destination: mobileState.destinationMasked })}
                </p>
                {mobileState.debugOTP ? (
                  <Alert className="tone-warning">
                    <AlertDescription>
                      {t("devCode")}: <code>{mobileState.debugOTP}</code>
                    </AlertDescription>
                  </Alert>
                ) : null}
                <FieldGroup className="gap-3">
                  <Field>
                    <FieldLabel htmlFor="mobileOTP">{t("otpCode")}</FieldLabel>
                    <Input
                      id="mobileOTP"
                      value={mobileState.otpCode}
                      onChange={(event) =>
                        setMobileState((current) => ({
                          ...current,
                          otpCode: event.target.value,
                        }))
                      }
                      placeholder={t("otpCodePlaceholder")}
                    />
                  </Field>
                </FieldGroup>
                <Button type="submit" variant="secondary" disabled={loadingField === "mobile"}>
                  {t("confirmMobile")}
                </Button>
              </form>
            ) : null}
          </FormSection>
        </SettingsCard>
      </section>
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

function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}
