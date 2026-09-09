import {
  LOCALE_COOKIE_MAX_AGE_SECONDS,
  LOCALE_COOKIE_NAME,
  resolveLocale,
  type AppLocale,
} from "@/lib/locale";
import { getLocaleDirection } from "@/lib/locale-direction";

export function serializeLocaleCookie(locale: AppLocale, isSecure: boolean): string {
  const attributes = [
    `${LOCALE_COOKIE_NAME}=${locale}`,
    "Path=/",
    `Max-Age=${LOCALE_COOKIE_MAX_AGE_SECONDS}`,
    "SameSite=Lax",
  ];

  if (isSecure) attributes.push("Secure");

  return attributes.join("; ");
}

/**
 * Persist the visitor's language choice so the next server render picks it up,
 * and apply the matching document direction immediately. The cookie is
 * intentionally readable by scripts: it carries no secret and the server needs
 * it on the very next request.
 */
export function setBrowserLocalePreference(locale: AppLocale): void {
  if (typeof document === "undefined") return;

  document.cookie = serializeLocaleCookie(locale, window.location.protocol === "https:");
  applyDocumentLocale(locale);
}

export function applyDocumentLocale(locale: string): void {
  if (typeof document === "undefined") return;

  const resolved = resolveLocale(locale);
  const element = document.documentElement;
  element.lang = resolved;
  element.dir = getLocaleDirection(resolved);
}
