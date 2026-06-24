import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { ADMIN_SESSION_COOKIE } from '@/lib/auth-constants';

export function proxy(request: NextRequest) {
  // This is only an optimistic redirect. Every API handler validates the DB-backed session.
  const token = request.cookies.get(ADMIN_SESSION_COOKIE)?.value;
  if (!token && request.nextUrl.pathname.startsWith('/admin')) {
    const url = new URL('/login', request.url);
    url.searchParams.set('next', request.nextUrl.pathname);
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ['/admin/:path*'],
};
