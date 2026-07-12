"use client";

import * as React from "react";

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

export type FormDialogSize = "sm" | "md" | "lg" | "xl";
export type FormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  footer?: React.ReactNode;
  size?: FormDialogSize;
  pending?: boolean;
  contentClassName?: string;
  bodyClassName?: string;
};

const sizeClasses: Record<FormDialogSize, string> = { sm: "sm:max-w-md", md: "sm:max-w-2xl", lg: "sm:max-w-4xl", xl: "sm:max-w-6xl" };

export function FormDialog({ open, onOpenChange, title, description, children, footer, size = "sm", pending = false, contentClassName, bodyClassName }: FormDialogProps) {
  function changeOpen(nextOpen: boolean) {
    if (pending && !nextOpen) return;
    onOpenChange(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogContent className={cn("max-h-[calc(100dvh-2rem)] overflow-y-auto", sizeClasses[size], contentClassName)} showCloseButton={!pending} onEscapeKeyDown={(event) => pending && event.preventDefault()} onPointerDownOutside={(event) => pending && event.preventDefault()} {...(!description ? { "aria-describedby": undefined } : {})}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <div className={cn("flex flex-col gap-4", bodyClassName)}>{children}</div>
        {footer ? <DialogFooter>{footer}</DialogFooter> : null}
      </DialogContent>
    </Dialog>
  );
}
