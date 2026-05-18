import { adminProxy } from '@/lib/admin-proxy';

export async function POST(request: Request) {
  const formData = await request.formData();
  const slug = formData.get('slug');
  const file = formData.get('file');
  if (typeof slug !== 'string' || !(file instanceof File)) {
    return Response.json({ error: 'slug and file are required' }, { status: 400 });
  }
  const outbound = new FormData();
  outbound.append('file', file);
  return adminProxy(`/providers/admin/${encodeURIComponent(slug)}/logo`, {
    method: 'POST',
    body: outbound,
  });
}
