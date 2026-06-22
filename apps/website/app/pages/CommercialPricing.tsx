'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Check, Minus } from 'lucide-react';

export type PublicPlanPrice = {
  billing_interval: 'monthly' | 'quarterly' | 'yearly';
  currency: string;
  amount_minor?: number | null;
  discount_percent: number;
};

export type PublicCommercialPlan = {
  slug: string;
  display_name: string;
  description?: string;
  prices?: PublicPlanPrice[];
  features?: Record<string, unknown>;
};

type BillingInterval = PublicPlanPrice['billing_interval'];

export function CommercialPricing({ plans }: { plans: PublicCommercialPlan[] }) {
  const [interval, setInterval] = useState<BillingInterval>('monthly');

  return (
    <main className="px-4 py-20 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <header className="mb-10 max-w-3xl">
          <p className="mb-3 text-sm font-semibold text-indigo-600 dark:text-indigo-400">Shahrah plans</p>
          <h1 className="text-4xl font-bold tracking-normal text-slate-950 dark:text-white sm:text-5xl">Pricing</h1>
          <p className="mt-4 text-lg text-slate-600 dark:text-slate-300">Monthly limits with quarterly and yearly billing discounts.</p>
        </header>

        <div className="mb-8 inline-flex rounded-lg border border-slate-200 bg-white p-1 dark:border-slate-800 dark:bg-slate-950">
          {(['monthly', 'quarterly', 'yearly'] as BillingInterval[]).map((item) => (
            <button key={item} type="button" onClick={() => setInterval(item)} className={`min-h-10 rounded-md px-4 text-sm font-medium ${interval === item ? 'bg-slate-900 text-white dark:bg-white dark:text-slate-950' : 'text-slate-600 dark:text-slate-300'}`}>
              {intervalLabel(item)}
            </button>
          ))}
        </div>

        {plans.length === 0 ? (
          <div className="rounded-lg border border-slate-200 px-5 py-10 text-center text-slate-600 dark:border-slate-800 dark:text-slate-300">Pricing is temporarily unavailable.</div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            {plans.map((plan) => {
              const price = plan.prices?.find((item) => item.billing_interval === interval);
              const isGrowth = plan.slug === 'growth';
              return (
                <article key={plan.slug} className={`flex min-h-[520px] flex-col rounded-lg border p-6 ${isGrowth ? 'border-indigo-500 bg-indigo-50/40 dark:bg-indigo-950/20' : 'border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950'}`}>
                  <div className="mb-6">
                    <div className="flex items-center justify-between gap-3">
                      <h2 className="text-xl font-semibold text-slate-950 dark:text-white">{plan.display_name}</h2>
                      {isGrowth ? <span className="text-xs font-semibold text-indigo-600 dark:text-indigo-400">Popular</span> : null}
                    </div>
                    <p className="mt-2 min-h-10 text-sm text-slate-600 dark:text-slate-400">{plan.description ?? plan.slug}</p>
                    <div className="mt-5 min-h-16">
                      <p className="text-2xl font-bold text-slate-950 dark:text-white">{formatPrice(price)}</p>
                      {price?.discount_percent ? <p className="mt-1 text-xs text-emerald-600">{price.discount_percent}% discount</p> : null}
                    </div>
                  </div>

                  <ul className="flex-1 space-y-3 text-sm text-slate-700 dark:text-slate-300">
                    <PlanFeature label={`${formatLimit(plan.features?.['gateway.requests.monthly'])} requests / month`} enabled />
                    <PlanFeature label={`${formatLimit(plan.features?.['users.max'])} users`} enabled />
                    <PlanFeature label="Semantic cache" enabled={plan.features?.['cache.semantic.enabled'] === true} />
                    <PlanFeature label="Rate and budget policies" enabled={plan.features?.['rate_limit_rules.enabled'] === true && plan.features?.['budget_limit_rules.enabled'] === true} />
                    <PlanFeature label="Virtual access tokens" enabled={plan.features?.['auth.vat.enabled'] === true} />
                    <PlanFeature label="Role based access" enabled={plan.features?.['rbac.enabled'] === true} />
                    <PlanFeature label="Organization units" enabled={plan.features?.['org_units.enabled'] === true} />
                  </ul>

                  <Link href={`/onboarding?plan=${encodeURIComponent(plan.slug)}`} className={`mt-6 inline-flex min-h-11 items-center justify-center rounded-lg px-4 text-sm font-semibold ${isGrowth ? 'bg-indigo-600 text-white hover:bg-indigo-700' : 'bg-slate-100 text-slate-950 hover:bg-slate-200 dark:bg-slate-800 dark:text-white dark:hover:bg-slate-700'}`}>
                    {price?.amount_minor == null ? 'Contact sales' : 'Select plan'}
                  </Link>
                </article>
              );
            })}
          </div>
        )}
      </div>
    </main>
  );
}

function PlanFeature({ label, enabled }: { label: string; enabled: boolean }) {
  const Icon = enabled ? Check : Minus;
  return <li className={`flex items-start gap-2 ${enabled ? '' : 'text-slate-400'}`}><Icon className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" /><span>{label}</span></li>;
}

function intervalLabel(value: BillingInterval) {
  if (value === 'quarterly') return '3 months';
  if (value === 'yearly') return 'Yearly';
  return 'Monthly';
}

function formatPrice(price?: PublicPlanPrice) {
  if (!price || price.amount_minor == null) return 'Custom pricing';
  return `${new Intl.NumberFormat('fa-IR').format(price.amount_minor)} ${price.currency}`;
}

function formatLimit(value: unknown) {
  if (value == null || value === 'custom') return 'Custom';
  if (typeof value === 'number') return new Intl.NumberFormat('en-US').format(value);
  return String(value);
}
