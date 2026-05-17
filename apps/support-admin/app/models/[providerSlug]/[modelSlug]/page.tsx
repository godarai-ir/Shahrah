import { apiGet } from '@/lib/api';
import { Badge } from '@/components/ui/badge';

type Pricing = { id: string; currency: string; region?: { slug: string }; inputPricePerMillionTokens?: number; outputPricePerMillionTokens?: number; cachedInputPricePerMillionTokens?: number };
type ModelDetail = { displayName: string; status: string; canonicalId: string; longDescription?: string; shortDescription?: string; provider?: { name: string }; contextWindow?: number; maxOutputTokens?: number; knowledgeCutoff?: string; pricing?: Pricing[] };

export default async function ModelDetailPage({ params }: { params: Promise<{ providerSlug: string; modelSlug: string }> }) {
  const { providerSlug, modelSlug } = await params;
  let model: ModelDetail | null = null;
  let error: string | null = null;
  try {
    model = await apiGet<ModelDetail>(`/models/${encodeURIComponent(`${providerSlug}/${modelSlug}`)}`);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load model';
  }

  if (!model) {
    return (
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold">Model</h1>
        <p className="text-sm text-amber-300">Unable to load model details.</p>
        {error ? <p className="text-xs text-zinc-500">{error}</p> : null}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-semibold">{model.displayName}</h1>
        <Badge>{model.status}</Badge>
      </div>
      <p className="text-sm text-zinc-500">{model.canonicalId}</p>
      <p className="text-zinc-300">{model.longDescription || model.shortDescription}</p>

      <section className="grid gap-2 rounded-xl border border-zinc-800 bg-zinc-900 p-4 text-sm">
        <div>Provider: {model.provider?.name}</div>
        <div>Context window: {model.contextWindow ?? '-'}</div>
        <div>Max output tokens: {model.maxOutputTokens ?? '-'}</div>
        <div>Knowledge cutoff: {model.knowledgeCutoff ? new Date(model.knowledgeCutoff).toLocaleDateString() : '-'}</div>
      </section>

      <section className="rounded-xl border border-zinc-800 bg-zinc-900 p-4">
        <h2 className="mb-2 text-lg font-medium">Pricing</h2>
        <div className="space-y-2 text-sm">
          {model.pricing?.map((p) => (
            <div key={p.id} className="rounded border border-zinc-800 p-2">
              <div>{p.currency} | {p.region?.slug ?? 'global'}</div>
              <div>Input: {p.inputPricePerMillionTokens ?? '-'} | Output: {p.outputPricePerMillionTokens ?? '-'} | Cached: {p.cachedInputPricePerMillionTokens ?? '-'}</div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
