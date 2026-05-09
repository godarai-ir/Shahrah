/**
 * AuthRepository - Port interface for authentication operations.
 * Defines the contract that all auth implementations must follow,
 * whether they use Prisma directly, HTTP API, or in-memory storage.
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

export interface AuthRepository {
  // CAPTCHA
  createCaptchaChallenge(): Promise<CaptchaChallenge>;
  consumeCaptcha(id: string, answer: string): Promise<ConsumeResult>;

  // OTP
  createOTPRequest(params: CreateOTPParams): Promise<CreateOTPResult>;
  verifyOTPRequest(params: VerifyOTPParams): Promise<VerifyResult>;

  // Session
  createSession(params: CreateSessionParams): Promise<TenantSession>;
  getSession(id: string): Promise<TenantSession | null>;
  revokeSession(id: string): Promise<void>;

  // User
  findOrCreateUser(params: FindOrCreateUserParams): Promise<TenantUser>;
  getUser(tenantSlug: string, userID: string): Promise<TenantUser | null>;
  updateUserContact(params: UpdateUserContactParams): Promise<TenantUser>;
  touchLastLogin(tenantSlug: string, userID: string): Promise<TenantUser | null>;

  // Utility
  normalizeIdentifier(raw: string): NormalizedIdentifier | null;
}
