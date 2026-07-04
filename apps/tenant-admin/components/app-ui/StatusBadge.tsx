import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type Status = "active" | "inactive" | "pending" | "success" | "warning" | "error" | "neutral";

const statusClasses: Record<Status, string> = {
  active: "border-success/30 bg-success/10 text-success",
  inactive: "border-border bg-muted text-muted-foreground",
  pending: "border-warning/30 bg-warning/10 text-warning",
  success: "border-success/30 bg-success/10 text-success",
  warning: "border-warning/30 bg-warning/10 text-warning",
  error: "border-destructive/30 bg-destructive/10 text-destructive",
  neutral: "border-border bg-secondary text-secondary-foreground",
};

type StatusBadgeProps = React.ComponentProps<typeof Badge> & {
  status?: Status;
};

export function StatusBadge({ status = "neutral", className, variant = "outline", ...props }: StatusBadgeProps) {
  return <Badge variant={variant} className={cn(statusClasses[status], className)} {...props} />;
}
