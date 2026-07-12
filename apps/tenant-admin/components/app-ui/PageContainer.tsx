import * as React from "react";

import { cn } from "@/lib/utils";

export type PageContainerSize = "wide" | "readable" | "full";
export type PageContainerSpacing = "compact" | "default" | "relaxed";
export type PageContainerPadding = "default" | "none";
export type PageContainerProps = React.ComponentProps<"div"> & {
  as?: "div" | "main" | "section";
  size?: PageContainerSize;
  spacing?: PageContainerSpacing;
  padding?: PageContainerPadding;
};

const sizeClasses: Record<PageContainerSize, string> = { wide: "max-w-7xl", readable: "max-w-5xl", full: "max-w-none" };
const spacingClasses: Record<PageContainerSpacing, string> = { compact: "gap-4", default: "gap-6", relaxed: "gap-8" };
const paddingClasses: Record<PageContainerPadding, string> = { default: "px-4 py-6 md:px-6", none: "p-0" };

export function PageContainer({ as: Component = "div", size = "wide", spacing = "default", padding = "default", className, ...props }: PageContainerProps) {
  return (
    <Component
      data-size={size}
      data-spacing={spacing}
      data-padding={padding}
      className={cn("mx-auto flex w-full min-w-0 flex-col", sizeClasses[size], spacingClasses[spacing], paddingClasses[padding], className)}
      {...props}
    />
  );
}
