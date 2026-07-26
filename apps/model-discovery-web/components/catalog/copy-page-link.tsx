"use client";

import { useState } from "react";
import { Check, Link2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { copyTextToClipboard } from "@/lib/clipboard";
import { cn } from "@/lib/utils";

export function CopyPageLink({
  label,
  copiedLabel,
  failedLabel,
  className,
}: {
  label: string;
  copiedLabel: string;
  failedLabel: string;
  className?: string;
}) {
  const [status, setStatus] = useState<"idle" | "copied" | "failed">("idle");

  async function copyCurrentPageLink() {
    try {
      await copyTextToClipboard(window.location.href);
      setStatus("copied");
    } catch {
      setStatus("failed");
    }

    window.setTimeout(() => setStatus("idle"), 1800);
  }

  const accessibleLabel = status === "copied" ? copiedLabel : label;

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="icon-sm"
        className={cn("shrink-0", className)}
        aria-label={accessibleLabel}
        title={accessibleLabel}
        onClick={copyCurrentPageLink}
      >
        {status === "copied" ? (
          <Check data-icon="inline-start" aria-hidden="true" />
        ) : (
          <Link2 data-icon="inline-start" aria-hidden="true" />
        )}
      </Button>
      <span className="sr-only" aria-live="polite" aria-atomic="true">
        {status === "copied"
          ? copiedLabel
          : status === "failed"
            ? failedLabel
            : ""}
      </span>
    </>
  );
}
