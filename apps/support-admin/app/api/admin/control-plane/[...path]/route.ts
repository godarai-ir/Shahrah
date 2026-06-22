import { NextResponse } from 'next/server';

import { getAdminTokenFromCookie } from '@/lib/admin-auth';

type RouteContext = {
  params: Promise<{ path?: string[] }>;
};

export const dynamic = 'force-dynamic';

export async function GET(_request: Request, context: RouteContext) {
  return proxyControlPlanePlatform(context, { method: 'GET' });
}

export async function POST(request: Request, context: RouteContext) {
  return proxyControlPlanePlatform(context, {
    method: 'POST',
    body: await request.text(),
    contentType: request.headers.get('content-type') ?? 'application/json',
  });
}

export async function PUT(request: Request, context: RouteContext) {
  return proxyControlPlanePlatform(context, {
    method: 'PUT',
    body: await request.text(),
    contentType: request.headers.get('content-type') ?? 'application/json',
  });
}

export async function DELETE(_request: Request, context: RouteContext) {
  return proxyControlPlanePlatform(context, { method: 'DELETE' });
}

async function proxyControlPlanePlatform(
  context: RouteContext,
  init: { method: 'GET' | 'POST' | 'PUT' | 'DELETE'; body?: string; contentType?: string },
) {
  const adminToken = await getAdminTokenFromCookie();
  if (!adminToken) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const masterKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY;
  if (!masterKey) {
    return NextResponse.json({ error: 'SHAHRAH_MASTER_KEY or MASTER_KEY is required.' }, { status: 500 });
  }

  const { path = [] } = await context.params;
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? 'http://127.0.0.1:4000';
  const upstreamPath = path.map(encodeURIComponent).join('/');
  const url = new URL(`/api/v1/platform/${upstreamPath}`, ensureTrailingSlash(baseURL));

  const response = await fetch(url, {
    method: init.method,
    headers: {
      authorization: `Bearer ${masterKey}`,
      accept: 'application/json',
      ...(init.body ? { 'content-type': init.contentType ?? 'application/json' } : {}),
    },
    body: init.body || undefined,
    cache: 'no-store',
  });
  const text = await response.text();
  return new NextResponse(text, {
    status: response.status,
    headers: { 'content-type': response.headers.get('content-type') ?? 'application/json' },
  });
}

function ensureTrailingSlash(value: string): string {
  return value.endsWith('/') ? value : `${value}/`;
}
