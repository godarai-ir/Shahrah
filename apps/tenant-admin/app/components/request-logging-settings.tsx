"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Ban,
  Database,
  Eye,
  FileSearch,
  type LucideIcon,
} from "lucide-react";

import { StatusOptionGroup } from "@/components/app-ui";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export type RequestLoggingSettings = {
  tenant_id?: string;
  logging_mode: "none" | "metadata-only" | "redacted-content" | "full-content";
  updated_at?: string;
  updated_by_user_id?: string;
};

type RequestLoggingSettingsProps = {
  initialSettings: RequestLoggingSettings | null;
  initialError?: string | null;
  active?: boolean;
  showHeader?: boolean;
  surface?: "card" | "plain";
};

const MODES: Array<{
  value: RequestLoggingSettings["logging_mode"];
  key: "none" | "metadataOnly" | "redactedContent" | "fullContent";
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
    icon: Database,
    tone: "info",
  },
  {
    value: "redacted-content",
    key: "redactedContent",
    icon: FileSearch,
    tone: "warning",
  },
  {
    value: "full-content",
    key: "fullContent",
    icon: Eye,
    tone: "success",
  },
];

export default function RequestLoggingSettingsPanel(props: RequestLoggingSettingsProps) {
  const t = useTranslations("settings.requestLogging");
  const surface = props.surface ?? "card";
  const showHeader = props.showHeader ?? true;
  const [loggingMode, setLoggingMode] = useState<RequestLoggingSettings["logging_mode"]>(
    props.initialSettings?.logging_mode ?? "metadata-only",
  );
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(props.initialError ?? null);
  const [updatedAt, setUpdatedAt] = useState<string | null>(props.initialSettings?.updated_at ?? null);
  const [updatedBy, setUpdatedBy] = useState<string | null>(props.initialSettings?.updated_by_user_id ?? null);
  const [hasResolvedInitialLoad, setHasResolvedInitialLoad] = useState(Boolean(props.initialSettings));

  useEffect(() => {
    if (!props.active || hasResolvedInitialLoad || props.initialSettings) return;

    let cancelled = false;
    queueMicrotask(() => {
      if (cancelled) return;
      setLoading(true);
      setError(null);
    });

    void requestJSON<RequestLoggingSettings>("/api/settings/request-logging")
      .then((payload) => {
        if (cancelled) return;
        applySettings(payload);
        setHasResolvedInitialLoad(true);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(normalizeError(err, t("messages.updateFailed")));
        setHasResolvedInitialLoad(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [hasResolvedInitialLoad, props.active, props.initialSettings, t]);

  function applySettings(settings: RequestLoggingSettings) {
    setLoggingMode(settings.logging_mode ?? "metadata-only");
    setUpdatedAt(settings.updated_at ?? null);
    setUpdatedBy(settings.updated_by_user_id ?? null);
  }

  async function save() {
    setSaving(true);
    setError(null);
    setMessage(null);

    try {
      const payload = await requestJSON<RequestLoggingSettings>("/api/settings/request-logging", {
        method: "PUT",
        body: JSON.stringify({ logging_mode: loggingMode }),
      });
      applySettings(payload);
      setMessage(t("messages.updated"));
    } catch (err) {
      setError(normalizeError(err, t("messages.updateFailed")));
    } finally {
      setSaving(false);
    }
  }

  const content = (
    <div className="flex flex-col gap-4">
      {loading && !hasResolvedInitialLoad ? (
        <Alert>
          <AlertDescription>{t("loading")}</AlertDescription>
        </Alert>
      ) : null}

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

      <div className="flex flex-col gap-3">
        <div className="flex flex-col gap-1">
          <p className="text-sm font-medium">{t("loggingModeLabel")}</p>
          <p className="text-xs text-muted-foreground">{t("modeBehaviorTitle")}</p>
        </div>

        <StatusOptionGroup
          label={t("loggingModeLabel")}
          value={loggingMode}
          onValueChange={setLoggingMode}
          disabled={saving || loading}
          options={MODES.map((item) => ({
            ...item,
            badge: item.value,
            title: t(`modes.${item.key}.title`),
            description: t(`modes.${item.key}.description`),
          }))}
        />
      </div>

      <div className="grid gap-1 text-xs text-muted-foreground">
        <p>{t("updatedAtLabel")}: {formatDate(updatedAt, t("notSet"))}</p>
        <p>{t("updatedByLabel")}: {updatedBy || t("notSet")}</p>
      </div>

      <Button onClick={() => void save()} disabled={saving || loading} className="self-start">
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
            <p className="text-sm text-muted-foreground">{t("description")}</p>
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
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
      ) : null}
      <CardContent className="flex flex-col gap-4">{content}</CardContent>
    </Card>
  );
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

function formatDate(value: string | null, fallback: string) {
  if (!value) return fallback;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toISOString().replace("T", " ").replace("Z", " UTC");
}
