export type WalletPayload = {
  id?: string;
  wallet_id?: string;
  tenant_id?: string;
  balance_irr?: number;
  pending_irr?: number;
};

export type WalletTransaction = {
  id: string;
  type: string;
  amount_irr: number;
  balance_after_irr?: number;
  status: string;
  created_at: string;
};

export type BillingProviderUsage = {
  provider_name: string;
  request_count: number;
  input_tokens: number;
  output_tokens: number;
  raw_cost_irr: number;
  discount_irr: number;
  cache_savings_irr: number;
  final_cost_irr: number;
  share: number;
};

export type BillingUsageSummary = {
  range: string;
  start: string;
  end: string;
  total_usage_irr: number;
  discount_irr: number;
  cache_savings_irr: number;
  gift_credit_irr: number;
  providers: BillingProviderUsage[];
};

export type MonthlyUsageDetail = BillingUsageSummary & {
  month: string;
};

export type BankTransferRequest = {
  id: string;
  public_id?: string;
  amount_irr: number;
  tracking_number: string;
  bank_transaction_at: string;
  receipt_file_url?: string;
  status: string;
  rejection_reason?: string;
  created_at: string;
};

export type BillingProfile = {
  id?: string;
  profile_type: string;
  data: Record<string, unknown>;
};

export type Invoice = {
  id: string;
  invoice_number: string;
  status: string;
  period_start: string;
  period_end: string;
  raw_usage_irr: number;
  discount_irr: number;
  gift_credit_irr: number;
  cache_savings_irr: number;
  final_amount_irr: number;
  net_adjustment_irr: number;
  created_at: string;
};

export type DiscountRedemption = {
  id: string;
  code: string;
  discount_type: string;
  value_irr?: number;
  percent?: number;
  status: string;
  used_amount_irr: number;
  created_at: string;
};

export function unwrapList<T>(payload: unknown): T[] {
  if (Array.isArray(payload)) return payload as T[];
  if (payload && typeof payload === "object" && Array.isArray((payload as { data?: unknown }).data)) {
    return (payload as { data: T[] }).data;
  }
  return [];
}

const faNumber = new Intl.NumberFormat("fa-IR");

export function formatIRR(value: number | undefined | null) {
  return `${faNumber.format(Math.round(Number(value ?? 0)))} ریال`;
}

export function formatToman(valueIRR: number | undefined | null) {
  return `${faNumber.format(Math.round(Number(valueIRR ?? 0) / 10))} تومان`;
}

export function formatDate(value: string | undefined | null) {
  if (!value) return "-";
  return new Date(value).toLocaleString("fa-IR");
}

export function formatDateOnly(value: string | undefined | null) {
  if (!value) return "-";
  return new Date(value).toLocaleDateString("fa-IR");
}

export function statusTone(status: string): "success" | "warning" | "error" | "pending" | "neutral" {
  if (["successful", "completed", "approved", "paid", "used", "active", "issued"].includes(status)) return "success";
  if (["pending", "pending_review", "draft"].includes(status)) return "pending";
  if (["failed", "rejected", "cancelled", "expired"].includes(status)) return "error";
  if (status === "refunded") return "warning";
  return "neutral";
}
