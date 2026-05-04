import TenantAdminDashboard from "@/app/components/tenant-admin-dashboard";
import AdminShell from "@/app/components/admin-shell";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { requireAuth } from "@/lib/auth/session";

type TenantKey = {
  id: string;
  key_alias: string;
  role: string;
  key_name?: string;
  models: string[];
  spend: number;
  blocked: boolean;
  expires_at?: string;
};

type TenantModel = {
  id: string;
  provider?: string;
  aliases?: string[];
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
  amount_usd?: number | null;
  status: string;
  created_at: string;
};
type WalletTransactionListResponse = {
  data?: WalletTransaction[];
};

type KeyListResponse = {
  data?: TenantKey[];
};

type ModelListResponse = {
  data?: TenantModel[];
};

export const dynamic = "force-dynamic";

export default async function Home() {
  const auth = await requireAuth("/login");
  const tenantSlug = auth.tenantSlug;

  const [walletResult, keysResult, modelsResult, txResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/wallet",
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/keys",
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/models",
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/wallet/transactions?limit=20",
    }),
  ]);

  const initialWallet = walletResult.ok
    ? (walletResult.payload as Wallet)
    : null;
  const initialKeys = keysResult.ok
    ? ((keysResult.payload as KeyListResponse).data ?? [])
    : [];
  const initialModels = modelsResult.ok
    ? ((modelsResult.payload as ModelListResponse).data ?? [])
    : [];
  const initialTransactions = txResult.ok
    ? ((txResult.payload as WalletTransactionListResponse).data ?? [])
    : [];

  const errorMessage =
    walletResult.errorMessage ??
    keysResult.errorMessage ??
    modelsResult.errorMessage ??
    txResult.errorMessage ??
    null;

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TenantAdminDashboard
        initialHost={`${tenantSlug}.localhost`}
        initialWallet={initialWallet}
        initialKeys={initialKeys}
        initialModels={initialModels}
        initialTransactions={initialTransactions}
        initialError={errorMessage}
        currentUser={currentUser}
      />
    </AdminShell>
  );
}
