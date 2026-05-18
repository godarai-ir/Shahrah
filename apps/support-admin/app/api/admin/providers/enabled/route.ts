import { adminProxy } from '@/lib/admin-proxy';

export async function POST(request: Request) {
  const body = await request.text();
  return adminProxy('/providers/admin/enabled', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body,
  });
}
