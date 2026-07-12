import * as React from "react";

import { formatDateTime, parseDateTime, type DateTimeInput } from "@/lib/date-time";
import { cn } from "@/lib/utils";

export type DateTimeValueProps = Omit<React.ComponentProps<"time">, "dateTime" | "children"> & {
  value: DateTimeInput;
  locale?: string;
  fallback?: string;
  format?: Intl.DateTimeFormatOptions;
};

export function DateTimeValue({ value, locale, fallback, format, className, ...props }: DateTimeValueProps) {
  const date = parseDateTime(value);
  return (
    <time dateTime={date?.toISOString()} className={cn("tabular-nums", className)} {...props}>
      {formatDateTime(date, { locale, fallback, format })}
    </time>
  );
}
