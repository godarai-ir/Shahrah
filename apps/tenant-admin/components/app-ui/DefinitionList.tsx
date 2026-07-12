import * as React from "react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type DefinitionListProps = React.ComponentProps<"dl"> & { columns?: 1 | 2 | 3 | 4 };

export type DefinitionListItemProps = {
  label: React.ReactNode;
  value: React.ReactNode;
  icon?: LucideIcon;
  mono?: boolean;
  wrap?: boolean;
  emphasis?: boolean;
  surface?: boolean;
  valueDir?: "rtl" | "ltr" | "auto";
  title?: string;
  className?: string;
};

const columnClasses = {
  1: "grid-cols-1",
  2: "grid-cols-1 sm:grid-cols-2",
  3: "grid-cols-1 sm:grid-cols-2 2xl:grid-cols-3",
  4: "grid-cols-1 sm:grid-cols-2 xl:grid-cols-4",
} as const;

export function DefinitionList({ columns = 1, className, ...props }: DefinitionListProps) {
  return <dl className={cn("grid gap-x-4 gap-y-3", columnClasses[columns], className)} {...props} />;
}

export function DefinitionListItem({
  label, value, icon: Icon, mono, wrap, emphasis, surface, valueDir, title, className,
}: DefinitionListItemProps) {
  const shouldWrap = Boolean(wrap);
  const resolvedTitle = title ?? (!shouldWrap && typeof value === "string" ? value : undefined);

  return (
    <div className={cn("flex min-w-0 items-start gap-2", surface && "rounded-lg border border-border bg-muted/20 px-3 py-2", className)}>
      {Icon ? (
        <span className="mt-0.5 shrink-0 text-muted-foreground [&_svg]:size-4" aria-hidden="true"><Icon /></span>
      ) : null}
      <div className="min-w-0">
        <dt className="text-xs text-muted-foreground">{label}</dt>
        <dd
          dir={valueDir}
          title={resolvedTitle}
          className={cn("text-sm font-medium", mono && "font-mono text-xs", emphasis && "font-semibold", shouldWrap ? "break-all" : "truncate")}
        >
          {value}
        </dd>
      </div>
    </div>
  );
}
