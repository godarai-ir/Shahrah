import assert from "node:assert/strict";
import test from "node:test";

import { resolveClientIPAddressOptions } from "./client-ip-config";

test("uses the ingress single-value client IP header by default", () => {
  assert.deepEqual(resolveClientIPAddressOptions({}), {
    ipAddressHeaders: ["x-real-ip"],
  });
});

test("normalizes configured headers and removes unsafe header names", () => {
  assert.deepEqual(
    resolveClientIPAddressOptions({
      BETTER_AUTH_IP_ADDRESS_HEADERS:
        " CF-Connecting-IP, x-forwarded-for, x-forwarded-for, invalid header ",
    }),
    {
      ipAddressHeaders: ["cf-connecting-ip", "x-forwarded-for"],
    },
  );
});

test("passes exact trusted proxy addresses to Better Auth", () => {
  assert.deepEqual(
    resolveClientIPAddressOptions({
      BETTER_AUTH_IP_ADDRESS_HEADERS: "x-forwarded-for",
      BETTER_AUTH_TRUSTED_PROXIES: "10.1.2.3, 192.0.2.0/24, 10.1.2.3",
    }),
    {
      ipAddressHeaders: ["x-forwarded-for"],
      trustedProxies: ["10.1.2.3", "192.0.2.0/24"],
    },
  );
});
