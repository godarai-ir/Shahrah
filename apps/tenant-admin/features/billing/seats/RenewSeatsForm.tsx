"use client";

import { useEffect, useId, useMemo, useState, type ReactNode } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { ArrowRight, CalendarClock, CircleCheck, Clock, CreditCard, Crown, Lock, Sparkles, Users } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { seatErrorMessage } from "@/lib/seats";

import { useBillingFormat } from "../use-billing-format";
import {
  createSeatOrder,
  previewSeatRenewal,
  updateSeatRenewal,
  type PaidSeatType,
  type SeatRenewalPreview,
  type SeatRenewalPreviewLine,
  type SeatRenewalPreviewMember,
  type SeatSummary,
} from "./api";
import { RenewalRules } from "./RenewalRules";
import { RenewalTimeline } from "./RenewalTimeline";

const SEAT_ICONS = { pro: Sparkles, max: Crown } as const satisfies Record<PaidSeatType, typeof Sparkles>;
const PAID_TYPES = ["pro", "max"] as const satisfies readonly PaidSeatType[];

type RenewSeatsFormProps = {
  summary: SeatSummary;
  /** The preview for renewing every seat; null when the server could not load it. */
  preview: SeatRenewalPreview | null;
  /** True while the renewal window is open, decided on the server. */
  windowOpen: boolean;
  /** Paying needs billing:manage. */
  canManage: boolean;
  paymentResult: "success" | "failed" | null;
  /** The request time, so the server and the browser draw the same period position. */
  nowMs: number;
};

type Counts = Record<PaidSeatType, number>;

/**
 * Renews the paid seats. The admin decides how many seats of each type to renew
 * only when renewing, so nothing is stored beforehand; the page shows which
 * members lose a seat because of that choice and what it costs, and paying stays
 * locked until the admin has seen, and where members are affected confirmed, who
 * loses a seat.
 */
export function RenewSeatsForm({ summary, preview: initialPreview, windowOpen, canManage, paymentResult, nowMs }: RenewSeatsFormProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const confirmID = useId();
  const lockReasonID = useId();

  const held = useMemo(() => countsOf(summary, "purchased"), [summary]);
  const saved = useMemo(() => countsOf(summary, "renewing"), [summary]);
  const [keep, setKeep] = useState<Record<PaidSeatType, string>>(() => stringCounts(held));
  const [preview, setPreview] = useState(initialPreview);
  const [previewFor, setPreviewFor] = useState<string | null>(initialPreview ? previewKey(initialPreview) : null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [paying, setPaying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const types = PAID_TYPES.filter((type) => held[type] > 0);
  const counts = parseCounts(keep, held);
  const key = counts ? countsKey(counts) : null;
  const shown = preview && key !== null && previewFor === key ? preview : null;
  const losingTotal = shown ? shown.lines.reduce((sum, line) => sum + line.losing.length, 0) : 0;
  const assignedTotal = shown ? shown.lines.reduce((sum, line) => sum + line.assigned, 0) : 0;
  const running = summary.subscription_kind === "seats" && Boolean(summary.period_end);
  const periodEnd = summary.period_end ?? "";
  const renewing = running && windowOpen && Boolean(summary.renewal_due);

  useEffect(() => {
    if (!renewing || counts === null || key === null || key === previewFor) return;

    const controller = new AbortController();
    const timer = window.setTimeout(() => {
      setPreviewError(null);
      previewSeatRenewal(counts, controller.signal)
        .then((next) => {
          setPreview(next);
          setPreviewFor(key);
          setConfirmed(false);
        })
        .catch((err: unknown) => {
          if (!controller.signal.aborted) setPreviewError(seatErrorMessage(t, err));
        });
    }, 250);

    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
    // counts is derived from key; the effect reruns only when the chosen numbers change.
  }, [key, previewFor, renewing]); // eslint-disable-line react-hooks/exhaustive-deps

  const paymentAlert =
    paymentResult === "success" ? (
      <Alert variant="success">
        <AlertTitle>{t("renew.payment.successTitle")}</AlertTitle>
        <AlertDescription>{t("renew.payment.successDescription")}</AlertDescription>
      </Alert>
    ) : paymentResult === "failed" ? (
      <Alert variant="destructive">
        <AlertTitle>{t("renew.payment.failedTitle")}</AlertTitle>
        <AlertDescription>{t("renew.payment.failedDescription")}</AlertDescription>
      </Alert>
    ) : null;

  const back = (
    <div>
      <Button asChild variant="ghost" size="sm">
        <Link href="/billing/seats">
          <ArrowRight data-icon="inline-start" aria-hidden="true" className="ltr:rotate-180" />
          {t("renew.back")}
        </Link>
      </Button>
    </div>
  );

  if (!running || types.length === 0) {
    return (
      <div className="flex flex-col gap-6">
        {back}
        {paymentAlert}
        <EmptyState
          icon={Users}
          title={t("renew.empty.title")}
          description={t("renew.empty.description")}
          primaryAction={
            canManage ? (
              <Button asChild>
                <Link href="/billing/seats/buy">{t("renew.empty.action")}</Link>
              </Button>
            ) : undefined
          }
        />
      </div>
    );
  }

  const timeline = <RenewalTimeline periodStart={summary.period_start} opensAt={summary.renewal_opens_at} periodEnd={periodEnd} nowMs={nowMs} />;
  const opensOn = summary.renewal_opens_at ? format.date(summary.renewal_opens_at) : format.date(periodEnd);

  if (!windowOpen) {
    return (
      <div className="flex animate-fade-in flex-col gap-6">
        {back}
        {paymentAlert}
        <StatusCard icon={Clock} title={t("renew.status.closedTitle")} description={t("renew.status.closed", { date: opensOn })}>
          {timeline}
        </StatusCard>
        <AppCard title={t("renew.current.title")} description={t("renew.current.description")}>
          <div className="flex flex-col divide-y divide-border">
            {types.map((type) => (
              <SeatStockRow key={type} type={type} summary={summary} />
            ))}
          </div>
        </AppCard>
        <RenewalRules />
      </div>
    );
  }

  if (!summary.renewal_due) {
    return (
      <div className="flex animate-fade-in flex-col gap-6">
        {back}
        {paymentAlert}
        <StatusCard
          icon={CircleCheck}
          title={t("renew.status.doneTitle")}
          description={t("renew.status.done", { date: format.date(initialPreview?.next_period_end ?? periodEnd) })}
        >
          {timeline}
        </StatusCard>
        <RenewalRules />
      </div>
    );
  }

  async function saveCounts(): Promise<boolean> {
    if (!counts) return false;

    try {
      for (const type of types) {
        if (counts[type] !== saved[type]) await updateSeatRenewal({ seat: type, quantity: counts[type] });
      }
    } catch (err) {
      setError(seatErrorMessage(t, err));

      return false;
    }

    return true;
  }

  async function pay() {
    setPaying(true);
    setError(null);

    if (await saveCounts()) {
      try {
        const order = await createSeatOrder({ kind: "renewal", return_url: `${window.location.origin}/billing/seats/renew` });
        if (order.payment_url) {
          window.location.assign(order.payment_url);

          return;
        }

        setError(t("buy.noPaymentURL"));
      } catch (err) {
        setError(seatErrorMessage(t, err));
      }
    }

    setPaying(false);
  }

  const toPay = shown?.payable_irr ?? 0;
  const canPay = canManage && toPay > 0 && (losingTotal === 0 || confirmed);
  const lockReason = !shown
    ? null
    : toPay <= 0
      ? counts && types.every((type) => counts[type] === 0)
        ? t("renew.summary.noSeatsKept")
        : t("renew.summary.nothingToPay")
      : losingTotal > 0 && !confirmed
        ? t("renew.summary.payLockedConfirm")
        : null;

  return (
    <div className="flex flex-col gap-6">
      {back}
      {paymentAlert}

      <StatusCard icon={CalendarClock} title={t("renew.status.openTitle")} description={t("renew.status.open", { date: format.date(periodEnd) })}>
        {timeline}
      </StatusCard>

      <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex min-w-0 flex-col gap-6">
          <AppCard title={t("renew.counts.title")} description={t("renew.counts.description")}>
            <div className="flex flex-col divide-y divide-border">
              {types.map((type) => (
                <SeatCountRow
                  key={type}
                  type={type}
                  value={keep[type]}
                  line={shown?.lines.find((line) => line.seat === type)}
                  summary={summary}
                  disabled={!canManage || paying}
                  onChange={(next) => setKeep((current) => ({ ...current, [type]: next }))}
                />
              ))}
            </div>
            {!canManage ? <p className="mt-4 text-sm text-muted-foreground">{t("renew.noAccess")}</p> : null}
          </AppCard>

          <AppCard title={t("renew.impact.title")}>
            <div className="flex flex-col gap-4">
              {shown ? (
                <p role="status" className="sr-only">
                  {losingTotal > 0 ? t("renew.impact.someTitle", { count: losingTotal }) : t("renew.impact.none")}
                </p>
              ) : null}

              {previewError ? (
                <Alert variant="destructive">
                  <AlertDescription>{previewError}</AlertDescription>
                </Alert>
              ) : null}

              {!shown && !previewError ? <Spinner label={t("renew.impact.loading")} /> : null}

              {shown ? <ImpactSummary preview={shown} losingTotal={losingTotal} assignedTotal={assignedTotal} /> : null}
            </div>

            {shown && losingTotal > 0 && canManage ? (
              <div className="mt-4 flex items-start gap-3 rounded-lg border border-border bg-muted/20 p-3 transition-colors duration-150 has-data-[state=checked]:border-primary-subtle-border has-data-[state=checked]:bg-primary-soft">
                <Checkbox id={confirmID} checked={confirmed} onCheckedChange={(value) => setConfirmed(value === true)} className="mt-0.5" />
                <Label htmlFor={confirmID} className="text-sm leading-6">
                  {t("renew.impact.confirm", { count: losingTotal })}
                </Label>
              </div>
            ) : null}
          </AppCard>

          <RenewalRules />
        </div>

        <AppCard title={t("renew.summary.title")} className="lg:sticky lg:top-20">
          <div className="flex flex-col gap-4">
            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}

            {shown ? (
              <DefinitionList>
                {shown.lines
                  .filter((line) => line.keep - line.renewed > 0)
                  .map((line) => (
                    <DefinitionListItem
                      key={line.seat}
                      label={t("renew.summary.line", { count: line.keep - line.renewed, seat: t(`seats.${line.seat}`) })}
                      value={<Amount value={format.toman(line.unit_price_irr * (line.keep - line.renewed))} />}
                    />
                  ))}
                <DefinitionListItem label={t("buy.price")} value={<Amount value={format.toman(shown.price_irr)} />} />
                <DefinitionListItem label={t("buy.tax", { percent: Number(shown.tax_percent) || 0 })} value={<Amount value={format.toman(shown.tax_irr)} />} />
                <DefinitionListItem className="border-t border-border pt-3" label={t("buy.payable")} emphasis value={<Amount value={format.toman(shown.payable_irr)} />} />
                <DefinitionListItem label={t("renew.summary.currentEnd")} value={format.date(shown.period_end)} />
                <DefinitionListItem label={t("renew.summary.newEnd")} value={format.date(shown.next_period_end)} />
              </DefinitionList>
            ) : (
              <Spinner label={t("renew.impact.loading")} />
            )}

            {lockReason ? (
              <p id={lockReasonID} className="flex animate-fade-in items-start gap-2 text-sm leading-6 text-muted-foreground">
                <Lock aria-hidden="true" className="mt-1 size-3.5 shrink-0" />
                {lockReason}
              </p>
            ) : null}

            {canManage && toPay > 0 ? (
              <Button
                type="button"
                className="group w-full"
                disabled={!canPay || paying || counts === null}
                aria-busy={paying || undefined}
                aria-describedby={lockReason ? lockReasonID : undefined}
                onClick={pay}
              >
                {paying ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <CreditCard data-icon="inline-start" className="transition-transform duration-150 group-enabled:group-hover:-rotate-6" />
                )}
                {t("renew.actions.pay", { amount: format.toman(toPay) })}
              </Button>
            ) : null}
          </div>
        </AppCard>
      </div>
    </div>
  );
}

/** The card that says whether renewal is open, with where the period stands. */
function StatusCard({ icon: Icon, title, description, children }: { icon: typeof Clock; title: string; description: string; children: ReactNode }) {
  return (
    <AppCard
      title={
        <span className="flex items-center gap-2">
          <span aria-hidden="true" className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary [&_svg]:size-4">
            <Icon />
          </span>
          {title}
        </span>
      }
      description={description}
    >
      {children}
    </AppCard>
  );
}

/** An amount that settles in when it is recalculated. */
function Amount({ value }: { value: string }) {
  return (
    <span key={value} className="inline-block animate-value-change tabular-nums">
      {value}
    </span>
  );
}

function SeatIdentity({ type }: { type: PaidSeatType }) {
  const t = useTranslations("billingSeats");
  const Icon = SEAT_ICONS[type];

  return (
    <div className="flex items-center gap-3">
      <span aria-hidden="true" className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary [&_svg]:size-4">
        <Icon />
      </span>
      <span className="text-sm font-semibold">{t(`seats.${type}`)}</span>
    </div>
  );
}

function SeatStats({ summary, type }: { summary: SeatSummary; type: PaidSeatType }) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const stock = summary.paid.find((item) => item.type === type);

  return (
    <p className="text-xs tabular-nums text-muted-foreground">
      {t("renew.seat.stats", {
        held: format.count(stock?.purchased ?? 0),
        assigned: format.count(stock?.assigned ?? 0),
        available: format.count(stock?.available ?? 0),
      })}
    </p>
  );
}

/** The seats held of one type, read-only, while renewal is not open yet. */
function SeatStockRow({ type, summary }: { type: PaidSeatType; summary: SeatSummary }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 py-3 first:pt-0 last:pb-0">
      <SeatIdentity type={type} />
      <SeatStats summary={summary} type={type} />
    </div>
  );
}

function SeatCountRow({
  type,
  value,
  line,
  summary,
  disabled,
  onChange,
}: {
  type: PaidSeatType;
  value: string;
  line?: SeatRenewalPreviewLine;
  summary: SeatSummary;
  disabled: boolean;
  onChange: (value: string) => void;
}) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const titleID = useId();
  const heldCount = summary.paid.find((item) => item.type === type)?.purchased ?? 0;
  const parsed = parseCount(value, heldCount);
  const kept = parsed ?? 0;

  return (
    <section aria-labelledby={titleID} className="flex flex-col gap-4 py-5 first:pt-0 last:pb-0">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div id={titleID}>
          <SeatIdentity type={type} />
        </div>
        <SeatStats summary={summary} type={type} />
      </div>

      <div className="flex flex-col gap-2">
        <div aria-hidden="true" className="flex h-2 overflow-hidden rounded-full bg-muted">
          <span className="flex-1 bg-primary transition-[flex-grow] duration-300 ease-out" style={{ flexGrow: kept }} />
          <span className="flex-1 border-s border-neutral-soft-border bg-neutral-soft transition-[flex-grow] duration-300 ease-out" style={{ flexGrow: heldCount - kept }} />
        </div>
        <p className="text-xs tabular-nums text-muted-foreground">{t("renew.seat.bar", { kept: format.count(kept), ends: format.count(heldCount - kept) })}</p>
      </div>

      <div className="flex items-end gap-2">
        <AppNumberField
          fieldClassName="flex-1"
          label={t("renew.seat.keep", { seat: t(`seats.${type}`) })}
          description={t("renew.seat.keepHint", { held: heldCount })}
          min={0}
          max={heldCount}
          inputMode="numeric"
          value={value}
          disabled={disabled}
          error={parsed === null ? t("renew.seat.invalid", { held: heldCount }) : undefined}
          onChange={(event) => onChange(event.target.value)}
        />
        <Button type="button" variant="outline" disabled={disabled || value === String(heldCount)} onClick={() => onChange(String(heldCount))}>
          {t("renew.seat.keepAll", { held: heldCount, seat: t(`seats.${type}`) })}
        </Button>
      </div>

      {line && line.renewed > 0 ? <p className="text-sm text-muted-foreground">{t("renew.seat.renewedNote", { count: line.renewed })}</p> : null}
    </section>
  );
}

function ImpactSummary({ preview, losingTotal, assignedTotal }: { preview: SeatRenewalPreview; losingTotal: number; assignedTotal: number }) {
  const t = useTranslations("billingSeats");

  if (losingTotal === 0) {
    return (
      <div className="flex animate-fade-in flex-col gap-2">
        <p className="flex items-center gap-2 text-sm font-semibold">
          <CircleCheck aria-hidden="true" className="size-4 text-success" />
          {t("renew.impact.none")}
        </p>
        <p className="text-sm leading-6 text-muted-foreground">{t("renew.impact.noneDescription")}</p>
        {assignedTotal > 0 ? <p className="text-sm leading-6 text-muted-foreground">{t("renew.impact.unpaid", { count: assignedTotal })}</p> : null}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <Alert variant="warning" className="animate-fade-in">
        <AlertTitle>{t("renew.impact.someTitle", { count: losingTotal })}</AlertTitle>
        <AlertDescription>{t("renew.impact.someDescription")}</AlertDescription>
      </Alert>

      {preview.lines
        .filter((line) => line.losing.length > 0)
        .map((line) => (
          <div key={line.seat} className="flex flex-col gap-2">
            <h4 className="text-sm font-semibold">{t("renew.impact.group", { seat: t(`seats.${line.seat}`), count: line.losing.length })}</h4>
            <ul className="flex flex-col divide-y divide-border rounded-lg border border-border">
              {line.losing.map((member, index) => (
                <MemberRow key={member.user_id} member={member} index={index} />
              ))}
            </ul>
          </div>
        ))}

      <p className="text-sm leading-6 text-muted-foreground">{t("renew.impact.rule")}</p>
      <p className="text-sm leading-6 text-muted-foreground">{t("renew.impact.unpaid", { count: assignedTotal })}</p>
    </div>
  );
}

function MemberRow({ member, index }: { member: SeatRenewalPreviewMember; index: number }) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const label = member.name || member.email;

  return (
    <li className="flex animate-fade-in flex-wrap items-center justify-between gap-2 p-3" style={{ animationDelay: `${Math.min(index, 6) * 40}ms` }}>
      <div className="flex min-w-0 items-center gap-3">
        <span aria-hidden="true" className="flex size-8 shrink-0 items-center justify-center rounded-full bg-neutral-soft text-xs font-semibold text-neutral-soft-foreground">
          {label ? label.trim().slice(0, 1).toUpperCase() : <Users className="size-4" />}
        </span>
        <div className="flex min-w-0 flex-col">
          <span className="truncate text-sm font-medium">{label || t("renew.impact.memberHidden")}</span>
          {member.name && member.email ? (
            <span dir="ltr" className="truncate text-start text-xs text-muted-foreground">
              {member.email}
            </span>
          ) : null}
          {member.assigned_at ? <span className="text-xs text-muted-foreground">{t("renew.impact.assignedAt", { date: format.date(member.assigned_at) })}</span> : null}
        </div>
      </div>
      <StatusBadge status={member.fallback === "free" ? "info" : "warning"}>
        {member.fallback === "free" ? t("renew.impact.fallbackFree") : t("renew.impact.fallbackUnassigned")}
      </StatusBadge>
    </li>
  );
}

function countsOf(summary: SeatSummary, field: "purchased" | "renewing"): Counts {
  const pick = (type: PaidSeatType) => summary.paid.find((item) => item.type === type)?.[field] ?? 0;

  return { pro: pick("pro"), max: pick("max") };
}

function stringCounts(counts: Counts): Record<PaidSeatType, string> {
  return { pro: String(counts.pro), max: String(counts.max) };
}

function countsKey(counts: Counts): string {
  return `${counts.pro}:${counts.max}`;
}

/** The counts a preview was computed for. */
function previewKey(preview: SeatRenewalPreview): string {
  const keep = (type: PaidSeatType) => preview.lines.find((line) => line.seat === type)?.keep ?? 0;

  return countsKey({ pro: keep("pro"), max: keep("max") });
}

function parseCount(value: string, held: number): number | null {
  if (!/^\d+$/.test(value.trim())) return null;

  const parsed = Number.parseInt(value, 10);

  return parsed <= held ? parsed : null;
}

function parseCounts(values: Record<PaidSeatType, string>, held: Counts): Counts | null {
  const pro = parseCount(values.pro, held.pro);
  const max = parseCount(values.max, held.max);

  return pro === null || max === null ? null : { pro, max };
}
