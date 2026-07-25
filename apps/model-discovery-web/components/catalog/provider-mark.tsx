"use client";

import { Building2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { resolveAssetUrl } from "@/lib/api";

export function ProviderMark({
  name,
  logoUrl,
  size = "default",
}: {
  name: string;
  logoUrl?: string | null;
  size?: "sm" | "default" | "lg";
}) {
  const src = resolveAssetUrl(logoUrl);

  return (
    <span
      className={cn(
        "relative flex shrink-0 items-center justify-center overflow-hidden border border-border bg-background text-muted-foreground",
        size === "sm" && "size-8 rounded-lg",
        size === "default" && "size-10 rounded-xl",
        size === "lg" && "size-12 rounded-xl",
      )}
    >
      <Building2
        className={cn(
          size === "sm" && "size-3.5",
          size === "default" && "size-4",
          size === "lg" && "size-5",
        )}
        aria-hidden="true"
      />
      {src ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src}
          alt=""
          className="absolute inset-0 size-full bg-background object-contain p-1.5"
          onError={(event) => {
            event.currentTarget.hidden = true;
          }}
        />
      ) : null}
      <span className="sr-only">{name}</span>
    </span>
  );
}
