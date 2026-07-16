import assert from "node:assert/strict";
import test from "node:test";

import { debugOTPResponseField, shouldExposeDebugOTP } from "./debug-otp.ts";

test("debug OTP exposure fails closed when NODE_ENV is unset", () => {
  assert.equal(
    shouldExposeDebugOTP({ SHAHRAH_EXPOSE_DEBUG_OTP: "true" }),
    false,
  );
  assert.deepEqual(
    debugOTPResponseField("123456", {
      SHAHRAH_EXPOSE_DEBUG_OTP: "true",
    }),
    {},
  );
});

test("debug OTP exposure requires an explicit opt-in", () => {
  assert.equal(shouldExposeDebugOTP({ NODE_ENV: "development" }), false);
  assert.equal(shouldExposeDebugOTP({ NODE_ENV: "test" }), false);
});

test("debug OTP exposure allows an explicit non-production opt-in", () => {
  assert.deepEqual(
    debugOTPResponseField("123456", {
      NODE_ENV: "development",
      SHAHRAH_EXPOSE_DEBUG_OTP: "true",
    }),
    { debug_otp: "123456" },
  );
  assert.equal(
    shouldExposeDebugOTP({
      NODE_ENV: "test",
      SHAHRAH_EXPOSE_DEBUG_OTP: "TRUE",
    }),
    true,
  );
});

test("debug OTP exposure rejects production and unknown environments", () => {
  for (const nodeEnv of ["production", "staging", "preview", ""]) {
    assert.equal(
      shouldExposeDebugOTP({
        NODE_ENV: nodeEnv,
        SHAHRAH_EXPOSE_DEBUG_OTP: "true",
      }),
      false,
    );
  }
});
