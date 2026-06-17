import test from "node:test";
import assert from "node:assert/strict";

import { decodeLoginIdentifierParam } from "./login-prefill.ts";

function encode(value: string) {
  return Buffer.from(value, "utf8")
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/g, "");
}

test("decodeLoginIdentifierParam decodes and normalizes an email identifier", () => {
  assert.equal(decodeLoginIdentifierParam(encode("User@Example.COM")), "user@example.com");
});

test("decodeLoginIdentifierParam decodes and normalizes a mobile identifier", () => {
  assert.equal(decodeLoginIdentifierParam(encode("0098912 345-6789")), "+989123456789");
});

test("decodeLoginIdentifierParam rejects unsafe or invalid values", () => {
  assert.equal(decodeLoginIdentifierParam(encode("<script@example.com>")), "");
  assert.equal(decodeLoginIdentifierParam(encode("not-a-login")), "");
});
