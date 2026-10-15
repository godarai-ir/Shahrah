import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";

/** The member's sign-in identity: email, mobile when mobile sign-in is enabled, or id. */
export function resolveIdentity(user: TenantAdminUser): string {
  return user.email || (mobileAuthUIEnabled ? user.mobile : "") || user.id;
}

/** The member's first and last name joined, or an empty string when neither is set. */
export function fullName(user: Pick<TenantAdminUser, "first_name" | "last_name">): string {
  return [user.first_name, user.last_name].map((part) => (part ?? "").trim()).filter(Boolean).join(" ");
}

/** The name to show for a member: their real name, else one guessed from the sign-in identity. */
export function memberDisplayName(user: TenantAdminUser): string {
  const identity = resolveIdentity(user);
  return fullName(user) || guessDisplayName(identity) || identity;
}

/** A readable name derived from an identity, such as the local part of an email. */
export function guessDisplayName(identity: string): string {
  const value = identity.trim();
  if (!value) return "";
  return value.includes("@") ? value.split("@")[0] : value;
}
