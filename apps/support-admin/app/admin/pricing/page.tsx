'use client';

import { useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';

type FeatureDefinition = {
  key: string;
  value_type: string;
  default_value: unknown;
  allow_override: boolean;
  category?: string;
};

type PlanVersion = {
  version: number;
  status: string;
};

type PlanPrice = {
  billing_interval: string;
  currency: string;
  amount_minor?: number | null;
  discount_percent: number;
};

type CommercialPlan = {
  slug: string;
  display_name: string;
  description?: string;
  active: boolean;
  versions?: PlanVersion[];
  prices?: PlanPrice[];
  features?: Record<string, unknown>;
};

const INTERVALS = ['monthly', 'quarterly', 'yearly'] as const;

export default function AdminPricingPage() {
  const [features, setFeatures] = useState<FeatureDefinition[]>([]);
  const [plans, setPlans] = useState<CommercialPlan[]>([]);
  const [selectedPlan, setSelectedPlan] = useState('');
  const [selectedVersion, setSelectedVersion] = useState('');
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const [planDraft, setPlanDraft] = useState({ slug: '', display_name: '', description: '' });
  const [versionDraft, setVersionDraft] = useState({ version: '', copy_from_version: '' });
  const [entitlementDraft, setEntitlementDraft] = useState({ feature_key: '', value: '' });
  const [priceDraft, setPriceDraft] = useState({ billing_interval: 'monthly', amount_minor: '', currency: 'IRR', discount_percent: '0' });
  const [overrideDraft, setOverrideDraft] = useState({ tenant_id: '', feature_key: '', value: '', reason: '' });

  useEffect(() => {
    void loadData();
  }, []);

  const activePlan = useMemo(
    () => plans.find((plan) => plan.slug === selectedPlan) ?? plans[0],
    [plans, selectedPlan],
  );
  const versions = activePlan?.versions ?? [];
  const activeVersion = Number(selectedVersion || versions[0]?.version || 0);

  async function loadData() {
    setLoading(true);
    setError('');
    try {
      const [featuresPayload, plansPayload] = await Promise.all([
        apiGet<{ data?: FeatureDefinition[] }>('features'),
        apiGet<{ data?: CommercialPlan[] }>('commercial-plans'),
      ]);
      const nextFeatures = featuresPayload.data ?? [];
      const nextPlans = plansPayload.data ?? [];
      setFeatures(nextFeatures);
      setPlans(nextPlans);
      setSelectedPlan((current) => current || nextPlans[0]?.slug || '');
      setSelectedVersion((current) => current || String(nextPlans[0]?.versions?.[0]?.version ?? ''));
      setEntitlementDraft((current) => ({ ...current, feature_key: current.feature_key || nextFeatures[0]?.key || '' }));
      setOverrideDraft((current) => ({ ...current, feature_key: current.feature_key || nextFeatures[0]?.key || '' }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pricing data');
    } finally {
      setLoading(false);
    }
  }

  async function createPlan(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submit(() => apiPost('commercial-plans', planDraft), 'Plan created.');
    setPlanDraft({ slug: '', display_name: '', description: '' });
  }

  async function createVersion(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!activePlan) return;
    await submit(
      () => apiPost(`commercial-plans/${activePlan.slug}/versions`, {
        version: Number(versionDraft.version),
        copy_from_version: versionDraft.copy_from_version ? Number(versionDraft.copy_from_version) : undefined,
      }),
      'Draft version created.',
    );
    setVersionDraft({ version: '', copy_from_version: '' });
  }

  async function publishVersion() {
    if (!activePlan || !activeVersion) return;
    await submit(
      () => apiPost(`commercial-plans/${activePlan.slug}/versions/${activeVersion}/publish`, {}),
      'Version published.',
    );
  }

  async function saveEntitlement(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!activePlan || !activeVersion) return;
    await submit(
      () => apiPost(`commercial-plans/${activePlan.slug}/versions/${activeVersion}/entitlements`, {
        feature_key: entitlementDraft.feature_key,
        value: parseJSONValue(entitlementDraft.value),
      }),
      'Entitlement saved.',
    );
  }

  async function savePrice(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!activePlan || !activeVersion) return;
    await submit(
      () => apiPost(`commercial-plans/${activePlan.slug}/versions/${activeVersion}/prices`, {
        billing_interval: priceDraft.billing_interval,
        currency: priceDraft.currency || 'IRR',
        amount_minor: priceDraft.amount_minor.trim() ? Number(priceDraft.amount_minor) : null,
        discount_percent: Number(priceDraft.discount_percent || 0),
      }),
      'Price saved.',
    );
  }

  async function saveOverride(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submit(
      () => apiPut('tenant-entitlement-overrides', {
        tenant_id: overrideDraft.tenant_id,
        feature_key: overrideDraft.feature_key,
        value: parseJSONValue(overrideDraft.value),
        reason: overrideDraft.reason,
      }),
      'Tenant override saved.',
    );
  }

  async function submit(action: () => Promise<unknown>, success: string) {
    setError('');
    setMessage('');
    try {
      await action();
      setMessage(success);
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed');
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <p className="text-sm text-zinc-400">Commercial plans</p>
          <h1 className="text-2xl font-semibold text-zinc-50">Plan Builder</h1>
        </div>
        <button className="rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100" onClick={() => void loadData()}>
          Refresh
        </button>
      </div>

      {loading ? <p className="text-sm text-zinc-400">Loading...</p> : null}
      {error ? <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-200">{error}</p> : null}
      {message ? <p className="rounded-md border border-emerald-900/50 bg-emerald-950/40 px-3 py-2 text-sm text-emerald-200">{message}</p> : null}

      <section className="grid gap-4 lg:grid-cols-[320px_1fr]">
        <div className="space-y-4 rounded-lg border border-zinc-800 bg-zinc-950 p-4">
          <h2 className="text-base font-semibold text-zinc-100">Plans</h2>
          <div className="space-y-2">
            {plans.map((plan) => (
              <button
                key={plan.slug}
                className={`w-full rounded-md border px-3 py-2 text-left text-sm ${activePlan?.slug === plan.slug ? 'border-cyan-500 bg-cyan-950/30 text-cyan-100' : 'border-zinc-800 text-zinc-200'}`}
                onClick={() => {
                  setSelectedPlan(plan.slug);
                  setSelectedVersion(String(plan.versions?.[0]?.version ?? ''));
                }}
              >
                <span className="block font-medium">{plan.display_name}</span>
                <span className="text-xs text-zinc-500">{plan.slug}</span>
              </button>
            ))}
          </div>
          <form className="space-y-2 border-t border-zinc-800 pt-4" onSubmit={createPlan}>
            <input className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="slug" value={planDraft.slug} onChange={(event) => setPlanDraft({ ...planDraft, slug: event.target.value })} />
            <input className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="display name" value={planDraft.display_name} onChange={(event) => setPlanDraft({ ...planDraft, display_name: event.target.value })} />
            <input className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="description" value={planDraft.description} onChange={(event) => setPlanDraft({ ...planDraft, description: event.target.value })} />
            <button className="w-full rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Create plan</button>
          </form>
        </div>

        <div className="space-y-4">
          <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-4">
            <div className="flex flex-wrap items-center gap-3">
              <select className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" value={selectedVersion} onChange={(event) => setSelectedVersion(event.target.value)}>
                {versions.map((version) => (
                  <option key={version.version} value={version.version}>
                    v{version.version} - {version.status}
                  </option>
                ))}
              </select>
              <button className="rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100" onClick={() => void publishVersion()}>
                Publish version
              </button>
            </div>
          </div>

          <form className="grid gap-3 rounded-lg border border-zinc-800 bg-zinc-950 p-4 md:grid-cols-4" onSubmit={createVersion}>
            <h2 className="text-base font-semibold text-zinc-100 md:col-span-4">Draft version</h2>
            <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="version" value={versionDraft.version} onChange={(event) => setVersionDraft({ ...versionDraft, version: event.target.value })} />
            <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="copy from version" value={versionDraft.copy_from_version} onChange={(event) => setVersionDraft({ ...versionDraft, copy_from_version: event.target.value })} />
            <button className="rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black md:col-span-2">Create draft</button>
          </form>

          <div className="grid gap-4 xl:grid-cols-2">
            <form className="space-y-3 rounded-lg border border-zinc-800 bg-zinc-950 p-4" onSubmit={saveEntitlement}>
              <h2 className="text-base font-semibold text-zinc-100">Entitlement</h2>
              <select className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" value={entitlementDraft.feature_key} onChange={(event) => setEntitlementDraft({ ...entitlementDraft, feature_key: event.target.value })}>
                {features.map((feature) => (
                  <option key={feature.key} value={feature.key}>{feature.key}</option>
                ))}
              </select>
              <input className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder='value JSON, e.g. true or 10000 or "custom"' value={entitlementDraft.value} onChange={(event) => setEntitlementDraft({ ...entitlementDraft, value: event.target.value })} />
              <button className="rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Save entitlement</button>
            </form>

            <form className="space-y-3 rounded-lg border border-zinc-800 bg-zinc-950 p-4" onSubmit={savePrice}>
              <h2 className="text-base font-semibold text-zinc-100">Price</h2>
              <select className="w-full rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" value={priceDraft.billing_interval} onChange={(event) => setPriceDraft({ ...priceDraft, billing_interval: event.target.value })}>
                {INTERVALS.map((interval) => <option key={interval} value={interval}>{interval}</option>)}
              </select>
              <div className="grid grid-cols-3 gap-2">
                <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="amount" value={priceDraft.amount_minor} onChange={(event) => setPriceDraft({ ...priceDraft, amount_minor: event.target.value })} />
                <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="currency" value={priceDraft.currency} onChange={(event) => setPriceDraft({ ...priceDraft, currency: event.target.value })} />
                <input className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm" placeholder="discount %" value={priceDraft.discount_percent} onChange={(event) => setPriceDraft({ ...priceDraft, discount_percent: event.target.value })} />
              </div>
              <button className="rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Save price</button>
            </form>
          </div>

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

          <div className="rounded-lg border border-zinc-800 bg-zinc-950 p-4">
            <h2 className="mb-3 text-base font-semibold text-zinc-100">Published values</h2>
            <pre className="max-h-80 overflow-auto rounded-md bg-black p-3 text-xs text-zinc-300">
              {JSON.stringify({ features: activePlan?.features ?? {}, prices: activePlan?.prices ?? [] }, null, 2)}
            </pre>
          </div>
        </div>
      </section>
    </div>
  );
}

async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, { cache: 'no-store' });
  return parseResponse<T>(response);
}

async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
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
