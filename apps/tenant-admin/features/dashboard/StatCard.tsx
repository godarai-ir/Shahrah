import { LucideIcon } from "lucide-react";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
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

export function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  color = "indigo",
  className,
}: StatCardProps) {
  const trendStatus = trend?.positive ? "success" : "error";
  const iconClassName = cn(
    color === "green" && "bg-success/10 text-success",
    color === "amber" && "bg-warning/10 text-warning",
    color === "red" && "tone-danger",
    color === "purple" && "surface-primary-soft",
  );

  return (
    <MetricCard
      label={title}
      value={value}
      description={subtitle}
      icon={Icon}
      className={className}
      iconClassName={iconClassName}
      trend={trend ? <StatusBadge status={trendStatus}>{trend.positive ? "↑" : "↓"} {Math.abs(trend.value)}%</StatusBadge> : undefined}
    />
  );
}
