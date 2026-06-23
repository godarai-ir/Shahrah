import { NextResponse } from 'next/server';
import { randomUUID } from 'crypto';

import { requireAdminSession } from '@/lib/admin-auth';
import { isSameOriginRequest } from '@/lib/server/request-security';

type RouteContext = {
  params: Promise<{ path?: string[] }>;
};

export const dynamic = 'force-dynamic';

export async function GET(_request: Request, context: RouteContext) {
  return proxyControlPlanePlatform(context, { method: 'GET' });
}

export async function POST(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  return proxyControlPlanePlatform(context, {
    method: 'POST',
    body: await request.text(),
    contentType: request.headers.get('content-type') ?? 'application/json',
  });
}

export async function PUT(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  return proxyControlPlanePlatform(context, {
    method: 'PUT',
    body: await request.text(),
    contentType: request.headers.get('content-type') ?? 'application/json',
  });
}

export async function DELETE(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  return proxyControlPlanePlatform(context, { method: 'DELETE' });
}

async function proxyControlPlanePlatform(
  context: RouteContext,
  init: { method: 'GET' | 'POST' | 'PUT' | 'DELETE'; body?: string; contentType?: string },
) {
  const session = await requireAdminSession('platform:admin');
  if (!session) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const serviceToken = process.env.CONTROL_PLANE_SERVICE_TOKEN;
  if (!serviceToken) {
    return NextResponse.json({ error: 'CONTROL_PLANE_SERVICE_TOKEN is required.' }, { status: 503 });
  }

  const { path = [] } = await context.params;
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? 'http://127.0.0.1:4000';
  const upstreamPath = path.map(encodeURIComponent).join('/');
  const url = new URL(`/api/v1/platform/${upstreamPath}`, ensureTrailingSlash(baseURL));

  const response = await fetch(url, {
    method: init.method,
    headers: {
      authorization: `Bearer ${serviceToken}`,
      'x-shahrah-actor-id': session.userId,
      'x-shahrah-actor-role': session.role,
      'x-request-id': randomUUID(),
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
