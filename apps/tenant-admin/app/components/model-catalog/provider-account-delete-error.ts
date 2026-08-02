export type ProviderAccountDeleteErrorCode =
  | "activeAccessKeys"
  | "forbidden"
  | "notFound"
  | "unknown";

export function classifyProviderAccountDeleteError(
  error: unknown,
): ProviderAccountDeleteErrorCode {
  const message = error instanceof Error ? error.message.trim().toLowerCase() : "";

  if (
    message.includes("provider account is referenced by active keys") ||
    (message.includes("provider account") && message.includes("active keys"))
  ) {
    return "activeAccessKeys";
  }
  if (message.includes("insufficient permissions") || message.includes("forbidden")) {
    return "forbidden";
  }
  if (message.includes("not found")) {
    return "notFound";
  }
  return "unknown";
}
