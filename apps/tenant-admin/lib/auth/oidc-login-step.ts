/** What the login page does with a pending OIDC request. */
export type OIDCLoginStep = "continue" | "reject" | "form";

export type OIDCLoginStepInput = {
  hasSession: boolean;
  prompt: "none" | "login" | "";
  reauthRequested: boolean;
};

/**
 * prompt=login and a rejected session both demand a fresh sign-in even when a session exists;
 * prompt=none must never show a form, so a visitor without a session is sent back with an error.
 */
export function decideOIDCLoginStep({ hasSession, prompt, reauthRequested }: OIDCLoginStepInput): OIDCLoginStep {
  if (prompt === "none") return hasSession && !reauthRequested ? "continue" : "reject";
  if (hasSession && prompt !== "login" && !reauthRequested) return "continue";
  return "form";
}
