import { adminProxy } from '@/lib/admin-proxy';

export async function GET(request: Request) {
  const query = new URL(request.url).search;
  return adminProxy(`/admin/models${query}`);
}

export async function POST(request: Request) {
  const body = await request.text();
  return adminProxy('/admin/models/upsert', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body,
  });
}
