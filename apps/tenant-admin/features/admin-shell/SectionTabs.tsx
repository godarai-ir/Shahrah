"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

import type { ResolvedNavTabs } from "@/features/admin-shell/navigation";
import { cn } from "@/lib/utils";

/** Route-backed tabs for a sidebar section. Each tab is a real link so legacy routes keep working as deep links. */
export function SectionTabs({ tabs }: { tabs: ResolvedNavTabs["tabs"] }) {
  const t = useTranslations("shell");

  if (tabs.length < 2) return null;

  return (
    <nav aria-label={t("nav.sectionTabs")} className="custom-scrollbar -mx-1 mb-4 overflow-x-auto px-1 md:mb-6">
      <ul className="flex w-max min-w-full items-center gap-1 border-b border-border">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          return (
            <li key={tab.key}>
              <Link
                href={tab.href}
                aria-current={tab.isCurrent ? "page" : undefined}
                className={cn(
                  "relative inline-flex h-10 items-center gap-2 whitespace-nowrap rounded-t-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-ring",
                  "after:absolute after:inset-x-2 after:-bottom-px after:h-0.5 after:rounded-full after:bg-primary after:opacity-0",
                  tab.isCurrent && "text-foreground after:opacity-100",
                )}
              >
                <Icon className="size-4 shrink-0" aria-hidden="true" />
                <span>{tab.title}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
