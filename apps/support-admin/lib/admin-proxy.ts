import { NextResponse } from 'next/server';
import { randomUUID } from 'crypto';
import { API_BASE } from '@/lib/api';
import {
  type AdminPermission,
  getAdminSession,
  roleHasPermission,
} from '@/lib/admin-auth';
import { isSameOriginFromHeaders } from '@/lib/server/request-security';

function requiredPermission(path: string, method: string): AdminPermission {
  if (method === 'GET') return 'catalog:read';
  if (path.startsWith('/providers/admin')) return 'provider:write';
  if (path.startsWith('/admin/import')) return 'import:execute';
  if (path.startsWith('/admin/indexing')) return 'indexing:execute';
  if (path.startsWith('/settings/display')) return 'display-settings:write';
  return 'catalog:write';
}

export async function adminProxy(path: string, init?: RequestInit) {
  const method = init?.method?.toUpperCase() ?? 'GET';
  const session = await getAdminSession();
  if (!session) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  if (!roleHasPermission(session.role, requiredPermission(path, method))) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }
  if (method !== 'GET' && method !== 'HEAD' && !(await isSameOriginFromHeaders())) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const serviceToken = process.env.MODEL_DISCOVERY_SERVICE_TOKEN;
  if (!serviceToken) {
    return NextResponse.json({ error: 'MODEL_DISCOVERY_SERVICE_TOKEN is required' }, { status: 503 });
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      ...(init?.headers || {}),
      authorization: `Bearer ${serviceToken}`,
      'x-shahrah-actor-id': session.userId,
      'x-shahrah-actor-role': session.role,
      'x-request-id': randomUUID(),
    },
    cache: 'no-store',
  });

  const body = await res.arrayBuffer();
  const headers = new Headers();
  headers.set('content-type', res.headers.get('content-type') || 'application/json');
  copyResponseHeader(res.headers, headers, 'cache-control');
  copyResponseHeader(res.headers, headers, 'content-disposition');
  copyResponseHeader(res.headers, headers, 'etag');
  copyResponseHeader(res.headers, headers, 'last-modified');

  return new NextResponse(statusAllowsBody(res.status) ? body : null, {
    status: res.status,
    headers,
  });
}

function copyResponseHeader(source: Headers, target: Headers, name: string) {
  const value = source.get(name);
  if (value) target.set(name, value);
}

function statusAllowsBody(status: number) {
  return status !== 204 && status !== 205 && status !== 304;
}
