"use client";

import { useState, type FormEvent } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { updateCreditSettings, type CreditSettings } from "./api";

const CREDIT_AMOUNT_PATTERN = /^\d+(\.\d{1,2})?$/;

type CreditSettingsCardProps = {
  initialSettings: CreditSettings;
  /** Changing settings needs billing:manage. */
  canManage: boolean;
};

/** Overage after the plan quota and add-on packs run out, with an optional credit cap. */
export function CreditSettingsCard({ initialSettings, canManage }: CreditSettingsCardProps) {
  const t = useTranslations("billing");
  const [settings, setSettings] = useState(initialSettings);
  const [enabled, setEnabled] = useState(initialSettings.overage_enabled);
  const [cap, setCap] = useState(initialSettings.overage_cap_credits ?? "");
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "error"; message: string } | null>(null);

  const trimmedCap = cap.trim();
  const capInvalid = trimmedCap !== "" && (!CREDIT_AMOUNT_PATTERN.test(trimmedCap) || Number(trimmedCap) <= 0);
  const dirty = enabled !== settings.overage_enabled || trimmedCap !== (settings.overage_cap_credits ?? "");
  const locked = !canManage || !settings.overage_allowed;

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (capInvalid) return;
    setSaving(true);
    setNotice(null);
    try {
      const next = await updateCreditSettings({ overage_enabled: enabled, overage_cap_credits: trimmedCap });
      setSettings(next);
      setEnabled(next.overage_enabled);
      setCap(next.overage_cap_credits ?? "");
      showFeedbackToast({ kind: "success", title: t("credits.settings.saved") });
    } catch (err) {
      setNotice({ kind: "error", message: err instanceof Error ? err.message : t("unknownError") });
    } finally {
      setSaving(false);
    }
  }

  return (
    <SettingsCard
      title={t("credits.settings.title")}
      description={t("credits.settings.description")}
      actions={
        <Button asChild variant="outline" size="sm">
          <Link href="/billing/credits">{t("credits.settings.buyPacks")}</Link>
        </Button>
      }
    >
      <form className="flex flex-col gap-4" onSubmit={save}>
        <OperationNotice notice={notice} labels={{ success: t("credits.settings.saved"), error: t("errorTitle"), warning: t("errorTitle"), info: t("credits.settings.saved") }} />
        <AppSwitchField
          label={t("credits.settings.overage")}
          description={settings.overage_allowed ? t("credits.settings.overageDescription") : t("credits.settings.overageNotAllowed")}
          checked={enabled}
          disabled={locked || saving}
          onCheckedChange={setEnabled}
        />
        <AppNumberField
          id="credit-overage-cap"
          label={t("credits.settings.cap")}
          description={t("credits.settings.capDescription")}
          placeholder={t("credits.settings.capPlaceholder")}
          min={0.01}
          step={0.01}
          inputMode="decimal"
          value={cap}
          onChange={(event) => setCap(event.target.value)}
          disabled={locked || saving || !enabled}
          error={capInvalid ? t("credits.settings.capInvalid") : undefined}
        />
        {canManage ? (
          <div>
            <Button type="submit" disabled={locked || saving || !dirty || capInvalid} aria-busy={saving || undefined}>
              {saving ? <Spinner data-icon="inline-start" /> : null}
              {saving ? t("credits.settings.saving") : t("credits.settings.save")}
            </Button>
          </div>
        ) : null}
      </form>
    </SettingsCard>
  );
}
