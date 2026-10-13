'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { controlPlaneRequest, formatDate, formatIRR, type TenantListResponse } from '@/lib/support-tenants';

export default function TenantsPage() {
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('');
  const [plan, setPlan] = useState('');
  const [payload, setPayload] = useState<TenantListResponse>({ data: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  async function load(cursor = '') {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ limit: '25' });
    if (query.trim()) params.set('query', query.trim());
    if (status) params.set('status', status);
    if (plan.trim()) params.set('plan', plan.trim());
    if (cursor) params.set('cursor', cursor);
    try {
      setPayload(await controlPlaneRequest<TenantListResponse>(`tenants?${params}`));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load tenants');
    } finally {
      setLoading(false);
    }
  }

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */
  useEffect(() => { void load(); }, []);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  return (
    <div className="flex flex-col gap-6">
      <div>
        <p className="text-sm text-zinc-400">Customer operations</p>
        <h1 className="text-2xl font-semibold text-zinc-50">Tenants</h1>
      </div>
      <Card>
        <form className="grid gap-3 md:grid-cols-[1fr_180px_180px_auto]" onSubmit={(event) => { event.preventDefault(); void load(); }}>
          <Input aria-label="Search tenants" placeholder="Slug, domain, or tenant ID" value={query} onChange={(event) => setQuery(event.target.value)} />
          <select aria-label="Tenant status" className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">All statuses</option><option value="active">Active</option><option value="suspended">Suspended</option>
          </select>
          <Input aria-label="Plan slug" placeholder="Plan slug" value={plan} onChange={(event) => setPlan(event.target.value)} />
          <Button type="submit" disabled={loading}>{loading ? 'Loading…' : 'Search'}</Button>
        </form>
      </Card>
      {error ? <p role="alert" className="rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}
      <div className="overflow-x-auto rounded-xl border border-zinc-800">
        <table className="w-full text-left text-sm">
          <thead className="bg-zinc-900 text-zinc-400"><tr><th className="p-3">Tenant</th><th className="p-3">Status</th><th className="p-3">Plan</th><th className="p-3">Wallet</th><th className="p-3">Period end</th></tr></thead>
          <tbody className="divide-y divide-zinc-800">
            {payload.data.map((tenant) => (
              <tr key={tenant.id} className="bg-zinc-950 hover:bg-zinc-900/70">
                <td className="p-3"><Link className="font-medium text-cyan-300 hover:underline" href={`/admin/tenants/${tenant.id}`}>{tenant.slug}</Link><div className="text-xs text-zinc-500">{tenant.domain}</div></td>
                <td className="p-3 capitalize">{tenant.status}</td><td className="p-3">{tenant.plan_display_name || 'No plan'}</td>
                <td className="p-3 tabular-nums">{formatIRR(tenant.wallet_balance_minor)}</td>
                <td className="p-3 text-zinc-400">{formatDate(tenant.period_end)}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {!loading && payload.data.length === 0 ? <p className="p-8 text-center text-sm text-zinc-400">No tenants match these filters.</p> : null}
      </div>
      {payload.next_cursor ? <Button className="self-end" onClick={() => void load(payload.next_cursor)}>Next page</Button> : null}
    </div>
  );
}
