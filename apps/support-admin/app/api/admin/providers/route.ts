import { adminProxy } from '@/lib/admin-proxy';

export async function GET() {
  return adminProxy('/providers/admin/list');
}

export async function POST(request: Request) {
  const body = await request.text();
  return adminProxy('/providers/admin/upsert', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body,
  });
}
