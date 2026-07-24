import { ModelsExplorer } from '@/components/catalog/models-explorer';
import { PageHero } from '@/components/page-hero';
import { Boxes } from 'lucide-react';
import { getTranslations } from 'next-intl/server';

export default async function ModelsPage() {
  const t = await getTranslations('Catalog');

  return (
    <div className="flex flex-col gap-6 sm:gap-8">
      <PageHero
        eyebrow={t('eyebrow')}
        title={t('title')}
        description={t('subtitle')}
        icon={Boxes}
      />
      <ModelsExplorer />
    </div>
  );
}

