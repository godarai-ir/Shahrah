import 'server-only';

import { createHash, randomBytes } from 'crypto';
import { cookies } from 'next/headers';
import { supportAdminDb } from '@/lib/server/db';
export { ADMIN_SESSION_COOKIE, SESSION_TTL_SECONDS } from '@/lib/auth-constants';
import { ADMIN_SESSION_COOKIE, SESSION_TTL_SECONDS } from '@/lib/auth-constants';

export type AdminPermission =
  | 'catalog:read'
  | 'catalog:write'
  | 'provider:write'
  | 'import:execute'
  | 'indexing:execute'
  | 'display-settings:write'
  | 'platform:admin';

export type AdminSession = {
  sessionId: string;
  userId: string;
  email: string;
  role: string;
  expiresAt: Date;
};

const ROLE_PERMISSIONS: Record<string, readonly AdminPermission[]> = {
  super_admin: [
    'catalog:read', 'catalog:write', 'provider:write', 'import:execute',
    'indexing:execute', 'display-settings:write', 'platform:admin',
  ],
  catalog_admin: [
    'catalog:read', 'catalog:write', 'provider:write', 'import:execute',
    'indexing:execute', 'display-settings:write',
  ],
  support_agent: ['catalog:read'],
};

function tokenHash(token: string): string {
  return createHash('sha256').update(token).digest('hex');
}

export function roleHasPermission(role: string, permission: AdminPermission): boolean {
  return ROLE_PERMISSIONS[role]?.includes(permission) ?? false;
}

export async function createAdminSession(userId: string): Promise<{ token: string; expiresAt: Date }> {
  const token = randomBytes(32).toString('base64url');
  const expiresAt = new Date(Date.now() + SESSION_TTL_SECONDS * 1000);
  await supportAdminDb.supportAdminSession.create({
    data: { tokenHash: tokenHash(token), userId, expiresAt },
  });
  return { token, expiresAt };
}

export async function getAdminSession(): Promise<AdminSession | null> {
  const store = await cookies();
  const token = store.get(ADMIN_SESSION_COOKIE)?.value;
  if (!token) return null;

  const session = await supportAdminDb.supportAdminSession.findUnique({
    where: { tokenHash: tokenHash(token) },
    include: { user: true },
  });
  if (!session || session.expiresAt <= new Date() || session.user.status !== 'active') return null;

  return {
    sessionId: session.id,
    userId: session.user.id,
    email: session.user.email,
    role: session.user.role,
    expiresAt: session.expiresAt,
  };
}

export async function deleteAdminSession(): Promise<void> {
  const store = await cookies();
  const token = store.get(ADMIN_SESSION_COOKIE)?.value;
  if (!token) return;
  await supportAdminDb.supportAdminSession.deleteMany({ where: { tokenHash: tokenHash(token) } });
}

export async function requireAdminSession(permission?: AdminPermission): Promise<AdminSession | null> {
  const session = await getAdminSession();
  if (!session || (permission && !roleHasPermission(session.role, permission))) return null;
  return session;
}
