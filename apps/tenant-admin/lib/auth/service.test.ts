import test from "node:test";
import assert from "node:assert/strict";

import { AuthService } from "./service.ts";
import type { AuthRepository } from "./repository.ts";
import type {
  CaptchaChallenge,
  ConsumeResult,
  CreateOTPParams,
  CreateOTPResult,
  CreateSessionParams,
  FindOrCreateUserParams,
  NormalizedIdentifier,
  TenantSession,
  TenantUser,
  UpdateUserContactParams,
  VerifyOTPParams,
  VerifyResult,
} from "./types.ts";

class FakeAuthRepository implements AuthRepository {
  public touchCallCount = 0;
  public shouldFailTouch = false;

  async createCaptchaChallenge(): Promise<CaptchaChallenge> {
    return {
      captcha_id: "captcha-1",
      image_data_url: "data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=",
      expires_in_seconds: 120,
    };
  }

  async consumeCaptcha(): Promise<ConsumeResult> {
    return { ok: true };
  }

  async createOTPRequest(params: CreateOTPParams): Promise<CreateOTPResult> {
    void params;
    return {
      request_id: "req-1",
      channel: "email",
      destination_masked: "a***@tenant1.com",
      expires_in_seconds: 300,
      debug_otp: "123456",
    };
  }

  async verifyOTPRequest(params: VerifyOTPParams): Promise<VerifyResult> {
    void params;
    return {
      ok: true,
      record: {
        id: "req-1",
        tenantSlug: "tenant1",
        purpose: "login",
        channel: "email",
        identifier: "admin@tenant1.com",
        otpCode: "123456",
        attempts: 0,
        expiresAtISO: new Date(Date.now() + 300_000).toISOString(),
        createdAtISO: new Date().toISOString(),
      },
    };
  }

  async createSession(params: CreateSessionParams): Promise<TenantSession> {
    return {
      id: "sess-1",
      tenantSlug: params.tenantSlug,
      userID: params.userID,
      createdAtISO: new Date().toISOString(),
      expiresAtISO: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    };
  }

  async getSession(): Promise<TenantSession | null> {
    return null;
  }

  async revokeSession(): Promise<void> {}

  async findOrCreateUser(params: FindOrCreateUserParams): Promise<TenantUser> {
    void params;
    return {
      id: "user-1",
      tenantSlug: "tenant1",
      email: "admin@tenant1.com",
      role: "tenant_admin",
      createdAtISO: new Date().toISOString(),
      updatedAtISO: new Date().toISOString(),
    };
  }

  async getUser(): Promise<TenantUser | null> {
    return null;
  }

  async updateUserContact(params: UpdateUserContactParams): Promise<TenantUser> {
    void params;
    throw new Error("not used");
  }

  async touchLastLogin(): Promise<TenantUser | null> {
    this.touchCallCount += 1;
    if (this.shouldFailTouch) {
      throw new Error("Auth API error: 404 Not Found");
    }
    return null;
  }

  normalizeIdentifier(raw: string): NormalizedIdentifier | null {
    if (raw.includes("@")) {
      return { channel: "email", normalized: raw.toLowerCase() };
    }
    return null;
  }
}

test("verifyOTPForLogin succeeds even when touch-last-login returns 404", async () => {
  const repo = new FakeAuthRepository();
  repo.shouldFailTouch = true;
  const service = new AuthService(repo);

  const result = await service.verifyOTPForLogin({
    tenantSlug: "tenant1",
    requestID: "req-1",
    otpCode: "123456",
  });

  assert.equal(result.ok, true);
  assert.equal(result.user?.email, "admin@tenant1.com");
  assert.equal(result.session?.id, "sess-1");
  assert.equal(repo.touchCallCount, 1);
});
