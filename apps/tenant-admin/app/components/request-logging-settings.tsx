"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  Ban,
  Database,
  Eye,
  FileSearch,
  type LucideIcon,
} from "lucide-react";

import { StatusOptionCard } from "@/app/components/status-option-card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

type RequestLoggingSettings = {
  tenant_id?: string;
  logging_mode: "none" | "metadata-only" | "redacted-content" | "full-content";
  updated_at?: string;
  updated_by_user_id?: string;
};

type RequestLoggingSettingsProps = {
  initialSettings: RequestLoggingSettings | null;
  initialError?: string | null;
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
  const [loggingMode, setLoggingMode] = useState<RequestLoggingSettings["logging_mode"]>(
    props.initialSettings?.logging_mode ?? "metadata-only",
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
      const payload = await requestJSON<RequestLoggingSettings>("/api/settings/request-logging", {
        method: "PUT",
        body: JSON.stringify({ logging_mode: loggingMode }),
      });
      setUpdatedAt(payload.updated_at ?? null);
      setUpdatedBy(payload.updated_by_user_id ?? null);
      setMessage(t("messages.updated"));
    } catch (err) {
      setError(normalizeError(err, t("messages.updateFailed")));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        <CardDescription>{t("description")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
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

        <div className="space-y-3">
          <div>
            <p className="text-sm font-medium">{t("loggingModeLabel")}</p>
            <p className="text-xs text-muted-foreground">{t("modeBehaviorTitle")}</p>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            {MODES.map((item) => {
              const selected = loggingMode === item.value;

              return (
                <StatusOptionCard
                  key={item.value}
                  onClick={() => setLoggingMode(item.value)}
                  disabled={saving}
                  icon={item.icon}
                  selected={selected}
                  badge={item.value}
                  title={t(`modes.${item.key}.title`)}
                  description={t(`modes.${item.key}.description`)}
                  tone={item.tone}
                />
              );
            })}
          </div>
        </div>

        <div className="grid gap-1 text-xs text-muted-foreground">
          <p>{t("updatedAtLabel")}: {formatDate(updatedAt, t("notSet"))}</p>
          <p>{t("updatedByLabel")}: {updatedBy || t("notSet")}</p>
        </div>

        <Button onClick={() => void save()} disabled={saving}>
          {saving ? t("saving") : t("save")}
        </Button>
      </CardContent>
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
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toISOString().replace("T", " ").replace("Z", " UTC");
}
