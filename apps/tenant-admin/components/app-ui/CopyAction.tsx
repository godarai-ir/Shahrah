"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";

import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { Button } from "@/components/ui/button";

export type CopyActionProps = Omit<React.ComponentProps<typeof Button>, "children" | "onClick"> & {
  value: string;
  copyLabel: string;
  copiedLabel: string;
  copyingLabel?: string;
  showLabel?: boolean;
  resetAfter?: number;
  onCopied?: () => void;
  onCopyError?: (error: unknown) => void;
};

export function CopyAction({ value, copyLabel, copiedLabel, copyingLabel = copyLabel, showLabel = false, resetAfter = 1600, onCopied, onCopyError, variant = "ghost", size, disabled, ...props }: CopyActionProps) {
  const [state, setState] = React.useState<"idle" | "copying" | "copied">("idle");
  const resetTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => () => {
    if (resetTimer.current) clearTimeout(resetTimer.current);
  }, []);

  async function copy() {
    if (!value || disabled || state === "copying") return;
    setState("copying");
    try {
      await copyTextToClipboard(value);
      setState("copied");
      onCopied?.();
      resetTimer.current = setTimeout(() => setState("idle"), resetAfter);
    } catch (error) {
      setState("idle");
      onCopyError?.(error);
    }
  }

  const label = state === "copying" ? copyingLabel : state === "copied" ? copiedLabel : copyLabel;
  return (
    <Button type="button" variant={variant} size={size ?? (showLabel ? "sm" : "icon-xs")} aria-label={label} title={label} aria-busy={state === "copying" || undefined} disabled={disabled || !value || state === "copying"} onClick={() => void copy()} {...props}>
      {state === "copied" ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
      {showLabel ? label : null}
      <span className="sr-only" role="status" aria-live="polite">{state === "copied" ? copiedLabel : ""}</span>
    </Button>
  );
}
