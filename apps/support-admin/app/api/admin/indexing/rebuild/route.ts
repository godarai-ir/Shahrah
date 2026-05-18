import { adminProxy } from '@/lib/admin-proxy';

export async function POST() {
  return adminProxy('/admin/indexing/rebuild', { method: 'POST', headers: { 'content-type': 'application/json' }, body: '{}' });
}
