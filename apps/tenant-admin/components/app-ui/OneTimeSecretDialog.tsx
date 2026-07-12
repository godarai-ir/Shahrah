"use client";

import * as React from "react";

import { CopyableValue, type CopyableValueLabels } from "@/components/app-ui/CopyableValue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";

export type OneTimeSecretDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  value: string | null;
  title: React.ReactNode;
  description: React.ReactNode;
  warning?: React.ReactNode;
  closeLabel: React.ReactNode;
  copyLabels: CopyableValueLabels;
  masked?: boolean;
  allowReveal?: boolean;
  onClear?: () => void;
  onCopied?: () => void;
  onCopyError?: (error: unknown) => void;
};

export function OneTimeSecretDialog({
  open,
  onOpenChange,
  value,
  title,
  description,
  warning,
  closeLabel,
  copyLabels,
  masked,
  allowReveal,
  onClear,
  onCopied,
  onCopyError,
}: OneTimeSecretDialogProps) {
  function handleOpenChange(nextOpen: boolean) {
    onOpenChange(nextOpen);
    if (!nextOpen) onClear?.();
  }

  return (
    <Dialog open={open && Boolean(value)} onOpenChange={handleOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-3">
          {warning ? (
            <Alert>
              <AlertDescription>{warning}</AlertDescription>
            </Alert>
          ) : null}
          <CopyableValue
            value={value ?? ""}
            labels={copyLabels}
            masked={masked}
            allowReveal={allowReveal}
            onCopied={onCopied}
            onCopyError={onCopyError}
          />
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
            {closeLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
