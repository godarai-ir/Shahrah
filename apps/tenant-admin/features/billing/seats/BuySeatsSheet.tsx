"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { CreditCard } from "lucide-react";

import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { useBillingFormat } from "../use-billing-format";
import { createSeatOrder, quoteSeats, type PaidSeatType, type SeatQuote } from "./api";
import { seatErrorMessage } from "./domain";

export type BuySeatsTarget = {
  seat: PaidSeatType;
  /** The member who gets one of the seats once paid. */
  assignUserID?: string;
  assignLabel?: string;
  /** The upgrade request the purchase approves. */
  requestID?: string;
};

type BuySeatsSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  target: BuySeatsTarget;
};

/** Buys paid seats for the rest of the billing period, priced pro rata, through the payment gateway. */
export function BuySeatsSheet({ open, onOpenChange, target }: BuySeatsSheetProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const [seat, setSeat] = useState<PaidSeatType>(target.seat);
  const [quantity, setQuantity] = useState("1");
  const [quote, setQuote] = useState<SeatQuote | null>(null);
  const [quoting, setQuoting] = useState(false);
  const [paying, setPaying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const forMember = Boolean(target.assignUserID);
  const count = Number.parseInt(quantity, 10);
  const validCount = Number.isInteger(count) && count >= 1 && count <= 1000;

  // The sheet is mounted per target, so its state starts fresh each time.
  useEffect(() => {
    if (!open || !validCount) return;

    let canceled = false;
    const timer = window.setTimeout(() => {
      setQuoting(true);
      quoteSeats({ seat, quantity: count })
        .then((next) => {
          if (!canceled) setQuote(next);
        })
        .catch((err: unknown) => {
          if (!canceled) {
            setQuote(null);
            setError(seatErrorMessage(t, err));
          }
        })
        .finally(() => {
          if (!canceled) setQuoting(false);
        });
    }, 250);

    return () => {
      canceled = true;
      window.clearTimeout(timer);
    };
  }, [count, open, seat, t, validCount]);

  async function pay() {
    setPaying(true);
    setError(null);
    try {
      const order = await createSeatOrder({
        kind: "purchase",
        seat,
        quantity: count,
        assign_user_id: target.assignUserID,
        request_id: target.requestID,
        return_url: `${window.location.origin}/billing/seats`,
      });
      if (order.payment_url) {
        window.location.assign(order.payment_url);
        return;
      }
      setError(t("buy.noPaymentURL"));
    } catch (err) {
      setError(seatErrorMessage(t, err));
    }
    setPaying(false);
  }

  // A quote for other inputs is never shown, so the amount on the button always matches the order.
  const shownQuote = quote && quote.seat === seat && quote.quantity === count ? quote : null;

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      pending={paying}
      title={forMember ? t("buy.titleForMember", { member: target.assignLabel ?? "" }) : t("buy.title")}
      description={t("buy.description")}
      footer={
        <Button type="button" className="w-full" disabled={!shownQuote || quoting || paying || !validCount} aria-busy={paying || undefined} onClick={pay}>
          {paying ? <Spinner data-icon="inline-start" /> : <CreditCard data-icon="inline-start" />}
          {shownQuote ? t("buy.pay", { amount: format.toman(shownQuote.payable_irr) }) : t("buy.payPending")}
        </Button>
      }
    >
      <div className="flex flex-col gap-5">
        <AppSelectField
          label={t("buy.seat")}
          value={seat}
          onValueChange={(value) => setSeat(value as PaidSeatType)}
          disabled={forMember}
          options={[
            { value: "pro", label: t("seats.pro") },
            { value: "max", label: t("seats.max") },
          ]}
        />
        <AppNumberField
          label={t("buy.quantity")}
          value={quantity}
          min={1}
          max={1000}
          inputMode="numeric"
          disabled={forMember}
          onChange={(event) => setQuantity(event.target.value)}
          error={validCount ? undefined : t("buy.quantityInvalid")}
          description={forMember ? t("buy.forMemberHint") : t("buy.quantityHint")}
        />

        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        {quoting && !shownQuote ? <Spinner aria-label={t("buy.quoting")} /> : null}

        {shownQuote ? (
          <DefinitionList>
            <DefinitionListItem
              label={t("buy.unitPrice", { interval: t(`intervals.${shownQuote.billing_interval === "yearly" ? "yearly" : "monthly"}`) })}
              value={<span className="tabular-nums">{format.toman(shownQuote.unit_period_price_irr)}</span>}
            />
            <DefinitionListItem label={t("buy.prorated")} value={t("buy.proratedValue", { remaining: shownQuote.remaining_days, total: shownQuote.period_days })} />
            <DefinitionListItem label={t("buy.periodEnd")} value={format.date(shownQuote.period_end)} />
            <DefinitionListItem label={t("buy.price")} value={<span className="tabular-nums">{format.toman(shownQuote.price_irr)}</span>} />
            <DefinitionListItem
              label={t("buy.tax", { percent: Number(shownQuote.tax_percent) || 0 })}
              value={<span className="tabular-nums">{format.toman(shownQuote.tax_irr)}</span>}
            />
            <DefinitionListItem label={t("buy.payable")} emphasis value={<span className="tabular-nums">{format.toman(shownQuote.payable_irr)}</span>} />
          </DefinitionList>
        ) : null}

        <p className="text-sm text-muted-foreground">{t("buy.rules")}</p>
      </div>
    </FormSheet>
  );
}
