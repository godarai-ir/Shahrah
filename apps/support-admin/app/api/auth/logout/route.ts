import { NextResponse } from 'next/server';
import { ADMIN_SESSION_COOKIE, deleteAdminSession } from '@/lib/admin-auth';
import { isSameOriginRequest } from '@/lib/server/request-security';

export async function POST(request: Request) {
  if (!isSameOriginRequest(request)) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }
  await deleteAdminSession();
  const response = NextResponse.json({ ok: true });
  response.cookies.set(ADMIN_SESSION_COOKIE, '', {
    httpOnly: true,
    sameSite: 'strict',
    secure: process.env.SUPPORT_ADMIN_COOKIE_SECURE
      ? process.env.SUPPORT_ADMIN_COOKIE_SECURE === 'true'
      : process.env.NODE_ENV === 'production',
    path: process.env.SUPPORT_ADMIN_COOKIE_PATH ?? '/',
    maxAge: 0,
  });
  return response;
}
