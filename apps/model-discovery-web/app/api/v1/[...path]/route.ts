import { NextResponse } from 'next/server';

const PUBLIC_ROOTS = new Set([
  'models',
  'providers',
  'facets',
  'categories',
  'capabilities',
  'supported-parameters',
  'modalities',
  'regions',
  'settings',
  'health',
  'ready',
]);

type RouteContext = { params: Promise<{ path: string[] }> };

export const dynamic = 'force-dynamic';

export async function GET(request: Request, context: RouteContext) {
  const { path } = await context.params;
  if (!isPublicReadPath(path)) {
    return NextResponse.json({ error: 'Not found' }, { status: 404 });
  }

  const baseURL =
    process.env.MODEL_DISCOVERY_API_BASE_INTERNAL ||
    (process.env.NODE_ENV === 'production'
      ? 'http://model-discovery:3002/api/v1'
      : 'http://localhost:3002/api/v1');
  const incoming = new URL(request.url);
  const upstream = new URL(`${baseURL.replace(/\/$/, '')}/${path.map(encodeURIComponent).join('/')}`);
  upstream.search = incoming.search;

  try {
    const response = await fetch(upstream, {
      headers: { accept: request.headers.get('accept') || '*/*' },
      cache: 'no-store',
    });
    return new NextResponse(response.body, {
      status: response.status,
      headers: {
        'content-type': response.headers.get('content-type') || 'application/octet-stream',
        'cache-control': response.headers.get('cache-control') || 'no-store',
      },
    });
  } catch {
    return NextResponse.json({ error: 'Model discovery service unavailable' }, { status: 502 });
  }
}

function isPublicReadPath(path: string[]): boolean {
  if (path.length === 0 || !PUBLIC_ROOTS.has(path[0])) return false;
  return !path.includes('admin');
}
