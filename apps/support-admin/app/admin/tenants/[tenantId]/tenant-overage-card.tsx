'use client';

import { useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { formatCredits, type CreditSettings, type TenantCreditOverview } from '@/lib/support-credits';
import { controlPlaneRequest } from '@/lib/support-tenants';

const CAP_PATTERN = /^\d+(\.\d{1,2})?$/;

/**
 * Lets support turn on overage for a tenant: once the plan quota and packs run
 * out, requests keep running as debt up to the cap. Tenants cannot change this
 * themselves because it extends them credit.
 */
export function TenantOverageCard({ tenantId, tenantSlug, settings, onSaved }: {
  tenantId: string; tenantSlug: string; settings: CreditSettings; onSaved: (overview: TenantCreditOverview) => void;
}) {
  const [enabled, setEnabled] = useState(settings.overage_enabled);
  const [cap, setCap] = useState(settings.overage_cap_credits ?? '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const trimmedCap = cap.trim();
  const dirty = enabled !== settings.overage_enabled || trimmedCap !== (settings.overage_cap_credits ?? '');
  const isCreditTenant = settings.billing_unit === 'credit';

  async function submit(event: FormEvent) {
    event.preventDefault(); setError('');
    if (trimmedCap && (!CAP_PATTERN.test(trimmedCap) || Number(trimmedCap) <= 0)) {
      setError('The cap must be a positive number of credits with at most two decimals, or empty for no cap.'); return;
    }
    if (enabled && !settings.overage_allowed) { setError('The tenant plan does not allow overage. Grant the credits.overage_allowed entitlement first.'); return; }
    const capText = trimmedCap ? `a cap of ${formatCredits(trimmedCap)}` : 'no cap';
    const confirmText = enabled
      ? `Allow ${tenantSlug} to keep running as debt with ${capText}? The debt is invoiced at list price.`
      : `Turn off overage for ${tenantSlug}? Requests are rejected once credits run out.`;
    if (!window.confirm(confirmText)) return;
    setBusy(true);
    try {
      const next = await controlPlaneRequest<TenantCreditOverview>(`tenants/${tenantId}/credits/overage`, {
        method: 'PUT', body: JSON.stringify({ overage_enabled: enabled, overage_cap_credits: trimmedCap }),
      });
      setEnabled(next.settings.overage_enabled); setCap(next.settings.overage_cap_credits ?? '');
      onSaved(next);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Saving overage failed'); }
    finally { setBusy(false); }
  }

  return (
    <Card>
      <h2 className="text-lg font-semibold">Overage</h2>
      <p className="mt-1 text-sm text-zinc-400">When the plan credits and packs run out, requests are rejected, or with overage on they continue as debt up to the cap. Only support can change this.</p>
      {error ? <p role="alert" className="mt-3 text-sm text-red-300">{error}</p> : null}
      {!isCreditTenant ? <p className="mt-3 text-sm text-zinc-500">The tenant is not billed in credits yet.</p> : (
        <form className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-end" onSubmit={submit}>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={enabled} disabled={busy || (!settings.overage_allowed && !settings.overage_enabled)} onChange={(e) => setEnabled(e.target.checked)} />
            Allow overage
          </label>
          <label className="flex flex-col gap-1 text-xs text-zinc-400 sm:w-64">
            Cap in credits (empty for no cap)
            <Input inputMode="decimal" placeholder="e.g. 500" value={cap} disabled={busy || !enabled} onChange={(e) => setCap(e.target.value)} />
          </label>
          <Button type="submit" disabled={busy || !dirty}>{busy ? 'Saving…' : 'Save overage'}</Button>
        </form>
      )}
      {isCreditTenant && !settings.overage_allowed ? <p className="mt-2 text-xs text-zinc-500">The tenant plan does not allow overage.</p> : null}
    </Card>
  );
}
