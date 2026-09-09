import { useTranslations } from "next-intl";
import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export const IDENTIFIER_PATTERN = "[A-Za-z0-9_-]+";

export function isValidIdentifier(value: string): boolean {
  return /^[A-Za-z0-9_-]+$/.test(value);
}

type IdentifierFieldProps = Omit<React.ComponentProps<typeof Input>, "dir" | "type"> & {
  label?: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  containerClassName?: string;
};

export function IdentifierField({ id, label, description, error, required, disabled, className, containerClassName, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid, ...props }: IdentifierFieldProps) {
  const t = useTranslations("common");
  const generatedID = React.useId();
  const inputID = id ?? generatedID;
  // `null` deliberately opts out of the hint; `undefined` takes the default.
  const resolvedDescription = description === undefined ? t("identifierField.description") : description;
  const descriptionID = resolvedDescription ? `${inputID}-description` : undefined;
  const errorID = error ? `${inputID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, errorID].filter(Boolean).join(" ") || undefined;
  const invalid = Boolean(error) || ariaInvalid === true || ariaInvalid === "true";

  return (
    <Field className={containerClassName} data-invalid={invalid || undefined} data-disabled={disabled || undefined}>
      {label ? <FieldLabel htmlFor={inputID}>{label}</FieldLabel> : null}
      <Input {...props} id={inputID} type="text" dir="ltr" pattern={props.pattern ?? IDENTIFIER_PATTERN} required={required} disabled={disabled} aria-describedby={describedBy} aria-invalid={invalid || undefined} className={cn("font-mono text-start", className)} />
      {resolvedDescription ? <FieldDescription id={descriptionID}>{resolvedDescription}</FieldDescription> : null}
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}
