import assert from "node:assert/strict";
import test from "node:test";

import { hideMobileIdentity } from "./mobile-auth-ui.ts";

test("hides mobile-shaped identity labels while mobile auth UI is disabled", () => {
  assert.equal(hideMobileIdentity("+989121234567", false), "");
  assert.equal(hideMobileIdentity("0912 123 4567", false), "");
  assert.equal(hideMobileIdentity("person@example.com", false), "person@example.com");
});

test("keeps mobile identities available when mobile auth UI is enabled", () => {
  assert.equal(hideMobileIdentity("+989121234567", true), "+989121234567");
});
