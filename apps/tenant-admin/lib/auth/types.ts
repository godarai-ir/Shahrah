/**
 * Shared types for the decoupled authentication layer.
 * Extracted from store.ts to be used across repository implementations.
 */

export type AuthChannel = "email" | "mobile";
export type OTPPurpose = "login" | "change_email" | "change_mobile";

export type TenantUser = {
  id: string;
  tenantSlug: string;
  email?: string;
  mobile?: string;
  socialProviders?: string[];
  role: string;
  active: boolean;
  createdAtISO: string;
  updatedAtISO: string;
  lastLoginAtISO?: string;
};

export type TenantSession = {
  id: string;
  tenantSlug: string;
  userID: string;
  createdAtISO: string;
  expiresAtISO: string;
};

export type OTPRecord = {
  id: string;
  tenantSlug: string;
  purpose: OTPPurpose;
  channel: AuthChannel;
  identifier: string;
  userID?: string;
  pendingValue?: string;
  otpCode: string;
  attempts: number;
  expiresAtISO: string;
  createdAtISO: string;
};

export type CaptchaChallenge = {
  captcha_id: string;
  image_data_url: string;
  expires_in_seconds: number;
};

export type ConsumeResult =
  | { ok: true }
  | { ok: false; error: string };

export type VerifyResult =
  | { ok: true; record: OTPRecord }
  | { ok: false; error: string };

// Input/output DTOs
export type CreateOTPParams = {
  tenantSlug: string;
  purpose: OTPPurpose;
  channel: AuthChannel;
  identifier: string;
  userID?: string;
  pendingValue?: string;
};

export type CreateOTPResult = {
  request_id: string;
  channel: AuthChannel;
  destination_masked: string;
  expires_in_seconds: number;
  debug_otp?: string;
};

export type VerifyOTPParams = {
  tenantSlug: string;
  requestID: string;
  otpCode: string;
  purpose: OTPPurpose;
  userID?: string;
};

export type CreateSessionParams = {
  tenantSlug: string;
  userID: string;
};

export type FindOrCreateUserParams = {
  tenantSlug: string;
  channel: AuthChannel;
  identifier: string;
};

export type UpdateUserContactParams = {
  tenantSlug: string;
  userID: string;
  field: "email" | "mobile";
  value: string;
};

export type NormalizedIdentifier = {
  channel: AuthChannel;
  normalized: string;
};

// JWT Token payload for token-based auth
export type TenantAdminTokenPayload = {
  sub: string;        // user ID
  tenant: string;     // tenant slug
  role: string;
  session_id: string;
  iat: number;
  exp: number;
};
