import { ModelsExplorer } from '@/components/catalog/models-explorer';
import { CatalogHero } from '@/components/catalog/catalog-hero';
import { apiGet } from '@/lib/api';
import { getTranslations } from 'next-intl/server';

export default async function HomePage() {
  const t = await getTranslations('Catalog');
  const stats = await getCatalogStats();

  return (
    <div className="flex flex-col">
      <CatalogHero
        active="models"
        title={t('heroTitle')}
        summary={t('heroSummary', stats)}
      />
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <ModelsExplorer />
      </div>
    </div>
  );
}

async function getCatalogStats() {
  try {
    const providers = await apiGet<Array<{ modelCount?: number }>>('/providers');
    return {
      modelCount: providers.reduce((sum, provider) => sum + (provider.modelCount ?? 0), 0),
      providerCount: providers.length,
    };
  } catch {
    return { modelCount: 0, providerCount: 0 };
  }
}
