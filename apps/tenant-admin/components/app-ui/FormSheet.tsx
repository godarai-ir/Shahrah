"use client";

import * as React from "react";

import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

export type FormSheetSize = "sm" | "md" | "lg";
export type FormSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  footer?: React.ReactNode;
  size?: FormSheetSize;
  side?: "inline-start" | "inline-end";
  pending?: boolean;
  closeLabel?: string;
  contentClassName?: string;
  bodyClassName?: string;
};

const sizeClasses = { sm: "sm:max-w-md", md: "sm:max-w-xl", lg: "sm:max-w-2xl" } as const;

export function FormSheet({ open, onOpenChange, title, description, children, footer, size = "md", side = "inline-end", pending, closeLabel, contentClassName, bodyClassName }: FormSheetProps) {
  return (
    <Sheet open={open} onOpenChange={(nextOpen) => !pending && onOpenChange(nextOpen)}>
      <SheetContent side={side} closeLabel={closeLabel} aria-busy={pending} className={cn("flex w-full flex-col overflow-hidden p-0", sizeClasses[size], contentClassName)}>
        <SheetHeader className="shrink-0 border-b border-border px-6 py-5 text-start">
          <SheetTitle>{title}</SheetTitle>
          {description ? <SheetDescription>{description}</SheetDescription> : null}
        </SheetHeader>
        <div className={cn("min-h-0 flex-1 overflow-y-auto px-6 py-5", bodyClassName)}>{children}</div>
        {footer ? <SheetFooter className="shrink-0 border-t border-border px-6 py-4">{footer}</SheetFooter> : null}
      </SheetContent>
    </Sheet>
  );
}
