import * as React from "react";

import { CopyAction, type CopyActionProps } from "@/components/app-ui/CopyAction";
import { CardDescription, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type EntityIdentityHeaderProps = {
  title: React.ReactNode;
  identifier?: string;
  description?: React.ReactNode;
  status?: React.ReactNode;
  actions?: React.ReactNode;
  leading?: React.ReactNode;
  copy?: Omit<CopyActionProps, "value">;
  className?: string;
  titleClassName?: string;
};

export function EntityIdentityHeader({ title, identifier, description, status, actions, leading, copy, className, titleClassName }: EntityIdentityHeaderProps) {
  return (
    <div data-slot="entity-identity-header" className={cn("flex min-w-0 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between", className)}>
      <div className="flex min-w-0 items-start gap-3">
        {leading ? <div className="shrink-0">{leading}</div> : null}
        <div className="flex min-w-0 flex-col gap-1">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <CardTitle className={cn("truncate", titleClassName)}>{title}</CardTitle>
            {status}
          </div>
          {identifier ? (
            <div className="flex min-w-0 items-center gap-1">
              <code className="truncate font-mono text-xs text-muted-foreground" dir="ltr">{identifier}</code>
              {copy ? <CopyAction value={identifier} {...copy} /> : null}
            </div>
          ) : null}
          {description ? <CardDescription>{description}</CardDescription> : null}
        </div>
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </div>
  );
}
