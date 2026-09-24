import { GuardrailAPIError } from "./api";
import type { GuardrailTranslator } from "./form";

/** Maps an API failure to localized copy; a 409 means the slug is already taken. */
export function localizedError(error: unknown, fallback: string, t: GuardrailTranslator) {
  if (!(error instanceof GuardrailAPIError)) {
    return error instanceof Error ? error.message : fallback;
  }
  if (error.status === 409) return t("messages.duplicate");
  if (error.status === 400) return t("messages.invalid");
  if (error.status === 403) return t("messages.forbidden");
  if (error.status === 404) return t("messages.notFound");
  if (error.status && error.status >= 500) return t("messages.serverError");

  return fallback;
}
