import AdminShell from "@/app/components/admin-shell";
import TopUpPanel from "@/app/components/topup-panel";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";

export default async function TopUpPage() {
  const auth = await requireAuth("/login?next=/billing/topup");
  const t = await getTranslations("topupPage");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["topupPage"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description") }}
    >
      <TopUpPanel initialError={null} />
    </AdminShell>
  );
}
