import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-all duration-200 ease-out disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring press-active",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary-hover shadow-sm hover:shadow-primary/30 active:bg-primary-active",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive-hover shadow-sm",
        outline:
          "border border-border bg-background hover:bg-accent hover:text-accent-foreground hover:border-accent transition-all",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary-hover shadow-xs",
        ghost: "hover:bg-accent hover:text-accent-foreground rounded-md",
        link: "text-primary underline-offset-4 hover:underline text-inherit",
        success:
          "bg-success text-success-foreground hover:opacity-90 shadow-sm",
        warning:
          "bg-warning text-warning-foreground hover:opacity-90 shadow-sm",
      },
      size: {
        default: "h-10 px-4 py-2 min-w-[44px]",
        sm: "h-8 rounded-md px-3 text-xs min-w-[44px]",
        lg: "h-12 rounded-xl px-6 text-base min-w-[48px]",
        icon: "h-10 w-10",
        "icon-xs": "size-6 rounded-md p-0 min-w-0",
        "icon-sm": "size-8 rounded-md p-0 min-w-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

function Button({
  className,
  variant,
  size,
  asChild = false,
  children,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  }) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      {...props}
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
    >
      {children}
    </Comp>
  );
}

export { Button, buttonVariants };
