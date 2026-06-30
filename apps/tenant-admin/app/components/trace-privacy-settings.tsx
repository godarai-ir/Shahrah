"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  Ban,
  CheckCircle2,
  FileStack,
  LockKeyhole,
  ScanSearch,
  type LucideIcon,
} from "lucide-react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type TracePrivacySettings = {
  tenant_id?: string;
  capture_policy: "none" | "metadata-only" | "sampled-redacted" | "full-secure";
  updated_at?: string;
  updated_by_user_id?: string;
};

type TracePrivacySettingsProps = {
  initialSettings: TracePrivacySettings | null;
  initialError?: string | null;
};

const POLICIES: Array<{
  value: TracePrivacySettings["capture_policy"];
  key: "none" | "metadataOnly" | "sampledRedacted" | "fullSecure";
  icon: LucideIcon;
  accent: string;
  accentSoft: string;
  accentText: string;
}> = [
  {
    value: "none",
    key: "none",
    icon: Ban,
    accent: "border-slate-400/70",
    accentSoft: "bg-slate-500/10",
    accentText: "text-slate-700 dark:text-slate-300",
  },
  {
    value: "metadata-only",
    key: "metadataOnly",
    icon: FileStack,
    accent: "border-sky-500/70",
    accentSoft: "bg-sky-500/10",
    accentText: "text-sky-700 dark:text-sky-300",
  },
  {
    value: "sampled-redacted",
    key: "sampledRedacted",
    icon: ScanSearch,
    accent: "border-amber-500/70",
    accentSoft: "bg-amber-500/10",
    accentText: "text-amber-700 dark:text-amber-300",
  },
  {
    value: "full-secure",
    key: "fullSecure",
    icon: LockKeyhole,
    accent: "border-emerald-500/70",
    accentSoft: "bg-emerald-500/10",
    accentText: "text-emerald-700 dark:text-emerald-300",
  },
];

export default function TracePrivacySettingsPanel(props: TracePrivacySettingsProps) {
  const t = useTranslations("settings.tracePrivacy");
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
        <CardDescription>
          {t("description")}
        </CardDescription>
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

        <div className="space-y-3">
          <div>
            <p className="text-sm font-medium">{t("capturePolicyLabel")}</p>
            <p className="text-xs text-muted-foreground">{t("policyBehaviorTitle")}</p>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            {POLICIES.map((item) => {
              const Icon = item.icon;
              const selected = capturePolicy === item.value;

              return (
                <button
                  key={item.value}
                  type="button"
                  onClick={() => setCapturePolicy(item.value)}
                  disabled={saving}
                  aria-pressed={selected}
                  className={cn(
                    "group rounded-xl border p-4 text-right transition-all",
                    "hover:shadow-sm disabled:cursor-not-allowed disabled:opacity-70",
                    selected
                      ? cn(item.accent, item.accentSoft, "ring-2 ring-primary/20")
                      : "border-border bg-background hover:border-primary/30 hover:bg-muted/30",
                  )}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div
                      className={cn(
                        "flex size-10 shrink-0 items-center justify-center rounded-lg border",
                        selected ? cn(item.accent, item.accentSoft) : "border-border bg-muted/40",
                      )}
                    >
                      <Icon className={cn("size-5", selected ? item.accentText : "text-muted-foreground")} />
                    </div>
                    <div className="flex min-w-0 items-center gap-2">
                      {selected ? <CheckCircle2 className="size-4 text-primary" /> : null}
                      <span
                        className={cn(
                          "rounded-full px-2 py-0.5 text-[11px] font-medium",
                          selected ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground",
                        )}
                      >
                        {item.value}
                      </span>
                    </div>
                  </div>
                  <div className="mt-3 space-y-1">
                    <p className="text-sm font-semibold text-foreground">{t(`policies.${item.key}.title`)}</p>
                    <p className="text-xs leading-6 text-muted-foreground">{t(`policies.${item.key}.description`)}</p>
                  </div>
                </button>
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
