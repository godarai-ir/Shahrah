import Link from "next/link";
import { Building2, ExternalLink } from "lucide-react";
import { getTranslations } from "next-intl/server";
import { apiGet } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { PageHero } from "@/components/page-hero";
import { ProviderModelsList } from "@/components/provider-models-list";

type ProviderDetail = {
  slug: string;
  name: string;
  description?: string;
  status: string;
  websiteUrl?: string;
  docsUrl?: string;
  models?: Array<{ id: string; slug: string; displayName: string }>;
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
  const tCommon = await getTranslations("Common");
  const tStatus = await getTranslations("Statuses");
  let provider: ProviderDetail | null = null;

  try {
    provider = await apiGet<ProviderDetail>(`/providers/${providerSlug}`);
  } catch {
    provider = null;
  }

  if (!provider) {
    return (
      <div className="space-y-4">
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
    <div className="flex flex-col gap-6 sm:gap-8">
      <PageHero
        eyebrow={t("eyebrow")}
        title={provider.name}
        description={provider.description || t("noDescription")}
        icon={Building2}
        afterTitle={
          <>
            <Badge className={getStatusClasses(provider.status)}>{statusLabel}</Badge>
            <Badge>{t("modelCount", { count: modelCount })}</Badge>
          </>
        }
        actions={
          <>
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
          </>
        }
      />

      <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <p className="text-xs text-muted-foreground">{tCommon("status")}</p>
          <p className="mt-1 text-lg font-semibold">{statusLabel}</p>
        </Card>
        <Card>
          <p className="text-xs text-muted-foreground">{t("totalModels")}</p>
          <p className="mt-1 text-lg font-semibold">{new Intl.NumberFormat("fa-IR").format(modelCount)}</p>
        </Card>
        <Card className="sm:col-span-2 lg:col-span-1">
          <p className="text-xs text-muted-foreground">{tCommon("slug")}</p>
          <p className="mt-1 truncate text-lg font-semibold"><bdi dir="ltr" data-technical>{provider.slug}</bdi></p>
        </Card>
      </section>

      <section className="flex flex-col gap-4">
        <div>
          <h2 className="text-base font-semibold md:text-lg">{tCommon("models")}</h2>
          <p className="text-sm leading-6 text-muted-foreground">{t("modelsDescription")}</p>
        </div>

        {sortedModels.length === 0 ? (
          <Card><p className="text-sm text-muted-foreground">{t("noModels")}</p></Card>
        ) : (
          <ProviderModelsList providerSlug={provider.slug} models={sortedModels} />
        )}
      </section>
    </div>
  );
}
