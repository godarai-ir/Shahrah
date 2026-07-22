import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, Boxes, Building2, CircleAlert } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { apiGet, resolveAssetUrl } from "@/lib/api";
import { cn } from "@/lib/utils";
import { ProviderLogo } from "./provider-logo";

type ProviderRow = {
  slug: string;
  name: string;
  logoUrl?: string | null;
  modelCount?: number;
  models?: Array<{ id: string }>;
};

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("Providers");

  return {
    title: t("metadataTitle"),
    description: t("metadataDescription"),
  };
}

export default async function ProvidersPage() {
  const t = await getTranslations("Providers");
  let providers: ProviderRow[] = [];
  let hasError = false;

  try {
    providers = await apiGet<ProviderRow[]>("/providers");
  } catch {
    hasError = true;
  }

  return (
    <div className="flex min-w-0 flex-col gap-6 sm:gap-8">
      <header className="flex items-start gap-4">
        <span className="hidden size-12 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary-soft-foreground sm:flex">
          <Building2 className="size-6" aria-hidden="true" />
        </span>
        <div className="flex min-w-0 flex-col items-start gap-2">
          <Badge className="border-primary-soft bg-primary-soft text-primary-soft-foreground">
            {t("eyebrow")}
          </Badge>
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{t("title")}</h1>
          <p className="max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
            {t("subtitle")}
          </p>
        </div>
      </header>

      <section aria-labelledby="providers-directory-title" className="flex min-w-0 flex-col gap-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex min-w-0 flex-col gap-1">
            <h2 id="providers-directory-title" className="text-base font-semibold sm:text-lg">
              {t("directoryTitle")}
            </h2>
            <p className="text-sm leading-6 text-muted-foreground">{t("directoryDescription")}</p>
          </div>
          {!hasError ? (
            <Badge className="self-start px-3 py-1 sm:self-auto">
              {t("providersFound", { count: providers.length })}
            </Badge>
          ) : null}
        </div>

        {hasError ? (
          <ProviderState tone="warning" message={t("loadError")} />
        ) : null}
        {!hasError && providers.length === 0 ? (
          <ProviderState tone="neutral" message={t("noProviders")} />
        ) : null}

        {!hasError && providers.length > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {providers.map((provider) => {
              const logoUrl = resolveAssetUrl(provider.logoUrl);
              const modelCount = provider.modelCount ?? provider.models?.length ?? 0;

              return (
                <Card
                  key={provider.slug}
                  className="group flex min-w-0 flex-col overflow-hidden p-0 transition-colors hover:border-primary/60"
                >
                  <Link
                    href={`/providers/${provider.slug}`}
                    className="flex h-full min-w-0 flex-col rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <CardHeader className="flex-row items-start gap-3">
                      <ProviderLogo
                        src={logoUrl}
                        alt={t("logoAlt", { name: provider.name })}
                      />
                      <div className="min-w-0 flex-1 pt-0.5">
                        <CardTitle className="truncate text-lg leading-7 transition-colors group-hover:text-primary">
                          <bdi dir="auto">{provider.name}</bdi>
                        </CardTitle>
                        <CardDescription>{t("activeModels", { count: modelCount })}</CardDescription>
                      </div>
                    </CardHeader>

                    <CardContent className="flex-1">
                      <dl className="grid grid-cols-2 gap-2">
                        <ProviderMetric label={t("slugLabel")} value={provider.slug} technical />
                        <ProviderMetric
                          label={t("modelsLabel")}
                          value={modelCount.toLocaleString("fa-IR")}
                        />
                      </dl>
                    </CardContent>

                    <CardFooter className="justify-between gap-3 border-t border-border bg-muted/30">
                      <span className="text-sm font-medium text-primary">{t("viewDetails")}</span>
                      <ArrowLeft
                        className="size-4 text-primary transition-transform group-hover:-translate-x-1"
                        aria-hidden="true"
                      />
                    </CardFooter>
                  </Link>
                </Card>
              );
            })}
          </div>
        ) : null}
      </section>
    </div>
  );
}

function ProviderMetric({
  label,
  value,
  technical = false,
}: {
  label: string;
  value: string;
  technical?: boolean;
}) {
  return (
    <div className="min-w-0 rounded-lg bg-muted/50 p-2.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd
        className="mt-1 truncate text-sm font-semibold"
        dir={technical ? "ltr" : undefined}
        data-technical={technical || undefined}
        title={value}
      >
        {value}
      </dd>
    </div>
  );
}

function ProviderState({ tone, message }: { tone: "warning" | "neutral"; message: string }) {
  return (
    <Card role={tone === "warning" ? "alert" : "status"} className="p-0">
      <CardHeader className="flex-row items-center gap-3">
        <span
          className={cn(
            "flex size-10 shrink-0 items-center justify-center rounded-xl",
            tone === "warning" && "bg-warning-soft text-warning-soft-foreground",
            tone === "neutral" && "bg-muted text-muted-foreground",
          )}
        >
          {tone === "warning" ? (
            <CircleAlert className="size-5" aria-hidden="true" />
          ) : (
            <Boxes className="size-5" aria-hidden="true" />
          )}
        </span>
        <CardTitle className="text-sm leading-6">{message}</CardTitle>
      </CardHeader>
    </Card>
  );
}
