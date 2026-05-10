import AdminShell from "@/app/components/admin-shell";
import { requireAuth } from "@/lib/auth/session";
import { BudgetLimitRuleForm } from "@/features/gateway-policies/budget-limiting/BudgetLimitRuleForm";

export const dynamic = "force-dynamic";

export default async function Page() {
  const auth = await requireAuth("/login?next=/gateway-policies/budget-limiting-rules/new");
  const tenantSlug = auth.tenantSlug;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };
  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <BudgetLimitRuleForm />
    </AdminShell>
  );
}
