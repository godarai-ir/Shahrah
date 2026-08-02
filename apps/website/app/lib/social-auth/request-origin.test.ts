import assert from "node:assert/strict";
import test from "node:test";

import { hasExpectedRequestOrigin, resolveRequestOrigin } from "./request-origin";

test("uses the public host when Next reconstructs an internal request URL", () => {
  const request = new Request("http://0.0.0.0:3010/auth/social/error", {
    headers: { host: "localhost:3010" },
  });

  assert.equal(resolveRequestOrigin(request), "http://localhost:3010");
});

test("uses the first forwarded host and protocol supplied by the ingress", () => {
  const request = new Request("http://website:3010/auth/social/error", {
    headers: {
      host: "website:3010",
      "x-forwarded-host": "auth.godarai.ir, website:3010",
      "x-forwarded-proto": "https, http",
    },
  });

  assert.equal(resolveRequestOrigin(request), "https://auth.godarai.ir");
});

test("falls back to the request URL when host headers are absent", () => {
  const request = new Request("http://localhost:3010/auth/social/error");

  assert.equal(resolveRequestOrigin(request), "http://localhost:3010");
});

test("accepts a browser origin when Next exposes an internal request URL", () => {
  const request = new Request("http://0.0.0.0:3010/api/onboarding", {
    headers: {
      host: "localhost:3010",
      origin: "http://localhost:3010",
    },
  });

  assert.equal(hasExpectedRequestOrigin(request, "http://localhost:3010"), true);
});

test("rejects a cross-origin browser request", () => {
  const request = new Request("http://0.0.0.0:3010/api/onboarding", {
    headers: {
      host: "localhost:3010",
      origin: "http://malicious.localhost:3010",
    },
  });

  assert.equal(hasExpectedRequestOrigin(request, "http://localhost:3010"), false);
});

test("rejects a request received on an unexpected host", () => {
  const request = new Request("http://0.0.0.0:3010/api/onboarding", {
    headers: {
      host: "other.localhost:3010",
      origin: "http://localhost:3010",
    },
  });

  assert.equal(hasExpectedRequestOrigin(request, "http://localhost:3010"), false);
});
