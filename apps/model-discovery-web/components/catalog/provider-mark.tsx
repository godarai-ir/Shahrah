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
  size?: "xs" | "sm" | "default" | "lg";
}) {
  const src = resolveAssetUrl(logoUrl);

  return (
    <span
      className={cn(
        "relative flex shrink-0 items-center justify-center overflow-hidden border border-border bg-catalog-hero-foreground text-catalog-hero",
        size === "xs" && "size-6 rounded-md",
        size === "sm" && "size-8 rounded-lg",
        size === "default" && "size-10 rounded-xl",
        size === "lg" && "size-18 rounded-xl",
      )}
    >
      <Building2
        className={cn(
          size === "xs" && "size-3",
          size === "sm" && "size-3.5",
          size === "default" && "size-4",
          size === "lg" && "size-7",
        )}
        aria-hidden="true"
      />
      {src ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src}
          alt=""
          className="absolute inset-0 size-full bg-catalog-hero-foreground object-contain p-1.5"
          onError={(event) => {
            event.currentTarget.hidden = true;
          }}
        />
      ) : null}
      <span className="sr-only">{name}</span>
    </span>
  );
}
