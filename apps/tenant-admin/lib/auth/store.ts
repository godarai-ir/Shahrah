import { Prisma } from "@prisma/client";
import { randomBytes, randomInt, randomUUID } from "crypto";

import { sendOTPEmail } from "@/lib/email";
import { prisma } from "@/lib/prisma";

export type AuthChannel = "email" | "mobile";
export type OTPPurpose = "login" | "change_email" | "change_mobile";

export type TenantUser = {
  id: string;
  tenantSlug: string;
  email?: string;
  mobile?: string;
  role: string;
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

type OTPRecord = {
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

type DBTenantUserRow = {
  id: string;
  tenantSlug: string;
  email: string | null;
  mobile: string | null;
  role: string | null;
  createdAt: Date;
  updatedAt: Date;
  lastLoginAt: Date | null;
};

const CAPTCHA_TTL_MS = 2 * 60 * 1000;
const OTP_TTL_MS = 5 * 60 * 1000;
const SESSION_TTL_MS = 7 * 24 * 60 * 60 * 1000;
const MAX_OTP_ATTEMPTS = 5;
let databaseReadyPromise: Promise<void> | null = null;

export type NormalizedIdentifier = {
  channel: AuthChannel;
  normalized: string;
};

export function normalizeIdentifier(raw: string): NormalizedIdentifier | null {
  const value = raw.trim();
  if (!value) {
    return null;
  }

  const emailCandidate = value.toLowerCase();
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailCandidate)) {
    return {
      channel: "email",
      normalized: emailCandidate,
    };
  }

  const compact = value.replace(/[\s\-()]/g, "");
  if (/^\+?\d{10,15}$/.test(compact)) {
    return {
      channel: "mobile",
      normalized: compact,
    };
  }

  return null;
}

export async function createCaptchaChallenge() {
  await cleanupExpiredRecords();

  const id = randomUUID();
  const answer = generateCaptchaText(5);
  const expiresAt = new Date(Date.now() + CAPTCHA_TTL_MS);

  await prisma.tenantAdminCaptcha.create({
    data: {
      id,
      answer,
      expiresAt,
    },
  });

  return {
    captcha_id: id,
    image_data_url: buildCaptchaDataURL(answer),
    expires_in_seconds: Math.floor(CAPTCHA_TTL_MS / 1000),
  };
}

export async function consumeCaptcha(captchaID: string, answer: string) {
  await cleanupExpiredRecords();

  const record = await prisma.tenantAdminCaptcha.findUnique({
    where: { id: captchaID },
  });
  if (!record) {
    return {
      ok: false as const,
      error: "کپچا نامعتبر است یا منقضی شده است.",
    };
  }

  await prisma.tenantAdminCaptcha.delete({ where: { id: captchaID } });

  if (record.expiresAt.getTime() <= Date.now()) {
    return {
      ok: false as const,
      error: "زمان کپچا تمام شده است.",
    };
  }

  const normalizedAnswer = answer.trim().toUpperCase();
  if (normalizedAnswer !== record.answer) {
    return {
      ok: false as const,
      error: "کپچا اشتباه است.",
    };
  }

  return {
    ok: true as const,
  };
}

export async function createOTPRequest(input: {
  tenantSlug: string;
  purpose: OTPPurpose;
  channel: AuthChannel;
  identifier: string;
  userID?: string;
  pendingValue?: string;
}) {
  await cleanupExpiredRecords();

  const id = randomUUID();
  const otpCode = `${randomInt(100000, 1000000)}`;
  const expiresAt = new Date(Date.now() + OTP_TTL_MS);

  await prisma.tenantAdminOtpRequest.create({
    data: {
      id,
      tenantSlug: input.tenantSlug,
      purpose: input.purpose,
      channel: input.channel,
      identifier: input.identifier,
      userID: input.userID,
      pendingValue: input.pendingValue,
      otpCode,
      attempts: 0,
      expiresAt,
    },
  });

  // جایگزین موقت برای سرویس ارسال OTP
  if (input.channel === "email") {
    await sendOTPEmail({
      to: input.identifier,
      otpCode,
      expiresInMinutes: Math.floor(OTP_TTL_MS / 60000),
      tenantSlug: input.tenantSlug,
    });
  } else {
    console.info(
      `[tenant-admin][otp] tenant=${input.tenantSlug} purpose=${input.purpose} channel=${input.channel} to=${input.identifier} code=${otpCode}`,
    );
  }

  return {
    request_id: id,
    channel: input.channel,
    destination_masked: maskIdentifier(input.channel, input.identifier),
    expires_in_seconds: Math.floor(OTP_TTL_MS / 1000),
    debug_otp: process.env.NODE_ENV === "production" ? undefined : otpCode,
  };
}

export async function verifyOTPRequest(input: {
  tenantSlug: string;
  requestID: string;
  otpCode: string;
  purpose: OTPPurpose;
  userID?: string;
}) {
  await cleanupExpiredRecords();

  const record = await prisma.tenantAdminOtpRequest.findUnique({
    where: { id: input.requestID },
  });
  if (!record) {
    return {
      ok: false as const,
      error: "درخواست OTP معتبر نیست یا منقضی شده است.",
    };
  }

  if (record.tenantSlug !== input.tenantSlug) {
    return {
      ok: false as const,
      error: "این درخواست OTP مربوط به Tenant فعلی نیست.",
    };
  }

  if (record.purpose !== input.purpose) {
    return {
      ok: false as const,
      error: "کاربرد OTP با درخواست فعلی هم‌خوانی ندارد.",
    };
  }

  if (input.userID && record.userID !== input.userID) {
    return {
      ok: false as const,
      error: "این درخواست OTP متعلق به کاربر فعلی نیست.",
    };
  }

  if (record.expiresAt.getTime() <= Date.now()) {
    await prisma.tenantAdminOtpRequest.delete({ where: { id: record.id } });
    return {
      ok: false as const,
      error: "کد OTP منقضی شده است.",
    };
  }

  const normalizedCode = input.otpCode.trim();
  if (normalizedCode !== record.otpCode) {
    const nextAttempts = record.attempts + 1;
    if (nextAttempts >= MAX_OTP_ATTEMPTS) {
      await prisma.tenantAdminOtpRequest.delete({ where: { id: record.id } });
      return {
        ok: false as const,
        error: "تعداد تلاش بیش از حد مجاز است. دوباره OTP بگیرید.",
      };
    }
    await prisma.tenantAdminOtpRequest.update({
      where: { id: record.id },
      data: { attempts: nextAttempts },
    });
    return {
      ok: false as const,
      error: "کد OTP اشتباه است.",
    };
  }

  await prisma.tenantAdminOtpRequest.delete({ where: { id: record.id } });
  return {
    ok: true as const,
    record: toPublicOTPRecord(record),
  };
}

export async function findOrCreateTenantUser(input: {
  tenantSlug: string;
  channel: AuthChannel;
  identifier: string;
}) {
  await cleanupExpiredRecords();

  const existing = await findTenantUserByIdentifier(
    input.tenantSlug,
    input.channel,
    input.identifier,
  );

  if (existing) {
    return toPublicTenantUser(existing);
  }
  throw new Error("tenant member not found");
}

export async function touchLastLogin(tenantSlug: string, userID: string) {
  await ensureDatabaseReady();

  const current = await findTenantUserByID(tenantSlug, userID);
  if (!current) {
    return null;
  }

  const updated = await updateTenantUserLastLogin(tenantSlug, userID);
  return toPublicTenantUser(updated);
}

export async function getTenantUser(tenantSlug: string, userID: string) {
  await cleanupExpiredRecords();
  const user = await findTenantUserByID(tenantSlug, userID);
  if (!user) {
    return null;
  }
  return toPublicTenantUser(user);
}

export async function updateTenantUserContact(input: {
  tenantSlug: string;
  userID: string;
  field: "email" | "mobile";
  value: string;
}) {
  await cleanupExpiredRecords();

  if (input.field === "mobile") {
    return {
      ok: false as const,
      error: "ویرایش شماره همراه در اسکیمای فعلی پشتیبانی نمی‌شود.",
    };
  }

  const user = await findTenantUserByID(input.tenantSlug, input.userID);
  if (!user) {
    return {
      ok: false as const,
      error: "کاربر در Tenant فعلی یافت نشد.",
    };
  }

  const conflict = await findTenantUserByIdentifier(
    input.tenantSlug,
    input.field,
    input.value,
  );
  if (conflict) {
    return {
      ok: false as const,
      error:
        input.field === "email"
          ? "این ایمیل قبلا ثبت شده است."
          : "این شماره همراه قبلا ثبت شده است.",
    };
  }

  try {
    const updated = await updateTenantUserContactField(
      input.tenantSlug,
      input.userID,
      input.field,
      input.value,
    );

    return {
      ok: true as const,
      user: toPublicTenantUser(updated),
    };
  } catch (error) {
    if (isUniqueViolation(error)) {
      return {
        ok: false as const,
        error:
          input.field === "email"
            ? "این ایمیل قبلا ثبت شده است."
            : "این شماره همراه قبلا ثبت شده است.",
      };
    }
    throw error;
  }
}

export async function createTenantSession(input: {
  tenantSlug: string;
  userID: string;
}) {
  await cleanupExpiredRecords();

  const session = await prisma.tenantAdminSession.create({
    data: {
      id: randomBytes(24).toString("hex"),
      tenantSlug: input.tenantSlug,
      userID: input.userID,
      expiresAt: new Date(Date.now() + SESSION_TTL_MS),
    },
  });

  return toPublicSession(session);
}

export async function getTenantSession(sessionID: string) {
  await cleanupExpiredRecords();
  const record = await prisma.tenantAdminSession.findUnique({
    where: {
      id: sessionID,
    },
  });
  if (!record) {
    return null;
  }

  if (record.expiresAt.getTime() <= Date.now()) {
    await prisma.tenantAdminSession.delete({ where: { id: record.id } });
    return null;
  }

  return toPublicSession(record);
}

export async function revokeTenantSession(sessionID: string) {
  await ensureDatabaseReady();

  await prisma.tenantAdminSession.deleteMany({
    where: { id: sessionID },
  });
}

async function cleanupExpiredRecords() {
  await ensureDatabaseReady();

  const now = new Date();
  await Promise.all([
    prisma.tenantAdminCaptcha.deleteMany({
      where: {
        expiresAt: { lte: now },
      },
    }),
    prisma.tenantAdminOtpRequest.deleteMany({
      where: {
        expiresAt: { lte: now },
      },
    }),
    prisma.tenantAdminSession.deleteMany({
      where: {
        expiresAt: { lte: now },
      },
    }),
  ]);
}

async function getDefaultRoleForNewTenantUser(tenantSlug: string) {
  void tenantSlug;
  const rows = await prisma.$queryRaw<{ count: bigint }[]>(Prisma.sql`
    SELECT COUNT(*)::bigint AS count
    FROM users u
    WHERE COALESCE(u.role, '') <> 'deactivated'
  `);
  const count = Number(rows[0]?.count ?? 0);
  return count === 0 ? "tenant_admin" : "member";
}

async function ensureTenantHasAdmin(input: {
  tenantSlug: string;
  candidateUserID: string;
}) {
  const adminRows = await prisma.$queryRaw<{ id: string }[]>(Prisma.sql`
    SELECT u.user_id::text AS id
    FROM users u
    WHERE COALESCE(u.role, '') IN ('tenant_admin', 'proxy_admin')
    LIMIT 1
  `);
  if (adminRows.length > 0) {
    return null;
  }

  const activeRows = await prisma.$queryRaw<{ count: bigint }[]>(Prisma.sql`
    SELECT COUNT(*)::bigint AS count
    FROM users u
    WHERE COALESCE(u.role, '') <> 'deactivated'
  `);
  const activeUsers = Number(activeRows[0]?.count ?? 0);
  if (activeUsers === 0) {
    return null;
  }

  const updated = await prisma.$queryRaw<DBTenantUserRow[]>(Prisma.sql`
    UPDATE users AS u
    SET role = 'tenant_admin', updated_at = NOW()
    WHERE u.user_id::text = ${input.candidateUserID}
    RETURNING
      u.user_id::text AS id,
      ${input.tenantSlug}::text AS "tenantSlug",
      u.email,
      NULL::text AS mobile,
      u.role,
      u.created_at AS "createdAt",
      u.updated_at AS "updatedAt",
      NULL::timestamptz AS "lastLoginAt"
  `);
  return updated[0] ?? null;
}

async function findTenantUserByIdentifier(
  tenantSlug: string,
  channel: AuthChannel,
  identifier: string,
) {
  if (channel === "mobile") {
    return null;
  }

  const rows = await prisma.$queryRaw<DBTenantUserRow[]>(Prisma.sql`
    SELECT
      u.user_id::text AS id,
      t.slug AS "tenantSlug",
      u.email,
      NULL::text AS mobile,
      u.role,
      u.created_at AS "createdAt",
      u.updated_at AS "updatedAt",
      NULL::timestamptz AS "lastLoginAt"
    FROM users u
    JOIN tenants t ON t.tenant_id = u.tenant_id
    WHERE t.slug = ${tenantSlug}
      AND LOWER(COALESCE(u.email, '')) = LOWER(${identifier})
      AND COALESCE(u.active, true) = true
      AND LOWER(COALESCE(u.role, '')) NOT IN ('deactivated', 'removed', 'suspended')
    LIMIT 1
  `);
  return rows[0] ?? null;
}

async function findTenantUserByID(tenantSlug: string, userID: string) {
  const rows = await prisma.$queryRaw<DBTenantUserRow[]>(Prisma.sql`
    SELECT
      u.user_id::text AS id,
      t.slug AS "tenantSlug",
      u.email,
      NULL::text AS mobile,
      u.role,
      u.created_at AS "createdAt",
      u.updated_at AS "updatedAt",
      NULL::timestamptz AS "lastLoginAt"
    FROM users u
    JOIN tenants t ON t.tenant_id = u.tenant_id
    WHERE t.slug = ${tenantSlug}
      AND u.user_id::text = ${userID}
      AND COALESCE(u.active, true) = true
      AND LOWER(COALESCE(u.role, '')) NOT IN ('deactivated', 'removed', 'suspended')
    LIMIT 1
  `);
  return rows[0] ?? null;
}

async function createTenantUserFromIdentifier(
  tenantSlug: string,
  channel: AuthChannel,
  identifier: string,
  role: string,
) {
  if (channel === "mobile") {
    throw new Error("mobile auth is not supported by current users schema");
  }

  const rows = await prisma.$queryRaw<DBTenantUserRow[]>(Prisma.sql`
    INSERT INTO users (
      user_alias, email, role, created_at, updated_at
    )
    SELECT
      NULL,
      ${identifier},
      ${role},
      NOW(),
      NOW()
    RETURNING
      user_id::text AS id,
      ${tenantSlug}::text AS "tenantSlug",
      email,
      NULL::text AS mobile,
      role,
      created_at AS "createdAt",
      updated_at AS "updatedAt",
      NULL::timestamptz AS "lastLoginAt"
  `);
  if (rows.length === 0) {
    throw new Error("tenant not found");
  }
  return rows[0];
}

async function updateTenantUserLastLogin(tenantSlug: string, userID: string) {
  const rows = await prisma.$queryRaw<DBTenantUserRow[]>(Prisma.sql`
    UPDATE users AS u
    SET updated_at = NOW()
    FROM tenants t
    WHERE t.tenant_id = u.tenant_id
      AND t.slug = ${tenantSlug}
      AND u.user_id::text = ${userID}
    RETURNING
      u.user_id::text AS id,
      t.slug AS "tenantSlug",
      u.email,
      NULL::text AS mobile,
      u.role,
      u.created_at AS "createdAt",
      u.updated_at AS "updatedAt",
      NULL::timestamptz AS "lastLoginAt"
  `);
  if (rows.length === 0) {
    throw new Error("user not found");
  }
  return rows[0];
}

async function updateTenantUserContactField(
  tenantSlug: string,
  userID: string,
  field: "email" | "mobile",
  value: string,
) {
  if (field === "mobile") {
    throw new Error("mobile contact update is not supported by current users schema");
  }

  const query =
    Prisma.sql`
      UPDATE users AS u
      SET email = ${value}, updated_at = NOW()
      FROM tenants t
      WHERE t.tenant_id = u.tenant_id
        AND t.slug = ${tenantSlug}
        AND u.user_id::text = ${userID}
      RETURNING
        u.user_id::text AS id,
        t.slug AS "tenantSlug",
        u.email,
        NULL::text AS mobile,
        u.role,
        u.created_at AS "createdAt",
        u.updated_at AS "updatedAt",
        NULL::timestamptz AS "lastLoginAt"
    `;
  const rows = await prisma.$queryRaw<DBTenantUserRow[]>(query);
  if (rows.length === 0) {
    throw new Error("user not found");
  }
  return rows[0];
}

async function ensureDatabaseReady() {
  if (!databaseReadyPromise) {
    databaseReadyPromise = (async () => {
      // Keep this auth store self-contained in dev/new environments.
      // These statements are idempotent and only create missing tables/indexes.
      await ensureTenantAdminAuthSchema();
      await prisma.$queryRaw`SELECT 1`;
    })();
  }

  try {
    await databaseReadyPromise;
  } catch (error) {
    databaseReadyPromise = null;
    throw error;
  }
}

async function ensureTenantAdminAuthSchema() {
  await prisma.$executeRawUnsafe(`
    ALTER TABLE IF EXISTS "tenant_admin_sessions"
      DROP CONSTRAINT IF EXISTS "tenant_admin_sessions_userID_fkey"
  `);
  await prisma.$executeRawUnsafe(`
    ALTER TABLE IF EXISTS "tenant_admin_otp_requests"
      DROP CONSTRAINT IF EXISTS "tenant_admin_otp_requests_userID_fkey"
  `);
  await prisma.$executeRawUnsafe(`
    DROP TABLE IF EXISTS "tenant_admin_users"
  `);

  await prisma.$executeRawUnsafe(`
    CREATE TABLE IF NOT EXISTS "tenant_admin_captchas" (
      "id" TEXT NOT NULL,
      "answer" TEXT NOT NULL,
      "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
      "expiresAt" TIMESTAMP(3) NOT NULL,
      CONSTRAINT "tenant_admin_captchas_pkey" PRIMARY KEY ("id")
    )
  `);

  await prisma.$executeRawUnsafe(`
    CREATE TABLE IF NOT EXISTS "tenant_admin_sessions" (
      "id" TEXT NOT NULL,
      "tenantSlug" TEXT NOT NULL,
      "userID" TEXT NOT NULL,
      "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
      "expiresAt" TIMESTAMP(3) NOT NULL,
      CONSTRAINT "tenant_admin_sessions_pkey" PRIMARY KEY ("id")
    )
  `);

  await prisma.$executeRawUnsafe(`
    CREATE TABLE IF NOT EXISTS "tenant_admin_otp_requests" (
      "id" TEXT NOT NULL,
      "tenantSlug" TEXT NOT NULL,
      "purpose" TEXT NOT NULL,
      "channel" TEXT NOT NULL,
      "identifier" TEXT NOT NULL,
      "userID" TEXT,
      "pendingValue" TEXT,
      "otpCode" TEXT NOT NULL,
      "attempts" INTEGER NOT NULL DEFAULT 0,
      "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
      "expiresAt" TIMESTAMP(3) NOT NULL,
      CONSTRAINT "tenant_admin_otp_requests_pkey" PRIMARY KEY ("id")
    )
  `);

  await prisma.$executeRawUnsafe(`
    CREATE INDEX IF NOT EXISTS "tenant_admin_captcha_expires_at_idx"
      ON "tenant_admin_captchas" ("expiresAt")
  `);
  await prisma.$executeRawUnsafe(`
    CREATE INDEX IF NOT EXISTS "tenant_admin_session_tenant_slug_idx"
      ON "tenant_admin_sessions" ("tenantSlug")
  `);
  await prisma.$executeRawUnsafe(`
    CREATE INDEX IF NOT EXISTS "tenant_admin_session_expires_at_idx"
      ON "tenant_admin_sessions" ("expiresAt")
  `);
  await prisma.$executeRawUnsafe(`
    CREATE INDEX IF NOT EXISTS "tenant_admin_otp_tenant_purpose_idx"
      ON "tenant_admin_otp_requests" ("tenantSlug", "purpose")
  `);
  await prisma.$executeRawUnsafe(`
    CREATE INDEX IF NOT EXISTS "tenant_admin_otp_expires_at_idx"
      ON "tenant_admin_otp_requests" ("expiresAt")
  `);
}

function toPublicTenantUser(user: {
  id: string;
  tenantSlug: string;
  email: string | null;
  mobile: string | null;
  role: string | null;
  createdAt: Date;
  updatedAt: Date;
  lastLoginAt: Date | null;
}): TenantUser {
  return {
    id: user.id,
    tenantSlug: user.tenantSlug,
    email: user.email ?? undefined,
    mobile: user.mobile ?? undefined,
    role: normalizeTenantRole(user.role),
    createdAtISO: user.createdAt.toISOString(),
    updatedAtISO: user.updatedAt.toISOString(),
    lastLoginAtISO: user.lastLoginAt?.toISOString(),
  };
}

function normalizeTenantRole(role: string | null | undefined) {
  const normalized = (role ?? "").trim().toLowerCase();
  if (!normalized || normalized === "user") {
    return "member";
  }
  return normalized;
}

function toPublicSession(session: {
  id: string;
  tenantSlug: string;
  userID: string;
  createdAt: Date;
  expiresAt: Date;
}): TenantSession {
  return {
    id: session.id,
    tenantSlug: session.tenantSlug,
    userID: session.userID,
    createdAtISO: session.createdAt.toISOString(),
    expiresAtISO: session.expiresAt.toISOString(),
  };
}

function toPublicOTPRecord(record: {
  id: string;
  tenantSlug: string;
  purpose: string;
  channel: string;
  identifier: string;
  userID: string | null;
  pendingValue: string | null;
  otpCode: string;
  attempts: number;
  expiresAt: Date;
  createdAt: Date;
}): OTPRecord {
  return {
    id: record.id,
    tenantSlug: record.tenantSlug,
    purpose: record.purpose as OTPPurpose,
    channel: record.channel as AuthChannel,
    identifier: record.identifier,
    userID: record.userID ?? undefined,
    pendingValue: record.pendingValue ?? undefined,
    otpCode: record.otpCode,
    attempts: record.attempts,
    expiresAtISO: record.expiresAt.toISOString(),
    createdAtISO: record.createdAt.toISOString(),
  };
}

function isUniqueViolation(error: unknown) {
  return (
    error instanceof Prisma.PrismaClientKnownRequestError &&
    error.code === "P2002"
  );
}

function generateCaptchaText(length: number) {
  const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let value = "";
  for (let i = 0; i < length; i += 1) {
    value += alphabet[randomInt(0, alphabet.length)];
  }
  return value;
}

function buildCaptchaDataURL(answer: string) {
  const width = 170;
  const height = 56;
  const chars = answer
    .split("")
    .map((char, index) => {
      const x = 18 + index * 29;
      const y = index % 2 === 0 ? 34 : 40;
      const rotate = randomInt(-20, 21);
      return `<text x="${x}" y="${y}" transform="rotate(${rotate} ${x} ${y})" font-size="28" fill="#0f172a" font-family="monospace">${char}</text>`;
    })
    .join("");

  const noise = Array.from({ length: 8 })
    .map(() => {
      const x1 = randomInt(0, width);
      const y1 = randomInt(0, height);
      const x2 = randomInt(0, width);
      const y2 = randomInt(0, height);
      const opacity = (randomInt(25, 60) / 100).toFixed(2);
      return `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="#475569" stroke-opacity="${opacity}" stroke-width="1" />`;
    })
    .join("");

  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
<rect width="${width}" height="${height}" rx="10" fill="#cbd5e1" />
${noise}
${chars}
</svg>`;

  return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

function maskIdentifier(channel: AuthChannel, identifier: string) {
  if (channel === "email") {
    const [local, domain = ""] = identifier.split("@");
    if (!local) {
      return `***@${domain}`;
    }
    if (local.length <= 2) {
      return `${local[0] ?? "*"}***@${domain}`;
    }
    return `${local.slice(0, 2)}***@${domain}`;
  }

  const compact = identifier.replace(/\s+/g, "");
  if (compact.length <= 4) {
    return "***";
  }
  return `${compact.slice(0, 3)}****${compact.slice(-3)}`;
}
