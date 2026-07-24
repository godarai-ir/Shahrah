import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

const ALLOWED_SORTS = new Set([
  "name",
  "provider",
  "availability",
  "context",
  "input_price",
  "output_price",
]);

export async function GET(request: NextRequest) {
  const url = new URL(request.url);
  const params = new URLSearchParams();

  const query = url.searchParams.get("q")?.trim();
  if (query) params.set("q", query.slice(0, 200));
  setNormalizedCSV(params, "providers", url.searchParams.get("providers"));
  setNormalizedCSV(params, "account_ids", url.searchParams.get("account_ids"));

  const sort = url.searchParams.get("sort")?.trim().toLowerCase();
  if (sort && ALLOWED_SORTS.has(sort)) params.set("sort", sort);
  const order = url.searchParams.get("order")?.trim().toLowerCase();
  if (order === "asc" || order === "desc") params.set("order", order);
  setBoundedInteger(params, "limit", url.searchParams.get("limit"), 1, 100);
  setBoundedInteger(params, "offset", url.searchParams.get("offset"), 0, Number.MAX_SAFE_INTEGER);

  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/tenant-provider-account-models${suffix}`,
  });
}

function setNormalizedCSV(params: URLSearchParams, key: string, raw: string | null) {
  const values = [...new Set((raw ?? "").split(",").map((value) => value.trim()).filter(Boolean))];
  if (values.length) params.set(key, values.join(","));
}

function setBoundedInteger(
  params: URLSearchParams,
  key: string,
  raw: string | null,
  min: number,
  max: number,
) {
  if (!raw || !/^\d+$/.test(raw.trim())) return;
  const value = Number(raw);
  if (Number.isSafeInteger(value) && value >= min && value <= max) {
    params.set(key, String(value));
  }
}
