'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  conversionWarning, formatCredits, formatToman, type CreditConversion, type TenantCreditConversionStatus,
} from '@/lib/support-credits';
import { controlPlaneRequest, formatDate, formatIRR } from '@/lib/support-tenants';

/**
 * Converts one tenant between IRR and credit billing. Converting moves the IRR
 * wallet into a migration lot, gives IRR budget rules a credit limit, and moves
 * the subscription to the plan's credit version; a rollback reverses the money.
 */
export function TenantConversionCard({ tenantId, tenantSlug, onChanged }: { tenantId: string; tenantSlug: string; onChanged: () => void }) {
  const [status, setStatus] = useState<TenantCreditConversionStatus>();
  const [reason, setReason] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  async function load() {
    try {
      setStatus(await controlPlaneRequest<TenantCreditConversionStatus>(`tenants/${tenantId}/credits/conversion`));
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load the conversion'); }
  }

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */
  useEffect(() => { void load(); }, [tenantId]);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  async function submit(path: string, confirmText: string, done: (result: CreditConversion) => string) {
    setError(''); setMessage('');
    if (!reason.trim()) { setError('A reason is required; it is written to the audit log.'); return; }
    if (!window.confirm(confirmText)) return;
    setBusy(true);
    try {
      const result = await controlPlaneRequest<CreditConversion>(path, { method: 'POST', body: JSON.stringify({ reason: reason.trim() }) });
      setReason(''); setMessage(done(result)); await load(); onChanged();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Conversion failed'); } finally { setBusy(false); }
  }

  if (!status) return error ? <p role="alert" className="text-sm text-red-300">{error}</p> : null;

  const { preview } = status;
  const isCredit = preview.billing_unit === 'credit';

  function convert() {
    const plan = preview.plan_slug && preview.new_plan_monthly_credits
      ? `\nSubscription moves to the credit version of ${preview.plan_slug} (${formatCredits(preview.new_plan_monthly_credits)} a month); the first period is free.`
      : '';
    void submit(`tenants/${tenantId}/credits/conversion`,
      `Convert ${tenantSlug} to credits?\n\nWallet ${formatIRR(preview.wallet_irr)} becomes ${formatCredits(preview.credits)}; ${formatIRR(preview.residue_irr)} stays in the wallet.${plan}`,
      (result) => `Converted: ${formatCredits(result.credits)} granted, ${result.rules_converted} rules given a credit limit.`);
  }

  function rollback() {
    void submit(`tenants/${tenantId}/credits/conversion/rollback`,
      `Move ${tenantSlug} back to IRR billing?\n\nRemaining credits go back to the IRR wallet at the price of the last conversion, and any debt is taken from the wallet.`,
      (result) => `Rolled back: ${formatIRR(result.converted_irr)} moved${result.uncollected_irr > 0 ? `, ${formatIRR(result.uncollected_irr)} of debt uncollected` : ''}.`);
  }

  return (
    <Card>
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h2 className="text-lg font-semibold">Credit conversion</h2>
          <p className="mt-1 text-sm text-zinc-400">Billed in <span className="font-medium uppercase text-zinc-200">{preview.billing_unit}</span> · list price {formatToman(preview.price_toman_per_credit)} per credit</p>
        </div>
      </div>
      {error ? <p role="alert" className="mt-3 rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}
      {message ? <p role="status" className="mt-3 rounded-md border border-emerald-900 bg-emerald-950/40 p-3 text-sm text-emerald-200">{message}</p> : null}

      <dl className="mt-4 grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
        <div><dt className="text-zinc-400">IRR wallet</dt><dd className="mt-1 font-medium">{formatIRR(preview.wallet_irr)}</dd></div>
        <div><dt className="text-zinc-400">Becomes</dt><dd className="mt-1 font-medium">{formatCredits(preview.credits)}</dd><dd className="text-xs text-zinc-500">{formatIRR(preview.residue_irr)} residue stays</dd></div>
        <div><dt className="text-zinc-400">Budget rules to convert</dt><dd className="mt-1 font-medium">{preview.rules_to_convert}</dd></div>
        <div><dt className="text-zinc-400">Plan</dt><dd className="mt-1 font-medium">{preview.plan_slug ?? '—'}</dd>{preview.new_plan_monthly_credits ? <dd className="text-xs text-zinc-500">{formatCredits(preview.new_plan_monthly_credits)} a month{preview.new_plan_price_irr != null ? ` · ${formatIRR(preview.new_plan_price_irr)} per ${preview.billing_interval === 'yearly' ? 'year' : 'period'} from renewal` : ''}</dd> : null}</div>
      </dl>

      {preview.warnings.length > 0 ? <ul className="mt-4 flex flex-col gap-1 text-sm text-amber-300">{preview.warnings.map((code) => <li key={code}>• {conversionWarning(code)}</li>)}</ul> : null}

      <div className="mt-4 flex flex-col gap-3 sm:flex-row">
        <Input placeholder="Required reason" value={reason} onChange={(e) => setReason(e.target.value)} aria-label="Conversion reason" />
        {isCredit ? null : <Button disabled={busy} onClick={convert}>Convert to credits</Button>}
        {isCredit && preview.credits !== '0.00' ? <Button disabled={busy} onClick={convert}>Convert remaining wallet</Button> : null}
        {isCredit ? <Button disabled={busy} className="bg-zinc-800 text-zinc-100 hover:bg-zinc-700" onClick={rollback}>Roll back to IRR</Button> : null}
      </div>

      {status.conversions.length > 0 ? (
        <div className="mt-5 overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400"><tr><th className="py-2">#</th><th>Direction</th><th>Credits</th><th>IRR</th><th>Price</th><th>Reason</th><th>Time</th></tr></thead>
            <tbody className="divide-y divide-zinc-800">{status.conversions.map((item) => (
              <tr key={item.id}>
                <td className="py-3">{item.sequence}</td>
                <td>{item.direction === 'to_credit' ? 'To credits' : 'Back to IRR'}</td>
                <td>{formatCredits(item.credits)}{item.debt !== '0.00' ? <span className="block text-xs text-amber-300">debt {formatCredits(item.debt)}</span> : null}</td>
                <td className="text-zinc-400">{formatIRR(item.converted_irr)}{item.uncollected_irr > 0 ? <span className="block text-xs text-amber-300">uncollected {formatIRR(item.uncollected_irr)}</span> : null}</td>
                <td className="text-zinc-400">{formatToman(item.price_toman_per_credit)}</td>
                <td className="text-zinc-400">{[item.reason, item.actor].filter(Boolean).join(' · ')}</td>
                <td className="text-zinc-400">{formatDate(item.created_at)}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : null}
    </Card>
  );
}
