import { describe, expect, it } from "vitest";

import { classifyProviderAccountDeleteError } from "./provider-account-delete-error";

describe("provider account delete error classification", () => {
  it("classifies the active access-key dependency response", () => {
    expect(
      classifyProviderAccountDeleteError(
        new Error("provider account is referenced by active keys; revoke or update keys first"),
      ),
    ).toBe("activeAccessKeys");
  });

  it("classifies other known deletion failures", () => {
    expect(classifyProviderAccountDeleteError(new Error("insufficient permissions"))).toBe(
      "forbidden",
    );
    expect(classifyProviderAccountDeleteError(new Error("provider account not found"))).toBe(
      "notFound",
    );
  });

  it("does not expose an unknown backend message as a translation", () => {
    expect(classifyProviderAccountDeleteError(new Error("unexpected upstream failure"))).toBe(
      "unknown",
    );
  });
});
