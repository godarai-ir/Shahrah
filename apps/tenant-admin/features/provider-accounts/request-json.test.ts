import { afterEach, describe, expect, it, vi } from "vitest";

import { getProviderCatalog } from "./api";

describe("provider account request utilities", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("does not expose an HTML error response as the error message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response("<!DOCTYPE html><html><body>Not Found</body></html>", {
          status: 404,
          headers: { "Content-Type": "text/html; charset=utf-8" },
        }),
      ),
    );

    const request = getProviderCatalog();

    await expect(request).rejects.toThrow("Request failed (404)");
    await expect(request).rejects.not.toThrow("DOCTYPE");
  });

  it("preserves a structured JSON error message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        Response.json(
          { error: { message: "Catalog is unavailable." } },
          { status: 503 },
        ),
      ),
    );

    await expect(getProviderCatalog()).rejects.toThrow(
      "Catalog is unavailable.",
    );
  });

  it("rejects a successful response that is not JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response("<html><body>Unexpected proxy response</body></html>", {
          status: 200,
          headers: { "Content-Type": "text/html" },
        }),
      ),
    );

    await expect(getProviderCatalog()).rejects.toThrow(
      "Unexpected response contract from /api/provider-catalog/providers",
    );
  });
});
