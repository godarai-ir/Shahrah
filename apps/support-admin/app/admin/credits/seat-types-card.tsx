'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { formatToman } from '@/lib/support-credits';
import { controlPlaneRequest } from '@/lib/support-tenants';

type SeatTypeOffer = {
  type: 'free' | 'pro' | 'max';
  monthly_credits: number;
  lifetime: boolean;
  pool_access: boolean;
  discount_percent: string;
  yearly_discount_percent: string;
  monthly_price_irr: number;
  yearly_price_irr: number;
  suggested_monthly_price_irr: number;
  suggested_yearly_price_irr: number;
  custom_monthly_price_irr?: number;
  custom_yearly_price_irr?: number;
  active: boolean;
};

type Draft = { credits: string; discount: string; yearly: string; monthlyPrice: string; yearlyPrice: string; active: boolean };

const IRR_PER_TOMAN = 10;
const tomanText = (irr?: number) => (irr ? String(irr / IRR_PER_TOMAN) : '');

/** Parses a price field in toman: empty clears the set price (0), otherwise a positive whole number. */
function priceIRR(raw: string): number | undefined {
  if (!raw.trim()) return 0;
  const toman = Number(raw.replaceAll(',', ''));
  return Number.isSafeInteger(toman) && toman > 0 ? toman * IRR_PER_TOMAN : undefined;
}

/**
 * Member seat types: the personal credits each seat gives and what buyers pay for it. Support sets
 * round prices by hand; the price from the credit list price and discounts is only a suggestion, used
 * when no price is set. A yearly price left empty is the monthly price less the yearly discount.
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
      yearly: String(Number(item.yearly_discount_percent)), monthlyPrice: tomanText(item.custom_monthly_price_irr),
      yearlyPrice: tomanText(item.custom_yearly_price_irr), active: item.active,
    }])));
  }

  useEffect(() => {
    controlPlaneRequest<{ data: SeatTypeOffer[] }>('seat-types')
      .then(apply)
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : 'Failed to load seat types'));
  }, []);

  function update(type: string, patch: Partial<Draft>) {
    setDrafts({ ...drafts, [type]: { ...drafts[type], ...patch } });
  }

  async function save(item: SeatTypeOffer) {
    const draft = drafts[item.type];
    const credits = Number(draft.credits);
    const discount = Number(draft.discount);
    const yearly = Number(draft.yearly);
    if (!Number.isSafeInteger(credits) || credits <= 0) { setError('Credits must be a positive whole number.'); return; }
    for (const value of [discount, yearly]) {
      if (!Number.isInteger(value) || value < 0 || value >= 100) { setError('Discounts must be whole percentages below 100.'); return; }
    }
    const body: Record<string, unknown> = { monthly_credits: credits, discount_percent: String(discount), yearly_discount_percent: String(yearly), active: draft.active };
    let summary = `${credits.toLocaleString()} credits`;
    if (!item.lifetime) {
      const monthlyPrice = priceIRR(draft.monthlyPrice);
      const yearlyPrice = priceIRR(draft.yearlyPrice);
      if (monthlyPrice === undefined || yearlyPrice === undefined) { setError('Prices must be positive whole numbers of toman, or empty for the suggested price.'); return; }
      body.monthly_price_irr = monthlyPrice;
      body.yearly_price_irr = yearlyPrice;
      summary += `, ${monthlyPrice ? formatToman(monthlyPrice / IRR_PER_TOMAN) : 'suggested price'} a month`;
      summary += `, ${yearlyPrice ? formatToman(yearlyPrice / IRR_PER_TOMAN) : 'derived price'} a year`;
    }
    if (!window.confirm(`Change the ${item.type} seat to ${summary}? New orders and grants use it at once.`)) return;
    setError(''); setMessage('');
    try {
      apply(await controlPlaneRequest<{ data: SeatTypeOffer[] }>(`seat-types/${item.type}`, { method: 'PUT', body: JSON.stringify(body) }));
      setMessage(`The ${item.type} seat was updated.`);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Update failed'); }
  }

  return (
    <Card>
      <h2 className="text-lg font-semibold">Member seats</h2>
      <p className="mt-1 text-sm text-zinc-400">
        Personal GodarChat credits per member and what a seat costs. Set round prices in toman; an empty price uses the suggestion from the credit list price and discounts. Buyers pay the price column.
      </p>
      {error ? <p role="alert" className="mt-3 text-sm text-red-300">{error}</p> : null}
      {message ? <p role="status" className="mt-3 text-sm text-emerald-300">{message}</p> : null}
      <div className="mt-4 overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="text-zinc-400">
            <tr><th className="py-2">Seat</th><th>Credits</th><th>Discount %</th><th>Yearly extra %</th><th>Monthly price (toman)</th><th>Yearly price (toman)</th><th>Active</th><th /></tr>
          </thead>
          <tbody className="divide-y divide-zinc-800">
            {types.map((item) => {
              const draft = drafts[item.type];
              return (
                <tr key={item.type} className="align-top">
                  <td className="py-3 capitalize">{item.type}{item.lifetime ? ' (once)' : ''}</td>
                  <td className="py-2"><Input className="w-28" inputMode="numeric" value={draft?.credits ?? ''} onChange={(e) => update(item.type, { credits: e.target.value })} /></td>
                  <td className="py-2"><Input className="w-20" inputMode="numeric" disabled={item.lifetime} value={draft?.discount ?? ''} onChange={(e) => update(item.type, { discount: e.target.value })} /></td>
                  <td className="py-2"><Input className="w-20" inputMode="numeric" disabled={item.lifetime} value={draft?.yearly ?? ''} onChange={(e) => update(item.type, { yearly: e.target.value })} /></td>
                  {(['monthly', 'yearly'] as const).map((interval) => {
                    const field = interval === 'monthly' ? 'monthlyPrice' : 'yearlyPrice';
                    const charged = interval === 'monthly' ? item.monthly_price_irr : item.yearly_price_irr;
                    const suggested = interval === 'monthly' ? item.suggested_monthly_price_irr : item.suggested_yearly_price_irr;
                    return (
                      <td key={interval} className="py-2">
                        {item.lifetime ? <span className="text-zinc-500">—</span> : (
                          <>
                            <Input className="w-36" inputMode="numeric" aria-label={`${item.type} ${interval} price in toman`} placeholder={tomanText(suggested)}
                              value={draft?.[field] ?? ''} onChange={(e) => update(item.type, { [field]: e.target.value })} />
                            <p className="mt-1 text-xs text-zinc-400">Pays {formatToman(charged / IRR_PER_TOMAN)}</p>
                            <p className="text-xs text-zinc-500">Suggested {formatToman(suggested / IRR_PER_TOMAN)}</p>
                          </>
                        )}
                      </td>
                    );
                  })}
                  <td className="py-3"><input type="checkbox" aria-label={`${item.type} active`} checked={draft?.active ?? false} onChange={(e) => update(item.type, { active: e.target.checked })} /></td>
                  <td className="py-2"><Button onClick={() => void save(item)}>Save</Button></td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
