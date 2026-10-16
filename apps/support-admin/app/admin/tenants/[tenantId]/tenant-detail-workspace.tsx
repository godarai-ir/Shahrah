'use client';

import Link from 'next/link';
import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { TenantCreditsPanel } from './tenant-credits-panel';
import {
  controlPlaneRequest, formatDate, formatIRR, type SubscriptionPaymentIntent, type TenantDetail, type TenantEvent, type WalletTransaction,
} from '@/lib/support-tenants';

const tabs = ['overview', 'subscription', 'wallet', 'credits', 'events', 'controls'] as const;
type Tab = typeof tabs[number];

export function TenantDetailWorkspace({ tenantId }: { tenantId: string }) {
  const [detail, setDetail] = useState<TenantDetail>();
  const [events, setEvents] = useState<TenantEvent[]>([]);
  const [transactions, setTransactions] = useState<WalletTransaction[]>([]);
  const [payments, setPayments] = useState<SubscriptionPaymentIntent[]>([]);
  const [tab, setTab] = useState<Tab>('overview');
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [adjustment, setAdjustment] = useState({ direction: 'credit', amount: '', reason: '', externalReference: '' });
  const [planChange, setPlanChange] = useState({ plan: '', interval: 'monthly', reason: '' });

  async function load() {
    setLoading(true); setError('');
    try {
      const [nextDetail, eventPayload, transactionPayload, paymentPayload] = await Promise.all([
        controlPlaneRequest<TenantDetail>(`tenants/${tenantId}`),
        controlPlaneRequest<{ data: TenantEvent[] }>(`tenants/${tenantId}/events?limit=50`),
        controlPlaneRequest<{ data: WalletTransaction[] }>(`tenants/${tenantId}/wallet/transactions?limit=50`),
        controlPlaneRequest<{ data: SubscriptionPaymentIntent[] }>(`tenants/${tenantId}/billing/payment-intents?limit=50`),
      ]);
      setDetail(nextDetail); setEvents(eventPayload.data ?? []); setTransactions(transactionPayload.data ?? []); setPayments(paymentPayload.data ?? []);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load tenant');
    } finally { setLoading(false); }
  }

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */
  useEffect(() => { void load(); }, [tenantId]);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  async function submitAdjustment(event: FormEvent) {
    event.preventDefault(); setError(''); setMessage('');
    const amount = Number(adjustment.amount);
    if (!Number.isSafeInteger(amount) || amount <= 0) { setError('Amount must be a positive whole-number IRR value.'); return; }
    if (!adjustment.reason.trim()) { setError('A reason is required.'); return; }
    if (!window.confirm(`${adjustment.direction === 'credit' ? 'Credit' : 'Debit'} ${formatIRR(amount)} for ${detail?.tenant.slug}?`)) return;
    try {
      await controlPlaneRequest(`tenants/${tenantId}/wallet/adjustments`, { method: 'POST', body: JSON.stringify({
        direction: adjustment.direction, amount_minor: amount, reason: adjustment.reason.trim(),
        external_reference: adjustment.externalReference.trim(), idempotency_key: crypto.randomUUID(),
      }) });
      setAdjustment({ direction: 'credit', amount: '', reason: '', externalReference: '' });
      setMessage('Wallet adjustment completed and audited.'); await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Adjustment failed'); }
  }

  async function changeStatus(target: 'suspend' | 'reactivate') {
    const reason = window.prompt(`Reason to ${target} ${detail?.tenant.slug}:`)?.trim();
    if (!reason) return;
    if (!window.confirm(`Confirm ${target} for tenant ${detail?.tenant.slug}.`)) return;
    try {
      await controlPlaneRequest(`tenants/${tenantId}/${target}`, { method: 'POST', body: JSON.stringify({ reason }) });
      setMessage(`Tenant ${target === 'suspend' ? 'suspended' : 'reactivated'}.`); setError(''); await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Status change failed'); }
  }

  async function changePlan(event: FormEvent) {
    event.preventDefault();
    if (!planChange.plan.trim() || !planChange.reason.trim()) { setError('Plan slug and reason are required.'); return; }
    if (!window.confirm(`Immediately assign ${planChange.plan} (${planChange.interval}) to ${detail?.tenant.slug}? No proration will be applied.`)) return;
    try {
      await controlPlaneRequest('tenant-subscriptions', { method: 'PUT', body: JSON.stringify({
        tenant_id: tenantId, plan_slug: planChange.plan.trim(), billing_interval: planChange.interval,
        status: 'active', metadata: { support_reason: planChange.reason.trim() },
      }) });
      setMessage('Subscription updated.'); setError(''); await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Subscription update failed'); }
  }

  async function setCancellation(cancel: boolean) {
    const reason = window.prompt(`${cancel ? 'Schedule cancellation' : 'Reverse cancellation'} reason:`)?.trim();
    if (!reason || !window.confirm(`${cancel ? 'Schedule cancellation at period end' : 'Keep this subscription active'} for ${detail?.tenant.slug}?`)) return;
    try {
      await controlPlaneRequest(`tenants/${tenantId}/subscription/cancel-at-period-end`, { method: cancel ? 'POST' : 'DELETE', body: JSON.stringify({ reason }) });
      setMessage(cancel ? 'Cancellation scheduled.' : 'Scheduled cancellation reversed.'); setError(''); await load();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Cancellation update failed'); }
  }

  if (loading && !detail) return <p className="text-sm text-zinc-400">Loading tenant workspace…</p>;
  if (!detail) return <p role="alert" className="text-sm text-red-300">{error || 'Tenant unavailable'}</p>;
  const tenant = detail.tenant;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div><Link href="/admin/tenants" className="text-sm text-cyan-300 hover:underline">← Tenants</Link><h1 className="mt-2 text-2xl font-semibold text-zinc-50">{tenant.slug}</h1><p className="text-sm text-zinc-400">{tenant.domain} · {tenant.id}</p>{detail.owner ? <p className="mt-2 text-sm text-zinc-300">Owner: {detail.owner.name || detail.owner.email || detail.owner.phone || detail.owner.user_id}{detail.owner.email && detail.owner.name ? ` · ${detail.owner.email}` : ''}{detail.owner.phone ? ` · ${detail.owner.phone}` : ''}</p> : <p className="mt-2 text-sm text-zinc-500">Owner contact unavailable</p>}</div>
        <div className="flex gap-2"><span className="rounded-full border border-zinc-700 px-3 py-1 text-sm capitalize">{tenant.status}</span><span className="rounded-full border border-zinc-700 px-3 py-1 text-sm">{tenant.plan_display_name || 'No plan'}</span></div>
      </div>
      {error ? <p role="alert" className="rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}
      {message ? <p role="status" className="rounded-md border border-emerald-900 bg-emerald-950/40 p-3 text-sm text-emerald-200">{message}</p> : null}
      <nav aria-label="Tenant sections" className="flex flex-wrap gap-2 border-b border-zinc-800 pb-3">
        {tabs.map((item) => <button key={item} className={`rounded-md px-3 py-2 text-sm capitalize ${tab === item ? 'bg-cyan-500 text-black' : 'bg-zinc-900 text-zinc-300'}`} onClick={() => setTab(item)}>{item === 'wallet' ? 'Wallet & billing' : item}</button>)}
      </nav>

      {tab === 'overview' ? <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Metric label="Tenant owner" value={detail.owner?.name || 'Not set'} detail={[detail.owner?.email, detail.owner?.phone].filter(Boolean).join(' · ') || 'No contact information'} />
        <Metric label="Wallet balance" value={formatIRR(tenant.wallet_balance_minor)} detail={`${formatIRR(tenant.wallet_pending_minor)} pending`} />
        <Metric label="Subscription" value={tenant.plan_display_name || 'No plan'} detail={tenant.subscription_status || 'Unavailable'} />
        <Metric label="Period end" value={formatDate(tenant.period_end)} detail={`Created ${formatDate(tenant.created_at)}`} />
      </div> : null}

      {tab === 'subscription' ? <div className="grid gap-4 lg:grid-cols-2">
        <Card><h2 className="text-lg font-semibold">Current subscription</h2><dl className="mt-4 grid grid-cols-2 gap-3 text-sm"><Term label="Plan" value={detail.subscription?.plan_display_name || 'Unavailable'} /><Term label="Status" value={detail.subscription?.status || 'Unavailable'} /><Term label="Interval" value={detail.subscription?.billing_interval || 'Unavailable'} /><Term label="Period" value={`${formatDate(detail.subscription?.current_period_start)} — ${formatDate(detail.subscription?.current_period_end)}`} /><Term label="Cancellation" value={detail.subscription?.cancel_at_period_end ? 'Scheduled' : 'Not scheduled'} /></dl>{detail.subscription ? <div className="mt-5">{detail.subscription.cancel_at_period_end ? <Button onClick={() => void setCancellation(false)}>Reverse cancellation</Button> : <Button className="bg-red-500 text-white hover:bg-red-400" onClick={() => void setCancellation(true)}>Cancel at period end</Button>}</div> : null}</Card>
        <Card><h2 className="text-lg font-semibold">Assign plan</h2><p className="mt-1 text-sm text-zinc-400">Immediate, non-prorated operator assignment.</p><form className="mt-4 flex flex-col gap-3" onSubmit={changePlan}><Input placeholder="Plan slug" value={planChange.plan} onChange={(e) => setPlanChange({ ...planChange, plan: e.target.value })} /><select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={planChange.interval} onChange={(e) => setPlanChange({ ...planChange, interval: e.target.value })}><option value="monthly">Monthly</option><option value="quarterly">Quarterly</option><option value="yearly">Yearly</option></select><Input placeholder="Required reason" value={planChange.reason} onChange={(e) => setPlanChange({ ...planChange, reason: e.target.value })} /><Button type="submit">Assign plan</Button></form></Card>
      </div> : null}

      {tab === 'wallet' ? <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <Card><h2 className="text-lg font-semibold">Manual adjustment</h2><form className="mt-4 flex flex-col gap-3" onSubmit={submitAdjustment}><select className="rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm" value={adjustment.direction} onChange={(e) => setAdjustment({ ...adjustment, direction: e.target.value })}><option value="credit">Credit</option><option value="debit">Debit</option></select><Input inputMode="numeric" placeholder="Whole IRR amount" value={adjustment.amount} onChange={(e) => setAdjustment({ ...adjustment, amount: e.target.value })} /><Input placeholder="Required reason" value={adjustment.reason} onChange={(e) => setAdjustment({ ...adjustment, reason: e.target.value })} /><Input placeholder="Ticket/reference (optional)" value={adjustment.externalReference} onChange={(e) => setAdjustment({ ...adjustment, externalReference: e.target.value })} /><Button type="submit">Review adjustment</Button></form></Card>
        <div className="flex flex-col gap-4"><Card><h2 className="text-lg font-semibold">Recent ledger</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Type</th><th>Amount</th><th>Status</th><th>Time</th></tr></thead><tbody className="divide-y divide-zinc-800">{transactions.map((tx) => <tr key={tx.id}><td className="py-3">{tx.type}</td><td>{formatIRR(tx.amount_irr)}</td><td>{tx.status}</td><td className="text-zinc-400">{formatDate(tx.created_at)}</td></tr>)}</tbody></table>{transactions.length === 0 ? <p className="py-6 text-sm text-zinc-400">No wallet transactions.</p> : null}</div></Card><Card><h2 className="text-lg font-semibold">Subscription payments</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Status</th><th>Amount</th><th>Provider</th><th>Time</th></tr></thead><tbody className="divide-y divide-zinc-800">{payments.map((payment) => <tr key={payment.id}><td className="py-3">{payment.status}</td><td>{formatIRR(payment.amount_minor)}</td><td>{payment.provider}</td><td className="text-zinc-400">{formatDate(payment.created_at)}</td></tr>)}</tbody></table>{payments.length === 0 ? <p className="py-6 text-sm text-zinc-400">No subscription payment intents.</p> : null}</div></Card></div>
      </div> : null}

      {tab === 'credits' ? <TenantCreditsPanel tenantId={tenantId} tenantSlug={tenant.slug} /> : null}

      {tab === 'events' ? <Card><h2 className="text-lg font-semibold">Important events</h2><ol className="mt-4 divide-y divide-zinc-800">{events.map((item) => <li key={item.id} className="py-4"><div className="flex flex-wrap justify-between gap-2"><span className="font-medium">{item.type}</span><time className="text-sm text-zinc-400">{formatDate(item.created_at)}</time></div><p className="mt-1 text-sm text-zinc-300">{item.reason || 'No operator reason'}</p><p className="mt-1 text-xs text-zinc-500">{item.actor_role || 'system'} {item.actor_id ? `· ${item.actor_id}` : ''}</p></li>)}</ol>{events.length === 0 ? <p className="mt-4 text-sm text-zinc-400">No tenant events recorded.</p> : null}</Card> : null}

      {tab === 'controls' ? <Card><h2 className="text-lg font-semibold">Account controls</h2><p className="mt-2 text-sm text-zinc-400">Status changes are audited and require a reason. Suspension blocks new tenant traffic through shared tenant middleware.</p><div className="mt-5">{tenant.status === 'active' ? <Button className="bg-red-500 text-white hover:bg-red-400" onClick={() => void changeStatus('suspend')}>Suspend tenant</Button> : <Button onClick={() => void changeStatus('reactivate')}>Reactivate tenant</Button>}</div></Card> : null}
    </div>
  );
}

function Metric({ label, value, detail }: { label: string; value: string; detail: string }) { return <Card><p className="text-sm text-zinc-400">{label}</p><p className="mt-2 text-xl font-semibold">{value}</p><p className="mt-1 text-xs text-zinc-500">{detail}</p></Card>; }
function Term({ label, value }: { label: string; value: string }) { return <div><dt className="text-zinc-500">{label}</dt><dd className="mt-1 text-zinc-100">{value}</dd></div>; }
