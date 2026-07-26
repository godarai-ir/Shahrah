import Image from "next/image";

export type GodarAILogoVariant = "auto" | "light" | "dark" | "mono-navy" | "mono-white";

export type GodarAILogoProps = {
  alt?: string;
  className?: string;
  decorative?: boolean;
  priority?: boolean;
  sizes?: string;
  variant?: GodarAILogoVariant;
};

const sources: Record<Exclude<GodarAILogoVariant, "auto">, string> = {
  light: "/catalog/brand/svg/godarai-mark-light.svg",
  dark: "/catalog/brand/svg/godarai-mark-dark.svg",
  "mono-navy": "/catalog/brand/svg/godarai-mark-mono-navy.svg",
  "mono-white": "/catalog/brand/svg/godarai-mark-mono-white.svg",
};

export function GodarAILogo({
  alt = "GodarAI",
  className = "h-8 w-auto max-w-full object-contain",
  decorative = false,
  priority = false,
  sizes,
  variant = "auto",
}: GodarAILogoProps) {
  const accessibleAlt = decorative ? "" : alt;

  if (variant !== "auto") {
    return (
      <Image
        src={sources[variant]}
        alt={accessibleAlt}
        aria-hidden={decorative || undefined}
        width={336}
        height={247}
        className={className}
        priority={priority}
        sizes={sizes}
        unoptimized
      />
    );
  }

  return (
    <>
      <Image
        src={sources.light}
        alt={accessibleAlt}
        aria-hidden={decorative || undefined}
        width={336}
        height={247}
        className={`${className} dark:hidden`}
        priority={priority}
        sizes={sizes}
        unoptimized
      />
      <Image
        src={sources.dark}
        alt={accessibleAlt}
        aria-hidden={decorative || undefined}
        width={336}
        height={247}
        className={`${className} hidden dark:block`}
        priority={priority}
        sizes={sizes}
        unoptimized
      />
    </>
  );
}
