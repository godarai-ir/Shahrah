"use client";

import { Building2 } from "lucide-react";

export function ProviderLogo({
  src,
  alt,
}: {
  src: string | null;
  alt: string;
}) {
  return (
    <span className="relative flex size-12 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-border bg-background text-muted-foreground">
      <Building2 data-provider-logo-fallback className="size-5" aria-hidden="true" />
      {src ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src}
          alt={alt}
          className="absolute inset-0 size-full bg-background object-contain p-2"
          onError={(event) => {
            event.currentTarget.hidden = true;
          }}
        />
      ) : null}
    </span>
  );
}
