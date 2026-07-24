import * as React from "react";

import { cn } from "@/lib/utils";

export type ShellPageMetadata = {
  title: string;
  description?: string;
  helpHref?: string;
};

type ShellPageHeadingProps = Omit<ShellPageMetadata, "helpHref"> & {
  as?: "h1" | "div";
  className?: string;
};

export function ShellPageHeading({
  title,
  description,
  as: Heading = "h1",
  className,
}: ShellPageHeadingProps) {
  return (
    <div data-slot="shell-page-heading" className={cn("min-w-0 flex-1", className)}>
      <Heading className="truncate text-base font-semibold leading-6 text-foreground md:text-lg">
        {title}
      </Heading>
      {description ? (
        <p className="truncate text-xs leading-5 text-muted-foreground" title={description}>
          {description}
        </p>
      ) : null}
    </div>
  );
}
