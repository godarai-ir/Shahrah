import { NextResponse } from 'next/server';
import { ADMIN_TOKEN_COOKIE } from '@/lib/admin-auth';

export async function POST() {
  const response = NextResponse.json({ ok: true });
  response.cookies.set(ADMIN_TOKEN_COOKIE, '', { httpOnly: true, path: '/', maxAge: 0 });
  return response;
}
