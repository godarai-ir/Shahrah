"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useDateTimeFormatter } from "@/hooks/use-locale-format";
import { APIClientError } from "@/lib/http/api-client";
import { previewSeatChange, SEAT_TYPES, seatErrorMessage, type PaidSeatType, type SeatMember, type SeatType } from "@/lib/seats";

import { setMemberSeat } from "./api";

type MemberSeatDialogProps = {
  member: SeatMember | null;
  /** Display name of the member, used in the title. */
  memberLabel: string;
  onOpenChange: (open: boolean) => void;
  /** Called after a change was applied or scheduled. */
  onChanged: () => void;
  periodEnd: string | null;
};

const PURCHASE_REQUIRED = "seat_purchase_required";

/** Changes one member's seat, explaining when the change applies before it is made. */
export function MemberSeatDialog({ member, memberLabel, onOpenChange, onChanged, periodEnd }: MemberSeatDialogProps) {
  const t = useTranslations("accessManagement");
  const tErrors = useTranslations("billingSeats");
  const router = useRouter();
  const dateFormatter = useDateTimeFormatter({ year: "numeric", month: "2-digit", day: "2-digit" });
  // The dialog is mounted per member, so the suggested seat is the next one up.
  const [target, setTarget] = useState<SeatType>(member?.seat === "pro" ? "max" : "pro");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [needsPurchase, setNeedsPurchase] = useState(false);

  const preview = useMemo(() => (member ? previewSeatChange(member, target) : null), [member, target]);

  async function apply() {
    if (!member) return;
    setPending(true);
    setError(null);
    setNeedsPurchase(false);
    try {
      await setMemberSeat(member.user_id, { seat: target });
      onChanged();
      onOpenChange(false);
    } catch (err) {
      if (err instanceof APIClientError && err.code === PURCHASE_REQUIRED) {
        setNeedsPurchase(true);
      } else {
        setError(seatErrorMessage(tErrors, err));
      }
    }
    setPending(false);
  }

  // Buying happens in billing, which opens checkout for one seat assigned to this member.
  function buyForMember(seat: PaidSeatType) {
    if (!member) return;
    router.push(`/billing/seats?buy=${seat}&assign=${encodeURIComponent(member.user_id)}`);
  }

  return (
    <FormDialog
      open={member !== null}
      onOpenChange={onOpenChange}
      pending={pending}
      title={t("seat.change.title", { member: memberLabel })}
      description={member ? t("seat.change.current", { seat: t(`form.seatOptions.${member.seat}`) }) : undefined}
      footer={
        needsPurchase && (target === "pro" || target === "max") ? (
          <Button type="button" onClick={() => buyForMember(target)}>
            {t("seat.change.buyAndAssign")}
          </Button>
        ) : (
          <Button type="button" disabled={!preview?.allowed || pending} aria-busy={pending || undefined} onClick={apply}>
            {pending ? <Spinner data-icon="inline-start" /> : null}
            {preview?.allowed && preview.timing === "period_end" ? t("seat.change.schedule") : t("seat.change.apply")}
          </Button>
        )
      }
    >
      <div className="flex flex-col gap-4">
        <AppSelectField
          label={t("seat.change.target")}
          value={target}
          onValueChange={(value) => {
            setTarget(value as SeatType);
            setNeedsPurchase(false);
            setError(null);
          }}
          options={SEAT_TYPES.map((seat) => ({ value: seat, label: t(`form.seatOptions.${seat}`) }))}
        />

        {preview && !preview.allowed ? (
          <Alert>
            <AlertDescription>{t(`seat.change.blocked.${preview.reason}`)}</AlertDescription>
          </Alert>
        ) : null}

        {preview?.allowed ? (
          <Alert>
            <AlertTitle>{t(preview.timing === "period_end" ? "seat.change.whenPeriodEnd" : "seat.change.whenNow")}</AlertTitle>
            <AlertDescription>
              {preview.timing === "period_end"
                ? t("seat.change.periodEndHint", { date: periodEnd ? dateFormatter.format(new Date(periodEnd)) : "—" })
                : preview.needsPaidSeat
                  ? t("seat.change.paidSeatHint")
                  : t("seat.change.nowHint")}
            </AlertDescription>
          </Alert>
        ) : null}

        {needsPurchase ? (
          <Alert>
            <AlertTitle>{t("seat.change.noStockTitle")}</AlertTitle>
            <AlertDescription>{t("seat.change.noStockDescription")}</AlertDescription>
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
