import { cache } from "react";

import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { z } from "zod";

export type PermissionDefinition = {
  key: string;
  group: string;
  description?: string;
};

export type MePermissionsResponse = {
  permissions?: string[];
  catalog?: PermissionDefinition[];
};

const mePermissionsSchema = z.object({
  permissions: z.array(z.string()).optional(),
  catalog: z.array(z.object({ key: z.string(), group: z.string(), description: z.string().optional() })).optional(),
});

const PERMISSIONS_TTL_MS = 60_000;
const MAX_CACHE_ENTRIES = 4000;

type CachedPermissionsEntry = {
  expiresAt: number;
  payload: MePermissionsResponse;
};

const permissionsTTLCache = new Map<string, CachedPermissionsEntry>();

const loadMePermissionsRequestScoped = cache(async (tenantSlug: string, userID: string): Promise<MePermissionsResponse> => {
  const cacheKey = `${tenantSlug.toLowerCase()}:${userID.toLowerCase()}`;
  const now = Date.now();
  const cached = permissionsTTLCache.get(cacheKey);
  if (cached && cached.expiresAt > now) {
    return cached.payload;
  }

  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/me/permissions",
    userID: userID.trim() || undefined,
  });
  if (!upstream.ok) {
    throw new Error(upstream.errorMessage ?? `Request failed (${upstream.status})`);
  }

  const parsed = mePermissionsSchema.safeParse(upstream.payload ?? {});
  if (!parsed.success) throw new Error("Unexpected permissions response contract");
  const payload = parsed.data;
  if (permissionsTTLCache.size >= MAX_CACHE_ENTRIES) {
    for (const [key, value] of permissionsTTLCache.entries()) {
      if (value.expiresAt <= now) {
        permissionsTTLCache.delete(key);
      }
    }
    if (permissionsTTLCache.size >= MAX_CACHE_ENTRIES) {
      const firstKey = permissionsTTLCache.keys().next().value;
      if (firstKey) permissionsTTLCache.delete(firstKey);
    }
  }
  permissionsTTLCache.set(cacheKey, {
    expiresAt: now + PERMISSIONS_TTL_MS,
    payload,
  });
  return payload;
});

export async function getCurrentMePermissions(tenantSlug: string, userID: string): Promise<MePermissionsResponse> {
  return loadMePermissionsRequestScoped(tenantSlug.trim(), userID.trim());
}

export function getNormalizedPermissionSet(payload: MePermissionsResponse): Set<string> {
  return new Set((payload.permissions ?? []).map((item) => item.trim().toLowerCase()));
}

export function invalidateMePermissionsCache(tenantSlug: string, userID: string): void {
  const cacheKey = `${tenantSlug.trim().toLowerCase()}:${userID.trim().toLowerCase()}`;
  permissionsTTLCache.delete(cacheKey);
}

export function invalidateAllMePermissionsCache(): void {
  permissionsTTLCache.clear();
}
