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

  async function handleCopy() {
    if (!value || isCopying || disabled) return;
    setIsCopying(true);
    try {
      await copyTextToClipboard(value);
      setIsCopied(true);
      onCopied?.();
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
      <InputGroup data-disabled={disabled || undefined}>
        <InputGroupInput
          value={value}
          readOnly
          disabled={disabled}
          type={isVisible ? "text" : "password"}
          dir="ltr"
          aria-label={labels.valueAriaLabel}
          className="font-mono text-xs"
        />
        <InputGroupAddon align="inline-end">
          {masked && allowReveal && labels.reveal && labels.hide ? (
            <InputGroupButton
              size="icon-xs"
              aria-label={isVisible ? labels.hide : labels.reveal}
              title={isVisible ? labels.hide : labels.reveal}
              disabled={disabled}
              onClick={() => setIsVisible((visible) => !visible)}
            >
              {isVisible ? <EyeOff /> : <Eye />}
            </InputGroupButton>
          ) : null}
          <InputGroupButton
            size="sm"
            aria-label={copyLabel}
            aria-busy={isCopying}
            disabled={disabled || isCopying || !value}
            onClick={() => void handleCopy()}
          >
            {isCopied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
            {copyLabel}
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
