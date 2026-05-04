import AdminShell from "@/app/components/admin-shell";
import TopUpPanel from "@/app/components/topup-panel";
import { requireAuth } from "@/lib/auth/session";
import { prisma } from "@/lib/prisma";

export const dynamic = "force-dynamic";

type WalletTopUpSettings = {
  min_amount_usd: number;
  max_amount_usd: number;
  tax_percent: number;
  fee_percent: number;
};

type Wallet = {
  id: string;
  tenant_id: string;
  balance_irr: number;
  pending_irr: number;
};

type WalletTransaction = {
  id: string;
  type: string;
  amount_irr: number;
  status: string;
  created_at: string;
};

type TenantRow = { tenant_id: string };
type WalletRow = { wallet_id: string; tenant_id: string; balance_irr: number; pending_irr: number };
type SettingsRow = WalletTopUpSettings;
type TransactionRow = { transaction_id: string; tx_type: string; amount_irr: number; status: string; created_at: Date };

export default async function TopUpPage() {
  const auth = await requireAuth("/login?next=/billing/topup");
  const tenantSlug = auth.tenantSlug;

  let initialError: string | null = null;
  let initialWallet: Wallet | null = null;
  let initialSettings: WalletTopUpSettings | null = null;
  let initialTransactions: WalletTransaction[] = [];

  try {
    const tenantRows = await prisma.$queryRaw<TenantRow[]>`
      SELECT tenant_id
      FROM tenants
      WHERE slug = ${tenantSlug}
      LIMIT 1
    `;

    const tenantID = tenantRows[0]?.tenant_id;
    if (!tenantID) {
      throw new Error("Tenant not found");
    }

    const [walletRows, settingsRows, transactionRows] = await Promise.all([
      prisma.$queryRaw<WalletRow[]>`
        SELECT wallet_id, tenant_id, balance_irr, pending_irr
        FROM wallets
        WHERE tenant_id = ${tenantID}
        LIMIT 1
      `,
      prisma.$queryRaw<SettingsRow[]>`
        SELECT min_amount_usd, max_amount_usd, tax_percent, fee_percent
        FROM wallet_topup_settings
        LIMIT 1
      `,
      prisma.$queryRaw<TransactionRow[]>`
        SELECT transaction_id, tx_type, amount_irr, status, created_at
        FROM wallet_transactions
        WHERE tenant_id = ${tenantID}
        ORDER BY created_at DESC
        LIMIT 20
      `,
    ]);

    const wallet = walletRows[0];
    if (wallet) {
      initialWallet = {
        id: wallet.wallet_id,
        tenant_id: wallet.tenant_id,
        balance_irr: Number(wallet.balance_irr ?? 0),
        pending_irr: Number(wallet.pending_irr ?? 0),
      };
    }

    initialSettings = settingsRows[0] ?? null;
    initialTransactions = transactionRows.map((row) => ({
      id: row.transaction_id,
      type: row.tx_type,
      amount_irr: Number(row.amount_irr ?? 0),
      status: row.status,
      created_at: row.created_at.toISOString(),
    }));
  } catch (error) {
    initialError = error instanceof Error ? error.message : "Failed to load wallet data";
  }

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TopUpPanel
        initialWallet={initialWallet}
        initialSettings={initialSettings}
        initialTransactions={initialTransactions}
        initialError={initialError}
      />
    </AdminShell>
  );
}
