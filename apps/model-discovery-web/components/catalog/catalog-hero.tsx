import Link from "next/link";
import Image from "next/image";
import { Boxes, Cable } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { cn } from "@/lib/utils";

export async function CatalogHero({
  active,
  title,
  summary,
}: {
  active: "models" | "providers";
  title: string;
  summary: React.ReactNode;
}) {
  const t = await getTranslations("CatalogShell");

  return (
    <header className="catalog-hero relative isolate overflow-hidden border-y border-catalog-hero-border bg-catalog-hero text-catalog-hero-foreground">
      <div className="catalog-hero-media absolute inset-0" aria-hidden="true">
        <Image
          src="/catalog/images/catalog-hero-ai-models.png"
          alt=""
          fill
          priority
          sizes="100vw"
          className="catalog-hero-image scale-105 object-cover blur-sm"
        />
        <div className="absolute inset-0 bg-catalog-hero/10" />
        <div className="catalog-hero-vignette absolute inset-0" />
      </div>

      <div className="relative z-10 mx-auto flex w-full max-w-7xl flex-col items-center px-4 pb-14 pt-12 text-center sm:px-6 sm:pb-16 sm:pt-16 lg:px-8">
        <div className="mb-5 flex size-11 items-center justify-center rounded-xl border border-catalog-hero-border bg-background/10 shadow-sm">
          {active === "models" ? (
            <Boxes className="size-5" aria-hidden="true" />
          ) : (
            <Cable className="size-5" aria-hidden="true" />
          )}
        </div>
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">{title}</h1>
        <p className="mt-3 max-w-3xl text-sm leading-7 text-catalog-hero-muted sm:text-base">
          {summary}
        </p>

        <nav
          aria-label={t("tabsLabel")}
          className="mt-7 inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-xl border border-catalog-hero-border bg-background/10 p-1"
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
