import { adminProxy } from '@/lib/admin-proxy';

export async function POST(
  _request: Request,
  context: { params: Promise<{ canonicalId: string }> },
) {
  const { canonicalId } = await context.params;
  return adminProxy(
    `/admin/models/${encodeURIComponent(canonicalId)}/publish`,
    { method: 'POST' },
  );
}
