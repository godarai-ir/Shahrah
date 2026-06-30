import { LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type StatCardProps = {
  title: string;
  value: string;
  subtitle?: string;
  icon?: LucideIcon;
  trend?: {
    value: number;
    positive?: boolean;
  };
  color?: "indigo" | "purple" | "green" | "amber" | "red";
  className?: string;
};

const colorMap = {
  indigo: {
    bg: "bg-primary/10",
    text: "text-primary",
    trend: "text-primary",
  },
  purple: {
    bg: "surface-primary-soft",
    text: "text-[var(--primary-soft-foreground)]",
    trend: "text-[var(--primary-soft-foreground)]",
  },
  green: {
    bg: "bg-success-muted",
    text: "text-success",
    trend: "text-success",
  },
  amber: {
    bg: "bg-warning-muted",
    text: "text-warning",
    trend: "text-warning",
  },
  red: {
    bg: "bg-destructive/10",
    text: "text-destructive",
    trend: "text-destructive",
  },
};

export function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  color = "indigo",
  className,
}: StatCardProps) {
  const colors = colorMap[color];

  return (
    <Card
      className={cn(
        "group card-hover",
        className,
      )}
    >
      <CardContent className="p-4 md:p-6">
        <div className="flex items-start justify-between gap-4">
          {/* Content */}
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-muted-foreground truncate">
              {title}
            </p>
            <p className="mt-2 text-2xl font-bold tracking-tight md:text-3xl">
              {value}
            </p>
            {subtitle && (
              <p className="mt-1 text-xs text-muted-foreground">
                {subtitle}
              </p>
            )}
            {trend && (
              <div className="mt-2 flex items-center gap-1">
                <span
                  className={cn(
                    "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                    trend.positive
                      ? "bg-success/10 text-success"
                      : "bg-destructive/10 text-destructive",
                  )}
                >
                  {trend.positive ? "↑" : "↓"} {Math.abs(trend.value)}%
                </span>
              </div>
            )}
          </div>

          {/* Icon */}
          {Icon && (
            <div
              className={cn(
                "flex h-10 w-10 items-center justify-center rounded-xl transition-transform duration-200 group-hover:scale-110",
                colors.bg,
                colors.text,
              )}
            >
              <Icon className="h-5 w-5" aria-hidden="true" />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
