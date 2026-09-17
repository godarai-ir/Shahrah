import { resolveSectionLandingTab } from "@/features/admin-shell/section-landing";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

import GuardrailsPage from "../guardrails/page";
import BudgetLimitRulesPage from "./budget-limiting-rules/page";
import RateLimitRulesPage from "./rate-limiting-rules/page";

export const dynamic = "force-dynamic";

/** Section landing for runtime policies. Renders the first permitted tab in place; the legacy routes stay as deep links. */
export default async function GatewayPoliciesPage() {
  const auth = await requireAuth("/login?next=/gateway-policies");
  const tab = await resolveSectionLandingTab("/gateway-policies", auth.tenantSlug, auth.user.id);

  if (tab === "rateLimitPolicies") return <RateLimitRulesPage />;
  if (tab === "budgetLimitPolicies") return <BudgetLimitRulesPage />;
  if (tab === "guardrails") return <GuardrailsPage />;

  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list", "policy:manage"]);
  return null;
}
