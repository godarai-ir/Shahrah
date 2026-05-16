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
  public revokeCallCount = 0;
  public active = true;
  public sessionUserID = "user-1";
  public createSessionCallCount = 0;
  public membershipAllowedForTenant = new Map<string, boolean>();
  public userByTenantAndID = new Map<string, TenantUser>();

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
    this.createSessionCallCount += 1;
    return {
      id: "sess-1",
      tenantSlug: params.tenantSlug,
      userID: params.userID,
      createdAtISO: new Date().toISOString(),
      expiresAtISO: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    };
  }

  async getSession(): Promise<TenantSession | null> {
    return {
      id: "sess-1",
      tenantSlug: "tenant1",
      userID: this.sessionUserID,
      createdAtISO: new Date().toISOString(),
      expiresAtISO: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    };
  }

  async revokeSession(): Promise<void> {
    this.revokeCallCount += 1;
  }

  async findOrCreateUser(params: FindOrCreateUserParams): Promise<TenantUser> {
    if (this.membershipAllowedForTenant.get(params.tenantSlug) === false) {
      throw new Error("tenant member not found");
    }
    return {
      id: "user-1",
      tenantSlug: params.tenantSlug,
      email: "admin@tenant1.com",
      role: "tenant_admin",
      active: this.active,
      createdAtISO: new Date().toISOString(),
      updatedAtISO: new Date().toISOString(),
    };
  }

  async getUser(tenantSlug: string, userID: string): Promise<TenantUser | null> {
    return this.userByTenantAndID.get(`${tenantSlug}:${userID}`) ?? null;
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

test("active tenant user can log in", async () => {
  const repo = new FakeAuthRepository();
  repo.active = true;
  const service = new AuthService(repo);

  const result = await service.verifyOTPForLogin({
    tenantSlug: "tenant1",
    requestID: "req-1",
    otpCode: "123456",
  });

  assert.equal(result.ok, true);
  assert.equal(result.user?.active, true);
});

test("deactivated tenant user cannot log in", async () => {
  const repo = new FakeAuthRepository();
  repo.active = false;
  const service = new AuthService(repo);

  const result = await service.verifyOTPForLogin({
    tenantSlug: "tenant1",
    requestID: "req-1",
    otpCode: "123456",
  });

  assert.equal(result.ok, false);
  assert.equal(result.status, 403);
});

test("deactivated tenant user with existing session cannot access auth context", async () => {
  const repo = new FakeAuthRepository();
  repo.sessionUserID = "user-1";
  repo.userByTenantAndID.set("tenant1:user-1", {
    id: "user-1",
    tenantSlug: "tenant1",
    email: "user1@tenant1.com",
    role: "member",
    active: false,
    createdAtISO: new Date().toISOString(),
    updatedAtISO: new Date().toISOString(),
  });
  const service = new AuthService(repo);

  const ctx = await service.getAuthContext("sess-1", "tenant1");
  assert.equal(ctx, null);
  assert.equal(repo.revokeCallCount, 1);
});

test("same email in another tenant is isolated", async () => {
  const repo = new FakeAuthRepository();
  repo.membershipAllowedForTenant.set("tenant2", false);
  const service = new AuthService(repo);

  const result = await service.verifyOTPForLogin({
    tenantSlug: "tenant2",
    requestID: "req-1",
    otpCode: "123456",
  });

  assert.equal(result.ok, false);
  assert.equal(result.status, 403);
  assert.equal(repo.createSessionCallCount, 0);
});

test("non-member cannot log in and no session is issued", async () => {
  const repo = new FakeAuthRepository();
  repo.membershipAllowedForTenant.set("tenant1", false);
  const service = new AuthService(repo);

  const result = await service.verifyOTPForLogin({
    tenantSlug: "tenant1",
    requestID: "req-1",
    otpCode: "123456",
  });

  assert.equal(result.ok, false);
  assert.equal(result.status, 403);
  assert.equal(repo.createSessionCallCount, 0);
});

test("reactivation restores access when user becomes active", async () => {
  const repo = new FakeAuthRepository();
  repo.sessionUserID = "user-1";
  repo.userByTenantAndID.set("tenant1:user-1", {
    id: "user-1",
    tenantSlug: "tenant1",
    email: "user1@tenant1.com",
    role: "member",
    active: true,
    createdAtISO: new Date().toISOString(),
    updatedAtISO: new Date().toISOString(),
  });
  const service = new AuthService(repo);

  const ctx = await service.getAuthContext("sess-1", "tenant1");
  assert.ok(ctx);
  assert.equal(ctx?.user.active, true);
});
