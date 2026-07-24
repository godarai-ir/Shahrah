import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { AlertCircle, CheckCircle, Info, TriangleAlert } from "lucide-react";

import { cn } from "@/lib/utils";

const alertVariants = cva(
  "relative w-full rounded-lg border p-4 text-sm [&>svg]:size-4 [&>svg]:shrink-0",
  {
    variants: {
      variant: {
        default:
          "border-border bg-card text-foreground [&>svg]:text-foreground",
        destructive:
          "border-destructive-soft-border bg-destructive-soft text-destructive-soft-foreground [&>svg]:text-destructive-soft-foreground",
        success:
          "border-success-soft-border bg-success-soft text-success-soft-foreground [&>svg]:text-success-soft-foreground",
        warning:
          "border-warning-soft-border bg-warning-soft text-warning-soft-foreground [&>svg]:text-warning-soft-foreground",
        info:
          "border-info-soft-border bg-info-soft text-info-soft-foreground [&>svg]:text-info-soft-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

function Alert({
  className,
  variant,
  icon,
  children,
  ...props
}: React.ComponentProps<"div"> &
  VariantProps<typeof alertVariants> & {
    icon?: React.ReactNode;
  }) {
  const defaultIcons: Record<string, React.ReactNode> = {
    destructive: <AlertCircle aria-hidden="true" />,
    success: <CheckCircle aria-hidden="true" />,
    warning: <TriangleAlert aria-hidden="true" />,
    info: <Info aria-hidden="true" />,
  };

  const resolvedIcon = icon ?? defaultIcons[variant ?? "default"];

  return (
    <div
      role="alert"
      data-slot="alert"
      className={cn(
        "flex items-start gap-3",
        alertVariants({ variant }),
        className,
      )}
      {...props}
    >
      {resolvedIcon && <span className="mt-0.5 flex-shrink-0">{resolvedIcon}</span>}
      <div className="flex-1">{children}</div>
    </div>
  );
}

function AlertTitle({ className, ...props }: React.ComponentProps<"h5">) {
  return (
    <h5
      data-slot="alert-title"
      className={cn("mb-1 font-semibold leading-none tracking-tight", className)}
      {...props}
    />
  );
}

function AlertDescription({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-description"
      className={cn("text-sm leading-relaxed [&_p]:leading-relaxed", className)}
      {...props}
    />
  );
}

export { Alert, AlertTitle, AlertDescription };
