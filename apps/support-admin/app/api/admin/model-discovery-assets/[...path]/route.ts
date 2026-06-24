import { NextResponse } from 'next/server';
import { adminProxy } from '@/lib/admin-proxy';

type RouteContext = { params: Promise<{ path: string[] }> };

export async function GET(_request: Request, context: RouteContext) {
  const { path } = await context.params;
  if (path.length < 3 || path[0] !== 'providers' || path[1] !== 'assets') {
    return NextResponse.json({ error: 'Not found' }, { status: 404 });
  }
  return adminProxy(`/${path.map(encodeURIComponent).join('/')}`);
}
