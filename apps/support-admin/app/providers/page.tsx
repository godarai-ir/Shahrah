import Link from 'next/link';
import { apiGet } from '@/lib/api';

type ProviderRow = { slug: string; name: string; models?: Array<{ id: string }> };

export default async function ProvidersPage() {
  let providers: ProviderRow[] = [];
  let error: string | null = null;
  try {
    providers = await apiGet<ProviderRow[]>('/providers');
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load providers';
  }
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Providers</h1>
      {error ? (
        <div className="rounded-xl border border-amber-700 bg-amber-950/40 p-4 text-sm text-amber-200">
          Failed to load providers. Ensure backend is running and `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` is correct.
          <div className="mt-1 text-xs text-amber-300">{error}</div>
        </div>
      ) : null}
      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
        {providers.map((p) => (
          <Link key={p.slug} href={`/providers/${p.slug}`} className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">
            <h2 className="text-lg font-medium">{p.name}</h2>
            <p className="text-xs text-zinc-500">{p.slug}</p>
            <p className="mt-2 text-sm text-zinc-400">Active models: {p.models?.length ?? 0}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
