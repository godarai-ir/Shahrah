"use client";

import { useEffect, useId, useState } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, CalendarDays, CalendarRange, CreditCard, Crown, Landmark, Sparkles } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { OptionCardGroup } from "@/components/app-ui/OptionCardGroup";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

import { useBillingFormat } from "../use-billing-format";
import { BANK_TRANSFER_PROVIDER, createSeatOrder, quoteSeats, type PaidSeatType, type SeatInterval, type SeatOffer, type SeatOrder, type SeatQuote } from "./api";
import { SeatTransferSheet } from "./SeatTransferSheet";
import { seatErrorMessage } from "@/lib/seats";

const PAID_SEATS = [
  { type: "pro", icon: Sparkles },
  { type: "max", icon: Crown },
] as const satisfies readonly { type: PaidSeatType; icon: typeof Sparkles }[];

export type BuySeatsTarget = {
  seat: PaidSeatType;
  /** The member who gets one of the seats once paid. */
  assignUserID?: string;
  assignLabel?: string;
  /** The upgrade request the purchase approves. */
  requestID?: string;
};

type BuySeatsFormProps = {
  target: BuySeatsTarget;
  /** The seat offers: the monthly credits and prices each paid seat comes with. */
  offers: SeatOffer[];
};

type PaymentMethod = "online" | typeof BANK_TRANSFER_PROVIDER;

/**
 * Buys paid seats through the payment gateway or by bank transfer. With a running seat subscription
 * the seats are priced pro rata to its period end; otherwise the purchase starts
 * the subscription, and the admin picks its billing interval here.
 */
export function BuySeatsForm({ target, offers }: BuySeatsFormProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const seatLabelID = useId();
  const intervalLabelID = useId();
  const methodLabelID = useId();
  const router = useRouter();
  const [seat, setSeat] = useState<PaidSeatType>(target.seat);
  const [quantity, setQuantity] = useState("1");
  const [billingInterval, setBillingInterval] = useState<SeatInterval>("monthly");
  // Unknown until the first quote: only a purchase that starts the seat
  // subscription may choose an interval; later seats follow the subscription.
  const [startsSubscription, setStartsSubscription] = useState(false);
  const [quote, setQuote] = useState<SeatQuote | null>(null);
  const [quoting, setQuoting] = useState(false);
  const [paying, setPaying] = useState(false);
  const [method, setMethod] = useState<PaymentMethod>("online");
  const [transferOrder, setTransferOrder] = useState<SeatOrder | null>(null);
  const [error, setError] = useState<string | null>(null);
  const forMember = Boolean(target.assignUserID);
  const count = Number.parseInt(quantity, 10);
  const validCount = Number.isInteger(count) && count >= 1 && count <= 1000;
  const offer = (type: PaidSeatType) => offers.find((item) => item.type === type);
  const selectedCredits = offer(seat)?.monthly_credits;

  useEffect(() => {
    if (!validCount) return;

    let canceled = false;
    const timer = window.setTimeout(() => {
      setQuoting(true);
      setError(null);
      quoteSeats({ seat, quantity: count, billing_interval: startsSubscription ? billingInterval : undefined })
        .then((next) => {
          if (canceled) return;
          setQuote(next);
          setStartsSubscription(next.starts_subscription);
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
  }, [count, billingInterval, seat, startsSubscription, t, validCount]);

  async function pay() {
    setPaying(true);
    setError(null);
    try {
      const order = await createSeatOrder({
        kind: "purchase",
        provider: method === BANK_TRANSFER_PROVIDER ? BANK_TRANSFER_PROVIDER : undefined,
        seat,
        quantity: count,
        billing_interval: startsSubscription ? billingInterval : undefined,
        assign_user_id: target.assignUserID,
        request_id: target.requestID,
        return_url: `${window.location.origin}/billing/seats`,
      });
      if (order.status === "AWAITING_TRANSFER") {
        setTransferOrder(order);
        setPaying(false);
        return;
      }
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
  const shownQuote =
    quote && quote.seat === seat && quote.quantity === count && (!quote.starts_subscription || quote.billing_interval === billingInterval) ? quote : null;

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Button asChild variant="ghost" size="sm">
          <Link href="/billing/seats">
            <ArrowRight data-icon="inline-start" aria-hidden="true" className="ltr:rotate-180" />
            {t("buy.back")}
          </Link>
        </Button>
      </div>

      <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <AppCard title={forMember ? t("buy.titleForMember", { member: target.assignLabel ?? "" }) : t("buy.formTitle")} description={t("buy.description")}>
          <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-2">
              <span id={seatLabelID} className="text-sm font-medium">{t("buy.seat")}</span>
              <OptionCardGroup
                aria-labelledby={seatLabelID}
                value={seat}
                onValueChange={setSeat}
                disabled={forMember}
                options={PAID_SEATS.map(({ type, icon }) => {
                  const item = offer(type);

                  return {
                    value: type,
                    icon,
                    title: t(`seats.${type}`),
                    description: item ? t("buy.monthlyCredits", { credits: format.count(item.monthly_credits) }) : undefined,
                    tip: item ? t("buy.monthlyPrice", { price: format.toman(item.monthly_price_irr) }) : undefined,
                  };
                })}
              />
            </div>

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

            {startsSubscription ? (
              <div className="flex flex-col gap-2">
                <span id={intervalLabelID} className="text-sm font-medium">{t("buy.interval")}</span>
                <OptionCardGroup
                  aria-labelledby={intervalLabelID}
                  density="compact"
                  value={billingInterval}
                  onValueChange={setBillingInterval}
                  options={[
                    {
                      value: "monthly",
                      icon: CalendarDays,
                      title: t("intervals.monthly"),
                      description: offer(seat) ? t("buy.monthlyPrice", { price: format.toman(offer(seat)?.monthly_price_irr) }) : undefined,
                    },
                    {
                      value: "yearly",
                      icon: CalendarRange,
                      title: t("intervals.yearly"),
                      description: offer(seat) ? t("buy.yearlyPrice", { price: format.toman(offer(seat)?.yearly_price_irr) }) : undefined,
                    },
                  ]}
                />
                <p className="text-xs leading-5 text-muted-foreground">{t("buy.intervalHint")}</p>
              </div>
            ) : null}

            {startsSubscription ? (
              <Alert>
                <AlertDescription>{t("buy.startsSubscription")}</AlertDescription>
              </Alert>
            ) : null}

            <div className="flex flex-col gap-2">
              <span id={methodLabelID} className="text-sm font-medium">{t("buy.method")}</span>
              <OptionCardGroup
                aria-labelledby={methodLabelID}
                density="compact"
                value={method}
                onValueChange={setMethod}
                options={[
                  { value: "online", icon: CreditCard, title: t("buy.methods.online"), description: t("buy.methods.onlineHint") },
                  { value: BANK_TRANSFER_PROVIDER, icon: Landmark, title: t("buy.methods.transfer"), description: t("buy.methods.transferHint") },
                ]}
              />
            </div>

            <p className="text-sm leading-6 text-muted-foreground">{t("buy.rules")}</p>
          </div>
        </AppCard>

        <AppCard title={t("buy.summaryTitle")} className="lg:sticky lg:top-20">
          <div className="flex flex-col gap-4">
            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}

            {quoting && !shownQuote ? <Spinner label={t("buy.quoting")} /> : null}

            {shownQuote ? (
              <DefinitionList>
                {selectedCredits !== undefined && validCount ? (
                  <DefinitionListItem
                    label={t("buy.creditsTotal")}
                    value={<span className="tabular-nums">{t("buy.creditsTotalValue", { credits: format.count(selectedCredits * count) })}</span>}
                  />
                ) : null}
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
                <DefinitionListItem
                  className="border-t border-border pt-3"
                  label={t("buy.payable")}
                  emphasis
                  value={<span className="tabular-nums">{format.toman(shownQuote.payable_irr)}</span>}
                />
              </DefinitionList>
            ) : null}

            <Button type="button" className="w-full" disabled={!shownQuote || quoting || paying || !validCount} aria-busy={paying || undefined} onClick={pay}>
              {paying ? <Spinner data-icon="inline-start" /> : method === BANK_TRANSFER_PROVIDER ? <Landmark data-icon="inline-start" /> : <CreditCard data-icon="inline-start" />}
              {!shownQuote ? t("buy.payPending") : t(method === BANK_TRANSFER_PROVIDER ? "buy.payByTransfer" : "buy.pay", { amount: format.toman(shownQuote.payable_irr) })}
            </Button>
          </div>
        </AppCard>
      </div>

      {/* The order is already saved; leaving the drawer unfinished sends the admin to the orders list to submit the receipt later. */}
      <SeatTransferSheet order={transferOrder} onClose={() => router.push("/billing/seats")} onSubmitted={() => router.push("/billing/seats")} />
    </div>
  );
}
