/**
 * ApiAuthRepository - HTTP adapter that calls Go control-plane for auth operations.
 * Implements AuthRepository by making REST calls to the backend.
 */

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
} from "./types";
import type { AuthRepository } from "./repository";
import { z } from "zod";

const captchaChallengeSchema = z.object({
  captcha_id: z.string(), image_data_url: z.string(), expires_in_seconds: z.number(),
});
const consumeResultSchema = z.discriminatedUnion("ok", [
  z.object({ ok: z.literal(true) }),
  z.object({ ok: z.literal(false), error: z.string() }),
]);
const createOTPResultSchema = z.object({
  request_id: z.string(), channel: z.enum(["email", "mobile"]), destination_masked: z.string(),
  expires_in_seconds: z.number(), debug_otp: z.string().optional(),
});
const rawOTPRecordSchema = z.object({
  id: z.string(), tenant_slug: z.string(), purpose: z.enum(["login", "change_email", "change_mobile"]),
  channel: z.enum(["email", "mobile"]), identifier: z.string(),
  user_id: z.string().nullish().transform((value) => value ?? undefined),
  pending_value: z.string().nullish().transform((value) => value ?? undefined),
  otp_code: z.string(), attempts: z.number(),
  expires_at_iso: z.string(), created_at_iso: z.string(),
});
const verifyOTPResponseSchema = z.object({
  ok: z.boolean(), error: z.string().optional(), record: rawOTPRecordSchema.optional(),
});
const rawSessionSchema = z.object({
  id: z.string(), tenant_slug: z.string(), user_id: z.string(), created_at_iso: z.string(), expires_at_iso: z.string(),
});
const rawUserSchema = z.object({
  id: z.string(), tenant_slug: z.string(), email: z.string().optional(), mobile: z.string().optional(),
  social_providers: z.array(z.string()).optional(), role: z.string(), active: z.boolean().optional(),
  created_at_iso: z.string(), updated_at_iso: z.string(), last_login_at_iso: z.string().optional(),
});

export class ApiAuthRepository implements AuthRepository {
  constructor(
    private baseURL: string,
    private apiKey: string,
  ) {}

  private headers(): Record<string, string> {
    return {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${this.apiKey}`,
    };
  }

  private async post<T>(path: string, body: unknown, schema: z.ZodType<T>): Promise<T> {
    const res = await fetch(`${this.baseURL}${path}`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
    const payload = await res.json().catch(() => null) as unknown;
    if (!res.ok) {
      throw new Error(
        `Auth API error: ${res.status} ${res.statusText}${payload ? ` - ${JSON.stringify(payload)}` : ""}`,
      );
    }
    const parsed = schema.safeParse(payload);
    if (!parsed.success) throw new Error(`Unexpected auth API response contract from ${path}`);
    return parsed.data;
  }

  private async get<T>(path: string, schema: z.ZodType<T>): Promise<T | null> {
    const res = await fetch(`${this.baseURL}${path}`, {
      method: "GET",
      headers: this.headers(),
    });
    if (!res.ok) {
      if (res.status === 404) return null;
      const errBody = await res.json().catch(() => null);
      throw new Error(
        `Auth API error: ${res.status} ${res.statusText}${errBody ? ` - ${JSON.stringify(errBody)}` : ""}`,
      );
    }
    const payload = await res.json().catch(() => null) as unknown;
    const parsed = schema.safeParse(payload);
    if (!parsed.success) throw new Error(`Unexpected auth API response contract from ${path}`);
    return parsed.data;
  }

  private async delete(path: string): Promise<void> {
    const res = await fetch(`${this.baseURL}${path}`, {
      method: "DELETE",
      headers: this.headers(),
    });
    if (!res.ok) {
      const errBody = await res.json().catch(() => null);
      throw new Error(
        `Auth API error: ${res.status} ${res.statusText}${errBody ? ` - ${JSON.stringify(errBody)}` : ""}`,
      );
    }
  }

  // CAPTCHA
  async createCaptchaChallenge(): Promise<CaptchaChallenge> {
    return this.post("/api/v1/auth/captcha/create", {}, captchaChallengeSchema);
  }

  async consumeCaptcha(id: string, answer: string): Promise<ConsumeResult> {
    return this.post("/api/v1/auth/captcha/consume", {
      captcha_id: id,
      answer,
    }, consumeResultSchema);
  }

  // OTP
  async createOTPRequest(params: CreateOTPParams): Promise<CreateOTPResult> {
    return this.post("/api/v1/auth/otp/request", {
      tenant_slug: params.tenantSlug,
      purpose: params.purpose,
      channel: params.channel,
      identifier: params.identifier,
      user_id: params.userID,
      pending_value: params.pendingValue,
    }, createOTPResultSchema);
  }

  async verifyOTPRequest(params: VerifyOTPParams): Promise<VerifyResult> {
    const res = await this.post("/api/v1/auth/otp/verify", {
      tenant_slug: params.tenantSlug,
      request_id: params.requestID,
      otp_code: params.otpCode,
      purpose: params.purpose,
      user_id: params.userID,
    }, verifyOTPResponseSchema);

    if (!res.ok || !res.record) {
      return { ok: false, error: res.error ?? "Verification failed" };
    }

    return {
      ok: true,
      record: {
        id: res.record.id,
        tenantSlug: res.record.tenant_slug,
        purpose: res.record.purpose,
        channel: res.record.channel,
        identifier: res.record.identifier,
        userID: res.record.user_id,
        pendingValue: res.record.pending_value,
        otpCode: res.record.otp_code,
        attempts: res.record.attempts,
        expiresAtISO: res.record.expires_at_iso,
        createdAtISO: res.record.created_at_iso,
      },
    };
  }

  // Session
  async createSession(params: CreateSessionParams): Promise<TenantSession> {
    const response = await this.post("/api/v1/auth/session", {
      tenant_slug: params.tenantSlug,
      user_id: params.userID,
    }, rawSessionSchema);

    return {
      id: response.id,
      tenantSlug: response.tenant_slug,
      userID: response.user_id,
      createdAtISO: response.created_at_iso,
      expiresAtISO: response.expires_at_iso,
    };
  }

  async getSession(id: string): Promise<TenantSession | null> {
    const response = await this.get(`/api/v1/auth/session/${id}`, rawSessionSchema);

    if (!response) {
      return null;
    }

    return {
      id: response.id,
      tenantSlug: response.tenant_slug,
      userID: response.user_id,
      createdAtISO: response.created_at_iso,
      expiresAtISO: response.expires_at_iso,
    };
  }

  async revokeSession(id: string): Promise<void> {
    return this.delete(`/api/v1/auth/session/${id}`);
  }

  // User
  async findOrCreateUser(params: FindOrCreateUserParams): Promise<TenantUser> {
    const response = await this.post("/api/v1/auth/user/find-or-create", {
      tenant_slug: params.tenantSlug,
      channel: params.channel,
      identifier: params.identifier,
    }, rawUserSchema);

    return mapTenantUser(response);
  }

  async getUser(tenantSlug: string, userID: string): Promise<TenantUser | null> {
    const response = await this.get(
      `/api/v1/auth/user/${encodeURIComponent(userID)}?tenant_slug=${encodeURIComponent(tenantSlug)}`,
      rawUserSchema,
    );

    return response ? mapTenantUser(response) : null;
  }

  async updateUserContact(params: UpdateUserContactParams): Promise<TenantUser> {
    const response = await this.post(
      `/api/v1/auth/user/${encodeURIComponent(params.userID)}/update-contact`,
      {
        tenant_slug: params.tenantSlug,
        field: params.field,
        value: params.value,
      },
      rawUserSchema,
    );

    return mapTenantUser(response);
  }

  async touchLastLogin(
    tenantSlug: string,
    userID: string,
  ): Promise<TenantUser | null> {
    const response = await this.post(
      `/api/v1/auth/user/${encodeURIComponent(userID)}/touch-last-login`,
      { tenant_slug: tenantSlug },
      rawUserSchema.nullable(),
    );

    return response ? mapTenantUser(response) : null;
  }

  // Utility
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
}

function mapTenantUser(input: {
  id: string;
  tenant_slug: string;
  email?: string;
  mobile?: string;
  social_providers?: string[];
  role: string;
  active?: boolean;
  created_at_iso: string;
  updated_at_iso: string;
  last_login_at_iso?: string;
}): TenantUser {
  return {
    id: input.id,
    tenantSlug: input.tenant_slug,
    email: input.email,
    mobile: input.mobile,
    socialProviders: Array.isArray(input.social_providers)
      ? input.social_providers.filter((provider): provider is string => typeof provider === "string")
      : undefined,
    role: input.role,
    active: input.active ?? true,
    createdAtISO: input.created_at_iso,
    updatedAtISO: input.updated_at_iso,
    lastLoginAtISO: input.last_login_at_iso,
  };
}
