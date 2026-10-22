'use client';

import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import Link from 'next/link';
import { Card } from '@/components/ui/card';
import { SeatTypesCard } from '../credits/seat-types-card';

type FeatureDefinition = {
  key: string;
  value_type: string;
  default_value: unknown;
  allow_override: boolean;
  category?: string;
};

/**
 * Seat pricing. Commercial plans are retired: every tenant gets every feature, and revenue comes
 * from paid seats and credit packs, so this page only edits seat types and tenant overrides.
 */
export default function AdminPricingPage() {
  const [features, setFeatures] = useState<FeatureDefinition[]>([]);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [overrideDraft, setOverrideDraft] = useState({ tenant_id: '', feature_key: '', value: '', reason: '' });

  useEffect(() => {
    apiGet<{ data?: FeatureDefinition[] }>('features')
      .then((payload) => {
        const nextFeatures = payload.data ?? [];
        setFeatures(nextFeatures);
        setOverrideDraft((current) => ({ ...current, feature_key: current.feature_key || nextFeatures.find((feature) => feature.allow_override)?.key || '' }));
      })
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'Failed to load features'));
  }, []);

  async function saveOverride(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setMessage('');
    try {
      await apiPut('tenant-entitlement-overrides', {
        tenant_id: overrideDraft.tenant_id,
        feature_key: overrideDraft.feature_key,
        value: parseJSONValue(overrideDraft.value),
        reason: overrideDraft.reason,
      });
      setMessage('Tenant override saved.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed');
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="text-sm text-zinc-400">Seat-only pricing</p>
        <h1 className="text-2xl font-semibold text-zinc-50">Pricing</h1>
      </div>

      <Card>
        <h2 className="text-base font-semibold text-zinc-100">Commercial plans are retired</h2>
        <p className="mt-2 text-sm text-zinc-400">
          Every tenant gets every feature. Tenants pay only for paid seats (the subscription, monthly or yearly, anchored to the first paid seat) and for credit packs that fund the shared pool used by the API.
          Up to 5 free seats replace the old plan user limit, and there is no trial. Seat prices follow the active credit list price; change it, the rate card and packs on the{' '}
          <Link href="/admin/credits" className="text-cyan-300 hover:underline">Credits</Link> page.
        </p>
      </Card>

      <SeatTypesCard />

      {error ? <p role="alert" className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-200">{error}</p> : null}
      {message ? <p role="status" className="rounded-md border border-emerald-900/50 bg-emerald-950/40 px-3 py-2 text-sm text-emerald-200">{message}</p> : null}

      <form className="space-y-3 rounded-lg border border-zinc-800 bg-zinc-950 p-4" onSubmit={saveOverride}>
        <h2 className="text-base font-semibold text-zinc-100">Tenant override</h2>
        <div className="grid gap-2 md:grid-cols-4">
          <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="tenant uuid" value={overrideDraft.tenant_id} onChange={(event) => setOverrideDraft({ ...overrideDraft, tenant_id: event.target.value })} />
          <select className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" value={overrideDraft.feature_key} onChange={(event) => setOverrideDraft({ ...overrideDraft, feature_key: event.target.value })}>
            {features.filter((feature) => feature.allow_override).map((feature) => (
              <option key={feature.key} value={feature.key}>{feature.key}</option>
            ))}
          </select>
          <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="value JSON" value={overrideDraft.value} onChange={(event) => setOverrideDraft({ ...overrideDraft, value: event.target.value })} />
          <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="reason" value={overrideDraft.reason} onChange={(event) => setOverrideDraft({ ...overrideDraft, reason: event.target.value })} />
        </div>
        <button className="rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Save override</button>
      </form>
    </div>
  );
}

async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, { cache: 'no-store' });
  return parseResponse<T>(response);
}

async function apiPut<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, {
    method: 'PUT',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
  return parseResponse<T>(response);
}

async function parseResponse<T>(response: Response): Promise<T> {
  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok) {
    const message = payload?.error?.message ?? payload?.message ?? payload?.error ?? `Request failed: ${response.status}`;
    throw new Error(String(message));
  }
  return payload as T;
}

function parseJSONValue(raw: string): unknown {
  const trimmed = raw.trim();
  if (!trimmed) return null;
  try {
    return JSON.parse(trimmed);
  } catch {
    return trimmed;
  }
}
