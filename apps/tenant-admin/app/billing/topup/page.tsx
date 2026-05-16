import AdminShell from "@/app/components/admin-shell";
import TopUpPanel from "@/app/components/topup-panel";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

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

type WalletPayload = {
  id?: string;
  wallet_id?: string;
  tenant_id?: string;
  balance_irr: number;
  pending_irr: number;
};

type SettingsPayload = WalletTopUpSettings;

type TransactionPayload = {
  id?: string;
  transaction_id?: string;
  type?: string;
  tx_type?: string;
  amount_irr: number;
  status: string;
  created_at: string;
};

type TransactionListPayload = {
  data?: TransactionPayload[];
};

export default async function TopUpPage() {
  const auth = await requireAuth("/login?next=/billing/topup");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  let initialError: string | null = null;
  let initialWallet: Wallet | null = null;
  let initialSettings: WalletTopUpSettings | null = null;
  let initialTransactions: WalletTransaction[] = [];

  try {
    const [walletResponse, settingsResponse, transactionsResponse] = await Promise.all([
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/wallet",
        userID: auth.user.id,
      }),
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/wallet/topup/settings",
        userID: auth.user.id,
      }),
      fetchControlPlaneJSON({
        tenantSlug,
        method: "GET",
        path: "/api/v1/wallet/transactions?limit=20",
        userID: auth.user.id,
      }),
    ]);

    if (!walletResponse.ok) {
      throw new Error(walletResponse.errorMessage ?? "Failed to load wallet");
    }
    if (!settingsResponse.ok) {
      throw new Error(settingsResponse.errorMessage ?? "Failed to load wallet settings");
    }
    if (!transactionsResponse.ok) {
      throw new Error(transactionsResponse.errorMessage ?? "Failed to load transactions");
    }

    const wallet = walletResponse.payload as WalletPayload;
    const settings = settingsResponse.payload as SettingsPayload;
    const txPayload = (transactionsResponse.payload ?? {}) as
      | TransactionListPayload
      | TransactionPayload[];
    const transactions = Array.isArray(txPayload)
      ? txPayload
      : Array.isArray(txPayload.data)
        ? txPayload.data
        : [];

    const walletID = wallet?.wallet_id ?? wallet?.id;
    if (walletID) {
      initialWallet = {
        id: walletID,
        tenant_id: wallet.tenant_id ?? "",
        balance_irr: Number(wallet.balance_irr ?? 0),
        pending_irr: Number(wallet.pending_irr ?? 0),
      };
    }

    initialSettings = settings ?? null;
    initialTransactions = transactions.map((row) => ({
      id: row.transaction_id ?? row.id ?? "",
      type: row.tx_type ?? row.type ?? "",
      amount_irr: Number(row.amount_irr ?? 0),
      status: row.status,
      created_at: row.created_at,
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
