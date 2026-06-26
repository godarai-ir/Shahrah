"use client";

import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";

export function Field(props: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label>{props.label}</Label>
      {props.children}
    </div>
  );
}

export function ConfigSummaryPills(props: { pills: string[]; emptyLabel: string }) {
  const filtered = props.pills.filter(Boolean);
  if (filtered.length === 0) {
    return <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.emptyLabel}</div>;
  }
  return (
    <div className="flex flex-wrap gap-2">
      {filtered.map((pill) => (
        <Badge key={pill} variant="outline">
          {pill}
        </Badge>
      ))}
    </div>
  );
}
