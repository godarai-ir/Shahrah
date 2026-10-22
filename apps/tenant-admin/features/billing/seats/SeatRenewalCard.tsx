"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { CreditCard } from "lucide-react";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { SettingsCard } from "@/components/app-ui/SettingsCard";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { useBillingFormat } from "../use-billing-format";
import { createSeatOrder, updateSeatRenewal, type PaidSeatType, type SeatSummary } from "./api";
import { seatErrorMessage } from "@/lib/seats";

type SeatRenewalCardProps = {
  summary: SeatSummary;
  canManage: boolean;
  onSummaryChange: (summary: SeatSummary) => void;
};

/**
 * Paid seats run to the end of the period they were paid for. After the plan
 * renews, this card pays for carrying them into the new period; the admin can
 * keep fewer seats, and the rest lapse.
 */
export function SeatRenewalCard({ summary, canManage, onSummaryChange }: SeatRenewalCardProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const [keep, setKeep] = useState<Record<PaidSeatType, string>>(() => ({
    pro: String(summary.paid.find((item) => item.type === "pro")?.renewing ?? 0),
    max: String(summary.paid.find((item) => item.type === "max")?.renewing ?? 0),
  }));
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const renewal = summary.renewal_due;

  async function saveKeep(seat: PaidSeatType) {
    setPending(seat);
    setError(null);
    try {
      onSummaryChange(await updateSeatRenewal({ seat, quantity: Number.parseInt(keep[seat], 10) || 0 }));
    } catch (err) {
      setError(seatErrorMessage(t, err));
    }
    setPending(null);
  }

  async function pay() {
    setPending("pay");
    setError(null);
    try {
      const order = await createSeatOrder({ kind: "renewal", return_url: `${window.location.origin}/billing/seats` });
      if (order.payment_url) {
        window.location.assign(order.payment_url);
        return;
      }
      setError(t("buy.noPaymentURL"));
    } catch (err) {
      setError(seatErrorMessage(t, err));
    }
    setPending(null);
  }

  const held = summary.paid.filter((item) => item.purchased > 0);
  if (held.length === 0 && !renewal) return null;

  return (
    <SettingsCard className="max-w-none" title={t("renewal.title")} description={t("renewal.description")}>
      <div className="flex flex-col gap-4">
        {renewal ? (
          <Alert>
            <AlertTitle>{t("renewal.dueTitle", { date: format.date(renewal.period_end) })}</AlertTitle>
            <AlertDescription>
              {t("renewal.dueDescription", {
                seats: renewal.lines.reduce((sum, line) => sum + line.quantity, 0),
                amount: format.toman(renewal.payable_irr),
              })}
            </AlertDescription>
          </Alert>
        ) : null}

        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {held.map((item) => (
            <div key={item.type} className="flex items-end gap-2">
              <AppNumberField
                fieldClassName="flex-1"
                label={t("renewal.keep", { seat: t(`seats.${item.type}`) })}
                description={t("renewal.keepHint", { held: item.purchased, assigned: item.assigned })}
                min={0}
                max={item.purchased}
                inputMode="numeric"
                value={keep[item.type]}
                onChange={(event) => setKeep((current) => ({ ...current, [item.type]: event.target.value }))}
                disabled={!canManage || pending !== null}
              />
              <Button type="button" variant="outline" disabled={!canManage || pending !== null} aria-busy={pending === item.type || undefined} onClick={() => saveKeep(item.type)}>
                {pending === item.type ? <Spinner data-icon="inline-start" /> : null}
                {t("renewal.saveKeep")}
              </Button>
            </div>
          ))}
        </div>

        {renewal && canManage ? (
          <div>
            <Button type="button" disabled={pending !== null} aria-busy={pending === "pay" || undefined} onClick={pay}>
              {pending === "pay" ? <Spinner data-icon="inline-start" /> : <CreditCard data-icon="inline-start" />}
              {t("renewal.pay", { amount: format.toman(renewal.payable_irr) })}
            </Button>
          </div>
        ) : null}
      </div>
    </SettingsCard>
  );
}
