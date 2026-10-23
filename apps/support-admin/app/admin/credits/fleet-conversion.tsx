'use client';

import { useEffect, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  blocksPlanMove, conversionWarning, formatCredits, type CreditConversionPreview, type CreditFleetConversionResult,
} from '@/lib/support-credits';
import { controlPlaneRequest, formatIRR } from '@/lib/support-tenants';

const PREVIEW_LIMIT = 200;

/**
 * Dry run and batch execution of the fleet cutover to credit billing. Every
 * IRR tenant is converted; nobody opts in or out (product decision P2).
 */
export function FleetConversion() {
  const [previews, setPreviews] = useState<CreditConversionPreview[]>();
  const [reason, setReason] = useState('');
  const [batch, setBatch] = useState('20');
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<CreditFleetConversionResult>();
  const [error, setError] = useState('');

  async function load() {
    try {
      const payload = await controlPlaneRequest<{ data: CreditConversionPreview[] }>(`credits/conversions/preview?limit=${PREVIEW_LIMIT}`);
      setPreviews(payload.data ?? []);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load the dry run'); }
  }

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, []);
  /* eslint-enable react-hooks/set-state-in-effect */

  async function convertBatch(event: FormEvent) {
    event.preventDefault(); setError('');
    const limit = Number(batch);
    if (!Number.isInteger(limit) || limit < 1 || limit > 100) { setError('Batch size must be a whole number from 1 to 100.'); return; }
    if (!reason.trim()) { setError('A reason is required; it is written to every tenant\'s audit log.'); return; }
    if (!window.confirm(`Convert the next ${limit} IRR tenants to credit billing? Wallets become credits, budget limits get credit limits, and subscriptions move to the credit plans.`)) return;
    setBusy(true);
    try {
      setResult(await controlPlaneRequest<CreditFleetConversionResult>('credits/conversions/fleet', {
        method: 'POST', body: JSON.stringify({ reason: reason.trim(), limit }),
      }));
      await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Fleet conversion failed'); } finally { setBusy(false); }
  }

  const needsAttention = (previews ?? []).filter((item) => item.warnings.length > 0).length;
  const planBlocked = (previews ?? []).filter(blocksPlanMove).length;

  return (
    <Card>
      <h2 className="text-lg font-semibold">Fleet conversion to credits</h2>
      <p className="mt-1 text-sm text-zinc-400">Dry run of every IRR tenant, oldest first. Each tenant converts in its own transaction, so a failure stops only that tenant; run batches until none are left.</p>
      {error ? <p role="alert" className="mt-3 rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}

      <div className="mt-4 grid gap-3 text-sm sm:grid-cols-3">
        <div><p className="text-zinc-400">IRR tenants{previews && previews.length === PREVIEW_LIMIT ? ` (first ${PREVIEW_LIMIT})` : ''}</p><p className="mt-1 text-xl font-semibold">{previews?.length ?? '—'}</p></div>
        <div><p className="text-zinc-400">With warnings</p><p className={`mt-1 text-xl font-semibold ${needsAttention > 0 ? 'text-amber-300' : ''}`}>{previews ? needsAttention : '—'}</p></div>
        <div><p className="text-zinc-400">Plan not moved</p><p className={`mt-1 text-xl font-semibold ${planBlocked > 0 ? 'text-amber-300' : ''}`}>{previews ? planBlocked : '—'}</p></div>
      </div>

      <form className="mt-4 flex flex-col gap-3 sm:flex-row" onSubmit={convertBatch}>
        <Input placeholder="Required reason, e.g. credit cutover" value={reason} onChange={(e) => setReason(e.target.value)} aria-label="Fleet conversion reason" />
        <label className="flex items-center gap-2 text-sm text-zinc-400">Batch<Input className="w-20" inputMode="numeric" value={batch} onChange={(e) => setBatch(e.target.value)} aria-label="Batch size" /></label>
        <Button type="submit" disabled={busy || !previews || previews.length === 0}>{busy ? 'Converting…' : 'Convert next batch'}</Button>
      </form>

      {result ? (
        <div role="status" className="mt-4 rounded-md border border-zinc-800 p-3 text-sm">
          <p>{result.converted.length} converted · {result.failed.length} failed · {result.remaining} IRR tenants left</p>
          {result.failed.length > 0 ? <ul className="mt-2 flex flex-col gap-1 text-red-300">{result.failed.map((item) => <li key={item.tenant_id}><Link className="hover:underline" href={`/admin/tenants/${item.tenant_id}`}>{item.tenant_slug || item.tenant_id}</Link>: {item.error}</li>)}</ul> : null}
        </div>
      ) : null}

      {previews && previews.length > 0 ? (
        <div className="mt-5 overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400"><tr><th className="py-2">Tenant</th><th>Wallet</th><th>Credits</th><th>Rules</th><th>Plan</th><th>Warnings</th></tr></thead>
            <tbody className="divide-y divide-zinc-800">{previews.map((item) => (
              <tr key={item.tenant_id}>
                <td className="py-3"><Link className="text-cyan-300 hover:underline" href={`/admin/tenants/${item.tenant_id}`}>{item.tenant_slug}</Link></td>
                <td>{formatIRR(item.wallet_irr)}{item.unsettled_irr > 0 ? <span className="block text-xs text-zinc-500">−{formatIRR(item.unsettled_irr)} unsettled</span> : null}</td>
                <td>{formatCredits(item.credits)}</td>
                <td>{item.rules_to_convert}</td>
                <td>{item.plan_slug ?? '—'}{item.new_plan_monthly_credits ? <span className="block text-xs text-zinc-500">{formatCredits(item.new_plan_monthly_credits)} a month</span> : null}</td>
                <td className="text-xs text-amber-300">{item.warnings.map(conversionWarning).join(' ')}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : null}
      {previews && previews.length === 0 ? <p className="mt-4 text-sm text-zinc-400">Every tenant is billed in credits.</p> : null}
    </Card>
  );
}
