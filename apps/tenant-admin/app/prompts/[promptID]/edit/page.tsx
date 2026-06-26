import AdminShell from "@/app/components/admin-shell";
import PromptBuilderPage from "@/app/components/prompt-builder-page";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

type PromptEditPageProps = {
  params: Promise<{ promptID: string }>;
};

export default async function PromptEditPage(props: PromptEditPageProps) {
  const auth = await requireAuth("/login?next=/prompts");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["policy:list", "policy:manage"]);

  const { promptID } = await props.params;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={auth.tenantSlug} currentUser={currentUser}>
      <PromptBuilderPage mode="edit" promptID={promptID} />
    </AdminShell>
  );
}
