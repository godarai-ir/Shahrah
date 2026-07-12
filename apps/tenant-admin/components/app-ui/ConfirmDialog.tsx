"use client";

import * as React from "react";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

type ConfirmDialogProps = {
  trigger?: React.ReactElement;
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: React.ReactNode;
  description?: React.ReactNode;
  confirmLabel: React.ReactNode;
  cancelLabel: React.ReactNode;
  pendingLabel?: React.ReactNode;
  onConfirm?: () => void | Promise<void>;
  onConfirmError?: (error: unknown) => void;
  destructive?: boolean;
  pending?: boolean;
  confirmDisabled?: boolean;
  closeOnConfirm?: boolean;
};

export function ConfirmDialog({
  trigger,
  open,
  defaultOpen = false,
  onOpenChange,
  title,
  description,
  confirmLabel,
  cancelLabel,
  pendingLabel,
  onConfirm,
  onConfirmError,
  destructive = true,
  pending = false,
  confirmDisabled = false,
  closeOnConfirm = true,
}: ConfirmDialogProps) {
  const [internalOpen, setInternalOpen] = React.useState(defaultOpen);
  const [internalPending, setInternalPending] = React.useState(false);
  const isControlled = open !== undefined;
  const isOpen = isControlled ? open : internalOpen;
  const isPending = pending || internalPending;

  function handleOpenChange(nextOpen: boolean) {
    if (isPending && !nextOpen) return;
    if (!isControlled) setInternalOpen(nextOpen);
    onOpenChange?.(nextOpen);
  }

  async function handleConfirm() {
    if (isPending || confirmDisabled) return;
    setInternalPending(true);
    try {
      await onConfirm?.();
      if (closeOnConfirm) handleOpenChange(false);
    } catch (error) {
      onConfirmError?.(error);
    } finally {
      setInternalPending(false);
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      {trigger ? <DialogTrigger asChild>{trigger}</DialogTrigger> : null}
      <DialogContent
        showCloseButton={!isPending}
        onEscapeKeyDown={(event) => isPending && event.preventDefault()}
        onPointerDownOutside={(event) => isPending && event.preventDefault()}
        {...(!description ? { "aria-describedby": undefined } : {})}
      >
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => handleOpenChange(false)} disabled={isPending}>
            {cancelLabel}
          </Button>
          <Button
            type="button"
            variant={destructive ? "destructive" : "default"}
            onClick={() => void handleConfirm()}
            disabled={isPending || confirmDisabled}
            aria-busy={isPending || undefined}
          >
            {isPending ? pendingLabel ?? confirmLabel : confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
