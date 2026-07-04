import * as React from "react";

import { cn } from "@/lib/utils";

type FormSectionProps = {
  title: React.ReactNode;
  description?: React.ReactNode;
  footer?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
};

export function FormSection({ title, description, footer, children, className }: FormSectionProps) {
  return (
    <section className={cn("grid gap-4 rounded-xl border border-border bg-card p-4 md:grid-cols-[minmax(0,16rem)_1fr] md:p-6", className)}>
      <div>
        <h2 className="text-base font-semibold text-card-foreground">{title}</h2>
        {description ? <p className="mt-1 text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      <div className="flex min-w-0 flex-col gap-4">
        {children}
        {footer ? <div className="flex flex-wrap items-center justify-end gap-2 border-t border-border pt-4">{footer}</div> : null}
      </div>
    </section>
  );
}
