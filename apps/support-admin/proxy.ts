import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { ADMIN_TOKEN_COOKIE } from '@/lib/admin-auth';

export function proxy(request: NextRequest) {
  const token = request.cookies.get(ADMIN_TOKEN_COOKIE)?.value;
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
