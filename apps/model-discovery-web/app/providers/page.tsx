import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { CatalogHero } from "@/components/catalog/catalog-hero";
import {
  ProvidersExplorer,
  type ProviderDirectoryItem,
} from "@/components/catalog/providers-explorer";
import { apiGet } from "@/lib/api";

type ProviderModelsResponse = {
  data: Array<{ mode?: string[] }>;
};

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("Providers");
  return {
    title: t("metadataTitle"),
    description: t("metadataDescription"),
  };
}

export default async function ProvidersPage({
  searchParams,
}: {
  searchParams?: Promise<{ section?: string | string[] }>;
}) {
  const t = await getTranslations("Providers");
  const params = await searchParams;
  const section = params?.section === "providers" ? "providers" : "endpoints";
  let providers: ProviderDirectoryItem[] = [];
  let hasError = false;

  try {
    providers = await apiGet<ProviderDirectoryItem[]>("/providers");
    providers = await enrichProviderModes(providers);
  } catch {
    hasError = true;
  }

  const stats = {
    providerCount: providers.length,
    modelCount: providers.reduce((sum, provider) => sum + (provider.modelCount ?? 0), 0),
    endpointCount: providers.reduce((sum, provider) => sum + provider.modes.length, 0),
  };

  return (
    <div className="flex flex-col">
      <CatalogHero
        active="providers"
        title={t("heroTitle")}
        summary={t("heroSummary", stats)}
        modelCount={stats.modelCount}
        providerCount={stats.providerCount}
        endpointCount={stats.endpointCount}
        activeProviderSection={section}
      />
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <ProvidersExplorer providers={providers} hasError={hasError} initialSection={section} />
      </div>
    </div>
  );
}

async function enrichProviderModes(
  providers: ProviderDirectoryItem[],
): Promise<ProviderDirectoryItem[]> {
  const missingModes = providers.filter(
    (provider) => !Array.isArray(provider.modes) || provider.modes.length === 0,
  );
  if (missingModes.length === 0) return providers;

  const modeResults = await Promise.allSettled(
    missingModes.map(async (provider) => {
      const response = await apiGet<ProviderModelsResponse>(
        `/providers/${provider.slug}/models`,
      );
      return [
        provider.slug,
        Array.from(
          new Set(
            response.data.flatMap((model) => model.mode ?? []).filter((mode) => mode !== "unknown"),
          ),
        ).sort(),
      ] as const;
    }),
  );

  const modesByProvider = new Map<string, string[]>();
  for (const result of modeResults) {
    if (result.status === "fulfilled") {
      modesByProvider.set(result.value[0], result.value[1]);
    }
  }

  return providers.map((provider) => ({
    ...provider,
    modes: provider.modes?.length
      ? provider.modes.filter((mode) => mode !== "unknown")
      : (modesByProvider.get(provider.slug) ?? []),
  }));
}
