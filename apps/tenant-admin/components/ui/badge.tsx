import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex min-h-6 items-center justify-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
  {
    variants: {
      variant: {
        default: "border-transparent bg-primary text-primary-foreground",
        secondary: "border-neutral-soft-border bg-neutral-soft text-neutral-soft-foreground",
        destructive:
          "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground",
        outline: "border-border bg-card text-foreground",
        success:
          "border-success-soft-border bg-success-soft text-success-soft-foreground",
        warning:
          "border-warning-soft-border bg-warning-soft text-warning-soft-foreground",
        info:
          "border-info-soft-border bg-info-soft text-info-soft-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

function Badge({
  className,
  variant,
  ...props
}: React.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return (
    <span
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
