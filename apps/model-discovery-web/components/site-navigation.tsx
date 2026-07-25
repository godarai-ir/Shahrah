"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Code2, ExternalLink } from "lucide-react";

const links = [
  { href: "https://godarai.ir/docs", key: "useApi", external: true },
  { href: "/compare", key: "compare", external: false },
  { href: "https://github.com/delpak/shahrah", key: "github", external: true },
] as const;

export function SiteNavigation() {
  const t = useTranslations("Navigation");

  return (
    <div className="scrollbar-none flex min-w-0 flex-1 items-center justify-end gap-1 overflow-x-auto">
      {links.map(({ href, key, external }) =>
        external ? (
          <a
            key={href}
            href={href}
            target="_blank"
            rel="noreferrer"
            className="hidden shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm font-medium text-catalog-hero-muted transition-colors hover:bg-background/10 hover:text-catalog-hero-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:inline-flex"
          >
            {key === "github" ? <Code2 aria-hidden="true" /> : null}
            {t(key)}
            {key !== "github" ? <ExternalLink className="size-3.5" aria-hidden="true" /> : null}
          </a>
        ) : (
          <Link
            key={href}
            href={href}
            className="hidden shrink-0 rounded-lg px-2.5 py-2 text-sm font-medium text-catalog-hero-muted transition-colors hover:bg-background/10 hover:text-catalog-hero-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring md:inline-flex"
          >
            {t(key)}
          </Link>
        ),
      )}
    </div>
  );
}
