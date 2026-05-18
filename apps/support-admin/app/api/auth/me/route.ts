import { NextResponse } from 'next/server';
import { getAdminTokenFromCookie } from '@/lib/admin-auth';

export async function GET() {
  const token = await getAdminTokenFromCookie();
  return NextResponse.json({ authenticated: Boolean(token) });
}
