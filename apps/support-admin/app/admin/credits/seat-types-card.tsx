'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { controlPlaneRequest, formatIRR } from '@/lib/support-tenants';

type SeatTypeOffer = {
  type: 'free' | 'pro' | 'max';
  monthly_credits: number;
  lifetime: boolean;
  pool_access: boolean;
  discount_percent: string;
  yearly_discount_percent: string;
  monthly_price_irr: number;
  yearly_price_irr: number;
  active: boolean;
};

type Draft = { credits: string; discount: string; yearly: string; active: boolean };

/**
 * Member seat types: the personal credits each seat gives and its discount off the credit list
 * price. Prices follow the active list price, so a quota or discount change reprices new orders.
 */
export function SeatTypesCard() {
  const [types, setTypes] = useState<SeatTypeOffer[]>([]);
  const [drafts, setDrafts] = useState<Record<string, Draft>>({});
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  function apply(payload: { data: SeatTypeOffer[] }) {
    setTypes(payload.data ?? []);
    setDrafts(Object.fromEntries((payload.data ?? []).map((item) => [item.type, {
      credits: String(item.monthly_credits), discount: String(Number(item.discount_percent)),
      yearly: String(Number(item.yearly_discount_percent)), active: item.active,
    }])));
  }

  useEffect(() => {
    controlPlaneRequest<{ data: SeatTypeOffer[] }>('seat-types')
      .then(apply)
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : 'Failed to load seat types'));
  }, []);

  async function save(type: string) {
    const draft = drafts[type];
    const credits = Number(draft.credits);
    const discount = Number(draft.discount);
    const yearly = Number(draft.yearly);
    if (!Number.isSafeInteger(credits) || credits <= 0) { setError('Credits must be a positive whole number.'); return; }
    for (const value of [discount, yearly]) {
      if (!Number.isInteger(value) || value < 0 || value >= 100) { setError('Discounts must be whole percentages below 100.'); return; }
    }
    if (!window.confirm(`Change the ${type} seat to ${credits.toLocaleString()} credits at ${discount}% off (${yearly}% more yearly)? New orders and grants use it at once.`)) return;
    setError(''); setMessage('');
    try {
      apply(await controlPlaneRequest<{ data: SeatTypeOffer[] }>(`seat-types/${type}`, {
        method: 'PUT', body: JSON.stringify({ monthly_credits: credits, discount_percent: String(discount), yearly_discount_percent: String(yearly), active: draft.active }),
      }));
      setMessage(`The ${type} seat was updated.`);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Update failed'); }
  }

  return (
    <Card>
      <h2 className="text-lg font-semibold">Member seats</h2>
      <p className="mt-1 text-sm text-zinc-400">Personal GodarChat credits per member. The free seat is granted once per member; Pro and Max renew monthly.</p>
      {error ? <p role="alert" className="mt-3 text-sm text-red-300">{error}</p> : null}
      {message ? <p role="status" className="mt-3 text-sm text-emerald-300">{message}</p> : null}
      <div className="mt-4 overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="text-zinc-400"><tr><th className="py-2">Seat</th><th>Credits</th><th>Discount %</th><th>Yearly extra %</th><th>Monthly price</th><th>Yearly price</th><th>Active</th><th /></tr></thead>
          <tbody className="divide-y divide-zinc-800">
            {types.map((item) => {
              const draft = drafts[item.type];
              return (
                <tr key={item.type}>
                  <td className="py-3 capitalize">{item.type}{item.lifetime ? ' (once)' : ''}</td>
                  <td><Input className="w-28" inputMode="numeric" value={draft?.credits ?? ''} onChange={(e) => setDrafts({ ...drafts, [item.type]: { ...draft, credits: e.target.value } })} /></td>
                  <td><Input className="w-20" inputMode="numeric" disabled={item.lifetime} value={draft?.discount ?? ''} onChange={(e) => setDrafts({ ...drafts, [item.type]: { ...draft, discount: e.target.value } })} /></td>
                  <td><Input className="w-20" inputMode="numeric" disabled={item.lifetime} value={draft?.yearly ?? ''} onChange={(e) => setDrafts({ ...drafts, [item.type]: { ...draft, yearly: e.target.value } })} /></td>
                  <td className="text-zinc-300">{item.lifetime ? '—' : formatIRR(item.monthly_price_irr)}</td>
                  <td className="text-zinc-300">{item.lifetime ? '—' : formatIRR(item.yearly_price_irr)}</td>
                  <td><input type="checkbox" aria-label={`${item.type} active`} checked={draft?.active ?? false} onChange={(e) => setDrafts({ ...drafts, [item.type]: { ...draft, active: e.target.checked } })} /></td>
                  <td><Button onClick={() => void save(item.type)}>Save</Button></td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
