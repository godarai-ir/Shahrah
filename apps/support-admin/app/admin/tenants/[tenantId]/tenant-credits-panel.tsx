'use client';

import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  formatCredits, parseCreditAmount, type CreditLedgerEntry, type CreditLot, type CreditLotAllocation, type TenantCreditOverview,
} from '@/lib/support-credits';
import { controlPlaneRequest, formatDate, formatIRR } from '@/lib/support-tenants';
import { TenantConversionCard } from './tenant-conversion-card';

const LEDGER_PAGE = 50;

/** Credit balance, lots, ledger, and audited corrections for one tenant. */
export function TenantCreditsPanel({ tenantId, tenantSlug }: { tenantId: string; tenantSlug: string }) {
  const [overview, setOverview] = useState<TenantCreditOverview>();
  const [lots, setLots] = useState<CreditLot[]>([]);
  const [entries, setEntries] = useState<CreditLedgerEntry[]>([]);
  const [hasOlder, setHasOlder] = useState(false);
  const [openLot, setOpenLot] = useState<{ lotId: string; allocations: CreditLotAllocation[] }>();
  const [adjustment, setAdjustment] = useState({ amount: '', reason: '', expiresAt: '' });
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  async function load() {
    setError('');
    try {
      const [nextOverview, lotPayload, ledgerPayload] = await Promise.all([
        controlPlaneRequest<TenantCreditOverview>(`tenants/${tenantId}/credits`),
        controlPlaneRequest<{ data: CreditLot[] }>(`tenants/${tenantId}/credits/lots`),
        controlPlaneRequest<{ data: CreditLedgerEntry[] }>(`tenants/${tenantId}/credits/ledger?limit=${LEDGER_PAGE}`),
      ]);
      setOverview(nextOverview); setLots(lotPayload.data ?? []);
      const page = ledgerPayload.data ?? [];
      setEntries(page); setHasOlder(page.length === LEDGER_PAGE);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load credits'); }
  }

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */
  useEffect(() => { void load(); }, [tenantId]);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  async function loadOlder() {
    const oldest = entries.at(-1);
    if (!oldest) return;
    try {
      const payload = await controlPlaneRequest<{ data: CreditLedgerEntry[] }>(
        `tenants/${tenantId}/credits/ledger?limit=${LEDGER_PAGE}&before=${encodeURIComponent(oldest.occurred_at)}`,
      );
      const page = payload.data ?? [];
      setEntries([...entries, ...page]); setHasOlder(page.length === LEDGER_PAGE);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load older entries'); }
  }

  async function toggleLot(lotId: string) {
    if (openLot?.lotId === lotId) { setOpenLot(undefined); return; }
    try {
      const payload = await controlPlaneRequest<{ data: CreditLotAllocation[] }>(`tenants/${tenantId}/credits/lots/${lotId}/allocations`);
      setOpenLot({ lotId, allocations: payload.data ?? [] });
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load lot movements'); }
  }

  async function submitAdjustment(event: FormEvent) {
    event.preventDefault(); setError(''); setMessage('');
    const amount = parseCreditAmount(adjustment.amount);
    if (!amount) { setError('Amount must be a non-zero number of credits with at most two decimals, for example 100 or -12.50.'); return; }
    if (!adjustment.reason.trim()) { setError('A reason is required.'); return; }
    const verb = amount.startsWith('-') ? 'Remove' : 'Grant';
    if (!window.confirm(`${verb} ${formatCredits(amount.replace('-', ''))} for ${tenantSlug}? This is written to the append-only ledger and cannot be edited.`)) return;
    try {
      await controlPlaneRequest(`tenants/${tenantId}/credits/adjustments`, { method: 'POST', body: JSON.stringify({
        amount_credits: amount, reason: adjustment.reason.trim(), idempotency_key: crypto.randomUUID(),
        ...(adjustment.expiresAt && !amount.startsWith('-') ? { expires_at: new Date(adjustment.expiresAt).toISOString() } : {}),
      }) });
      setAdjustment({ amount: '', reason: '', expiresAt: '' });
      setMessage('Credit adjustment written and audited.'); await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Adjustment failed'); }
  }

  if (!overview) return error ? <p role="alert" className="text-sm text-red-300">{error}</p> : <p className="text-sm text-zinc-400">Loading credits…</p>;

  return (
    <div className="flex flex-col gap-4">
      {error ? <p role="alert" className="rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}
      {message ? <p role="status" className="rounded-md border border-emerald-900 bg-emerald-950/40 p-3 text-sm text-emerald-200">{message}</p> : null}

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card><p className="text-sm text-zinc-400">Billing unit</p><p className="mt-2 text-xl font-semibold uppercase">{overview.billing_unit}</p><p className="mt-1 text-xs text-zinc-500">Changed only by a conversion below</p></Card>
        <Card><p className="text-sm text-zinc-400">Available</p><p className="mt-2 text-xl font-semibold">{formatCredits(overview.available_credits)}</p><p className="mt-1 text-xs text-zinc-500">Across unexpired lots</p></Card>
        <Card><p className="text-sm text-zinc-400">Debt</p><p className={`mt-2 text-xl font-semibold ${Number(overview.debt_credits) > 0 ? 'text-amber-300' : ''}`}>{formatCredits(overview.debt_credits)}</p><p className="mt-1 text-xs text-zinc-500">Repaid by the next grant</p></Card>
        <Card><p className="text-sm text-zinc-400">Overage</p><p className="mt-2 text-xl font-semibold">{overview.settings.overage_enabled ? 'Enabled' : 'Off'}</p><p className="mt-1 text-xs text-zinc-500">{overview.settings.overage_allowed ? 'Allowed by plan' : 'Not allowed by plan'}{overview.settings.overage_cap_credits ? ` · cap ${formatCredits(overview.settings.overage_cap_credits)}` : ''}</p></Card>
      </div>

      <TenantConversionCard tenantId={tenantId} tenantSlug={tenantSlug} onChanged={() => void load()} />

      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <Card><h2 className="text-lg font-semibold">Manual credit adjustment</h2><p className="mt-1 text-sm text-zinc-400">Positive amounts create a lot; negative amounts are spent from lots in consumption order.</p><form className="mt-4 flex flex-col gap-3" onSubmit={submitAdjustment}><Input inputMode="decimal" placeholder="Credits, e.g. 100 or -12.50" value={adjustment.amount} onChange={(e) => setAdjustment({ ...adjustment, amount: e.target.value })} /><Input placeholder="Required reason" value={adjustment.reason} onChange={(e) => setAdjustment({ ...adjustment, reason: e.target.value })} /><label className="text-xs text-zinc-400">Expiry for granted credits (optional)<Input type="date" value={adjustment.expiresAt} onChange={(e) => setAdjustment({ ...adjustment, expiresAt: e.target.value })} /></label><Button type="submit">Review adjustment</Button></form></Card>

        <Card><h2 className="text-lg font-semibold">Lots</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Source</th><th>Remaining</th><th>Granted</th><th>Expires</th><th>Paid</th><th /></tr></thead><tbody className="divide-y divide-zinc-800">{lots.map((lot) => <LotRow key={lot.lot_id} lot={lot} open={openLot?.lotId === lot.lot_id} allocations={openLot?.lotId === lot.lot_id ? openLot.allocations : []} onToggle={() => void toggleLot(lot.lot_id)} />)}</tbody></table>{lots.length === 0 ? <p className="py-6 text-sm text-zinc-400">No credit lots.</p> : null}</div></Card>
      </div>

      <Card><h2 className="text-lg font-semibold">Ledger</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Kind</th><th>Amount</th><th>Detail</th><th>Time</th></tr></thead><tbody className="divide-y divide-zinc-800">{entries.map((entry) => <tr key={entry.entry_id}><td className="py-3">{entry.kind}</td><td className={entry.amount_credits.startsWith('-') ? 'text-zinc-300' : 'text-emerald-300'}>{formatCredits(entry.amount_credits)}</td><td className="text-zinc-400">{entryDetail(entry)}</td><td className="text-zinc-400">{formatDate(entry.occurred_at)}</td></tr>)}</tbody></table>{entries.length === 0 ? <p className="py-6 text-sm text-zinc-400">No ledger entries.</p> : null}{hasOlder ? <Button className="mt-4" onClick={() => void loadOlder()}>Load older entries</Button> : null}</div></Card>
    </div>
  );
}

function LotRow({ lot, open, allocations, onToggle }: { lot: CreditLot; open: boolean; allocations: CreditLotAllocation[]; onToggle: () => void }) {
  return (
    <>
      <tr>
        <td className="py-3">{lot.source}</td>
        <td>{formatCredits(lot.remaining_credits)}</td>
        <td className="text-zinc-400">{formatCredits(lot.amount_credits)}</td>
        <td className="text-zinc-400">{lot.expires_at ? formatDate(lot.expires_at) : 'Never'}</td>
        <td className="text-zinc-400">{lot.purchase_price_irr != null ? formatIRR(lot.purchase_price_irr) : '—'}</td>
        <td><button type="button" className="text-cyan-300 hover:underline" aria-expanded={open} onClick={onToggle}>{open ? 'Hide' : 'Movements'}</button></td>
      </tr>
      {open ? <tr><td colSpan={6} className="bg-zinc-950 px-3 py-3"><ul className="flex flex-col gap-1 text-xs text-zinc-300">{allocations.map((item, index) => <li key={`${item.entry_id}-${index}`}>{formatDate(item.occurred_at)} · {item.kind} · {formatCredits(item.amount_credits)}</li>)}{allocations.length === 0 ? <li className="text-zinc-500">No movements.</li> : null}</ul></td></tr> : null}
    </>
  );
}

function entryDetail(entry: CreditLedgerEntry): string {
  if (entry.kind === 'consume') {
    const parts = [entry.product, entry.model, entry.billing_mode === 'external_billed' && entry.credit_equivalent ? `BYOK fee on ${formatCredits(entry.credit_equivalent)}` : ''];
    return parts.filter(Boolean).join(' · ') || entry.request_id || '—';
  }
  return [entry.reason, entry.actor].filter(Boolean).join(' · ') || '—';
}
