import * as React from "react";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import type { DateTimeInput } from "@/lib/date-time";
import { cn } from "@/lib/utils";

export type SettingsAuditMetaProps = { updatedAt: DateTimeInput; updatedBy?: React.ReactNode; updatedAtLabel: React.ReactNode; updatedByLabel: React.ReactNode; fallback: string; locale?: string; className?: string };

export function SettingsAuditMeta({ updatedAt, updatedBy, updatedAtLabel, updatedByLabel, fallback, locale, className }: SettingsAuditMetaProps) {
  return <dl className={cn("grid gap-1 text-xs text-muted-foreground", className)}><div className="flex flex-wrap gap-1"><dt>{updatedAtLabel}:</dt><dd><DateTimeValue value={updatedAt} locale={locale} fallback={fallback} /></dd></div><div className="flex flex-wrap gap-1"><dt>{updatedByLabel}:</dt><dd>{updatedBy || fallback}</dd></div></dl>;
}
