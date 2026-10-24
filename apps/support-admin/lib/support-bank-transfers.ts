// Contracts for the control-plane bank transfer review queue (packages/contracts/platform/billing_financial.go).

export type BankTransferKind = 'wallet_topup' | 'credit_pack' | 'seat';
export type BankTransferStatus = 'pending_review' | 'approved' | 'rejected';

export type BankTransferItem = {
  id: string;
  public_id: string;
  tenant_id: string;
  tenant_slug: string;
  kind: BankTransferKind;
  amount_irr: number;
  tracking_number: string;
  bank_transaction_at: string;
  receipt_file_url?: string;
  status: BankTransferStatus;
  rejection_reason?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
};

export type BankTransferList = { data: BankTransferItem[]; total: number };

export const kindLabels: Record<BankTransferKind, string> = {
  wallet_topup: 'Wallet top-up',
  credit_pack: 'Credit pack',
  seat: 'Seat order',
};

/** Receipt images are private; this proxied route checks the operator's permission before streaming one. */
export function receiptHref(item: BankTransferItem): string | undefined {
  const fileName = item.receipt_file_url?.split('/').pop();
  if (!fileName) return undefined;
  return `/api/admin/control-plane/tenants/${encodeURIComponent(item.tenant_id)}/wallet/bank-transfer-receipts/${encodeURIComponent(fileName)}`;
}
