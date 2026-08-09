import { afterEach, describe, expect, it, vi } from "vitest";
import { z } from "zod";

import { APIContractError, requestParsedJSON } from "./api-client";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("requestParsedJSON", () => {
  it("parses a successful response through its runtime schema", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ id: "item-1" }))));

    await expect(requestParsedJSON(z.object({ id: z.string() }), "/api/items/1"))
      .resolves.toEqual({ id: "item-1" });
  });

  it("exposes stable transport error metadata", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ error: { code: "item_missing", message: "Item not found" } }),
      { status: 404 },
    )));

    const request = requestParsedJSON(z.object({ id: z.string() }), "/api/items/missing");
    await expect(request).rejects.toMatchObject({
      name: "APIClientError",
      status: 404,
      code: "item_missing",
      message: "Item not found",
    });
  });

  it("rejects responses that violate the endpoint contract", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ id: 42 }))));

    await expect(requestParsedJSON(z.object({ id: z.string() }), "/api/items/1"))
      .rejects.toBeInstanceOf(APIContractError);
  });
});
