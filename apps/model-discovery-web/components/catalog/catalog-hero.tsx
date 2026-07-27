import Link from "next/link";
import { Boxes, Cable } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export async function CatalogHero({
  active,
  title,
  summary,
  modelCount,
  providerCount,
  endpointCount,
  activeProviderSection = "providers",
  compact = false,
}: {
  active: "models" | "providers";
  title: string;
  summary: React.ReactNode;
  modelCount?: number;
  providerCount?: number;
  endpointCount?: number;
  activeProviderSection?: "providers" | "endpoints";
  compact?: boolean;
}) {
  const t = await getTranslations("CatalogShell");
  const formatter = new Intl.NumberFormat("fa-IR");

  return (
    <header className="catalog-hero relative isolate overflow-hidden border-y border-border bg-card text-card-foreground">
      <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
        <div className="catalog-hero-mesh absolute inset-0" />
        <div className="catalog-hero-orbit catalog-hero-orbit-start absolute -start-16 top-1/4 size-56 rounded-full border border-primary/15 sm:size-72" />
        <div className="catalog-hero-orbit catalog-hero-orbit-end absolute -end-10 top-0 size-44 rounded-full border border-primary/20 sm:size-64" />
        <span className="catalog-hero-node absolute start-1/4 top-1/4 size-2 rounded-full bg-primary/60 shadow-sm" />
        <span className="catalog-hero-node catalog-hero-node-delayed absolute end-1/4 bottom-1/4 size-3 rounded-full border border-primary/30 bg-card shadow-sm" />
      </div>

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
              "flex shrink-0 items-center justify-center rounded-xl border border-primary/15 bg-primary-soft text-primary-soft-foreground shadow-sm",
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
                "text-muted-foreground",
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
            "inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-xl border border-border bg-background/80 p-1 shadow-sm backdrop-blur",
            compact ? "self-start sm:shrink-0 sm:self-auto" : "mt-7",
          )}
        >
          <CatalogTabLink href="/models" active={active === "models"}>
            <span>{t("modelsTab")}</span>
            {modelCount !== undefined ? <Badge>{formatter.format(modelCount)}</Badge> : null}
          </CatalogTabLink>
          <CatalogTabLink
            href="/providers?section=providers"
            active={active === "providers" && activeProviderSection === "providers"}
          >
            <span>{t("providersTab")}</span>
            {providerCount !== undefined ? <Badge>{formatter.format(providerCount)}</Badge> : null}
          </CatalogTabLink>
          {endpointCount !== undefined ? (
            <CatalogTabLink
              href="/providers?section=endpoints"
              active={active === "providers" && activeProviderSection === "endpoints"}
            >
              <span>{t("operationsTab")}</span>
              <Badge>{formatter.format(endpointCount)}</Badge>
            </CatalogTabLink>
          ) : null}
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
        "inline-flex shrink-0 items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active
          ? "bg-primary text-primary-foreground shadow-sm"
          : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
      )}
    >
      {children}
    </Link>
  );
}
