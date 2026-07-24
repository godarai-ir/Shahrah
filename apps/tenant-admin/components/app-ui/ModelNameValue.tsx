import { useTranslations } from "next-intl";
import * as React from "react";

import { CopyAction } from "@/components/app-ui/CopyAction";
import { cn } from "@/lib/utils";

export type ModelNameValueProps = Omit<React.ComponentProps<"div">, "children" | "dir"> & {
  value: string;
  copyValue?: string;
  fallback?: React.ReactNode;
  onCopyError?: (error: unknown) => void;
};

export function ModelNameValue({ value, copyValue = value, fallback = "-", onCopyError, className, ...props }: ModelNameValueProps) {
  const t = useTranslations("common");

  if (!value) return <>{fallback}</>;

  return (
    <div dir="ltr" className={cn("flex min-w-0 max-w-full items-center justify-start gap-1 text-start", className)} {...props}>
      <code className="block min-w-0 truncate whitespace-nowrap font-mono text-xs" title={value}>{value}</code>
      <CopyAction
        value={copyValue}
        copyLabel={t("modelName.copy")}
        copyingLabel={t("modelName.copying")}
        copiedLabel={t("modelName.copied")}
        onCopyError={onCopyError}
        className="shrink-0"
      />
    </div>
  );
}
