"use client";

import * as React from "react";

import { AppTextField, type AppTextFieldProps } from "@/components/app-ui/AppTextField";
import { cn } from "@/lib/utils";

export const EMAIL_PATTERN = "[^\\s@]+@[^\\s@]+\\.[^\\s@]+";

export function isValidEmail(value: string): boolean {
  return new RegExp(`^${EMAIL_PATTERN}$`).test(value.trim());
}

export type AppEmailFieldProps = Omit<AppTextFieldProps, "dir" | "inputMode" | "pattern" | "type"> & {
  invalidMessage: React.ReactNode;
  requiredMessage?: React.ReactNode;
};

export function AppEmailField({
  autoComplete = "email",
  autoCapitalize = "none",
  className,
  defaultValue,
  error,
  invalidMessage,
  onBlur,
  onChange,
  onInvalid,
  required,
  requiredMessage,
  spellCheck = false,
  value,
  ...props
}: AppEmailFieldProps) {
  const [internalValue, setInternalValue] = React.useState(() => String(defaultValue ?? ""));
  const [showValidation, setShowValidation] = React.useState(false);
  const currentValue = value === undefined ? internalValue : String(value ?? "");
  const trimmedValue = currentValue.trim();
  const validationError = showValidation
    ? !trimmedValue && required
      ? requiredMessage ?? invalidMessage
      : trimmedValue && !isValidEmail(trimmedValue)
        ? invalidMessage
        : undefined
    : undefined;

  return (
    <AppTextField
      {...props}
      type="email"
      dir="ltr"
      inputMode="email"
      autoComplete={autoComplete}
      autoCapitalize={autoCapitalize}
      spellCheck={spellCheck}
      pattern={EMAIL_PATTERN}
      required={required}
      value={value}
      defaultValue={defaultValue}
      error={error ?? validationError}
      className={cn("text-start", className)}
      onChange={(event) => {
        if (value === undefined) setInternalValue(event.currentTarget.value);
        onChange?.(event);
      }}
      onBlur={(event) => {
        setShowValidation(true);
        onBlur?.(event);
      }}
      onInvalid={(event) => {
        event.preventDefault();
        setShowValidation(true);
        onInvalid?.(event);
      }}
    />
  );
}
