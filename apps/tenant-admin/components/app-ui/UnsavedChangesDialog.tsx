import * as React from "react";

import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";

export type UnsavedChangesDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  stayLabel: React.ReactNode;
  discardLabel: React.ReactNode;
  discardingLabel?: React.ReactNode;
  onDiscard: () => void | Promise<void>;
  onDiscardError?: (error: unknown) => void;
};

export function UnsavedChangesDialog({ open, onOpenChange, title, description, stayLabel, discardLabel, discardingLabel, onDiscard, onDiscardError }: UnsavedChangesDialogProps) {
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      description={description}
      cancelLabel={stayLabel}
      confirmLabel={discardLabel}
      pendingLabel={discardingLabel ?? discardLabel}
      onConfirm={onDiscard}
      onConfirmError={onDiscardError}
      destructive
    />
  );
}
