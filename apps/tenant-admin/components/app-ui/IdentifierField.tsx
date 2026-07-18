import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export const IDENTIFIER_PATTERN = "[A-Za-z0-9_-]+";
export const DEFAULT_IDENTIFIER_DESCRIPTION = "شناسه را بدون فاصله و با حروف انگلیسی، عدد، خط تیره یا زیرخط وارد کنید.";

export function isValidIdentifier(value: string): boolean {
  return /^[A-Za-z0-9_-]+$/.test(value);
}

type IdentifierFieldProps = Omit<React.ComponentProps<typeof Input>, "dir" | "type"> & {
  label?: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  containerClassName?: string;
};

export function IdentifierField({ id, label, description = DEFAULT_IDENTIFIER_DESCRIPTION, error, required, disabled, className, containerClassName, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid, ...props }: IdentifierFieldProps) {
  const generatedID = React.useId();
  const inputID = id ?? generatedID;
  const descriptionID = description ? `${inputID}-description` : undefined;
  const errorID = error ? `${inputID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, errorID].filter(Boolean).join(" ") || undefined;
  const invalid = Boolean(error) || ariaInvalid === true || ariaInvalid === "true";

  return (
    <Field className={containerClassName} data-invalid={invalid || undefined} data-disabled={disabled || undefined}>
      {label ? <FieldLabel htmlFor={inputID}>{label}</FieldLabel> : null}
      <Input {...props} id={inputID} type="text" dir="ltr" pattern={props.pattern ?? IDENTIFIER_PATTERN} required={required} disabled={disabled} aria-describedby={describedBy} aria-invalid={invalid || undefined} className={cn("font-mono text-start", className)} />
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}
