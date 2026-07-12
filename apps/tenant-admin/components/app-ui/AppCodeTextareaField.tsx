import * as React from "react";

import { AppTextareaField, type AppTextareaFieldProps } from "@/components/app-ui/AppTextField";
import { cn } from "@/lib/utils";

export type AppCodeTextareaFieldProps = AppTextareaFieldProps;

export function AppCodeTextareaField({ className, dir = "ltr", spellCheck = false, ...props }: AppCodeTextareaFieldProps) {
  return (
    <AppTextareaField
      dir={dir}
      spellCheck={spellCheck}
      className={cn("min-h-40 text-left font-mono text-xs leading-6", className)}
      {...props}
    />
  );
}
