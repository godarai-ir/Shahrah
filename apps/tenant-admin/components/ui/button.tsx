import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-[background-color,border-color,color,box-shadow,transform] duration-150 ease-out disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring press-active",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-sm hover:bg-primary-hover hover:shadow-primary active:bg-primary-active",
        destructive:
          "bg-destructive text-destructive-foreground shadow-xs hover:bg-destructive-hover",
        outline:
          "border border-input-border bg-card text-foreground shadow-xs hover:border-primary-subtle-border hover:bg-primary-soft hover:text-primary-soft-foreground",
        secondary:
          "bg-secondary text-secondary-foreground shadow-xs hover:bg-secondary-hover",
        ghost: "text-foreground hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
        success:
          "bg-success text-success-foreground shadow-xs hover:brightness-95",
        warning:
          "bg-warning text-warning-foreground shadow-xs hover:brightness-95",
      },
      size: {
        default: "h-10 px-4 py-2 min-w-[44px]",
        sm: "h-8 rounded-md px-3 text-xs min-w-[44px]",
        lg: "h-11 px-6 text-base min-w-[48px]",
        icon: "size-10",
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
