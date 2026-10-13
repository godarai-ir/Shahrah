import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";

/** The member's sign-in identity: email, mobile when mobile sign-in is enabled, or id. */
export function resolveIdentity(user: TenantAdminUser): string {
  return user.email || (mobileAuthUIEnabled ? user.mobile : "") || user.id;
}

/** A readable name derived from an identity, such as the local part of an email. */
export function guessDisplayName(identity: string): string {
  const value = identity.trim();
  if (!value) return "";
  return value.includes("@") ? value.split("@")[0] : value;
}
