import { describe, expect, it } from "vitest";

import nextConfig, { buildContentSecurityPolicy } from "./next.config";

describe("tenant-admin response security headers", () => {
  it("sets browser security headers on every response", async () => {
    expect(nextConfig.headers).toBeTypeOf("function");

    const rules = await nextConfig.headers!();
    const globalRule = rules.find((rule) => rule.source === "/:path*");
    const headers = new Map(
      globalRule?.headers.map((header) => [header.key.toLowerCase(), header.value]),
    );

    expect(headers.get("content-security-policy")).toContain(
      "frame-ancestors 'none'",
    );
    expect(headers.get("strict-transport-security")).toContain(
      "max-age=63072000",
    );
    expect(headers.get("x-content-type-options")).toBe("nosniff");
    expect(headers.get("x-frame-options")).toBe("DENY");
    expect(headers.get("referrer-policy")).toBe("strict-origin-when-cross-origin");
    expect(headers.get("permissions-policy")).toContain("camera=()");
  });

  it("prevents caching for every API response", async () => {
    expect(nextConfig.headers).toBeTypeOf("function");

    const rules = await nextConfig.headers!();
    const apiRule = rules.find((rule) => rule.source === "/api/:path*");

    expect(apiRule?.headers).toContainEqual({
      key: "Cache-Control",
      value: "no-store",
    });
  });

  it("does not enable development script or connection relaxations in production", () => {
    const policy = buildContentSecurityPolicy(true);

    expect(policy).not.toContain("'unsafe-eval'");
    expect(policy).toContain("connect-src 'self';");
  });
});
