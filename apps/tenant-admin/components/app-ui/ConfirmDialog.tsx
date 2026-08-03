"use client";

import * as React from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";

type ConfirmDialogProps = {
  trigger?: React.ReactElement;
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  confirmLabel: React.ReactNode;
  cancelLabel: React.ReactNode;
  pendingLabel?: React.ReactNode;
  closeLabel?: string;
  onConfirm?: () => void | Promise<void>;
  onConfirmError?: (error: unknown) => void;
  destructive?: boolean;
  pending?: boolean;
  confirmDisabled?: boolean;
  closeOnConfirm?: boolean;
  feedback?: React.ReactNode;
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
  feedback,
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
    <AlertDialog open={isOpen} onOpenChange={handleOpenChange}>
      {trigger ? <AlertDialogTrigger asChild>{trigger}</AlertDialogTrigger> : null}
      <AlertDialogContent
        onEscapeKeyDown={(event) => isPending && event.preventDefault()}
      >
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        {feedback}
        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button type="button" variant="outline" disabled={isPending}>
              {cancelLabel}
            </Button>
          </AlertDialogCancel>
          <AlertDialogAction asChild>
            <Button
              type="button"
              variant={destructive ? "destructive" : "default"}
              onClick={(event) => {
                event.preventDefault();
                void handleConfirm();
              }}
              disabled={isPending || confirmDisabled}
              aria-busy={isPending || undefined}
            >
              {isPending ? pendingLabel ?? confirmLabel : confirmLabel}
            </Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
