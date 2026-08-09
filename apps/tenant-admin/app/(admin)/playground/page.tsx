import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import PlaygroundPage from "@/features/playground/PlaygroundPage";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function PlaygroundRoute() {
  const auth = await requireAuth("/login?next=/playground");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["playground:read", "playground:run", "policy:list", "policy:manage"]);
  const t = await getTranslations("playground");

  return (
    <AdminShell messageNamespaces={["playground"]}
      tenantSlug={auth.tenantSlug}
      currentUser={{ id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role }}
      pageHeader={{ title: t("title"), description: t("description") }}
      contentPadding="none"
    >
      <PlaygroundPage />
    </AdminShell>
  );
}
