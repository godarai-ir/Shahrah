import Link from 'next/link';
import { apiGet } from '@/lib/api';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';

type ProviderDetail = {
  slug: string;
  name: string;
  description?: string;
  status: string;
  websiteUrl?: string;
  docsUrl?: string;
  models?: Array<{ id: string; slug: string; displayName: string }>;
};

function formatStatusLabel(status: string) {
  return status.replaceAll('_', ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function getStatusClasses(status: string) {
  const s = status.toLowerCase();
  if (s === 'active') return 'border-emerald-500/40 bg-emerald-500/10 text-emerald-300';
  if (s === 'maintenance') return 'border-amber-500/40 bg-amber-500/10 text-amber-300';
  if (s === 'inactive' || s === 'disabled') return 'border-rose-500/40 bg-rose-500/10 text-rose-300';
  return 'border-zinc-600 bg-zinc-800/60 text-zinc-200';
}

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
      <div className="space-y-4">
        <h1 className="text-2xl font-semibold">Provider</h1>
        <Card className="border-amber-700/50 bg-amber-950/30">
          <p className="text-sm text-amber-200">Unable to load provider details.</p>
          {error ? <p className="mt-1 text-xs text-amber-300/80">{error}</p> : null}
        </Card>
      </div>
    );
  }

  const modelCount = provider.models?.length ?? 0;
  const sortedModels = [...(provider.models ?? [])].sort((a, b) => a.displayName.localeCompare(b.displayName));

  return (
    <div className="space-y-6">
      <section className="relative overflow-hidden rounded-2xl border border-zinc-800 bg-gradient-to-br from-zinc-900 via-zinc-900 to-cyan-950/40 p-5 sm:p-7">
        <div className="pointer-events-none absolute -right-24 -top-24 h-56 w-56 rounded-full bg-cyan-500/10 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-20 left-0 h-44 w-44 rounded-full bg-sky-500/10 blur-3xl" />

        <div className="relative flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge className={getStatusClasses(provider.status)}>{formatStatusLabel(provider.status)}</Badge>
              <Badge className="border-zinc-700 bg-zinc-800/80 text-zinc-200">{modelCount} models</Badge>
            </div>
            <div>
              <h1 className="text-3xl font-semibold tracking-tight text-zinc-100">{provider.name}</h1>
              <p className="mt-1 text-sm text-zinc-400">/{provider.slug}</p>
            </div>
            <p className="max-w-2xl text-sm leading-6 text-zinc-300">{provider.description || 'No provider description is available yet.'}</p>
          </div>

          <div className="grid w-full gap-2 sm:w-auto">
            {provider.websiteUrl ? (
              <a
                href={provider.websiteUrl}
                target="_blank"
                rel="noreferrer"
                className="rounded-lg border border-zinc-700 bg-zinc-900/80 px-3 py-2 text-sm text-zinc-200 transition hover:border-cyan-500 hover:text-cyan-200"
              >
                Visit Website
              </a>
            ) : null}
            {provider.docsUrl ? (
              <a
                href={provider.docsUrl}
                target="_blank"
                rel="noreferrer"
                className="rounded-lg border border-zinc-700 bg-zinc-900/80 px-3 py-2 text-sm text-zinc-200 transition hover:border-cyan-500 hover:text-cyan-200"
              >
                View API Docs
              </a>
            ) : null}
            <Link
              href="/providers"
              className="rounded-lg border border-zinc-700 bg-zinc-900/80 px-3 py-2 text-sm text-zinc-300 transition hover:border-zinc-500 hover:text-zinc-100"
            >
              Back to Providers
            </Link>
          </div>
        </div>
      </section>

      <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <Card className="bg-zinc-900/50">
          <p className="text-xs uppercase tracking-wide text-zinc-500">Status</p>
          <p className="mt-1 text-lg font-semibold text-zinc-100">{formatStatusLabel(provider.status)}</p>
        </Card>
        <Card className="bg-zinc-900/50">
          <p className="text-xs uppercase tracking-wide text-zinc-500">Total Models</p>
          <p className="mt-1 text-lg font-semibold text-zinc-100">{modelCount}</p>
        </Card>
        <Card className="bg-zinc-900/50 sm:col-span-2 lg:col-span-1">
          <p className="text-xs uppercase tracking-wide text-zinc-500">Slug</p>
          <p className="mt-1 truncate text-lg font-semibold text-zinc-100">{provider.slug}</p>
        </Card>
      </section>

      <section className="space-y-3">
        <div className="flex items-end justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold text-zinc-100">Models</h2>
            <p className="text-sm text-zinc-400">Explore available models and open the full model profile.</p>
          </div>
        </div>

        {sortedModels.length === 0 ? (
          <Card>
            <p className="text-sm text-zinc-400">No models are currently attached to this provider.</p>
          </Card>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {sortedModels.map((m) => (
              <Link
                key={m.id}
                href={`/models/${provider.slug}/${m.slug}`}
                className="group rounded-xl border border-zinc-800 bg-zinc-900/70 p-4 transition hover:-translate-y-0.5 hover:border-cyan-500/70 hover:bg-zinc-900"
              >
                <p className="text-sm font-medium text-zinc-100 transition group-hover:text-cyan-200">{m.displayName}</p>
                <p className="mt-1 text-xs text-zinc-500">/{m.slug}</p>
                <p className="mt-3 text-xs text-cyan-300/80">View model details</p>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

