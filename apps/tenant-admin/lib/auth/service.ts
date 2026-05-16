/**
 * AuthService - Orchestrator for authentication operations.
 * Depends on AuthRepository port and NotificationPort (email/sms).
 * Provides high-level auth flows used by API routes.
 */

import type { AuthRepository } from "./repository";
import type {
  CaptchaChallenge,
  ConsumeResult,
  CreateOTPResult,
  NormalizedIdentifier,
  OTPRecord,
  TenantSession,
  TenantUser,
} from "./types";

export type RequestOTPInput = {
  identifier: string;
  captchaID: string;
  captchaAnswer: string;
  tenantSlug: string;
};

export type VerifyOTPLoginInput = {
  tenantSlug: string;
  requestID: string;
  otpCode: string;
};

export type UpdateContactInput = {
  tenantSlug: string;
  userID: string;
  field: "email" | "mobile";
  value: string;
};

export type OTPVerificationResult = {
  ok: boolean;
  status?: number;
  error?: string;
  record?: OTPRecord;
  session?: TenantSession;
  user?: TenantUser;
};

export class AuthService {
  constructor(private repo: AuthRepository) {}

  // CAPTCHA flows
  async createCaptcha(): Promise<CaptchaChallenge> {
    return this.repo.createCaptchaChallenge();
  }

  async consumeCaptcha(captchaID: string, answer: string): Promise<ConsumeResult> {
    return this.repo.consumeCaptcha(captchaID, answer);
  }

  // OTP flows
  async requestOTP(input: RequestOTPInput): Promise<CreateOTPResult> {
    // Validate captcha first
    const captchaResult = await this.consumeCaptcha(input.captchaID, input.captchaAnswer);
    if (!captchaResult.ok) {
      throw new Error(captchaResult.error);
    }

    const normalized = this.repo.normalizeIdentifier(input.identifier);
    if (!normalized) {
      throw new Error("ایمیل یا شماره همراه معتبر وارد کنید.");
    }

    return this.repo.createOTPRequest({
      tenantSlug: input.tenantSlug,
      purpose: "login",
      channel: normalized.channel,
      identifier: normalized.normalized,
    });
  }

  async verifyOTPForLogin(input: VerifyOTPLoginInput): Promise<OTPVerificationResult> {
    const verifyResult = await this.repo.verifyOTPRequest({
      tenantSlug: input.tenantSlug,
      requestID: input.requestID,
      otpCode: input.otpCode,
      purpose: "login",
    });

    if (!verifyResult.ok) {
      return { ok: false, error: verifyResult.error };
    }

    // Create or find user
    const record = verifyResult.record;
    const user = await this.repo.findOrCreateUser({
      tenantSlug: input.tenantSlug,
      channel: record.channel,
      identifier: record.identifier,
    });
    if (!user.active) {
      return {
        ok: false,
        status: 403,
        error: "این حساب کاربری غیرفعال است.",
      };
    }

    // Touch last login is best-effort and must not block successful login.
    try {
      await this.repo.touchLastLogin(input.tenantSlug, user.id);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      console.warn(
        `[tenant-admin][auth] touch-last-login failed tenant=${input.tenantSlug} user=${user.id}: ${message}`,
      );
    }

    // Create session
    const session = await this.repo.createSession({
      tenantSlug: input.tenantSlug,
      userID: user.id,
    });

    return { ok: true, record, session, user };
  }

  // Session management
  async getSession(sessionID: string): Promise<TenantSession | null> {
    return this.repo.getSession(sessionID);
  }

  async revokeSession(sessionID: string): Promise<void> {
    return this.repo.revokeSession(sessionID);
  }

  async getAuthContext(sessionID: string, hintedTenantSlug: string | null): Promise<{
    tenantSlug: string;
    session: TenantSession;
    user: TenantUser;
  } | null> {
    const session = await this.repo.getSession(sessionID);
    if (!session) return null;

    const tenantSlug =
      hintedTenantSlug === session.tenantSlug ? hintedTenantSlug : session.tenantSlug;

    const user = await this.repo.getUser(tenantSlug, session.userID);
    if (!user) {
      await this.repo.revokeSession(session.id);
      return null;
    }
    if (!user.active) {
      await this.repo.revokeSession(session.id);
      return null;
    }

    return { tenantSlug, session, user };
  }

  // User management
  async getUser(tenantSlug: string, userID: string): Promise<TenantUser | null> {
    return this.repo.getUser(tenantSlug, userID);
  }

  async updateContact(input: UpdateContactInput): Promise<TenantUser> {
    return this.repo.updateUserContact({
      tenantSlug: input.tenantSlug,
      userID: input.userID,
      field: input.field,
      value: input.value,
    });
  }

  async requestContactChangeOTP(input: {
    tenantSlug: string;
    userID: string;
    field: "email" | "mobile";
    identifier: string;
  }): Promise<CreateOTPResult> {
    const normalized = this.repo.normalizeIdentifier(input.identifier);
    if (!normalized) {
      throw new Error("ایمیل یا شماره همراه معتبر وارد کنید.");
    }

    return this.repo.createOTPRequest({
      tenantSlug: input.tenantSlug,
      purpose: input.field === "email" ? "change_email" : "change_mobile",
      channel: normalized.channel,
      identifier: normalized.normalized,
      userID: input.userID,
      pendingValue: normalized.normalized,
    });
  }

  async verifyContactChangeOTP(input: {
    tenantSlug: string;
    userID: string;
    requestID: string;
    otpCode: string;
    field: "email" | "mobile";
  }): Promise<TenantUser> {
    const verifyResult = await this.repo.verifyOTPRequest({
      tenantSlug: input.tenantSlug,
      requestID: input.requestID,
      otpCode: input.otpCode,
      purpose: input.field === "email" ? "change_email" : "change_mobile",
      userID: input.userID,
    });

    if (!verifyResult.ok) {
      throw new Error(verifyResult.error);
    }

    const pendingValue = verifyResult.record.pendingValue ?? verifyResult.record.identifier;
    return this.repo.updateUserContact({
      tenantSlug: input.tenantSlug,
      userID: input.userID,
      field: input.field,
      value: pendingValue,
    });
  }

  // Utility
  normalizeIdentifier(raw: string): NormalizedIdentifier | null {
    return this.repo.normalizeIdentifier(raw);
  }
}
