import { describe, expect, it } from "vitest";

import { buildCreatedConnectionHref, getCreateStepState, shouldPersistCreateAccount } from "./create-flow";

describe("provider connection create flow", () => {
  it("keeps intermediate continue actions free of server account creation", () => {
    expect(shouldPersistCreateAccount("continue-details")).toBe(false);
    expect(shouldPersistCreateAccount("continue-models")).toBe(false);
    expect(shouldPersistCreateAccount("continue-access")).toBe(false);
    expect(shouldPersistCreateAccount("credential-test")).toBe(true);
    expect(shouldPersistCreateAccount("activate")).toBe(true);
  });

  it("marks visited steps complete without depending on a server account id", () => {
    expect(getCreateStepState({ activeStep: "models", step: "details", completedThrough: 1 })).toBe("complete");
    expect(getCreateStepState({ activeStep: "models", step: "models", completedThrough: 1 })).toBe("active");
    expect(getCreateStepState({ activeStep: "models", step: "access", completedThrough: 1 })).toBe("idle");
  });

  it("preserves the confirmed account name in the success redirect", () => {
    expect(buildCreatedConnectionHref("Anthropic", "anthropic-brisk-orbit")).toBe(
      "/provider-accounts?tab=connected&provider=anthropic&created=anthropic-brisk-orbit",
    );
  });
});
