import assert from "node:assert/strict";
import test from "node:test";

import { decideOIDCLoginStep } from "./oidc-login-step.ts";

test("an existing session continues to the application", () => {
  assert.equal(decideOIDCLoginStep({ hasSession: true, prompt: "", reauthRequested: false }), "continue");
});

test("a visitor without a session sees the login form", () => {
  assert.equal(decideOIDCLoginStep({ hasSession: false, prompt: "", reauthRequested: false }), "form");
});

test("prompt=login and a rejected session force a fresh sign-in", () => {
  assert.equal(decideOIDCLoginStep({ hasSession: true, prompt: "login", reauthRequested: false }), "form");
  assert.equal(decideOIDCLoginStep({ hasSession: true, prompt: "", reauthRequested: true }), "form");
});

test("prompt=none never shows a form", () => {
  assert.equal(decideOIDCLoginStep({ hasSession: false, prompt: "none", reauthRequested: false }), "reject");
  assert.equal(decideOIDCLoginStep({ hasSession: true, prompt: "none", reauthRequested: false }), "continue");
  assert.equal(
    decideOIDCLoginStep({ hasSession: true, prompt: "none", reauthRequested: true }),
    "reject",
    "a rejected session with prompt=none must not loop back to the continue route",
  );
});
