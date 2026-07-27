import { getTranslations } from "next-intl/server";
import { CatalogHero } from "@/components/catalog/catalog-hero";
import { ModelsExplorer } from "@/components/catalog/models-explorer";
import { apiGet } from "@/lib/api";

export async function ModelsCatalogPage() {
  const t = await getTranslations("Catalog");
  const stats = await getCatalogStats();

  return (
    <div className="flex flex-col">
      <CatalogHero
        active="models"
        compact
        title={t("heroTitle")}
        summary={t("heroSummary", stats)}
        modelCount={stats.modelCount}
      />
      <div className="mx-auto w-full max-w-7xl px-4 py-5 sm:px-6 sm:py-6 lg:px-8">
        <ModelsExplorer />
      </div>
    </div>
  );
}

async function getCatalogStats() {
  try {
    const providers = await apiGet<Array<{ modelCount?: number }>>("/providers");
    return {
      modelCount: providers.reduce((sum, provider) => sum + (provider.modelCount ?? 0), 0),
      providerCount: providers.length,
    };
  } catch {
    return { modelCount: 0, providerCount: 0 };
  }
}
