import { RuleAPIError, type RuleErrorCode } from "@/features/gateway-policies/rate-limiting/ruleApi";

const CODED_ERRORS: ReadonlySet<RuleErrorCode> = new Set([
  "saveRejected",
  "invalidBody",
  "duplicateRuleId",
  "internal",
  "unauthenticated",
  "forbidden",
  "notFound",
]);

/**
 * Display copy for a rule API failure. Coded errors are translated; an upstream
 * message that carries its own explanation is shown as-is, and anything else
 * falls back to the caller's generic message.
 */
export function resolveRuleErrorMessage(
  error: unknown,
  translateCode: (code: RuleErrorCode) => string,
  fallback: string,
): string {
  if (error instanceof RuleAPIError) {
    if (CODED_ERRORS.has(error.code)) return translateCode(error.code);
    return error.message || fallback;
  }

  return error instanceof Error && error.message ? error.message : fallback;
}
