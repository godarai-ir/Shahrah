"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import {
  Ban,
  Database,
  Eye,
  FileSearch,
  type LucideIcon,
} from "lucide-react";

import { SettingsAuditMeta } from "@/components/app-ui/SettingsAuditMeta";
import { StatusOptionGroup } from "@/components/app-ui/StatusOptionGroup";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  getRequestLoggingSettings,
  updateRequestLoggingSettings,
  type RequestLoggingSettings,
} from "@/features/logs/api";

export type { RequestLoggingSettings } from "@/features/logs/api";

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
  const locale = useLocale();
  const surface = props.surface ?? "card";
  const showHeader = props.showHeader ?? true;
  const [loggingMode, setLoggingMode] = useState<RequestLoggingSettings["logging_mode"]>(
    props.initialSettings?.logging_mode ?? "metadata-only",
  );
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
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

    void getRequestLoggingSettings()
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

    try {
      const payload = await updateRequestLoggingSettings(loggingMode);
      applySettings(payload);
      showFeedbackToast({ kind: "success", title: t("messages.updated"), options: { id: "request-logging-updated" } });
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

      <SettingsAuditMeta updatedAt={updatedAt} updatedBy={updatedBy} updatedAtLabel={t("updatedAtLabel")} updatedByLabel={t("updatedByLabel")} fallback={t("notSet")} locale={locale} />

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

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
