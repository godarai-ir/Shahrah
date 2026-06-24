import { NextResponse } from 'next/server';
import { z } from 'zod';
import {
  ADMIN_SESSION_COOKIE,
  createAdminSession,
  SESSION_TTL_SECONDS,
} from '@/lib/admin-auth';
import { supportAdminDb } from '@/lib/server/db';
import { verifyPassword } from '@/lib/server/password';
import { isSameOriginRequest } from '@/lib/server/request-security';

const loginSchema = z.object({
  email: z.string().trim().toLowerCase().email().max(254),
  password: z.string().min(1).max(1024),
});

const MAX_FAILED_ATTEMPTS = 5;
const LOCKOUT_MS = 15 * 60 * 1000;

export async function POST(request: Request) {
  if (!isSameOriginRequest(request)) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const parsed = loginSchema.safeParse(await request.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: 'Invalid payload' }, { status: 400 });
  }

  const user = await supportAdminDb.supportAdminUser.findUnique({
    where: { email: parsed.data.email },
  });
  const passwordValid = await verifyPassword(parsed.data.password, user?.passwordHash);
  const now = new Date();
  const locked = Boolean(user?.lockedUntil && user.lockedUntil > now);

  if (!user || !passwordValid || user.status !== 'active' || locked) {
    if (user && !locked) {
      const failedUser = await supportAdminDb.supportAdminUser.update({
        where: { id: user.id },
        data: {
          failedLoginCount: { increment: 1 },
        },
      });
      if (failedUser.failedLoginCount >= MAX_FAILED_ATTEMPTS) {
        await supportAdminDb.supportAdminUser.update({
          where: { id: user.id },
          data: { lockedUntil: new Date(Date.now() + LOCKOUT_MS) },
        });
      }
    }
    return NextResponse.json({ error: 'Invalid credentials' }, { status: 401 });
  }

  await supportAdminDb.$transaction(async (tx) => {
    await tx.supportAdminUser.update({
      where: { id: user.id },
      data: { failedLoginCount: 0, lockedUntil: null, lastLoginAt: now },
    });
    await tx.supportAdminSession.deleteMany({ where: { expiresAt: { lte: now } } });
  });
  const session = await createAdminSession(user.id);

  const response = NextResponse.json({ ok: true });
  response.cookies.set(ADMIN_SESSION_COOKIE, session.token, {
    httpOnly: true,
    sameSite: 'strict',
    secure: process.env.SUPPORT_ADMIN_COOKIE_SECURE
      ? process.env.SUPPORT_ADMIN_COOKIE_SECURE === 'true'
      : process.env.NODE_ENV === 'production',
    path: process.env.SUPPORT_ADMIN_COOKIE_PATH ?? '/',
    maxAge: SESSION_TTL_SECONDS,
  });
  return response;
}
