"use client";

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { apiGet, resolveAssetUrl } from '@/lib/api';
import { formatDisplayPrice } from '@/lib/display-settings';
import { useDisplaySettings } from '@/components/display-settings-provider';
import { CompareToggleButton } from '@/components/compare/compare-toggle-button';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

type ModelRow = {
  canonicalId: string;
  slug: string;
  name: string;
  provider: { slug: string; name: string; logoUrl?: string | null };
  modelSeries?: { slug: string; name: string; logoUrl?: string | null } | null;
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

type FacetOption = { slug: string; name: string };
type FacetsResponse = {
  providers?: FacetOption[];
  inputModalities?: FacetOption[];
  input_modalities?: FacetOption[];
  outputModalities?: FacetOption[];
  output_modalities?: FacetOption[];
  data?: {
    providers?: FacetOption[];
    inputModalities?: FacetOption[];
    input_modalities?: FacetOption[];
    outputModalities?: FacetOption[];
    output_modalities?: FacetOption[];
  };
};

export function ModelsExplorer() {
  const [q, setQ] = useState('');
  const [provider, setProvider] = useState('');
  const [inputModality, setInputModality] = useState('');
  const [outputModality, setOutputModality] = useState('');
  const [minContextWindow, setMinContextWindow] = useState('');
  const [maxContextWindow, setMaxContextWindow] = useState('');
  const [minPromptPrice, setMinPromptPrice] = useState('');
  const [maxPromptPrice, setMaxPromptPrice] = useState('');

  const params = useMemo(() => {
    const sp = new URLSearchParams();
    if (q) sp.set('q', q);
    if (provider) sp.set('provider', provider);
    if (inputModality) sp.set('input_modalities', inputModality);
    if (outputModality) sp.set('output_modalities', outputModality);
    if (minContextWindow) sp.set('min_context_window', minContextWindow);
    if (maxContextWindow) sp.set('max_context_window', maxContextWindow);
    if (minPromptPrice) sp.set('min_input_price', minPromptPrice);
    if (maxPromptPrice) sp.set('max_input_price', maxPromptPrice);
    sp.set('limit', '24');
    return sp.toString();
  }, [q, provider, inputModality, outputModality, minContextWindow, maxContextWindow, minPromptPrice, maxPromptPrice]);

  const modelsQ = useQuery({ queryKey: ['models', params], queryFn: () => apiGet<{ data: ModelRow[]; pagination: { total: number } }>(`/models?${params}`) });
  const facetsQ = useQuery({ queryKey: ['facets'], queryFn: () => apiGet<FacetsResponse>('/facets') });
  const { settings } = useDisplaySettings();

  const providers = facetsQ.data?.providers ?? facetsQ.data?.data?.providers ?? [];
  const inputModalities =
    facetsQ.data?.inputModalities ??
    facetsQ.data?.input_modalities ??
    facetsQ.data?.data?.inputModalities ??
    facetsQ.data?.data?.input_modalities ??
    [];
  const outputModalities =
    facetsQ.data?.outputModalities ??
    facetsQ.data?.output_modalities ??
    facetsQ.data?.data?.outputModalities ??
    facetsQ.data?.data?.output_modalities ??
    [];

  const activeFilterCount = [q, provider, inputModality, outputModality, minContextWindow, maxContextWindow, minPromptPrice, maxPromptPrice].filter(Boolean).length;

  return (
    <div className="space-y-4">
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <Input placeholder="Search model/provider/description" value={q} onChange={(e) => setQ(e.target.value)} />
        <select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={provider} onChange={(e) => setProvider(e.target.value)}>
          <option value="">All Providers</option>
          {providers.map((p) => <option key={p.slug} value={p.slug}>{p.name}</option>)}
        </select>
        <select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={inputModality} onChange={(e) => setInputModality(e.target.value)}>
          <option value="">All input modalities</option>
          {inputModalities.map((m) => <option key={m.slug} value={m.slug}>{m.name}</option>)}
        </select>
        <select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={outputModality} onChange={(e) => setOutputModality(e.target.value)}>
          <option value="">All output modalities</option>
          {outputModalities.map((m) => <option key={m.slug} value={m.slug}>{m.name}</option>)}
        </select>
        <Input type="number" placeholder="Context min" value={minContextWindow} onChange={(e) => setMinContextWindow(e.target.value)} />
        <Input type="number" placeholder="Context max" value={maxContextWindow} onChange={(e) => setMaxContextWindow(e.target.value)} />
        <Input type="number" step="0.0001" placeholder="Prompt price min ($/1M)" value={minPromptPrice} onChange={(e) => setMinPromptPrice(e.target.value)} />
        <Input type="number" step="0.0001" placeholder="Prompt price max ($/1M)" value={maxPromptPrice} onChange={(e) => setMaxPromptPrice(e.target.value)} />
        <Button type="button" onClick={() => { setQ(''); setProvider(''); setInputModality(''); setOutputModality(''); setMinContextWindow(''); setMaxContextWindow(''); setMinPromptPrice(''); setMaxPromptPrice(''); }}>Clear Filters ({activeFilterCount})</Button>
      </div>

      <p className="text-sm text-zinc-400">{modelsQ.data?.pagination.total ?? 0} models found</p>

      {modelsQ.isLoading ? <p>Loading models...</p> : null}
      {modelsQ.isError ? <p className="text-red-400">Failed to load models.</p> : null}
      {facetsQ.isError ? <p className="text-amber-400">Failed to load filter options.</p> : null}
      {!modelsQ.isLoading && modelsQ.data?.data.length === 0 ? <p className="text-zinc-400">No matching models.</p> : null}

      <div className="grid gap-3 lg:grid-cols-2">
        {modelsQ.data?.data.map((m) => (
          <Card key={m.canonicalId}>
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">{m.name}</h3>
                <p className="text-xs text-zinc-500">{m.canonicalId}</p>
                <div className="mt-1 flex items-center gap-2 text-xs text-zinc-500">
                  {resolveAssetUrl(m.provider.logoUrl) ? (
                    <img
                      src={resolveAssetUrl(m.provider.logoUrl) || ''}
                      alt={m.provider.name}
                      className="h-4 w-4 rounded"
                    />
                  ) : null}
                  {resolveAssetUrl(m.modelSeries?.logoUrl) ? (
                    <img
                      src={resolveAssetUrl(m.modelSeries?.logoUrl) || ''}
                      alt={m.modelSeries?.name || 'Model series'}
                      className="h-4 w-4 rounded"
                    />
                  ) : null}
                </div>
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
              <span>Input: {formatDisplayPrice(m.pricing?.inputPerMillionTokens, settings)}</span>
            </div>
            <div className="mt-3 flex flex-wrap gap-1">
              {m.capabilities.slice(0, 5).map((c) => <Badge key={c}>{c}</Badge>)}
            </div>
            <div className="mt-3 flex items-center gap-2">
              <Link href={`/models/${m.provider.slug}/${m.slug}`} className="inline-block text-sm text-cyan-400 hover:text-cyan-300">View details</Link>
              <CompareToggleButton canonicalId={m.canonicalId} compact />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
