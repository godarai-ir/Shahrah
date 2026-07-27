import { adminProxy } from '@/lib/admin-proxy';

export async function PUT(
  request: Request,
  context: { params: Promise<{ canonicalId: string }> },
) {
  const { canonicalId } = await context.params;
  return adminProxy(
    `/admin/models/${encodeURIComponent(canonicalId)}/pricing`,
    {
      method: 'PUT',
      headers: { 'content-type': 'application/json' },
      body: await request.text(),
    },
  );
}
