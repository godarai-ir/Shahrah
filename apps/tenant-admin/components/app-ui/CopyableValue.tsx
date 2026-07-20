"use client";

import * as React from "react";
import { Check, Copy, Eye, EyeOff } from "lucide-react";

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@/components/ui/input-group";
import { cn } from "@/lib/utils";

export type CopyableValueLabels = {
  valueAriaLabel: string;
  copy: string;
  copying: string;
  copied: string;
  reveal?: string;
  hide?: string;
};

export type CopyableValueProps = {
  value: string;
  labels: CopyableValueLabels;
  masked?: boolean;
  allowReveal?: boolean;
  disabled?: boolean;
  onCopied?: () => void;
  onCopyError?: (error: unknown) => void;
  className?: string;
};

export function CopyableValue({
  value,
  masked = false,
  ...props
}: CopyableValueProps) {
  return <CopyableValueControl key={`${masked ? "masked" : "visible"}:${value}`} value={value} masked={masked} {...props} />;
}

function CopyableValueControl({
  value,
  labels,
  masked = false,
  allowReveal = false,
  disabled,
  onCopied,
  onCopyError,
  className,
}: CopyableValueProps) {
  const [isCopying, setIsCopying] = React.useState(false);
  const [isCopied, setIsCopied] = React.useState(false);
  const [isVisible, setIsVisible] = React.useState(!masked);
  const resetTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => () => {
    if (resetTimer.current) clearTimeout(resetTimer.current);
  }, []);

  async function handleCopy() {
    if (!value || isCopying || disabled) return;
    if (resetTimer.current) clearTimeout(resetTimer.current);
    setIsCopying(true);
    try {
      await copyTextToClipboard(value);
      setIsCopied(true);
      onCopied?.();
      resetTimer.current = setTimeout(() => setIsCopied(false), 1600);
    } catch (error) {
      setIsCopied(false);
      onCopyError?.(error);
    } finally {
      setIsCopying(false);
    }
  }

  const copyLabel = isCopying ? labels.copying : isCopied ? labels.copied : labels.copy;

  return (
    <div className={cn("flex min-w-0 flex-col gap-2", className)}>
      <InputGroup data-disabled={disabled || undefined} className="bg-background">
        <InputGroupInput
          value={value}
          readOnly
          disabled={disabled}
          type={isVisible ? "text" : "password"}
          dir="ltr"
          aria-label={labels.valueAriaLabel}
          className="select-all truncate font-mono text-xs tabular-nums text-start"
        />
        <InputGroupAddon align="inline-end" className="gap-1 pe-1">
          {masked && allowReveal && labels.reveal && labels.hide ? (
            <InputGroupButton
              size="icon-sm"
              aria-label={isVisible ? labels.hide : labels.reveal}
              title={isVisible ? labels.hide : labels.reveal}
              disabled={disabled}
              onClick={() => setIsVisible((visible) => !visible)}
            >
              {isVisible ? <EyeOff /> : <Eye />}
            </InputGroupButton>
          ) : null}
          <InputGroupButton
            size="icon-sm"
            aria-label={copyLabel}
            title={copyLabel}
            aria-busy={isCopying}
            disabled={disabled || isCopying || !value}
            onClick={() => void handleCopy()}
          >
            {isCopied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
          </InputGroupButton>
        </InputGroupAddon>
      </InputGroup>
      <span className="sr-only" role="status" aria-live="polite">
        {isCopied ? labels.copied : ""}
      </span>
    </div>
  );
}

export async function copyTextToClipboard(value: string) {
  if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }

  if (typeof document === "undefined") {
    throw new Error("Clipboard is unavailable.");
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  try {
    textarea.focus();
    textarea.select();
    if (!document.execCommand("copy")) {
      throw new Error("Clipboard copy failed.");
    }
  } finally {
    document.body.removeChild(textarea);
  }
}
