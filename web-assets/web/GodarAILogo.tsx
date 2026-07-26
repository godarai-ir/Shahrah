import Image from "next/image";

export function GodarAILogo({ className = "h-8 w-auto" }: { className?: string }) {
  return (
    <picture>
      <source
        srcSet="/brand/svg/godarai-mark-dark.svg"
        media="(prefers-color-scheme: dark)"
      />
      <Image
        src="/brand/svg/godarai-mark-light.svg"
        alt="GodarAI"
        width={336}
        height={247}
        className={className}
        priority
      />
    </picture>
  );
}
