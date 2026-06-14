"use client";

import Link from "next/link";

import { Button } from "@/components/ui/button";
import type { Readiness } from "./types";
export function EmptyOverview(props: {
  title: string;
  description: string;
  setupLabel: string;
  setupHref: string;
  openExistingLabel: string;
  onOpenFirst: () => void;
  hasAccount: boolean;
  createDisabled?: boolean;
}) {
  return (
    <section className="space-y-3 rounded-lg border border-dashed p-6 text-center">
      <h3 className="text-lg font-semibold">{props.title}</h3>
      <p className="text-sm text-muted-foreground">{props.description}</p>
      <div className="flex flex-wrap items-center justify-center gap-2">
        <Button type="button" asChild disabled={props.createDisabled}>
          <Link href={props.setupHref}>{props.setupLabel}</Link>
        </Button>
        {props.hasAccount ? (
          <Button type="button" variant="outline" onClick={props.onOpenFirst}>{props.openExistingLabel}</Button>
        ) : null}
      </div>
    </section>
  );
}

export function SummaryKPI(props: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-muted/20 p-3">
      <div className="text-xs text-muted-foreground">{props.label}</div>
      <div className="truncate text-sm font-semibold">{props.value}</div>
    </div>
  );
}

export function ReadinessBadge(props: {
  readiness: Readiness;
  labels: Record<Readiness, string>;
}) {
  const mapping: Record<Readiness, { label: string; className: string }> = {
    ready: { label: props.labels.ready, className: "bg-emerald-100 text-emerald-700" },
    needs_credentials: { label: props.labels.needs_credentials, className: "bg-amber-100 text-amber-700" },
    needs_access: { label: props.labels.needs_access, className: "bg-amber-100 text-amber-700" },
    pricing_incomplete: { label: props.labels.pricing_incomplete, className: "bg-sky-100 text-sky-700" },
  };
  const view = mapping[props.readiness];
  return <span className={`rounded-full px-2 py-1 text-[10px] font-semibold ${view.className}`}>{view.label}</span>;
}


