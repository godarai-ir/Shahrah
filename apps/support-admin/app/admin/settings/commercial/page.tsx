"use client";

import { FormEvent, useEffect, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

type SignupSettings = {
  trial_days: number;
  signup_credit_irr: number;
  updated_by?: string;
  updated_at: string;
};

export default function CommercialSignupSettingsPage() {
  const [settings, setSettings] = useState<SignupSettings | null>(null);
  const [credit, setCredit] = useState('2000000');
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
		void (async () => {
			setError(null);
			const response = await fetch('/api/admin/control-plane/commercial-signup-settings', { cache: 'no-store' });
			const payload = await response.json().catch(() => null) as SignupSettings | { error?: string } | null;
			if (!response.ok || !payload || !('signup_credit_irr' in payload)) {
				setError(payload && 'error' in payload ? payload.error ?? 'Unable to load settings.' : 'Unable to load settings.');
				return;
			}
			setSettings(payload);
			setCredit(String(payload.signup_credit_irr));
		})();
  }, []);

  async function save(event: FormEvent) {
    event.preventDefault();
    const amount = Number(credit);
    if (!Number.isSafeInteger(amount) || amount < 0) {
      setError('Signup credit must be a non-negative whole IRR amount.');
      return;
    }
    setSaving(true);
    setError(null);
    setMessage(null);
    const response = await fetch('/api/admin/control-plane/commercial-signup-settings', {
      method: 'PUT',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ signup_credit_irr: amount }),
    });
    const payload = await response.json().catch(() => null) as SignupSettings | { message?: string; error?: string } | null;
    setSaving(false);
    if (!response.ok || !payload || !('signup_credit_irr' in payload)) {
      setError(payload && ('message' in payload || 'error' in payload) ? payload.message ?? payload.error ?? 'Save failed.' : 'Save failed.');
      return;
    }
    setSettings(payload);
    setCredit(String(payload.signup_credit_irr));
    setMessage('Signup credit updated. New tenants will receive this amount.');
  }

  return (
    <div className="max-w-2xl space-y-4">
      <div>
        <h1 className="text-2xl font-semibold">Commercial Signup Settings</h1>
        <p className="mt-1 text-sm text-zinc-400">Controls the promotional wallet credit granted once during tenant onboarding.</p>
      </div>
      {error ? <p className="rounded-md border border-red-800 bg-red-950/40 p-3 text-sm text-red-300">{error}</p> : null}
      {message ? <p className="rounded-md border border-emerald-800 bg-emerald-950/40 p-3 text-sm text-emerald-300">{message}</p> : null}
      <Card>
        <form className="space-y-4" onSubmit={save}>
          <div>
            <label className="mb-2 block text-sm font-medium" htmlFor="signup-credit">Signup credit (IRR)</label>
            <Input id="signup-credit" type="number" min="0" max="9007199254740991" step="1" value={credit} onChange={(event) => setCredit(event.target.value)} />
            <p className="mt-2 text-xs text-zinc-500">Default: 2,000,000 IRR. Existing tenant balances are not changed.</p>
          </div>
          <div className="rounded-md bg-zinc-950 p-3 text-sm text-zinc-300">
            Trial duration: <strong>{settings?.trial_days ?? 15} days</strong>
            {settings?.updated_at ? <span className="mt-1 block text-xs text-zinc-500">Last updated {new Date(settings.updated_at).toLocaleString()}</span> : null}
          </div>
          <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save settings'}</Button>
        </form>
      </Card>
    </div>
  );
}
