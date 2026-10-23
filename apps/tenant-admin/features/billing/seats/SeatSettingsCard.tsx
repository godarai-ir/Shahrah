"use client";

import { useState, type FormEvent } from "react";
import { useTranslations } from "next-intl";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { updateCreditSettings, type CreditSettings } from "../credits/api";
import { seatErrorMessage } from "@/lib/seats";

const CREDIT_AMOUNT_PATTERN = /^\d+(\.\d{1,2})?$/;

function invalidAmount(value: string) {
  const trimmed = value.trim();
  return trimmed !== "" && (!CREDIT_AMOUNT_PATTERN.test(trimmed) || Number(trimmed) <= 0);
}

/** Upgrade requests and the default monthly pool limit of Pro and Max members. */
export function SeatSettingsCard({ initialSettings, canManage }: { initialSettings: CreditSettings; canManage: boolean }) {
  const t = useTranslations("billingSeats");
  const [settings, setSettings] = useState(initialSettings);
  const [requests, setRequests] = useState(initialSettings.seat_upgrade_requests_enabled ?? true);
  const [proLimit, setProLimit] = useState(initialSettings.seat_pool_limit_pro_credits ?? "");
  const [maxLimit, setMaxLimit] = useState(initialSettings.seat_pool_limit_max_credits ?? "");
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState<{ kind: "error"; message: string } | null>(null);

  const dirty =
    requests !== (settings.seat_upgrade_requests_enabled ?? true) ||
    proLimit.trim() !== (settings.seat_pool_limit_pro_credits ?? "") ||
    maxLimit.trim() !== (settings.seat_pool_limit_max_credits ?? "");
  const invalid = invalidAmount(proLimit) || invalidAmount(maxLimit);

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (invalid) return;
    setSaving(true);
    setNotice(null);
    try {
      const next = await updateCreditSettings({
        seat_upgrade_requests_enabled: requests,
        seat_pool_limit_pro_credits: proLimit.trim(),
        seat_pool_limit_max_credits: maxLimit.trim(),
      });
      setSettings(next);
      setRequests(next.seat_upgrade_requests_enabled ?? true);
      setProLimit(next.seat_pool_limit_pro_credits ?? "");
      setMaxLimit(next.seat_pool_limit_max_credits ?? "");
      showFeedbackToast({ kind: "success", title: t("settings.saved") });
    } catch (err) {
      setNotice({ kind: "error", message: seatErrorMessage(t, err) });
    } finally {
      setSaving(false);
    }
  }

  return (
    <SettingsCard className="max-w-none" title={t("settings.title")} description={t("settings.description")}>
      <form className="flex flex-col gap-4" onSubmit={save}>
        <OperationNotice notice={notice} labels={{ success: t("settings.saved"), error: t("errors.title"), warning: t("errors.title"), info: t("settings.saved") }} />
        <AppSwitchField
          label={t("settings.requests")}
          description={t("settings.requestsDescription")}
          checked={requests}
          disabled={!canManage || saving}
          onCheckedChange={setRequests}
        />
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <AppNumberField
            label={t("settings.proLimit")}
            description={t("settings.limitDescription")}
            placeholder={t("settings.noLimit")}
            min={0.01}
            step={0.01}
            value={proLimit}
            onChange={(event) => setProLimit(event.target.value)}
            disabled={!canManage || saving}
            error={invalidAmount(proLimit) ? t("settings.limitInvalid") : undefined}
          />
          <AppNumberField
            label={t("settings.maxLimit")}
            description={t("settings.limitDescription")}
            placeholder={t("settings.noLimit")}
            min={0.01}
            step={0.01}
            value={maxLimit}
            onChange={(event) => setMaxLimit(event.target.value)}
            disabled={!canManage || saving}
            error={invalidAmount(maxLimit) ? t("settings.limitInvalid") : undefined}
          />
        </div>
        {canManage ? (
          <div>
            <Button type="submit" disabled={saving || !dirty || invalid} aria-busy={saving || undefined}>
              {saving ? <Spinner data-icon="inline-start" /> : null}
              {saving ? t("settings.saving") : t("settings.save")}
            </Button>
          </div>
        ) : null}
      </form>
    </SettingsCard>
  );
}
