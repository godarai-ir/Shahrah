import { useTranslations } from "next-intl";
import * as React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export type ExpirationDurationExample = {
  value: string;
  label: React.ReactNode;
};

export type ExpirationDurationFieldProps = Omit<React.ComponentProps<typeof Input>, "type" | "dir"> & {
  label?: React.ReactNode;
  description?: React.ReactNode;
  error?: React.ReactNode;
  examples?: ExpirationDurationExample[];
  examplesLabel?: React.ReactNode;
  fieldClassName?: string;
};

export function ExpirationDurationField({
  label,
  description,
  error,
  examples,
  examplesLabel,
  fieldClassName,
  id,
  disabled,
  className,
  "aria-describedby": ariaDescribedBy,
  ...props
}: ExpirationDurationFieldProps) {
  const t = useTranslations("common");
  const generatedID = React.useId();
  const controlID = id ?? generatedID;
  const descriptionID = description ? `${controlID}-description` : undefined;
  const examplesID = `${controlID}-examples`;
  const errorID = error ? `${controlID}-error` : undefined;
  const describedBy = [ariaDescribedBy, descriptionID, examplesID, errorID].filter(Boolean).join(" ") || undefined;
  const resolvedExamples = examples ?? [
    { value: "30m", label: t("expirationDuration.thirtyMinutes") },
    { value: "24h", label: t("expirationDuration.twentyFourHours") },
    { value: "7d", label: t("expirationDuration.sevenDays") },
    { value: "30d", label: t("expirationDuration.thirtyDays") },
  ];

  return (
    <Field data-invalid={Boolean(error) || undefined} data-disabled={disabled || undefined} className={fieldClassName}>
      <FieldLabel htmlFor={controlID}>{label ?? t("expirationDuration.label")}</FieldLabel>
      <Input
        {...props}
        id={controlID}
        type="text"
        dir="ltr"
        disabled={disabled}
        aria-invalid={Boolean(error) || undefined}
        aria-describedby={describedBy}
        autoComplete="off"
        spellCheck={false}
        className={cn("font-mono text-start", className)}
      />
      {description ? <FieldDescription id={descriptionID}>{description}</FieldDescription> : null}
      <FieldDescription id={examplesID}>
        {examplesLabel ?? t("expirationDuration.examplesLabel")}{" "}
        {resolvedExamples.map((example, index) => (
          <React.Fragment key={example.value}>
            {index > 0 ? t("expirationDuration.separator") : null}
            <code dir="ltr" className="font-mono text-foreground">{example.value}</code>{" "}
            <span>({example.label})</span>
          </React.Fragment>
        ))}
      </FieldDescription>
      {error ? <FieldError id={errorID}>{error}</FieldError> : null}
    </Field>
  );
}
