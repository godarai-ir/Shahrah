/**
 * InMemoryAuthRepository - In-memory implementation for unit tests.
 * All state is ephemeral and lost when the instance is garbage collected.
 */

import { randomBytes, randomInt, randomUUID } from "crypto";

import type { AuthRepository } from "./repository";
import type {
  AuthChannel,
  CaptchaChallenge,
  ConsumeResult,
  CreateOTPParams,
  CreateOTPResult,
  CreateSessionParams,
  FindOrCreateUserParams,
  NormalizedIdentifier,
  OTPRecord,
  TenantSession,
  TenantUser,
  UpdateUserContactParams,
  VerifyOTPParams,
  VerifyResult,
} from "./types";

const CAPTCHA_TTL_MS = 2 * 60 * 1000;
const OTP_TTL_MS = 5 * 60 * 1000;
const SESSION_TTL_MS = 7 * 24 * 60 * 60 * 1000;
const MAX_OTP_ATTEMPTS = 5;

export class InMemoryAuthRepository implements AuthRepository {
  private captchas = new Map<string, { answer: string; expiresAt: Date }>();
  private otpRequests = new Map<string, OTPRecord & { expiresAt: Date }>();
  private sessions = new Map<string, TenantSession & { expiresAt: Date }>();
  private users = new Map<string, TenantUser>();

  // Test helpers
  clear() {
    this.captchas.clear();
    this.otpRequests.clear();
    this.sessions.clear();
    this.users.clear();
  }

  addTestUser(user: TenantUser) {
    this.users.set(`${user.tenantSlug}:${user.id}`, user);
  }

  // CAPTCHA
  async createCaptchaChallenge(): Promise<CaptchaChallenge> {
    const id = randomUUID();
    const answer = this.generateCaptchaText(5);
    const expiresAt = new Date(Date.now() + CAPTCHA_TTL_MS);
    this.captchas.set(id, { answer, expiresAt });
    return {
      captcha_id: id,
      image_data_url: `data:image/svg+xml;base64,${Buffer.from(`<svg><text>${answer}</text></svg>`).toString("base64")}`,
      expires_in_seconds: Math.floor(CAPTCHA_TTL_MS / 1000),
    };
  }

  async consumeCaptcha(id: string, answer: string): Promise<ConsumeResult> {
    const entry = this.captchas.get(id);
    if (!entry) return { ok: false, error: "کپچا نامعتبر است یا منقضی شده است." };
    this.captchas.delete(id);
    if (entry.expiresAt.getTime() <= Date.now()) {
      return { ok: false, error: "زمان کپچا تمام شده است." };
    }
    if (answer.trim().toUpperCase() !== entry.answer) {
      return { ok: false, error: "کپچا اشتباه است." };
    }
    return { ok: true };
  }

  // OTP
  async createOTPRequest(input: CreateOTPParams): Promise<CreateOTPResult> {
    const id = randomUUID();
    const otpCode = `${randomInt(100000, 1000000)}`;
    const expiresAt = new Date(Date.now() + OTP_TTL_MS);
    const record: OTPRecord & { expiresAt: Date } = {
      id,
      tenantSlug: input.tenantSlug,
      purpose: input.purpose,
      channel: input.channel,
      identifier: input.identifier,
      userID: input.userID,
      pendingValue: input.pendingValue,
      otpCode,
      attempts: 0,
      expiresAtISO: expiresAt.toISOString(),
      createdAtISO: new Date().toISOString(),
      expiresAt,
    };
    this.otpRequests.set(id, record);
    // Skip email sending in memory implementation
    return {
      request_id: id,
      channel: input.channel,
      destination_masked: this.maskIdentifier(input.channel, input.identifier),
      expires_in_seconds: Math.floor(OTP_TTL_MS / 1000),
      debug_otp: otpCode, // Always expose for testing
    };
  }

  async verifyOTPRequest(input: VerifyOTPParams): Promise<VerifyResult> {
    const record = this.otpRequests.get(input.requestID);
    if (!record) return { ok: false, error: "درخواست OTP معتبر نیست یا منقضی شده است." };
    if (record.tenantSlug !== input.tenantSlug) {
      return { ok: false, error: "این درخواست OTP مربوط به Tenant فعلی نیست." };
    }
    if (record.purpose !== input.purpose) {
      return { ok: false, error: "کاربرد OTP با درخواست فعلی هم‌خوانی ندارد." };
    }
    if (input.userID && record.userID !== input.userID) {
      return { ok: false, error: "این درخواست OTP متعلق به کاربر فعلی نیست." };
    }
    if (record.expiresAt.getTime() <= Date.now()) {
      this.otpRequests.delete(input.requestID);
      return { ok: false, error: "کد OTP منقضی شده است." };
    }
    const normalizedCode = input.otpCode.trim();
    if (normalizedCode !== record.otpCode) {
      const nextAttempts = record.attempts + 1;
      if (nextAttempts >= MAX_OTP_ATTEMPTS) {
        this.otpRequests.delete(input.requestID);
        return { ok: false, error: "تعداد تلاش بیش از حد مجاز است. دوباره OTP بگیرید." };
      }
      record.attempts = nextAttempts;
      this.otpRequests.set(input.requestID, record);
      return { ok: false, error: "کد OTP اشتباه است." };
    }
    this.otpRequests.delete(input.requestID);
    const publicRecord = Object.fromEntries(
      Object.entries(record).filter(([key]) => key !== "expiresAt"),
    ) as OTPRecord;
    return { ok: true, record: publicRecord };
  }

  // Session
  async createSession(input: CreateSessionParams): Promise<TenantSession> {
    const id = randomBytes(24).toString("hex");
    const expiresAt = new Date(Date.now() + SESSION_TTL_MS);
    const session: TenantSession & { expiresAt: Date } = {
      id,
      tenantSlug: input.tenantSlug,
      userID: input.userID,
      createdAtISO: new Date().toISOString(),
      expiresAtISO: expiresAt.toISOString(),
      expiresAt,
    };
    this.sessions.set(id, session);
    return { id: session.id, tenantSlug: session.tenantSlug, userID: session.userID, createdAtISO: session.createdAtISO, expiresAtISO: session.expiresAtISO };
  }

  async getSession(sessionID: string): Promise<TenantSession | null> {
    const session = this.sessions.get(sessionID);
    if (!session) return null;
    if (session.expiresAt.getTime() <= Date.now()) {
      this.sessions.delete(sessionID);
      return null;
    }
    return { id: session.id, tenantSlug: session.tenantSlug, userID: session.userID, createdAtISO: session.createdAtISO, expiresAtISO: session.expiresAtISO };
  }

  async revokeSession(sessionID: string): Promise<void> {
    this.sessions.delete(sessionID);
  }

  // User
  async findOrCreateUser(input: FindOrCreateUserParams): Promise<TenantUser> {
    const key = this.findUserKey(input.tenantSlug, input.channel, input.identifier);
    if (key) {
      return this.users.get(key)!;
    }
    const user: TenantUser = {
      id: randomUUID(),
      tenantSlug: input.tenantSlug,
      email: input.channel === "email" ? input.identifier : undefined,
      mobile: input.channel === "mobile" ? input.identifier : undefined,
      role: this.users.size === 0 ? "tenant_admin" : "tenant_member",
      createdAtISO: new Date().toISOString(),
      updatedAtISO: new Date().toISOString(),
    };
    this.users.set(`${user.tenantSlug}:${user.id}`, user);
    // Also index by identifier
    this.users.set(this.makeIdentifierKey(input.tenantSlug, input.channel, input.identifier), user);
    return user;
  }

  async getUser(tenantSlug: string, userID: string): Promise<TenantUser | null> {
    return this.users.get(`${tenantSlug}:${userID}`) ?? null;
  }

  async updateUserContact(input: UpdateUserContactParams): Promise<TenantUser> {
    const key = `${input.tenantSlug}:${input.userID}`;
    const user = this.users.get(key);
    if (!user) throw new Error("کاربر یافت نشد.");
    if (input.field === "email") {
      user.email = input.value;
    } else {
      user.mobile = input.value;
    }
    user.updatedAtISO = new Date().toISOString();
    this.users.set(key, user);
    return user;
  }

  async touchLastLogin(tenantSlug: string, userID: string): Promise<TenantUser | null> {
    const key = `${tenantSlug}:${userID}`;
    const user = this.users.get(key);
    if (!user) return null;
    user.lastLoginAtISO = new Date().toISOString();
    user.updatedAtISO = new Date().toISOString();
    this.users.set(key, user);
    return user;
  }

  normalizeIdentifier(raw: string): NormalizedIdentifier | null {
    const value = raw.trim();
    if (!value) return null;
    const emailCandidate = value.toLowerCase();
    if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailCandidate)) {
      return { channel: "email", normalized: emailCandidate };
    }
    const compact = value.replace(/[\s\-()]/g, "");
    if (/^\+?\d{10,15}$/.test(compact)) {
      return { channel: "mobile", normalized: compact };
    }
    return null;
  }

  // Private helpers
  private findUserKey(tenantSlug: string, channel: AuthChannel, identifier: string): string | null {
    for (const [key, user] of this.users.entries()) {
      if (key.includes(":")) {
        if (user.tenantSlug === tenantSlug && user.id === key.split(":")[1]) {
          if (channel === "email" && user.email === identifier) return key;
          if (channel === "mobile" && user.mobile === identifier) return key;
        }
      }
    }
    return null;
  }

  private makeIdentifierKey(tenantSlug: string, channel: AuthChannel, identifier: string): string {
    return `${tenantSlug}:${channel}:${identifier}`;
  }

  private generateCaptchaText(length: number): string {
    const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
    let value = "";
    for (let i = 0; i < length; i += 1) {
      value += alphabet[randomInt(0, alphabet.length)];
    }
    return value;
  }

  private maskIdentifier(channel: AuthChannel, identifier: string): string {
    if (channel === "email") {
      const [local, domain = ""] = identifier.split("@");
      if (!local) return `***@${domain}`;
      if (local.length <= 2) return `${local[0] ?? "*"}***@${domain}`;
      return `${local.slice(0, 2)}***@${domain}`;
    }
    const compact = identifier.replace(/\s+/g, "");
    if (compact.length <= 4) return "***";
    return `${compact.slice(0, 3)}****${compact.slice(-3)}`;
  }
}
