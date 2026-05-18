import { cookies } from 'next/headers';

export const ADMIN_TOKEN_COOKIE = 'support_admin_token';

export async function getAdminTokenFromCookie(): Promise<string | null> {
  const store = await cookies();
  return store.get(ADMIN_TOKEN_COOKIE)?.value ?? null;
}
