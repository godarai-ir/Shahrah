"use client";

import * as React from "react";

import { AppTextField, type AppTextFieldProps } from "@/components/app-ui/AppTextField";
import { cn } from "@/lib/utils";

const LETTER_PATTERN = /\p{Letter}/u;
const RTL_SCRIPT_PATTERN = /[\p{Script=Adlam}\p{Script=Arabic}\p{Script=Hanifi_Rohingya}\p{Script=Hebrew}\p{Script=Mandaic}\p{Script=Nko}\p{Script=Samaritan}\p{Script=Syriac}\p{Script=Thaana}]/u;

export type DisplayNameDirection = "ltr" | "rtl";

export function resolveDisplayNameDirection(value: string): DisplayNameDirection | undefined {
  for (const character of value) {
    if (!LETTER_PATTERN.test(character)) continue;
    return RTL_SCRIPT_PATTERN.test(character) ? "rtl" : "ltr";
  }
  return undefined;
}

export type DisplayNameFieldProps = Omit<AppTextFieldProps, "dir" | "type">;

export function DisplayNameField({
  autoCapitalize = "words",
  className,
  defaultValue,
  onChange,
  value,
  ...props
}: DisplayNameFieldProps) {
  const [internalValue, setInternalValue] = React.useState(() => String(defaultValue ?? ""));
  const currentValue = value === undefined ? internalValue : String(value ?? "");
  const direction = resolveDisplayNameDirection(currentValue) ?? "auto";

  return (
    <AppTextField
      {...props}
      type="text"
      dir={direction}
      autoCapitalize={autoCapitalize}
      value={value}
      defaultValue={defaultValue}
      className={cn("text-start", className)}
      onChange={(event) => {
        if (value === undefined) setInternalValue(event.currentTarget.value);
        onChange?.(event);
      }}
    />
  );
}
