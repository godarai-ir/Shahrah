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
  headerClassName?: string;
  titleContainerClassName?: string;
  actionsClassName?: string;
  contentClassName?: string;
};

export function AppCard({
  title,
  description,
  actions,
  footer,
  children,
  className,
  headerClassName,
  titleContainerClassName,
  actionsClassName,
  contentClassName,
}: AppCardProps) {
  const hasHeader = title || description || actions;

  return (
    <Card className={cn("card-hover", className)}>
      {hasHeader ? (
        <CardHeader className={cn("flex flex-col items-start justify-between gap-3 sm:flex-row sm:gap-4", headerClassName)}>
          <div className={cn("min-w-0", titleContainerClassName)}>
            {title ? <CardTitle>{title}</CardTitle> : null}
            {description ? <CardDescription className="mt-1">{description}</CardDescription> : null}
          </div>
          {actions ? <div className={cn("flex shrink-0 flex-wrap items-center gap-2", actionsClassName)}>{actions}</div> : null}
        </CardHeader>
      ) : null}
      {children ? <CardContent className={cn(!hasHeader && "pt-4 md:pt-6", contentClassName)}>{children}</CardContent> : null}
      {footer ? <CardFooter>{footer}</CardFooter> : null}
    </Card>
  );
}
