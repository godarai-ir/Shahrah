import * as React from "react";

import { AppCard } from "@/components/app-ui/AppCard";
import type { PermissionOption } from "@/components/app-ui/PermissionEditor";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type PermissionSummaryLabels = {
  selectedPermissions: string;
  coveredGroups: string;
  clearAll: string;
  empty: string;
};

export type PermissionSummaryProps = {
  permissions: PermissionOption[];
  value: string[];
  title: React.ReactNode;
  description?: React.ReactNode;
  labels: PermissionSummaryLabels;
  onClear?: () => void;
  formatNumber?: (value: number) => string;
  className?: string;
};

export function PermissionSummary({
  permissions,
  value,
  title,
  description,
  labels,
  onClear,
  formatNumber = String,
  className,
}: PermissionSummaryProps) {
  const stats = getPermissionSelectionSummary(value, permissions);

  return (
    <AppCard
      title={title}
      description={description}
      className={cn("shadow-none", className)}
      contentClassName="flex flex-col gap-3"
    >
      <SummaryRow label={labels.selectedPermissions} value={formatNumber(stats.count)} />
      <SummaryRow label={labels.coveredGroups} value={formatNumber(stats.groupCount)} variant="outline" />
      {value.length > 0 && onClear ? (
        <Button type="button" size="sm" variant="outline" onClick={onClear}>
          {labels.clearAll}
        </Button>
      ) : null}
      {value.length === 0 ? <p className="text-xs leading-5 text-muted-foreground">{labels.empty}</p> : null}
    </AppCard>
  );
}

export function getPermissionSelectionSummary(selected: string[], permissions: PermissionOption[]) {
  const permissionByKey = new Map(permissions.map((permission) => [permission.key, permission] as const));
  const groups = new Set(selected.map((key) => permissionByKey.get(key)?.group ?? "other"));
  return { count: selected.length, groupCount: groups.size };
}

function SummaryRow({ label, value, variant = "secondary" }: { label: React.ReactNode; value: string; variant?: "secondary" | "outline" }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <Badge variant={variant}>{value}</Badge>
    </div>
  );
}
