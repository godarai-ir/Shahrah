import assert from "node:assert/strict";
import test from "node:test";

import { normalizeOnboardingLoginURL } from "./onboarding-login-url";

test("rewrites a social handoff URL to the configured local tenant origin", () => {
  const payload = normalizeOnboardingLoginURL(
    {
      tenant: { slug: "google-e2e" },
      login_url: "https://google-e2e.localhost/auth/handoff?code=secret",
    },
    "http://{tenant}.localhost:3000",
  ) as { login_url: string };

  assert.equal(
    payload.login_url,
    "http://google-e2e.localhost:3000/auth/handoff?code=secret",
  );
});

test("uses the configured production tenant origin without weakening HTTPS", () => {
  const payload = normalizeOnboardingLoginURL(
    {
      tenant: { slug: "acme" },
      login_url: "https://acme.internal/auth/handoff?code=secret",
    },
    "https://{tenant}.godarai.ir",
  ) as { login_url: string };

  assert.equal(
    payload.login_url,
    "https://acme.godarai.ir/auth/handoff?code=secret",
  );
});

test("also normalizes the email onboarding login path", () => {
  const payload = normalizeOnboardingLoginURL(
    {
      tenant: { slug: "acme" },
      login_url: "https://acme.internal/login",
    },
    "http://{tenant}.localhost:3000",
  ) as { login_url: string };

  assert.equal(payload.login_url, "http://acme.localhost:3000/login");
});

test("leaves non-provisioning payloads unchanged", () => {
  const payload = { available: true };

  assert.equal(
    normalizeOnboardingLoginURL(payload, "http://{tenant}.localhost:3000"),
    payload,
  );
});
