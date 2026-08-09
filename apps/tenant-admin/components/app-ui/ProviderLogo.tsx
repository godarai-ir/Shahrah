"use client";

import Image from "next/image";
import { useState } from "react";

import { getProviderFallbackLogoPath, getProviderLogoSrc } from "@/lib/provider-branding";
export function ProviderLogo(props: { provider: string; displayName: string; logoURL?: string; size?: number; decorative?: boolean }) {
  const size = props.size ?? 18;
  const fallbackSrc = getProviderFallbackLogoPath(props.provider);
  const [hasError, setHasError] = useState(false);
  const src = hasError ? fallbackSrc : getProviderLogoSrc(props.provider, props.logoURL);

  return (
    <Image
      src={src}
      alt={props.decorative ? "" : props.displayName}
      aria-hidden={props.decorative || undefined}
      width={size}
      height={size}
      className="shrink-0 rounded-sm object-contain"
      unoptimized
      onError={() => setHasError(true)}
    />
  );
}


