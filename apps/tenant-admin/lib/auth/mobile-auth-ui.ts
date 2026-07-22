export const mobileAuthUIEnabled =
  process.env.NEXT_PUBLIC_MOBILE_AUTH_UI_ENABLED?.trim().toLowerCase() === "true";

export function hideMobileIdentity(
  value: string | null | undefined,
  isMobileUIEnabled = mobileAuthUIEnabled,
): string {
  const normalized = value?.trim() ?? "";
  if (isMobileUIEnabled || !looksLikeMobileIdentifier(normalized)) {
    return normalized;
  }
  return "";
}

function looksLikeMobileIdentifier(value: string): boolean {
  const compact = value.replace(/[\s\-()]/g, "");
  return /^\+?\d{10,15}$/.test(compact);
}
