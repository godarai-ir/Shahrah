import assert from "node:assert/strict";
import test from "node:test";

import {
  normalizeSocialReturnPath,
  resolveCentralSocialAuthBaseURL,
} from "./social-auth-url.ts";

test("resolveCentralSocialAuthBaseURL requires an explicit production origin", () => {
  assert.equal(resolveCentralSocialAuthBaseURL({ NODE_ENV: "production" }), null);
  assert.equal(
    resolveCentralSocialAuthBaseURL({
      NODE_ENV: "production",
      GODARAI_AUTH_BASE_URL: "https://auth.godarai.ir/some/path",
    })?.toString(),
    "https://auth.godarai.ir/",
  );
});

test("resolveCentralSocialAuthBaseURL rejects unsafe URLs", () => {
  assert.equal(
    resolveCentralSocialAuthBaseURL({
      NODE_ENV: "production",
      GODARAI_AUTH_BASE_URL: "javascript:alert(1)",
    }),
    null,
  );
  assert.equal(
    resolveCentralSocialAuthBaseURL({
      NODE_ENV: "production",
      GODARAI_AUTH_BASE_URL: "https://user:password@auth.godarai.ir",
    }),
    null,
  );
});

test("normalizeSocialReturnPath only permits same-origin paths", () => {
  assert.equal(normalizeSocialReturnPath("/settings/account"), "/settings/account");
  assert.equal(normalizeSocialReturnPath("https://evil.example"), "/");
  assert.equal(normalizeSocialReturnPath("//evil.example"), "/");
});
