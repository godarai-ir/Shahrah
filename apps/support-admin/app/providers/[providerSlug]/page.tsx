import Link from 'next/link';
import { apiGet } from '@/lib/api';

type ProviderDetail = { slug: string; name: string; description?: string; status: string; websiteUrl?: string; docsUrl?: string; models?: Array<{ id: string; slug: string; displayName: string }> };

export default async function ProviderDetailPage({ params }: { params: Promise<{ providerSlug: string }> }) {
  const { providerSlug } = await params;
  let provider: ProviderDetail | null = null;
  let error: string | null = null;
  try {
    provider = await apiGet<ProviderDetail>(`/providers/${providerSlug}`);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load provider';
  }

  if (!provider) {
    return (
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold">Provider</h1>
        <p className="text-sm text-amber-300">Unable to load provider details.</p>
        {error ? <p className="text-xs text-zinc-500">{error}</p> : null}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">{provider.name}</h1>
      <p className="text-zinc-400">{provider.description || 'No description.'}</p>
      <div className="grid gap-2 rounded-xl border border-zinc-800 bg-zinc-900 p-4 text-sm">
        <div>Status: {provider.status}</div>
        <div>Website: {provider.websiteUrl || '-'}</div>
        <div>Docs: {provider.docsUrl || '-'}</div>
      </div>
      <h2 className="text-lg font-semibold">Models</h2>
      <div className="grid gap-2">
        {provider.models?.map((m) => (
          <Link key={m.id} href={`/models/${provider.slug}/${m.slug}`} className="rounded border border-zinc-800 bg-zinc-900 px-3 py-2 hover:border-cyan-500">{m.displayName}</Link>
        ))}
      </div>
    </div>
  );
}
