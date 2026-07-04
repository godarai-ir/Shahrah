import * as React from "react";

import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type AppCardProps = {
  title?: React.ReactNode;
  description?: React.ReactNode;
  actions?: React.ReactNode;
  footer?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
  contentClassName?: string;
};

export function AppCard({ title, description, actions, footer, children, className, contentClassName }: AppCardProps) {
  const hasHeader = title || description || actions;

  return (
    <Card className={cn("card-hover", className)}>
      {hasHeader ? (
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="min-w-0">
            {title ? <CardTitle>{title}</CardTitle> : null}
            {description ? <CardDescription className="mt-1">{description}</CardDescription> : null}
          </div>
          {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
        </CardHeader>
      ) : null}
      {children ? <CardContent className={cn(!hasHeader && "pt-4 md:pt-6", contentClassName)}>{children}</CardContent> : null}
      {footer ? <CardFooter>{footer}</CardFooter> : null}
    </Card>
  );
}
