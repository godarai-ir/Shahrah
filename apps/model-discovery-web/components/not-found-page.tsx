import Link from "next/link";
import { Boxes, House, RouteOff, SearchX, Sparkles } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { GodarAILogo } from "@/components/godarai-logo";

export type NotFoundCopy = {
  eyebrow: string;
  title: string;
  description: string;
  hint: string;
  homeAction: string;
  providersAction: string;
  brand: string;
  brandAria: string;
};

export function NotFoundPage({
  copy,
  showBrand = false,
}: {
  copy: NotFoundCopy;
  showBrand?: boolean;
}) {
  return (
    <div className={showBrand ? "flex min-h-screen flex-col" : "flex flex-col"}>
      {showBrand ? (
        <header className="border-b border-border bg-background/80 backdrop-blur-xl">
          <div className="mx-auto flex h-16 w-full max-w-7xl items-center px-4 sm:px-6 lg:px-8">
            <a
              href="/catalog"
              aria-label={copy.brandAria}
              className="flex items-center gap-2.5 rounded-xl font-bold tracking-tight text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <span className="flex size-10 items-center justify-center rounded-xl border border-card-border bg-card shadow-xs">
                <GodarAILogo
                  decorative
                  priority
                  className="h-8 w-auto max-w-full object-contain"
                  sizes="2rem"
                />
              </span>
              <span>{copy.brand}</span>
            </a>
          </div>
        </header>
      ) : null}

      <main className="mx-auto flex w-full max-w-7xl flex-1 items-center px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
        <section
          aria-labelledby="not-found-title"
          className="relative isolate grid w-full overflow-hidden rounded-3xl border border-card-border bg-card shadow-xl lg:grid-cols-2"
        >
          <div
            aria-hidden="true"
            className="pointer-events-none absolute -start-20 -top-20 size-72 rounded-full bg-primary-soft blur-3xl"
          />

          <div className="relative flex flex-col items-start justify-center gap-6 p-6 sm:p-10 lg:p-14">
            <Badge className="gap-1.5 border-primary-soft bg-primary-soft text-primary-soft-foreground">
              <Sparkles aria-hidden="true" />
              {copy.eyebrow}
            </Badge>

            <div className="flex max-w-xl flex-col gap-3">
              <h1
                id="not-found-title"
                className="font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl"
              >
                {copy.title}
              </h1>
              <p className="text-base leading-8 text-muted-foreground sm:text-lg">
                {copy.description}
              </p>
            </div>

            <div className="flex flex-wrap gap-2">
              <Button asChild size="lg">
                {showBrand ? (
                  <a href="/catalog">
                    <House data-icon="inline-start" aria-hidden="true" />
                    {copy.homeAction}
                  </a>
                ) : (
                  <Link href="/">
                    <House data-icon="inline-start" aria-hidden="true" />
                    {copy.homeAction}
                  </Link>
                )}
              </Button>
              <Button asChild size="lg" variant="outline">
                {showBrand ? (
                  <a href="/catalog/providers">
                    <Boxes data-icon="inline-start" aria-hidden="true" />
                    {copy.providersAction}
                  </a>
                ) : (
                  <Link href="/providers">
                    <Boxes data-icon="inline-start" aria-hidden="true" />
                    {copy.providersAction}
                  </Link>
                )}
              </Button>
            </div>

            <div className="flex max-w-lg items-start gap-3 rounded-xl border border-border bg-muted/60 p-4 text-sm leading-6 text-muted-foreground">
              <SearchX className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden="true" />
              <p>{copy.hint}</p>
            </div>
          </div>

          <div className="relative flex min-h-80 items-center justify-center overflow-hidden border-t border-card-border bg-muted/40 p-8 lg:min-h-96 lg:border-s lg:border-t-0">
            <div
              aria-hidden="true"
              className="absolute size-72 rounded-full border border-dashed border-primary/25"
            />
            <div
              aria-hidden="true"
              className="absolute size-52 rounded-full border border-primary/20"
            />
            <div
              aria-hidden="true"
              className="absolute size-32 rounded-full bg-primary-soft"
            />

            <div className="relative flex flex-col items-center gap-5 text-center">
              <div className="flex size-16 items-center justify-center rounded-2xl border border-primary/20 bg-card text-primary shadow-lg">
                <RouteOff className="size-8" aria-hidden="true" />
              </div>
              <p
                aria-hidden="true"
                className="font-heading text-8xl font-bold leading-none tracking-tighter text-primary sm:text-9xl"
              >
                ۴۰۴
              </p>
              <div
                aria-hidden="true"
                className="h-px w-32 bg-gradient-to-l from-transparent via-primary/60 to-transparent"
              />
            </div>

            <span
              aria-hidden="true"
              className="absolute end-8 top-8 size-3 rounded-full bg-primary shadow-sm"
            />
            <span
              aria-hidden="true"
              className="absolute bottom-10 start-10 size-2 rounded-full bg-success-soft-foreground"
            />
          </div>
        </section>
      </main>
    </div>
  );
}
