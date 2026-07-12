import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

type SharedFieldProps = {
  label: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  fieldClassName?: string;
};

export type AppTextFieldProps = SharedFieldProps & React.ComponentProps<typeof Input>;
export type AppTextareaFieldProps = SharedFieldProps & React.ComponentProps<typeof Textarea>;

export function AppTextField({ label, description, error, fieldClassName, id, disabled, className, ...props }: AppTextFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID}>{label}</FieldLabel>
      <Input id={controlID} disabled={disabled} aria-invalid={Boolean(error) || undefined} className={className} {...props} />
      {description ? <FieldDescription>{description}</FieldDescription> : null}
      {error ? <FieldError>{error}</FieldError> : null}
    </Field>
  );
}

export function AppTextareaField({ label, description, error, fieldClassName, id, disabled, className, ...props }: AppTextareaFieldProps) {
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID}>{label}</FieldLabel>
      <Textarea id={controlID} disabled={disabled} aria-invalid={Boolean(error) || undefined} className={className} {...props} />
      {description ? <FieldDescription>{description}</FieldDescription> : null}
      {error ? <FieldError>{error}</FieldError> : null}
    </Field>
  );
}
