"use client";

import * as React from "react";

import { Field, FieldContent, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";

export type AppSwitchFieldProps = Omit<React.ComponentProps<typeof Switch>, "id"> & {
  id?: string;
  label: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  surface?: boolean;
  fieldClassName?: string;
};

export function AppSwitchField({ id, label, description, error, surface = true, fieldClassName, disabled, ...props }: AppSwitchFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [descriptionID, errorID].filter(Boolean).join(" ") || undefined;

  return (
    <Field
      orientation="horizontal"
      data-invalid={Boolean(error) || undefined}
      data-disabled={disabled || undefined}
      className={cn(surface && "rounded-xl bg-muted/30 px-4 py-3", fieldClassName)}
    >
      <FieldContent>
        <FieldLabel htmlFor={controlID}>{label}</FieldLabel>
        {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
        {error ? <FieldError id={errorID}>{error}</FieldError> : null}
      </FieldContent>
      <Switch id={controlID} disabled={disabled} aria-invalid={Boolean(error) || undefined} aria-describedby={describedBy} {...props} />
    </Field>
  );
}
