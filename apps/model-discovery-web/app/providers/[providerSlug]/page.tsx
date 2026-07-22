import Link from "next/link";
import { getTranslations } from "next-intl/server";
import { apiGet } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

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
    <div className="space-y-6">
      <section className="relative overflow-hidden rounded-2xl border border-card-border bg-card p-5 sm:p-7">
        <div className="relative flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge className={getStatusClasses(provider.status)}>{statusLabel}</Badge>
              <Badge>{t("modelCount", { count: modelCount })}</Badge>
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{provider.name}</h1>
              <p className="mt-1 text-sm text-muted-foreground"><bdi dir="ltr" data-technical>/{provider.slug}</bdi></p>
            </div>
            <p dir="auto" className="max-w-2xl text-sm leading-6 text-muted-foreground">{provider.description || t("noDescription")}</p>
          </div>

          <div className="grid w-full gap-2 sm:w-auto">
            {provider.websiteUrl ? (
              <a href={provider.websiteUrl} target="_blank" rel="noreferrer" className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground transition-colors hover:border-primary hover:text-primary">
                {t("visitWebsite")}
              </a>
            ) : null}
            {provider.docsUrl ? (
              <a href={provider.docsUrl} target="_blank" rel="noreferrer" className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground transition-colors hover:border-primary hover:text-primary">
                {t("viewApiDocs")}
              </a>
            ) : null}
            <Link href="/providers" className="rounded-lg border border-border bg-background px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground">
              {t("backToProviders")}
            </Link>
          </div>
        </div>
      </section>

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

      <section className="space-y-3">
        <div>
          <h2 className="text-base font-semibold md:text-lg">{tCommon("models")}</h2>
          <p className="text-sm leading-6 text-muted-foreground">{t("modelsDescription")}</p>
        </div>

        {sortedModels.length === 0 ? (
          <Card><p className="text-sm text-muted-foreground">{t("noModels")}</p></Card>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {sortedModels.map((model) => (
              <Link key={model.id} href={`/models/${provider.slug}/${model.slug}`} className="group rounded-xl border border-card-border bg-card p-4 transition hover:-translate-y-0.5 hover:border-primary hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                <p className="text-sm font-medium transition-colors group-hover:text-primary">{model.displayName}</p>
                <p className="mt-1 text-xs text-muted-foreground"><bdi dir="ltr" data-technical>/{model.slug}</bdi></p>
                <p className="mt-3 text-xs text-primary">{t("viewModelDetails")}</p>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
