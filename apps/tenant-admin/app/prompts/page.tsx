import AdminShell from "@/app/components/admin-shell";
import PromptManagement from "@/app/components/prompt-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function PromptsPage() {
  const auth = await requireAuth("/login?next=/prompts");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list", "policy:manage"]);

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={auth.tenantSlug} currentUser={currentUser}>
      <PromptManagement />
    </AdminShell>
  );
}
