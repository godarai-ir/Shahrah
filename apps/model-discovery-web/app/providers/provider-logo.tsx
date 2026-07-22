"use client";

import { useEffect, useState } from "react";
import { Building2 } from "lucide-react";

export function ProviderLogo({
  src,
  alt,
}: {
  src: string | null;
  alt: string;
}) {
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    setIsLoaded(false);
    if (!src) return;

    const image = new Image();
    image.onload = () => setIsLoaded(true);
    image.src = src;

    return () => {
      image.onload = null;
    };
  }, [src]);

  return (
    <span className="flex size-12 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-border bg-background text-muted-foreground">
      {src && isLoaded ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src}
          alt={alt}
          className="size-full object-contain p-2"
        />
      ) : (
        <Building2 data-provider-logo-fallback className="size-5" aria-hidden="true" />
      )}
    </span>
  );
}
