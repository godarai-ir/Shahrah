"use client";

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/lib/api';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

type ModelRow = {
  canonicalId: string;
  slug: string;
  name: string;
  provider: { slug: string; name: string };
  description?: string;
  status: string;
  contextWindow?: number;
  maxOutputTokens?: number;
  inputModalities: string[];
  outputModalities: string[];
  capabilities: string[];
  pricing?: { inputPerMillionTokens?: number; outputPerMillionTokens?: number } | null;
  free: boolean;
};

export function ModelsExplorer() {
  const [q, setQ] = useState('');
  const [provider, setProvider] = useState('');
  const [outputModality, setOutputModality] = useState('');

  const params = useMemo(() => {
    const sp = new URLSearchParams();
    if (q) sp.set('q', q);
    if (provider) sp.set('provider', provider);
    if (outputModality) sp.set('output_modalities', outputModality);
    sp.set('limit', '24');
    return sp.toString();
  }, [q, provider, outputModality]);

  const modelsQ = useQuery({ queryKey: ['models', params], queryFn: () => apiGet<{ data: ModelRow[]; pagination: { total: number } }>(`/models?${params}`) });
  const facetsQ = useQuery({ queryKey: ['facets'], queryFn: () => apiGet<{ providers: Array<{ slug: string; name: string }>; outputModalities: Array<{ slug: string; name: string }> }>('/facets') });

  const activeFilterCount = [q, provider, outputModality].filter(Boolean).length;

  return (
    <div className="space-y-4">
      <div className="grid gap-3 md:grid-cols-4">
        <Input placeholder="Search model/provider/description" value={q} onChange={(e) => setQ(e.target.value)} />
        <select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={provider} onChange={(e) => setProvider(e.target.value)}>
          <option value="">All providers</option>
          {facetsQ.data?.providers.map((p) => <option key={p.slug} value={p.slug}>{p.name}</option>)}
        </select>
        <select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={outputModality} onChange={(e) => setOutputModality(e.target.value)}>
          <option value="">All output modalities</option>
          {facetsQ.data?.outputModalities.map((m) => <option key={m.slug} value={m.slug}>{m.name}</option>)}
        </select>
        <Button type="button" onClick={() => { setQ(''); setProvider(''); setOutputModality(''); }}>Clear Filters ({activeFilterCount})</Button>
      </div>

      <p className="text-sm text-zinc-400">{modelsQ.data?.pagination.total ?? 0} models found</p>

      {modelsQ.isLoading ? <p>Loading models...</p> : null}
      {modelsQ.isError ? <p className="text-red-400">Failed to load models.</p> : null}
      {!modelsQ.isLoading && modelsQ.data?.data.length === 0 ? <p className="text-zinc-400">No matching models.</p> : null}

      <div className="grid gap-3 lg:grid-cols-2">
        {modelsQ.data?.data.map((m) => (
          <Card key={m.canonicalId}>
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">{m.name}</h3>
                <p className="text-xs text-zinc-500">{m.canonicalId}</p>
              </div>
              <div className="flex gap-1">
                <Badge>{m.status}</Badge>
                {m.free ? <Badge className="border-emerald-400 text-emerald-300">free</Badge> : null}
              </div>
            </div>
            <p className="mt-2 text-sm text-zinc-300">{m.description || 'No description.'}</p>
            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-zinc-400">
              <span>Provider: {m.provider.name}</span>
              <span>Context: {m.contextWindow ?? '-'}</span>
              <span>Max output: {m.maxOutputTokens ?? '-'}</span>
              <span>Input $/1M: {m.pricing?.inputPerMillionTokens ?? '-'}</span>
            </div>
            <div className="mt-3 flex flex-wrap gap-1">
              {m.capabilities.slice(0, 5).map((c) => <Badge key={c}>{c}</Badge>)}
            </div>
            <Link href={`/models/${m.provider.slug}/${m.slug}`} className="mt-3 inline-block text-sm text-cyan-400 hover:text-cyan-300">View details</Link>
          </Card>
        ))}
      </div>
    </div>
  );
}
