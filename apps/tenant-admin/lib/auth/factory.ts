/**
 * Factory for creating AuthRepository implementation.
 * tenant-admin auth always goes through control-plane APIs.
 */

import { ApiAuthRepository } from "./api-auth-repository";
import type { AuthRepository } from "./repository";
import { AuthService } from "./service";

let cachedInstance: AuthRepository | null = null;

export function getAuthRepository(): AuthRepository {
  if (cachedInstance) return cachedInstance;

  const baseURL =
    process.env.CONTROL_PLANE_URL ??
    process.env.SHAHRAH_CONTROL_PLANE_URL ??
    "http://localhost:8080";
  const apiKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY ?? "";

  if (!apiKey) {
    throw new Error(
      "[tenant-admin][auth] SHAHRAH_MASTER_KEY (or MASTER_KEY) is required when using API auth.",
    );
  }

  cachedInstance = new ApiAuthRepository(baseURL, apiKey);
  return cachedInstance;
}

/**
 * Reset the cached instance (useful for tests).
 */
export function resetAuthRepository() {
  cachedInstance = null;
}

let cachedService: AuthService | null = null;

export function getAuthService(): AuthService {
  if (cachedService) return cachedService;
  cachedService = new AuthService(getAuthRepository());
  return cachedService;
}

export function resetAuthService() {
  cachedService = null;
}
