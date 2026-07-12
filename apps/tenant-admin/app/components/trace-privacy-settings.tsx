"use client";

import { useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  Ban,
  FileStack,
  LockKeyhole,
  ScanSearch,
  type LucideIcon,
} from "lucide-react";

import { SettingsAuditMeta, StatusOptionGroup } from "@/components/app-ui";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type TracePrivacySettings = {
  tenant_id?: string;
  capture_policy: "none" | "metadata-only" | "sampled-redacted" | "full-secure";
  updated_at?: string;
  updated_by_user_id?: string;
};

type TracePrivacySettingsProps = {
  initialSettings: TracePrivacySettings | null;
  initialError?: string | null;
  showHeader?: boolean;
  surface?: "card" | "plain";
  currentUserID?: string;
  currentUserLabel?: string;
};

const POLICIES: Array<{
  value: TracePrivacySettings["capture_policy"];
  key: "none" | "metadataOnly" | "sampledRedacted" | "fullSecure";
  icon: LucideIcon;
  tone: "neutral" | "info" | "warning" | "success";
}> = [
  {
    value: "none",
    key: "none",
    icon: Ban,
    tone: "neutral",
  },
  {
    value: "metadata-only",
    key: "metadataOnly",
    icon: FileStack,
    tone: "info",
  },
  {
    value: "sampled-redacted",
    key: "sampledRedacted",
    icon: ScanSearch,
    tone: "warning",
  },
  {
    value: "full-secure",
    key: "fullSecure",
    icon: LockKeyhole,
    tone: "success",
  },
];

export default function TracePrivacySettingsPanel(props: TracePrivacySettingsProps) {
  const t = useTranslations("settings.tracePrivacy");
  const locale = useLocale();
  const surface = props.surface ?? "card";
  const showHeader = props.showHeader ?? true;
  const [capturePolicy, setCapturePolicy] = useState<TracePrivacySettings["capture_policy"]>(
    props.initialSettings?.capture_policy ?? "metadata-only",
  );
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(props.initialError ?? null);
  const [updatedAt, setUpdatedAt] = useState<string | null>(props.initialSettings?.updated_at ?? null);
  const [updatedBy, setUpdatedBy] = useState<string | null>(props.initialSettings?.updated_by_user_id ?? null);

  async function save() {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const payload = await requestJSON<TracePrivacySettings>("/api/settings/trace-privacy", {
        method: "PUT",
        body: JSON.stringify({ capture_policy: capturePolicy }),
      });
      setUpdatedAt(payload.updated_at ?? null);
      setUpdatedBy(payload.updated_by_user_id ?? props.currentUserID ?? null);
      setMessage(t("messages.updated"));
    } catch (err) {
      setError(normalizeError(err, t("messages.updateFailed")));
    } finally {
      setSaving(false);
    }
  }

  const updatedByDisplay = formatUpdatedBy({
    updatedBy,
    currentUserID: props.currentUserID,
    currentUserLabel: props.currentUserLabel,
    fallback: t("notSet"),
  });

  const content = (
    <div className="flex flex-col gap-4">
        {message ? (
          <Alert className="surface-primary-soft">
            <AlertDescription>{message}</AlertDescription>
          </Alert>
        ) : null}
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        <StatusOptionGroup
          label={t("title")}
          value={capturePolicy}
          onValueChange={setCapturePolicy}
          disabled={saving}
          options={POLICIES.map((item) => ({
            ...item,
            badge: item.value,
            title: t(`policies.${item.key}.title`),
            description: t(`policies.${item.key}.description`),
          }))}
        />

        <SettingsAuditMeta updatedAt={updatedAt} updatedBy={updatedByDisplay} updatedAtLabel={t("updatedAtLabel")} updatedByLabel={t("updatedByLabel")} fallback={t("notSet")} locale={locale} />

        <Button onClick={() => void save()} disabled={saving}>
          {saving ? t("saving") : t("save")}
        </Button>
    </div>
  );

  if (surface === "plain") {
    return (
      <div className="flex flex-col gap-4">
        {showHeader ? (
          <div className="flex flex-col gap-1">
            <h2 className="text-base font-semibold">{t("title")}</h2>
            <p className="text-sm leading-6 text-muted-foreground">{t("description")}</p>
          </div>
        ) : null}
        {content}
      </div>
    );
  }

  return (
    <Card>
      {showHeader ? (
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>
            {t("description")}
          </CardDescription>
        </CardHeader>
      ) : null}
      <CardContent className={cn("flex flex-col gap-4", !showHeader && "pt-6")}>
        {content}
      </CardContent>
    </Card>
  );
}

function formatUpdatedBy({
  updatedBy,
  currentUserID,
  currentUserLabel,
  fallback,
}: {
  updatedBy: string | null;
  currentUserID?: string;
  currentUserLabel?: string;
  fallback: string;
}) {
  if (!updatedBy) return fallback;
  if (currentUserID && updatedBy === currentUserID) return currentUserLabel ?? fallback;
  if (looksLikeUUID(updatedBy)) return fallback;
  return updatedBy;
}

function looksLikeUUID(value: string) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  }
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessage(payload: unknown): string | null {
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
