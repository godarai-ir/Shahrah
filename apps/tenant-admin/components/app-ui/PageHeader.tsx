import * as React from "react";

import { cn } from "@/lib/utils";

type PageHeaderProps = {
  title: React.ReactNode;
  description?: React.ReactNode;
  eyebrow?: React.ReactNode;
  titleAddon?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
  titleClassName?: string;
  descriptionClassName?: string;
};

export function PageHeader({ title, description, eyebrow, titleAddon, actions, className, titleClassName, descriptionClassName }: PageHeaderProps) {
  return (
    <header className={cn("flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between", className)}>
      <div className="min-w-0 flex-1">
        {eyebrow ? <p className="mb-2 text-xs font-medium text-muted-foreground">{eyebrow}</p> : null}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h1 className={cn("text-2xl font-bold tracking-tight text-foreground md:text-3xl", titleClassName)}>{title}</h1>
          {titleAddon ? <div className="flex shrink-0 items-center gap-2">{titleAddon}</div> : null}
        </div>
        {description ? <div className={cn("mt-2 max-w-3xl text-sm leading-6 text-muted-foreground", descriptionClassName)}>{description}</div> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  );
}
