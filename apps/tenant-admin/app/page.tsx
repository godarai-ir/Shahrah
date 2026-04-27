import TenantAdminDashboard from "@/app/components/tenant-admin-dashboard";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { extractTenantSlugFromHost } from "@/lib/tenant";
import { requireAuth } from "@/lib/auth/session";
import { headers } from "next/headers";

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

type KeyListResponse = {
  data?: TenantKey[];
};

type ModelListResponse = {
  data?: TenantModel[];
};

export const dynamic = "force-dynamic";

export default async function Home() {
  const auth = await requireAuth("/login");

  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const [walletResult, keysResult, modelsResult] = await Promise.all([
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

  const errorMessage =
    walletResult.errorMessage ??
    keysResult.errorMessage ??
    modelsResult.errorMessage ??
    null;

  return (
    <TenantAdminDashboard
      initialHost={incomingHost}
      initialWallet={initialWallet}
      initialKeys={initialKeys}
      initialModels={initialModels}
      initialError={errorMessage}
      currentUser={{
        id: auth.user.id,
        email: auth.user.email,
        mobile: auth.user.mobile,
        role: auth.user.role,
      }}
    />
  );
}
