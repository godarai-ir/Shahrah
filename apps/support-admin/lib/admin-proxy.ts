import { NextResponse } from 'next/server';
import { API_BASE } from '@/lib/api';
import { getAdminTokenFromCookie } from '@/lib/admin-auth';

export async function adminProxy(path: string, init?: RequestInit) {
  const token = await getAdminTokenFromCookie();
  if (!token) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      ...(init?.headers || {}),
      authorization: `Bearer ${token}`,
    },
    cache: 'no-store',
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { 'content-type': res.headers.get('content-type') || 'application/json' },
  });
}
