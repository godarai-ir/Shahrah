import * as React from "react";

import { cn } from "@/lib/utils";

type FormSectionProps = {
  title?: React.ReactNode;
  description?: React.ReactNode;
  actions?: React.ReactNode;
  footer?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
};

export function FormSection({ title, description, actions, footer, children, className }: FormSectionProps) {
  const hasHeader = Boolean(title || description);

  return (
    <section className={cn("grid gap-4 rounded-xl border border-border bg-card p-4 md:p-6", hasHeader && "md:grid-cols-[minmax(0,16rem)_1fr]", className)}>
      {hasHeader ? (
        <div className="flex items-start justify-between gap-3">
          <div>
            {title ? <h2 className="text-base font-semibold text-card-foreground">{title}</h2> : null}
            {description ? <p className="mt-1 text-sm leading-6 text-muted-foreground">{description}</p> : null}
          </div>
          {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
        </div>
      ) : null}
      <div className="flex min-w-0 flex-col gap-4">
        {children}
        {footer ? <div className="flex flex-wrap items-center justify-end gap-2 border-t border-border pt-4">{footer}</div> : null}
      </div>
    </section>
  );
}
