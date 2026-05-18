import { adminProxy } from '@/lib/admin-proxy';

export async function GET() {
  return adminProxy('/admin/models/audit-log');
}
