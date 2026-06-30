"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";

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
}> = [
  { value: "none", key: "none" },
  { value: "metadata-only", key: "metadataOnly" },
  { value: "redacted-content", key: "redactedContent" },
  { value: "full-content", key: "fullContent" },
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
          <Alert className="border-primary/30 bg-primary/5">
            <AlertDescription>{message}</AlertDescription>
          </Alert>
        ) : null}

        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        <div className="grid gap-2">
          <Label htmlFor="requestLoggingMode">{t("loggingModeLabel")}</Label>
          <select
            id="requestLoggingMode"
            className="h-10 rounded-md border bg-background px-3 text-sm"
            value={loggingMode}
            onChange={(event) => setLoggingMode(event.target.value as RequestLoggingSettings["logging_mode"])}
            disabled={saving}
          >
            {MODES.map((item) => (
              <option key={item.value} value={item.value}>
                {t(`modes.${item.key}.title`)}
              </option>
            ))}
          </select>
        </div>

        <div className="rounded-md border p-3 text-xs text-muted-foreground">
          <p className="font-medium text-foreground">{t("modeBehaviorTitle")}</p>
          <ul className="mt-2 list-disc space-y-1 pl-4">
            {MODES.map((item) => (
              <li key={item.value}>
                <span className="font-medium">{item.value}</span>: {t(`modes.${item.key}.description`)}
              </li>
            ))}
          </ul>
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
