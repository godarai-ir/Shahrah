import Link from "next/link";
import { Boxes, Cable } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { cn } from "@/lib/utils";

export async function CatalogHero({
  active,
  title,
  summary,
  compact = false,
}: {
  active: "models" | "providers";
  title: string;
  summary: React.ReactNode;
  compact?: boolean;
}) {
  const t = await getTranslations("CatalogShell");

  return (
    <header className="catalog-hero relative isolate overflow-hidden border-y border-catalog-hero-border bg-gradient-to-bl from-primary/40 via-catalog-hero to-catalog-hero text-catalog-hero-foreground">
      <div
        className="pointer-events-none absolute inset-0 bg-gradient-to-tr from-transparent via-primary/10 to-primary/25"
        aria-hidden="true"
      />

      <div
        className={cn(
          "relative z-10 mx-auto flex w-full max-w-7xl flex-col px-4 sm:px-6 lg:px-8",
          compact
            ? "gap-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:py-6"
            : "items-center pb-14 pt-12 text-center sm:pb-16 sm:pt-16",
        )}
      >
        <div
          className={cn(
            "flex min-w-0",
            compact ? "items-center gap-3 text-start" : "flex-col items-center",
          )}
        >
          <div
            className={cn(
              "flex shrink-0 items-center justify-center rounded-xl border border-catalog-hero-border bg-background/10 shadow-sm",
              compact ? "size-10" : "mb-5 size-11",
            )}
          >
            {active === "models" ? (
              <Boxes className="size-5" aria-hidden="true" />
            ) : (
              <Cable className="size-5" aria-hidden="true" />
            )}
          </div>
          <div className="min-w-0">
            <h1
              className={cn(
                "font-bold tracking-tight",
                compact ? "text-xl sm:text-2xl" : "text-3xl sm:text-4xl",
              )}
            >
              {title}
            </h1>
            <p
              className={cn(
                "text-catalog-hero-muted",
                compact
                  ? "mt-1 text-sm leading-6"
                  : "mt-3 max-w-3xl text-sm leading-7 sm:text-base",
              )}
            >
              {summary}
            </p>
          </div>
        </div>

        <nav
          aria-label={t("tabsLabel")}
          className={cn(
            "inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-xl border border-catalog-hero-border bg-background/10 p-1",
            compact ? "self-start sm:shrink-0 sm:self-auto" : "mt-7",
          )}
        >
          <CatalogTabLink href="/models" active={active === "models"}>
            {t("modelsTab")}
          </CatalogTabLink>
          <CatalogTabLink href="/providers" active={active === "providers"}>
            {t("providersTab")}
          </CatalogTabLink>
        </nav>
      </div>
    </header>
  );
}

function CatalogTabLink({
  href,
  active,
  children,
}: {
  href: string;
  active: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={cn(
        "shrink-0 rounded-lg px-4 py-2 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active
          ? "bg-background text-foreground shadow-sm"
          : "text-catalog-hero-muted hover:bg-background/10 hover:text-catalog-hero-foreground",
      )}
    >
      {children}
    </Link>
  );
}
