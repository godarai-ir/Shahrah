import * as React from "react";
import { RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

export type RefreshButtonProps = Omit<React.ComponentProps<typeof Button>, "children" | "size" | "aria-label"> & {
  label: string;
  loadingLabel?: string;
  isLoading?: boolean;
};

export function RefreshButton({
  label,
  loadingLabel,
  isLoading = false,
  disabled,
  title,
  ...props
}: RefreshButtonProps) {
  const accessibleLabel = isLoading ? loadingLabel ?? label : label;

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      aria-label={accessibleLabel}
      aria-busy={isLoading || undefined}
      title={title ?? accessibleLabel}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? <Spinner /> : <RefreshCw aria-hidden="true" />}
    </Button>
  );
}
