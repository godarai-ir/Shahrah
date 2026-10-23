"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

import { useBillingFormat } from "../use-billing-format";

const DAY_MS = 24 * 60 * 60 * 1000;

type RenewalTimelineProps = {
  periodStart?: string | null;
  /** When the renewal window opens. */
  opensAt?: string | null;
  periodEnd: string;
  /** The request time, so the server and the browser draw the same position. */
  nowMs: number;
};

/**
 * Where the current period stands: how much has passed, where the renewal
 * window sits, and how many days are left. The bar fills once on arrival so the
 * position reads as progress; the dates and the days left say the same in text.
 */
export function RenewalTimeline({ periodStart, opensAt, periodEnd, nowMs }: RenewalTimelineProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const [arrived, setArrived] = useState(false);

  useEffect(() => {
    const frame = requestAnimationFrame(() => setArrived(true));

    return () => cancelAnimationFrame(frame);
  }, []);

  const start = periodStart ? Date.parse(periodStart) : Number.NaN;
  const end = Date.parse(periodEnd);
  const opens = opensAt ? Date.parse(opensAt) : Number.NaN;
  const span = end - start;
  const drawable = Number.isFinite(span) && span > 0;
  const clamp = (value: number) => Math.min(1, Math.max(0, value));
  const elapsed = drawable ? clamp((nowMs - start) / span) : 0;
  const windowStart = drawable && Number.isFinite(opens) ? clamp((opens - start) / span) : 1;
  const daysLeft = Math.max(0, Math.ceil((end - nowMs) / DAY_MS));

  return (
    <div className="flex flex-col gap-3">
      <p className="text-sm font-semibold tabular-nums">{daysLeft > 0 ? t("renew.timeline.daysLeft", { days: daysLeft }) : t("renew.timeline.lastDay")}</p>

      {drawable ? (
        <div className="relative flex h-2 overflow-hidden rounded-full bg-muted" role="img" aria-label={t("renew.timeline.label")}>
          <span
            aria-hidden="true"
            className="flex-1 bg-neutral-soft"
            style={{ flexGrow: windowStart }}
          />
          <span
            aria-hidden="true"
            className="flex-1 border-s border-warning-soft-border bg-warning-soft"
            style={{ flexGrow: 1 - windowStart }}
          />
          <span
            aria-hidden="true"
            className="absolute inset-y-0 start-0 rounded-full bg-primary transition-[width] duration-700 ease-out"
            style={{ width: `${arrived ? elapsed * 100 : 0}%` }}
          />
        </div>
      ) : null}

      <dl className="grid grid-cols-3 gap-2 text-xs">
        <div className="flex min-w-0 flex-col gap-0.5">
          <dt className="text-muted-foreground">{t("renew.timeline.start")}</dt>
          <dd className="font-medium">{periodStart ? format.date(periodStart) : "—"}</dd>
        </div>
        <div className="flex min-w-0 flex-col items-center gap-0.5 text-center">
          <dt className="text-muted-foreground">{t("renew.timeline.opens")}</dt>
          <dd className="font-medium">{opensAt ? format.date(opensAt) : "—"}</dd>
        </div>
        <div className="flex min-w-0 flex-col items-end gap-0.5 text-end">
          <dt className="text-muted-foreground">{t("renew.timeline.end")}</dt>
          <dd className="font-medium">{format.date(periodEnd)}</dd>
        </div>
      </dl>
    </div>
  );
}
