"use client";

import Link from "next/link";
import {
  ArrowUpRight,
  CheckCircle2,
  CircleDollarSign,
  Gauge,
  KeyRound,
  Layers3,
  Lock,
  Network,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
  Users,
} from "lucide-react";

import { HelpLink } from "@/components/app-ui/HelpLink";
import { ProviderLogo } from "@/components/app-ui/ProviderLogo";
import { Button } from "@/components/ui/button";
import type {
  QuickStartAction,
  QuickStartExploreCardID,
  QuickStartJourneyStep,
  QuickStartStepID,
  QuickStartStepStatus,
} from "@/features/quick-start/journey";
import type { ProviderCatalogProvider } from "@/lib/contracts/provider-catalog";
import { cn } from "@/lib/utils";

export type TranslateValues = Record<string, string | number | Date>;
export type TranslateFn = (key: string, values?: TranslateValues) => string;

export function HeaderDescription(props: {
  docsBaseHref: string;
  t: TranslateFn;
}) {
  return (
    <>
      {props.t("subtitle")}{" "}
      <HelpLink href={`${props.docsBaseHref}/quick-start`} />
    </>
  );
}

export function StepMarker(props: { label: string; status: QuickStartStepStatus }) {
  return (
    <span
      className={cn(
        "absolute start-0 top-1 flex size-10 items-center justify-center rounded-lg border text-sm font-semibold",
        props.status === "done"
          ? "border-success/30 bg-success-muted text-success"
          : props.status === "active"
            ? "border-primary/30 bg-primary/10 text-primary"
            : "border-border bg-muted/40 text-muted-foreground",
      )}
    >
      {props.status === "done" ? <CheckCircle2 className="size-5" /> : props.label}
    </span>
  );
}

export function QuickStartStepActive(props: {
  docsBaseHref: string;
  featuredProviders: ProviderCatalogProvider[];
  onAction: () => void;
  onSkip: () => void;
  step: QuickStartJourneyStep;
  t: TranslateFn;
}) {
  const { step, t } = props;

  return (
    <div className="overflow-hidden rounded-lg border border-card-border bg-card">
      <span className="block h-0.5 bg-primary" aria-hidden="true" />
      <div className="flex flex-col gap-3 p-4">
        <h3 className="text-base font-semibold md:text-lg">{t(`steps.${step.id}.title`)}</h3>
        <p className="text-sm leading-7 text-muted-foreground">{t(`steps.${step.id}.description`)}</p>

        {props.featuredProviders.length > 0 ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {props.featuredProviders.map((providerItem) => (
              <Link
                key={providerItem.provider}
                href={`/provider-accounts/new?provider=${encodeURIComponent(providerItem.provider)}`}
                className="group flex items-center gap-2 rounded-lg border border-border bg-background px-3 py-3 transition-colors hover:border-primary/30 hover:bg-accent"
              >
                <span className="flex size-10 shrink-0 items-center justify-center rounded-md bg-muted">
                  <ProviderLogo
                    provider={providerItem.provider}
                    displayName={providerItem.display_name}
                    logoURL={providerItem.logo_url}
                    size={24}
                  />
                </span>
                <span className="min-w-0 flex-1 text-sm font-medium leading-6">{providerItem.display_name}</span>
                <ArrowUpRight className="size-4 shrink-0 text-muted-foreground transition group-hover:text-foreground rtl:-rotate-90" aria-hidden="true" />
              </Link>
            ))}
          </div>
        ) : null}

        <div className="flex flex-wrap items-center gap-2">
          <StepActionButton stepID={step.id} status={step.status} action={step.action} t={t} onAction={props.onAction} />
          <Button type="button" variant="ghost" size="sm" onClick={props.onSkip}>
            {t("steps.skip")}
          </Button>
          <StepLearnLink stepID={step.id} t={t} docsBaseHref={props.docsBaseHref} />
        </div>

        <p className="text-xs leading-6 text-muted-foreground">{t(`steps.${step.id}.learn`)}</p>
      </div>
    </div>
  );
}

export function QuickStartStepDone(props: { onAction: () => void; step: QuickStartJourneyStep; t: TranslateFn }) {
  const { step, t } = props;

  return (
    <div className="flex min-h-12 flex-wrap items-center gap-x-3 gap-y-1 py-2">
      <span className="text-sm font-medium">{t(`steps.${step.id}.title`)}</span>
      {step.result ? (
        <span className="min-w-0 truncate font-mono text-xs text-muted-foreground" dir="ltr" title={step.result}>
          {step.result}
        </span>
      ) : null}
      <StepManageLink action={step.action} stepID={step.id} t={t} onAction={props.onAction} />
    </div>
  );
}

export function QuickStartStepPending(props: { onResume: () => void; step: QuickStartJourneyStep; t: TranslateFn }) {
  const { step, t } = props;
  const reason =
    step.status === "skipped"
      ? t("locked.skipped")
      : step.status === "locked"
        ? t("locked.afterStep", { index: step.waitsForIndex ?? 1 })
        : step.blockedBy === "feature"
          ? t("locked.feature")
          : t("locked.permission");
  const reasonTitle = step.blockedBy ? t(`blocked.${step.blockedBy}`) : undefined;

  return (
    <div className="flex min-h-12 flex-wrap items-center gap-x-3 gap-y-1 rounded-lg border border-border px-3 py-2">
      <span className="text-sm text-muted-foreground">{t(`steps.${step.id}.title`)}</span>
      <span className="ms-auto flex items-center gap-1.5 text-xs text-muted-foreground" title={reasonTitle}>
        <Lock className="size-3.5" aria-hidden="true" />
        {reason}
      </span>
      {step.status === "skipped" ? (
        <Button type="button" variant="link" size="sm" className="h-auto p-0 text-xs" onClick={props.onResume}>
          {t("steps.resume")}
        </Button>
      ) : null}
    </div>
  );
}

export function StepLearnLink(props: { docsBaseHref: string; stepID: QuickStartStepID; t: TranslateFn }) {
  const docsSlug = docsSlugByStep[props.stepID];
  if (!docsSlug) return null;

  return (
    <HelpLink href={`${props.docsBaseHref}${docsSlug}`} className="border-transparent bg-transparent px-0">
      {props.t("steps.learnMore")}
    </HelpLink>
  );
}

export const docsSlugByStep: Partial<Record<QuickStartStepID, string>> = {
  connect_provider: "/quick-start",
  create_pat: "/guides/api-keys",
  inspect_activity: "/observability-logs",
};

export function StepManageLink(props: {
  action: QuickStartAction;
  onAction: () => void;
  stepID: QuickStartStepID;
  t: TranslateFn;
}) {
  const label = props.t("steps.donePrefix");
  // The step-specific wording stays as the accessible name so "Manage" is never ambiguous.
  const title = props.t(`steps.${props.stepID}.manageAction`);

  if (props.action.type === "link") {
    return (
      <Button type="button" asChild variant="link" size="sm" className="ms-auto h-auto p-0 text-xs">
        <Link href={props.action.href} title={title} aria-label={title} onClick={props.onAction}>
          {label}
        </Link>
      </Button>
    );
  }

  return (
    <Button
      type="button"
      variant="link"
      size="sm"
      className="ms-auto h-auto p-0 text-xs"
      title={title}
      aria-label={title}
      onClick={props.onAction}
    >
      {label}
    </Button>
  );
}

export function StepActionButton(props: {
  action: QuickStartAction;
  onAction: () => void;
  status: QuickStartStepStatus;
  stepID: QuickStartStepID;
  t: TranslateFn;
}) {
  const label = props.t(`steps.${props.stepID}.action`);

  if (props.action.type === "link") {
    return (
      <Button type="button" asChild size="sm">
        <Link href={props.action.href} onClick={props.onAction}>
          <span>{label}</span>
          <ArrowUpRight className="shrink-0 rtl:-rotate-90" data-icon="inline-end" />
        </Link>
      </Button>
    );
  }

  return (
    <Button type="button" size="sm" onClick={props.onAction}>
      <span>{label}</span>
      {props.action.type === "anchor" ? (
        <TerminalSquare data-icon="inline-end" />
      ) : (
        <KeyRound data-icon="inline-end" />
      )}
    </Button>
  );
}

export function ExploreChip(props: { cardID: QuickStartExploreCardID; t: TranslateFn }) {
  const Icon = exploreIconMap[props.cardID];

  return (
    <Link
      href={exploreHrefMap[props.cardID]}
      className="group inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-1.5 text-xs font-medium transition-colors hover:border-primary/30 hover:bg-accent"
    >
      <Icon className="size-3.5 text-primary" aria-hidden="true" />
      <span>{props.t(`explore.cards.${props.cardID}.title`)}</span>
      <ArrowUpRight className="size-3.5 text-muted-foreground transition group-hover:text-foreground rtl:-rotate-90" aria-hidden="true" />
    </Link>
  );
}

export const exploreHrefMap: Record<QuickStartExploreCardID, string> = {
  provider_accounts: "/provider-accounts",
  virtual_users: "/virtual-users",
  teams: "/teams",
  guardrails: "/guardrails",
  prompts: "/prompts",
  metrics: "/metrics",
  logs: "/logs",
  billing_plan: "/billing/plan",
};

export const exploreIconMap: Record<QuickStartExploreCardID, typeof Network> = {
  provider_accounts: Network,
  virtual_users: KeyRound,
  teams: Users,
  guardrails: ShieldCheck,
  prompts: Sparkles,
  metrics: Gauge,
  logs: Layers3,
  billing_plan: CircleDollarSign,
};

