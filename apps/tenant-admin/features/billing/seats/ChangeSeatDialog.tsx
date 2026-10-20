"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { APIClientError } from "@/lib/http/api-client";

import { useBillingFormat } from "../use-billing-format";
import { purchaseRequiredQuote, setMemberSeat, SEAT_TYPES, type PaidSeatType, type SeatMember, type SeatType } from "./api";
import { previewSeatChange, seatErrorMessage } from "./domain";

type ChangeSeatDialogProps = {
  member: SeatMember | null;
  onOpenChange: (open: boolean) => void;
  /** Called after a change was applied or scheduled. */
  onChanged: () => void;
  /** Opens checkout for one seat of this type for the member. */
  onBuyForMember: (member: SeatMember, seat: PaidSeatType) => void;
  periodEnd: string | null | undefined;
};

/** Changes one member's seat, explaining when the change applies before it is made. */
export function ChangeSeatDialog({ member, onOpenChange, onChanged, onBuyForMember, periodEnd }: ChangeSeatDialogProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  // The dialog is mounted per member, so the suggested seat is the next one up.
  const [target, setTarget] = useState<SeatType>(member?.seat === "pro" ? "max" : member?.seat === "max" ? "pro" : "pro");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [needsPurchase, setNeedsPurchase] = useState<{ payableIRR: number | null } | null>(null);

  const preview = useMemo(() => (member ? previewSeatChange(member, target) : null), [member, target]);

  async function apply() {
    if (!member) return;
    setPending(true);
    setError(null);
    setNeedsPurchase(null);
    try {
      await setMemberSeat(member.user_id, { seat: target });
      onChanged();
      onOpenChange(false);
    } catch (err) {
      const quote = err instanceof APIClientError ? purchaseRequiredQuote(err.payload) : undefined;
      if (quote !== undefined) {
        setNeedsPurchase({ payableIRR: quote?.payable_irr ?? null });
      } else {
        setError(seatErrorMessage(t, err));
      }
    }
    setPending(false);
  }

  const memberLabel = member ? member.name || member.email : "";

  return (
    <FormDialog
      open={member !== null}
      onOpenChange={onOpenChange}
      pending={pending}
      title={t("change.title", { member: memberLabel })}
      description={member ? t("change.current", { seat: t(`seats.${member.seat}`) }) : undefined}
      footer={
        needsPurchase && member && (target === "pro" || target === "max") ? (
          <Button type="button" onClick={() => onBuyForMember(member, target)}>
            {needsPurchase.payableIRR !== null ? t("change.buyAndAssignFor", { amount: format.toman(needsPurchase.payableIRR) }) : t("change.buyAndAssign")}
          </Button>
        ) : (
          <Button type="button" disabled={!preview?.allowed || pending} aria-busy={pending || undefined} onClick={apply}>
            {pending ? <Spinner data-icon="inline-start" /> : null}
            {preview?.allowed && preview.timing === "period_end" ? t("change.schedule") : t("change.apply")}
          </Button>
        )
      }
    >
      <div className="flex flex-col gap-4">
        <AppSelectField
          label={t("change.target")}
          value={target}
          onValueChange={(value) => {
            setTarget(value as SeatType);
            setNeedsPurchase(null);
            setError(null);
          }}
          options={SEAT_TYPES.map((seat) => ({ value: seat, label: t(`seats.${seat}`) }))}
        />

        {preview && !preview.allowed ? (
          <Alert>
            <AlertDescription>{t(`change.blocked.${preview.reason}`)}</AlertDescription>
          </Alert>
        ) : null}

        {preview?.allowed ? (
          <Alert>
            <AlertTitle>{t(preview.timing === "period_end" ? "change.whenPeriodEnd" : "change.whenNow")}</AlertTitle>
            <AlertDescription>
              {preview.timing === "period_end"
                ? t("change.periodEndHint", { date: periodEnd ? format.date(periodEnd) : "—" })
                : preview.needsPaidSeat
                  ? t("change.paidSeatHint")
                  : t("change.nowHint")}
            </AlertDescription>
          </Alert>
        ) : null}

        {needsPurchase ? (
          <Alert>
            <AlertTitle>{t("change.noStockTitle")}</AlertTitle>
            <AlertDescription>{t("change.noStockDescription")}</AlertDescription>
          </Alert>
        ) : null}

        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
      </div>
    </FormDialog>
  );
}
