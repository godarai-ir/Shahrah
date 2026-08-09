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

import { SettingsAuditMeta } from "@/components/app-ui/SettingsAuditMeta";
import { StatusOptionGroup } from "@/components/app-ui/StatusOptionGroup";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { updateTracePrivacySettings, type TracePrivacySettings } from "./api";
export type { TracePrivacySettings } from "./api";

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
  const [error, setError] = useState<string | null>(props.initialError ?? null);
  const [updatedAt, setUpdatedAt] = useState<string | null>(props.initialSettings?.updated_at ?? null);
  const [updatedBy, setUpdatedBy] = useState<string | null>(props.initialSettings?.updated_by_user_id ?? null);

  async function save() {
    setSaving(true);
    setError(null);
    try {
      const payload = await updateTracePrivacySettings(capturePolicy);
      setUpdatedAt(payload.updated_at ?? null);
      setUpdatedBy(payload.updated_by_user_id ?? props.currentUserID ?? null);
      showFeedbackToast({ kind: "success", title: t("messages.updated"), options: { id: "trace-privacy-updated" } });
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

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
