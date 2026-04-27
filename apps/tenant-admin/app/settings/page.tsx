import { requireAuth } from "@/lib/auth/session";
import ProfileSettings from "@/app/components/profile-settings";

export const dynamic = "force-dynamic";

export default async function SettingsPage() {
  const auth = await requireAuth("/login?next=/settings");

  return (
    <ProfileSettings
      initialUser={{
        id: auth.user.id,
        email: auth.user.email,
        mobile: auth.user.mobile,
        role: auth.user.role,
      }}
    />
  );
}
