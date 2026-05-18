import { adminProxy } from '@/lib/admin-proxy';

export async function GET() {
  return adminProxy('/admin/models');
}

export async function POST(request: Request) {
  const body = await request.text();
  return adminProxy('/admin/models/upsert', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body,
  });
}
