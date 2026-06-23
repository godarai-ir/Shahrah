import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { ADMIN_SESSION_COOKIE, APP_BASE_PATH, withBasePath } from '@/lib/auth-constants';

export function proxy(request: NextRequest) {
  // This is only an optimistic redirect. Every API handler validates the DB-backed session.
  const token = request.cookies.get(ADMIN_SESSION_COOKIE)?.value;
  const adminPath = `${APP_BASE_PATH}/admin`;
  if (!token && request.nextUrl.pathname.startsWith(adminPath)) {
    const url = new URL(withBasePath('/login'), request.url);
    url.searchParams.set('next', request.nextUrl.pathname.slice(APP_BASE_PATH.length));
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ['/admin/:path*'],
};
