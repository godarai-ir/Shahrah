import { adminProxy } from '@/lib/admin-proxy';

export async function POST(request: Request) {
  const body = await request.json().catch(() => null) as { canonicalId?: string; enabled?: boolean } | null;
  if (!body?.canonicalId || typeof body.enabled !== 'boolean') {
    return Response.json({ error: 'canonicalId and enabled are required' }, { status: 400 });
  }
  return adminProxy(`/admin/models/${encodeURIComponent(body.canonicalId)}/enabled`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ enabled: body.enabled }),
  });
}
