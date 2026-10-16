export type TenantSummary = {
  id: string;
  slug: string;
  domain: string;
  status: string;
  created_at: string;
  plan_slug?: string;
  plan_display_name?: string;
  subscription_status?: string;
  period_end?: string;
  wallet_balance_minor: number;
  wallet_pending_minor: number;
};

export type TenantListResponse = { data: TenantSummary[]; next_cursor?: string };

export type Subscription = {
  id: string;
  plan_slug?: string;
  plan_display_name?: string;
  version?: number;
  status: string;
  billing_interval: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
};

export type TenantDetail = {
  tenant: TenantSummary;
  owner?: { user_id: string; name?: string; email?: string; phone?: string };
  subscription?: Subscription;
  entitlements?: { features: Record<string, unknown>; overrides?: Record<string, unknown> };
  wallet?: { id: string; tenant_id: string; balance_irr: number; pending_irr: number };
};

export type WalletTransaction = {
  id: string; type: string; amount_irr: number; status: string; description?: string; created_at: string;
};

export type TenantEvent = {
  id: string; type: string; severity: string; actor_id?: string; actor_role?: string; reason?: string;
  request_id?: string; payload?: Record<string, unknown>; created_at: string;
};

export type SubscriptionPaymentIntent = {
  id: string; public_id: string; status: string; provider: string; currency: string; amount_minor: number;
  discount_percent: number; ref_id?: string; failure_reason?: string; created_at: string; completed_at?: string;
};

export async function controlPlaneRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, {
    ...init,
    headers: { 'content-type': 'application/json', ...(init?.headers ?? {}) },
    cache: 'no-store',
  });
  const payload = await response.json().catch(() => ({})) as { error?: string | { message?: string; code?: string }; message?: string };
  if (!response.ok) throw new Error(controlPlaneErrorMessage(payload) || `Request failed (${response.status})`);
  return payload as T;
}

export function formatIRR(value: number): string {
  return `${new Intl.NumberFormat('fa-IR').format(Math.round(value))} IRR`;
}

export function formatDate(value?: string): string {
  if (!value) return 'Unavailable';
  return new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

/** Control-plane errors are either `{error: "text"}` or `{error: {code, message}}`. */
export function controlPlaneErrorMessage(payload: { error?: string | { message?: string; code?: string }; message?: string }): string | undefined {
  if (typeof payload.error === 'string') return payload.error;
  if (payload.error && typeof payload.error === 'object') return payload.error.message || payload.error.code;
  return payload.message;
}
