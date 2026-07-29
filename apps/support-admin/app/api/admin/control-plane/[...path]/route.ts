import { NextResponse } from 'next/server';
import { randomUUID } from 'crypto';

import { type AdminPermission, requireAdminSession } from '@/lib/admin-auth';
import { isSameOriginRequest } from '@/lib/server/request-security';

type RouteContext = {
  params: Promise<{ path?: string[] }>;
};

export const dynamic = 'force-dynamic';

export async function GET(request: Request, context: RouteContext) {
	return proxyControlPlanePlatform(context, { method: 'GET', search: new URL(request.url).search });
}

export async function POST(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  const contentType = request.headers.get('content-type') ?? 'application/json';
  return proxyControlPlanePlatform(context, {
    method: 'POST',
    body: contentType.includes('multipart/form-data') ? request.body ?? undefined : await request.text(),
    contentType,
    idempotencyKey: request.headers.get('idempotency-key') ?? undefined,
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

export async function PATCH(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  return proxyControlPlanePlatform(context, {
    method: 'PATCH',
    body: await request.text(),
    contentType: request.headers.get('content-type') ?? 'application/json',
    idempotencyKey: request.headers.get('idempotency-key') ?? undefined,
  });
}

export async function DELETE(request: Request, context: RouteContext) {
  if (!isSameOriginRequest(request)) return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  const body = await request.text();
  return proxyControlPlanePlatform(context, {
    method: 'DELETE',
    body,
    contentType: request.headers.get('content-type') ?? 'application/json',
  });
}

async function proxyControlPlanePlatform(
  context: RouteContext,
	init: { method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'; body?: BodyInit; contentType?: string; search?: string; idempotencyKey?: string },
) {
  const { path = [] } = await context.params;
  const permission = controlPlanePermission(path, init.method);
  const session = await requireAdminSession(permission);
  if (!session) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const serviceToken = process.env.CONTROL_PLANE_SERVICE_TOKEN;
  if (!serviceToken) {
    return NextResponse.json({ error: 'CONTROL_PLANE_SERVICE_TOKEN is required.' }, { status: 503 });
  }

  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? 'http://127.0.0.1:4000';
  const upstreamPath = path.map(encodeURIComponent).join('/');
	const url = new URL(`/api/v1/platform/${upstreamPath}`, ensureTrailingSlash(baseURL));
	url.search = init.search ?? '';
  const requestId = randomUUID();

  const fetchInit: RequestInit & { duplex?: 'half' } = {
    method: init.method,
    headers: {
      authorization: `Bearer ${serviceToken}`,
      'x-shahrah-actor-id': session.userId,
      'x-shahrah-actor-role': session.role,
      'x-request-id': requestId,
      accept: 'application/json',
      ...(init.body ? { 'content-type': init.contentType ?? 'application/json' } : {}),
      ...(init.idempotencyKey ? { 'idempotency-key': init.idempotencyKey } : {}),
    },
    body: init.body || undefined,
    cache: 'no-store',
  };
  if (init.body instanceof ReadableStream) fetchInit.duplex = 'half';
  const response = await fetch(url, fetchInit);
  const responseHeaders = new Headers({
    'content-type': response.headers.get('content-type') ?? 'application/json',
    'x-request-id': response.headers.get('x-request-id') ?? requestId,
  });
  const contentDisposition = response.headers.get('content-disposition');
  const cacheControl = response.headers.get('cache-control');
  const noSniff = response.headers.get('x-content-type-options');
  if (contentDisposition) responseHeaders.set('content-disposition', contentDisposition);
  if (cacheControl) responseHeaders.set('cache-control', cacheControl);
  if (noSniff) responseHeaders.set('x-content-type-options', noSniff);

  return new NextResponse(response.body, {
    status: response.status,
    headers: responseHeaders,
  });
}

function controlPlanePermission(path: string[], method: string): AdminPermission {
  const joined = path.join('/');
  if (joined.startsWith('support/ticket-attachments')) return 'support_ticket:read';
  if (joined.startsWith('support/tickets')) {
    if (method === 'GET') {
      if (joined.endsWith('/events')) return 'audit:read';
      return 'support_ticket:read';
    }
    if (joined.endsWith('/messages')) return 'support_ticket:reply';
    if (joined.endsWith('/triage')) return 'support_ticket:triage';
    if (joined.endsWith('/reopen')) return 'support_ticket:status_write';
    return 'support_ticket:status_write';
  }
  if (joined.startsWith('provider-connections') || joined.startsWith('provider-connection-schemas')) {
    if (method === 'GET') return 'platform_credentials:read_metadata';
    if (joined.includes('/test')) return 'platform_credentials:test';
    if (joined.includes('/activate') || joined.includes('/rollback')) return 'platform_credentials:activate';
    if (joined.includes('/disable') || joined.includes('/revoke')) return 'platform_credentials:disable';
    return 'platform_credentials:create_version';
  }
  if (joined.startsWith('tenants/')) {
    if (method === 'GET') {
      if (joined.includes('/wallet')) return 'wallet:read';
      if (joined.includes('/billing')) return 'billing:read';
      if (joined.includes('/events')) return 'audit:read';
      return 'tenant:read';
    }
    if (joined.includes('/wallet/adjustments')) return 'wallet:adjust';
    if (joined.endsWith('/suspend') || joined.endsWith('/reactivate')) return 'tenant:status:write';
    if (joined.includes('/subscription/')) return 'subscription:write';
  }
  if (joined === 'tenants' && method === 'GET') return 'tenant:read';
  if (joined === 'tenant-subscriptions') return 'subscription:write';
	if (joined === 'commercial-signup-settings') return method === 'GET' ? 'catalog:read' : 'platform:admin';
  if (method === 'GET') return 'catalog:read';
  if (joined.startsWith('commercial-plans') || joined.startsWith('features') || joined.startsWith('tenant-entitlement-overrides')) return 'catalog:write';
  return 'platform:admin';
}

function ensureTrailingSlash(value: string): string {
  return value.endsWith('/') ? value : `${value}/`;
}
