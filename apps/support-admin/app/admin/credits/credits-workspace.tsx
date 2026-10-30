'use client';

import { useEffect, useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  formatToman, type CreditPackVersion, type CreditPriceRevision, type CreditPriceStatus, type CreditRateCardVersion,
} from '@/lib/support-credits';
import { controlPlaneRequest, formatDate } from '@/lib/support-tenants';
import { SeatTypesCard } from './seat-types-card';

/** Credit price policy, rate-card multipliers, and the add-on pack catalog. */
export function CreditsWorkspace() {
  const [price, setPrice] = useState<CreditPriceStatus>();
  const [rateCards, setRateCards] = useState<CreditRateCardVersion[]>([]);
  const [packs, setPacks] = useState<CreditPackVersion[]>([]);
  const [multiplier, setMultiplier] = useState({ model: '', multiplier: '', reason: '' });
  const [changelog, setChangelog] = useState({ fa: '', en: '' });
  const [pack, setPack] = useState({ slug: '', credits: '', discount: '0', months: '12', sort: '100' });
  const [listPrice, setListPriceForm] = useState({ price: '', reason: '' });
  const [decisionReason, setDecisionReason] = useState('');
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  async function load() {
    setError('');
    try {
      const [nextPrice, cards, packList] = await Promise.all([
        controlPlaneRequest<CreditPriceStatus>('credits/price-policy'),
        controlPlaneRequest<{ data: CreditRateCardVersion[] }>('credits/rate-cards'),
        controlPlaneRequest<{ data: CreditPackVersion[] }>('credits/packs'),
      ]);
      setPrice(nextPrice); setRateCards(cards.data ?? []); setPacks(packList.data ?? []);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Failed to load credit settings'); }
  }

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, []);
  /* eslint-enable react-hooks/set-state-in-effect */

  async function run(action: () => Promise<unknown>, done: string) {
    setError(''); setMessage('');
    try { await action(); setMessage(done); await load(); } catch (cause) { setError(cause instanceof Error ? cause.message : 'Request failed'); }
  }

  function decide(revision: CreditPriceRevision, approve: boolean) {
    const reason = decisionReason.trim();
    if (!reason) { setError('A reason is required to approve or reject the proposal.'); return; }
    if (!window.confirm(approve
      ? `Approve ${formatToman(revision.price_toman_per_credit)} per credit? Add-on pack prices change immediately; plan prices publish after the 30-day notice.`
      : `Reject ${formatToman(revision.price_toman_per_credit)} per credit?`)) return;
    void run(() => controlPlaneRequest(`credits/price-revisions/${revision.id}/${approve ? 'approve' : 'reject'}`, { method: 'POST', body: JSON.stringify({ reason }) })
      .then(() => setDecisionReason('')), approve ? 'Price approved; plan versions are scheduled.' : 'Proposal rejected.');
  }

  function setListPrice(event: FormEvent) {
    event.preventDefault();
    const priceToman = Number(listPrice.price.replaceAll(',', ''));
    const reason = listPrice.reason.trim();
    if (!Number.isSafeInteger(priceToman) || priceToman <= 0) { setError('The list price must be a positive whole number of toman.'); return; }
    if (!reason) { setError('A reason is required to set the list price.'); return; }
    if (!window.confirm(`Set ${formatToman(priceToman)} per credit? This replaces any open proposal. Add-on pack prices and seat prices without a set price change immediately; plan prices publish after the 30-day notice.`)) return;
    void run(() => controlPlaneRequest('credits/price-revisions', { method: 'POST', body: JSON.stringify({ price_toman_per_credit: priceToman, reason }) })
      .then(() => setListPriceForm({ price: '', reason: '' })), 'List price set; plan versions are scheduled.');
  }

  const draft = rateCards.find((card) => card.status === 'draft');

  function saveMultiplier(event: FormEvent) {
    event.preventDefault();
    if (!draft) return;
    if (!multiplier.model.trim() || !multiplier.multiplier.trim() || !multiplier.reason.trim()) { setError('Model, multiplier, and reason are required.'); return; }
    void run(() => controlPlaneRequest(`credits/rate-cards/${draft.version}/multipliers`, { method: 'PUT', body: JSON.stringify({
      model: multiplier.model.trim(), multiplier: multiplier.multiplier.trim(), reason: multiplier.reason.trim(),
    }) }).then(() => setMultiplier({ model: '', multiplier: '', reason: '' })), 'Multiplier saved on the draft.');
  }

  function publishDraft(event: FormEvent) {
    event.preventDefault();
    if (!draft) return;
    if (!changelog.fa.trim()) { setError('A Persian changelog entry is required; it is shown publicly.'); return; }
    if (!window.confirm(`Publish rate card v${draft.version}? If any model gets more expensive it is scheduled 30 days ahead instead.`)) return;
    void run(() => controlPlaneRequest(`credits/rate-cards/${draft.version}/publish`, { method: 'POST', body: JSON.stringify({
      changelog_fa: changelog.fa.trim(), changelog_en: changelog.en.trim(),
    }) }).then(() => setChangelog({ fa: '', en: '' })), 'Rate card published or scheduled.');
  }

  function createPack(event: FormEvent) {
    event.preventDefault();
    const credits = Number(pack.credits);
    if (!pack.slug.trim() || !Number.isSafeInteger(credits) || credits <= 0) { setError('Slug and a positive whole number of credits are required.'); return; }
    void run(() => controlPlaneRequest('credits/packs', { method: 'POST', body: JSON.stringify({
      slug: pack.slug.trim(), credits, discount_percent: pack.discount.trim(), validity_months: Number(pack.months), sort_order: Number(pack.sort),
    }) }).then(() => setPack({ slug: '', credits: '', discount: '0', months: '12', sort: '100' })), 'Pack draft created.');
  }

  function publishPack(item: CreditPackVersion) {
    if (!window.confirm(`Publish ${item.slug} v${item.version}? New orders use it immediately; the current version is archived.`)) return;
    void run(() => controlPlaneRequest(`credits/packs/${encodeURIComponent(item.slug)}/versions/${item.version}/publish`, { method: 'POST', body: '{}' }), 'Pack published.');
  }

  return (
    <div className="flex flex-col gap-6">
      <div><h1 className="text-2xl font-semibold text-zinc-50">Credits</h1><p className="text-sm text-zinc-400">List price policy, rate-card multipliers, add-on packs, and member seats. Every change is audited.</p></div>
      {/* Sticky so feedback from forms further down the page stays in view. */}
      {error ? <p role="alert" className="sticky top-0 z-10 rounded-md border border-red-900 bg-red-950/40 p-3 text-sm text-red-200">{error}</p> : null}
      {message ? <p role="status" className="sticky top-0 z-10 rounded-md border border-emerald-900 bg-emerald-950/40 p-3 text-sm text-emerald-200">{message}</p> : null}

      <section className="grid gap-4 lg:grid-cols-3">
        <Card><p className="text-sm text-zinc-400">Active list price</p><p className="mt-2 text-xl font-semibold">{price?.active ? formatToman(price.active.price_toman_per_credit) : '—'}</p><p className="mt-1 text-xs text-zinc-500">Base {price?.active ? formatToman(price.active.base_usd_toman) : '—'} per USD · ±{price?.active?.band_pct ?? '10'}% band</p><form className="mt-3 flex flex-col gap-2" onSubmit={setListPrice}><Input inputMode="numeric" aria-label="New list price per credit, in toman" placeholder="New price per credit (toman)" value={listPrice.price} onChange={(e) => setListPriceForm({ ...listPrice, price: e.target.value })} /><Input aria-label="Reason for the new list price" placeholder="Reason (required)" value={listPrice.reason} onChange={(e) => setListPriceForm({ ...listPrice, reason: e.target.value })} /><Button type="submit">Set price</Button></form></Card>
        <Card><p className="text-sm text-zinc-400">7-day average</p><p className="mt-2 text-xl font-semibold">{price ? (price.seven_day_average_usd_toman ? formatToman(price.seven_day_average_usd_toman) : 'Not enough samples') : '—'}</p><p className="mt-1 text-xs text-zinc-500">Needs one sample on each of the last 7 days</p></Card>
        <Card><p className="text-sm text-zinc-400">Proposal</p>{price?.proposed ? <><p className="mt-2 text-xl font-semibold">{formatToman(price.proposed.price_toman_per_credit)}</p><p className="mt-1 text-xs text-zinc-500">From {formatToman(price.proposed.previous_price_toman ?? 0)} · average {formatToman(price.proposed.observed_avg_usd_toman ?? price.proposed.base_usd_toman)}</p><Input className="mt-3" aria-label="Reason for the proposal decision" placeholder="Reason (required)" value={decisionReason} onChange={(e) => setDecisionReason(e.target.value)} /><div className="mt-2 flex gap-2"><Button onClick={() => decide(price.proposed as CreditPriceRevision, true)}>Approve</Button><Button className="bg-zinc-800 text-zinc-100 hover:bg-zinc-700" onClick={() => decide(price.proposed as CreditPriceRevision, false)}>Reject</Button></div></> : <p className="mt-2 text-sm text-zinc-400">No open proposal.</p>}</Card>
      </section>

      <Card><h2 className="text-lg font-semibold">Price history</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Price</th><th>Status</th><th>Add-ons from</th><th>Plans from</th><th>Decision</th></tr></thead><tbody className="divide-y divide-zinc-800">{(price?.revisions ?? []).map((rev) => <tr key={rev.id}><td className="py-3">{formatToman(rev.price_toman_per_credit)}</td><td>{rev.status}</td><td className="text-zinc-400">{formatDate(rev.effective_from)}</td><td className="text-zinc-400">{formatDate(rev.plan_effective_from)}</td><td className="text-zinc-400">{[rev.decision_reason, rev.decided_by].filter(Boolean).join(' · ') || '—'}</td></tr>)}</tbody></table>{price && (price.revisions ?? []).length === 0 ? <p className="py-6 text-sm text-zinc-400">No price revisions yet.</p> : null}</div></Card>

      <section className="grid gap-4 lg:grid-cols-[1fr_380px]">
        <Card><div className="flex flex-wrap items-center justify-between gap-2"><h2 className="text-lg font-semibold">Rate cards</h2>{draft ? null : <Button onClick={() => void run(() => controlPlaneRequest('credits/rate-cards', { method: 'POST', body: '{}' }), 'Draft created from the published card.')}>Start draft</Button>}</div><div className="mt-4 flex flex-col gap-4">{rateCards.map((card) => <div key={card.version} className="rounded-md border border-zinc-800 p-3"><div className="flex flex-wrap justify-between gap-2 text-sm"><span className="font-medium">v{card.version} · {card.status}</span><span className="text-zinc-400">{card.status === 'draft' ? 'Not published' : `Effective ${formatDate(card.effective_from)}`}</span></div>{card.changelog_fa ? <p className="mt-1 text-sm text-zinc-300" dir="rtl">{card.changelog_fa}</p> : null}<ul className="mt-2 flex flex-col gap-1 text-xs text-zinc-400">{card.multipliers.map((item) => <li key={item.model}><span className="text-zinc-200">{item.model}</span> ×{item.multiplier} — {item.reason}</li>)}{card.multipliers.length === 0 ? <li>All models at the provider list price.</li> : null}</ul></div>)}</div></Card>
        {draft ? <div className="flex flex-col gap-4"><Card><h2 className="text-lg font-semibold">Draft v{draft.version}: multiplier</h2><form className="mt-4 flex flex-col gap-3" onSubmit={saveMultiplier}><Input aria-label="Model" placeholder="Model, provider/model" value={multiplier.model} onChange={(e) => setMultiplier({ ...multiplier, model: e.target.value })} /><Input inputMode="decimal" aria-label="Multiplier" placeholder="Multiplier, e.g. 1.15" value={multiplier.multiplier} onChange={(e) => setMultiplier({ ...multiplier, multiplier: e.target.value })} /><Input aria-label="Reason" placeholder="Reason (required): real cost difference" value={multiplier.reason} onChange={(e) => setMultiplier({ ...multiplier, reason: e.target.value })} /><Button type="submit">Save multiplier</Button></form></Card><Card><h2 className="text-lg font-semibold">Publish draft v{draft.version}</h2><form className="mt-4 flex flex-col gap-3" onSubmit={publishDraft}><Input dir="rtl" aria-label="Public changelog in Persian" placeholder="Public changelog (fa, required)" value={changelog.fa} onChange={(e) => setChangelog({ ...changelog, fa: e.target.value })} /><Input aria-label="Public changelog in English" placeholder="Public changelog (en, optional)" value={changelog.en} onChange={(e) => setChangelog({ ...changelog, en: e.target.value })} /><Button type="submit">Publish</Button></form></Card></div> : null}
      </section>

      <section className="grid gap-4 lg:grid-cols-[1fr_380px]">
        <Card><h2 className="text-lg font-semibold">Add-on packs</h2><div className="mt-4 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="text-zinc-400"><tr><th className="py-2">Pack</th><th>Credits</th><th>Discount</th><th>Validity</th><th>Status</th><th /></tr></thead><tbody className="divide-y divide-zinc-800">{packs.map((item) => <tr key={`${item.slug}-${item.version}`}><td className="py-3">{item.slug} v{item.version}</td><td>{item.credits.toLocaleString()}</td><td>{item.discount_percent}%</td><td>{item.validity_months} months</td><td>{item.status}</td><td>{item.status === 'draft' ? <Button onClick={() => publishPack(item)}>Publish</Button> : null}</td></tr>)}</tbody></table>{packs.length === 0 ? <p className="py-6 text-sm text-zinc-400">No add-on packs yet. Create a draft to start.</p> : null}</div><p className="mt-3 text-xs text-zinc-500">Prices follow the active list price less the pack discount; orders keep the price they were placed at.</p></Card>
        <Card><h2 className="text-lg font-semibold">New pack version</h2><form className="mt-4 flex flex-col gap-3" onSubmit={createPack}><Input aria-label="Pack slug" placeholder="Slug, e.g. credits-5000" value={pack.slug} onChange={(e) => setPack({ ...pack, slug: e.target.value })} /><Input inputMode="numeric" aria-label="Pack credits" placeholder="Credits" value={pack.credits} onChange={(e) => setPack({ ...pack, credits: e.target.value })} /><label className="text-xs text-zinc-400">Discount off the list price (%)<Input inputMode="decimal" value={pack.discount} onChange={(e) => setPack({ ...pack, discount: e.target.value })} /></label><label className="text-xs text-zinc-400">Validity (months)<Input inputMode="numeric" value={pack.months} onChange={(e) => setPack({ ...pack, months: e.target.value })} /></label><label className="text-xs text-zinc-400">Sort order<Input inputMode="numeric" value={pack.sort} onChange={(e) => setPack({ ...pack, sort: e.target.value })} /></label><Button type="submit">Create draft</Button></form></Card>
      </section>

      <SeatTypesCard />
    </div>
  );
}
