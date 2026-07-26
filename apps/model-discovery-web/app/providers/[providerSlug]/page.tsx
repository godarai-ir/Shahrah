import Link from "next/link";
import { ExternalLink } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { apiGet } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ProviderMark } from "@/components/catalog/provider-mark";
import { ProviderModelsList, type ProviderModel } from "@/components/provider-models-list";

type ProviderDetail = {
  slug: string;
  name: string;
  description?: string;
  status: string;
  websiteUrl?: string;
  docsUrl?: string;
  logoUrl?: string;
  models?: ProviderModel[];
};

type StatusKey = "active" | "inactive" | "disabled" | "maintenance";

function getStatusKey(status: string): StatusKey | null {
  const normalized = status.toLowerCase();
  if (normalized === "active" || normalized === "inactive" || normalized === "disabled" || normalized === "maintenance") {
    return normalized;
  }
  return null;
}

function getStatusClasses(status: string) {
  const normalized = status.toLowerCase();
  if (normalized === "active") return "border-success-soft bg-success-soft text-success-soft-foreground";
  if (normalized === "maintenance") return "border-warning-soft bg-warning-soft text-warning-soft-foreground";
  if (normalized === "inactive" || normalized === "disabled") return "border-destructive-soft bg-destructive-soft text-destructive-soft-foreground";
  return "border-border bg-muted text-muted-foreground";
}

export default async function ProviderDetailPage({ params }: { params: Promise<{ providerSlug: string }> }) {
  const { providerSlug } = await params;
  const t = await getTranslations("ProviderDetail");
  const tStatus = await getTranslations("Statuses");
  let provider: ProviderDetail | null = null;

  try {
    provider = await apiGet<ProviderDetail>(`/providers/${providerSlug}`);
  } catch {
    provider = null;
  }

  if (!provider) {
    return (
      <div className="mx-auto w-full max-w-7xl space-y-4 px-4 py-8 sm:px-6">
        <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{t("title")}</h1>
        <Card className="border-warning-soft bg-warning-soft text-warning-soft-foreground">
          <p className="text-sm" role="alert">{t("loadError")}</p>
        </Card>
      </div>
    );
  }

  const modelCount = provider.models?.length ?? 0;
  const sortedModels = [...(provider.models ?? [])].sort((a, b) => a.displayName.localeCompare(b.displayName, "fa"));
  const statusKey = getStatusKey(provider.status);
  const statusLabel = statusKey ? tStatus(statusKey) : provider.status;

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-8 sm:gap-8 sm:px-6 lg:py-10">
      <section className="rounded-2xl border border-border bg-card p-5 sm:p-7">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
          <div className="flex min-w-0 items-start gap-4">
            <ProviderMark name={provider.name} logoUrl={provider.logoUrl} size="lg" />
            <div className="min-w-0">
              <p className="text-sm font-medium text-primary">{t("eyebrow")}</p>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
                  <bdi dir="auto">{provider.name}</bdi>
                </h1>
                <Badge className={getStatusClasses(provider.status)}>{statusLabel}</Badge>
                <Badge>{t("modelCount", { count: modelCount })}</Badge>
              </div>
              <p className="mt-3 max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">
                {provider.description || t("noDescription")}
              </p>
              <code className="mt-3 inline-block text-xs text-muted-foreground" dir="ltr">{provider.slug}</code>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {provider.websiteUrl ? (
              <a href={provider.websiteUrl} target="_blank" rel="noreferrer" className="inline-flex items-center gap-2 rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent">
                {t("visitWebsite")}
                <ExternalLink className="size-4" aria-hidden="true" />
              </a>
            ) : null}
            {provider.docsUrl ? (
              <a href={provider.docsUrl} target="_blank" rel="noreferrer" className="inline-flex items-center gap-2 rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent">
                {t("viewApiDocs")}
                <ExternalLink className="size-4" aria-hidden="true" />
              </a>
            ) : null}
            <Link href="/providers" className="inline-flex items-center rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground">
              {t("backToProviders")}
            </Link>
          </div>
        </div>
      </section>

      {sortedModels.length === 0 ? (
        <Card><p className="text-sm text-muted-foreground">{t("noModels")}</p></Card>
      ) : (
        <ProviderModelsList providerSlug={provider.slug} models={sortedModels} docsUrl={provider.docsUrl} />
      )}
    </div>
  );
}
