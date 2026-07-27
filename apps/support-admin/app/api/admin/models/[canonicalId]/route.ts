import { adminProxy } from '@/lib/admin-proxy';

export async function GET(
  _request: Request,
  context: { params: Promise<{ canonicalId: string }> },
) {
  const { canonicalId } = await context.params;
  return adminProxy(`/admin/models/detail/${encodeURIComponent(canonicalId)}`);
}
