import { ModelsExplorer } from '@/components/catalog/models-explorer';
import { Badge } from '@/components/ui/badge';
import { Boxes } from 'lucide-react';
import { getTranslations } from 'next-intl/server';

export default async function HomePage() {
  const t = await getTranslations('Catalog');

  return (
    <div className="flex flex-col gap-6 sm:gap-8">
      <header className="flex items-start gap-4">
        <span className="hidden size-12 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary-soft-foreground sm:flex">
          <Boxes className="size-6" aria-hidden="true" />
        </span>
        <div className="flex min-w-0 flex-col items-start gap-2">
          <Badge className="border-primary-soft bg-primary-soft text-primary-soft-foreground">{t('eyebrow')}</Badge>
          <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{t('title')}</h1>
          <p className="max-w-3xl text-sm leading-7 text-muted-foreground sm:text-base">{t('subtitle')}</p>
        </div>
      </header>
      <ModelsExplorer />
    </div>
  );
}
