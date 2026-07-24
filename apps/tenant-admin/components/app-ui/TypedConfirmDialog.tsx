"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

export type TypedConfirmDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  expectedValue: string;
  inputLabel: React.ReactNode;
  cancelLabel: React.ReactNode;
  closeLabel?: string;
  confirmLabel: React.ReactNode;
  pendingLabel: React.ReactNode;
  onConfirm: () => void | Promise<void>;
  onConfirmError?: (error: unknown) => void;
  warning?: React.ReactNode;
  disabled?: boolean;
  inputDir?: "rtl" | "ltr" | "auto";
};

export function TypedConfirmDialog(props: TypedConfirmDialogProps) {
  return <TypedConfirmDialogState key={`${props.open ? "open" : "closed"}:${props.expectedValue}`} {...props} />;
}

function TypedConfirmDialogState({
  open,
  onOpenChange,
  title,
  description,
  expectedValue,
  inputLabel,
  cancelLabel,
  closeLabel,
  confirmLabel,
  pendingLabel,
  onConfirm,
  onConfirmError,
  warning,
  disabled,
  inputDir = "ltr",
}: TypedConfirmDialogProps) {
  const inputID = React.useId();
  const [typedValue, setTypedValue] = React.useState("");
  const [pending, setPending] = React.useState(false);
  const matches = typedValue.trim() === expectedValue;

  function changeOpen(nextOpen: boolean) {
    if (pending) return;
    if (!nextOpen) setTypedValue("");
    onOpenChange(nextOpen);
  }

  async function confirm(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!matches || disabled || pending) return;

    setPending(true);
    try {
      await onConfirm();
      setTypedValue("");
      onOpenChange(false);
    } catch (error) {
      onConfirmError?.(error);
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogContent aria-busy={pending} closeLabel={closeLabel ?? (typeof cancelLabel === "string" ? cancelLabel : undefined)} showCloseButton={!pending}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        {warning}
        <form className="flex flex-col gap-4" onSubmit={confirm}>
          <FieldGroup>
            <Field data-disabled={pending || disabled || undefined}>
              <FieldLabel htmlFor={inputID}>{inputLabel}</FieldLabel>
              <Input
                id={inputID}
                value={typedValue}
                onChange={(event) => setTypedValue(event.target.value)}
                placeholder={expectedValue}
                dir={inputDir}
                autoComplete="off"
                autoFocus
                disabled={pending || disabled}
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => changeOpen(false)} disabled={pending}>
              {cancelLabel}
            </Button>
            <Button type="submit" variant="destructive" disabled={!matches || disabled || pending} aria-busy={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : null}
              {pending ? pendingLabel : confirmLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
