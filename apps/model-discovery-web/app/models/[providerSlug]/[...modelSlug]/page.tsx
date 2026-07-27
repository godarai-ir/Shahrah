import { apiGet } from '@/lib/api';
import { getTranslations } from 'next-intl/server';
import { CatalogBreadcrumb } from '@/components/catalog/catalog-breadcrumb';
import { ModelDetailView } from '@/components/catalog/model-detail-view';
import type { PricingProfile, PricingSummary } from '@/lib/pricing';

type ModelDetail = {
  canonicalId: string;
  slug: string;
  displayName: string;
  shortDescription?: string | null;
  longDescription?: string | null;
  status: string;
  lifecycleStatus?: string | null;
  contextWindow?: number | null;
  maxOutputTokens?: number | null;
  knowledgeCutoff?: string | null;
  releaseDate?: string | null;
  deprecatedAt?: string | null;
  tokenizer?: string | null;
  license?: string | null;
  officialDocsUrl?: string | null;
  provider?: {
    slug?: string | null;
    name?: string | null;
    status?: string | null;
    docsUrl?: string | null;
    websiteUrl?: string | null;
    privacyPolicyUrl?: string | null;
    supportsByok?: boolean | null;
    supportsRegionalRouting?: boolean | null;
  } | null;
  pricingStatus?: 'unknown' | 'priced' | 'free';
  pricingProfiles?: PricingProfile[] | null;
  pricingSummary?: PricingSummary | null;
  inputModalities?: Array<{
    modality?: { slug?: string | null; name?: string | null } | null;
  }> | null;
  outputModalities?: Array<{
    modality?: { slug?: string | null; name?: string | null } | null;
  }> | null;
  supportedParams?: Array<{
    supportedParameter?: { slug?: string | null; name?: string | null } | null;
  }> | null;
  capabilities?: Array<{
    capability?: { slug?: string | null; name?: string | null } | null;
  }> | null;
  regions?: Array<{
    region?: { slug?: string | null; name?: string | null } | null;
  }> | null;
  sources?: Array<{
    id?: string;
    title?: string | null;
    url?: string | null;
    sourceType?: string | null;
    notes?: string | null;
  }> | null;
  changelog?: Array<{
    id?: string;
    changeType?: string | null;
    title?: string | null;
    description?: string | null;
    changedAt?: string | null;
    createdBy?: string | null;
  }> | null;
  benchmarks?: Array<{
    id?: string;
    score?: number | null;
    percentile?: number | null;
    evaluatedAt?: string | null;
    benchmark?: { slug?: string | null; name?: string | null } | null;
  }> | null;
};

export default async function ModelDetailPage({
  params,
}: {
  params: Promise<{ providerSlug: string; modelSlug: string[] }>;
}) {
  const { providerSlug, modelSlug } = await params;
  const modelSlugPath = (modelSlug || []).join('/');
  const requestedCanonicalId = `${providerSlug}/${modelSlugPath}`;
  const tNavigation = await getTranslations('Navigation');

  let model: ModelDetail | null = null;
  let error: string | null = null;
  try {
    model = await apiGet<ModelDetail>(
      `/models/${encodeURIComponent(requestedCanonicalId)}`,
    );
  } catch (err) {
    error = err instanceof Error ? err.message : null;
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-5 px-4 py-8 sm:px-6 lg:py-10">
      <CatalogBreadcrumb
        providersLabel={tNavigation('providers')}
        providerSlug={model?.provider?.slug || providerSlug}
        providerLabel={model?.provider?.name || model?.provider?.slug || providerSlug}
        modelLabel={model?.displayName || modelSlugPath}
      />
      <ModelDetailView
        model={model}
        requestedCanonicalId={requestedCanonicalId}
        fetchError={error}
      />
    </div>
  );
}


