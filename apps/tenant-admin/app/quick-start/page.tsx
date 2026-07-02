import AdminShell from "@/app/components/admin-shell";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function QuickStartPage() {
  const auth = await requireAuth("/login?next=/quick-start");
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={auth.tenantSlug} currentUser={currentUser}>
      <div className="min-h-[calc(100svh-6rem)]" />
    </AdminShell>
  );
}
