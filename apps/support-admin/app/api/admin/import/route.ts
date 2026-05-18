import { adminProxy } from '@/lib/admin-proxy';

export async function POST(request: Request) {
  const body = await request.text();
  return adminProxy('/admin/import/catalog-json', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body,
  });
}
