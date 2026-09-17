import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type Status = "active" | "inactive" | "pending" | "success" | "warning" | "error" | "info" | "neutral";

const statusClasses: Record<Status, string> = {
  active: "border-success-soft-border bg-success-soft text-success-soft-foreground",
  inactive: "border-neutral-soft-border bg-neutral-soft text-neutral-soft-foreground",
  pending: "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
  success: "border-success-soft-border bg-success-soft text-success-soft-foreground",
  warning: "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
  error: "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground",
  info: "border-info-soft-border bg-info-soft text-info-soft-foreground",
  neutral: "border-neutral-soft-border bg-neutral-soft text-neutral-soft-foreground",
};

type StatusBadgeProps = React.ComponentProps<typeof Badge> & {
  status?: Status;
};

export function StatusBadge({ status = "neutral", className, variant = "outline", ...props }: StatusBadgeProps) {
  return <Badge variant={variant} className={cn(statusClasses[status], className)} {...props} />;
}
