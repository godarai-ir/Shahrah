"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { ScrollText, Users } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { Button } from "@/components/ui/button";

/** The rule keys under `billingSeats.renew.rules.items`, in the order an admin meets them. */
const RULES = ["counts", "noAuto", "pays", "notRenewed", "order", "fallback", "unpaid", "chooseYourself"] as const;

/**
 * The rules of seat renewal, stated before anything is paid: when it opens,
 * what it buys, what happens to seats that are not renewed, and which members
 * lose theirs. Shown in every state of the page.
 */
export function RenewalRules() {
  const t = useTranslations("billingSeats");

  return (
    <AppCard
      title={
        <span className="flex items-center gap-2">
          <ScrollText aria-hidden="true" className="size-4 text-muted-foreground" />
          {t("renew.rules.title")}
        </span>
      }
      description={t("renew.rules.description")}
    >
      <ol className="flex flex-col divide-y divide-border">
        {RULES.map((rule, index) => (
          <li key={rule} className="flex items-start gap-3 py-3 first:pt-0 last:pb-0">
            <span
              aria-hidden="true"
              className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-primary-soft text-xs font-semibold tabular-nums text-primary-soft-foreground"
            >
              {index + 1}
            </span>
            <div className="flex min-w-0 flex-col gap-2">
              <p className="text-sm leading-6">{t(`renew.rules.items.${rule}`)}</p>
              {rule === "chooseYourself" ? (
                <Button asChild variant="outline" size="sm" className="self-start">
                  <Link href="/users">
                    <Users data-icon="inline-start" />
                    {t("renew.impact.manageMembers")}
                  </Link>
                </Button>
              ) : null}
            </div>
          </li>
        ))}
      </ol>
    </AppCard>
  );
}
