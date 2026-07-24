"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

const links = [
  { href: "/models", key: "models" },
  { href: "/providers", key: "providers" },
  { href: "/compare", key: "compare" },
] as const;

export function SiteNavigation() {
  const pathname = usePathname();
  const t = useTranslations("Navigation");

  return (
    <div className="scrollbar-none flex min-w-0 flex-1 items-center gap-1 overflow-x-auto rounded-xl border border-border bg-muted/60 p-1 sm:flex-none">
      {links.map(({ href, key }) => {
        const active = pathname === href || pathname.startsWith(`${href}/`);

        return (
          <Link
            key={href}
            href={href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "shrink-0 rounded-lg px-2.5 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3 sm:text-sm",
              active && "bg-background text-foreground shadow-sm",
            )}
          >
            {t(key)}
          </Link>
        );
      })}
    </div>
  );
}
