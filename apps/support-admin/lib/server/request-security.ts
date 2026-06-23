import 'server-only';

import { headers } from 'next/headers';

function sameOrigin(origin: string | null, hostHeader: string | null): boolean {
  if (!origin) return false;
  try {
    const configuredOrigin = process.env.SUPPORT_ADMIN_ORIGIN;
    if (configuredOrigin) return new URL(origin).origin === new URL(configuredOrigin).origin;
    if (process.env.NODE_ENV === 'production' || !hostHeader) return false;
    return new URL(origin).host === hostHeader.split(',')[0]?.trim();
  } catch {
    return false;
  }
}

export function isSameOriginRequest(request: Request): boolean {
  const origin = request.headers.get('origin');
  const host = request.headers.get('x-forwarded-host') ?? request.headers.get('host');
  return sameOrigin(origin, host);
}

export async function isSameOriginFromHeaders(): Promise<boolean> {
  const values = await headers();
  return sameOrigin(
    values.get('origin'),
    values.get('x-forwarded-host') ?? values.get('host'),
  );
}
