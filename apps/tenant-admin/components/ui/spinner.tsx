import * as React from "react";
import { Loader2 } from "lucide-react";

import { cn } from "@/lib/utils";

export type SpinnerProps = Omit<React.ComponentProps<typeof Loader2>, "aria-label"> & {
  label?: string;
  size?: "sm" | "default" | "lg";
};

const sizeClasses = { sm: "size-3", default: "size-4", lg: "size-6" } as const;

export function Spinner({ label, size = "default", className, ...props }: SpinnerProps) {
  return (
    <Loader2
      data-slot="spinner"
      role={label ? "status" : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      className={cn("animate-spin", sizeClasses[size], className)}
      {...props}
    />
  );
}
