import { NextResponse } from 'next/server';
import { getAdminSession, roleHasPermission } from '@/lib/admin-auth';
import {
  getResendReceivingClient,
  ResendReceivingConfigurationError,
} from '@/lib/server/resend-receiving';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

export async function GET(request: Request) {
  const session = await getAdminSession();
  if (!session) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  if (!roleHasPermission(session.role, 'inbound_email:read')) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const { searchParams } = new URL(request.url);
  const limit = parseLimit(searchParams.get('limit'));
  if (limit === null) {
    return NextResponse.json({ error: 'limit must be an integer between 1 and 100' }, { status: 400 });
  }

  const after = searchParams.get('after')?.trim();
  if (after && after.length > 200) {
    return NextResponse.json({ error: 'after cursor is too long' }, { status: 400 });
  }

  try {
    const resend = getResendReceivingClient();
    const { data, error } = await resend.emails.receiving.list({
      limit,
      ...(after ? { after } : {}),
    });

    if (error) return resendErrorResponse(error);

    return NextResponse.json(data, { headers: { 'cache-control': 'no-store' } });
  } catch (error) {
    if (error instanceof ResendReceivingConfigurationError) {
      return NextResponse.json({ error: error.message }, { status: 503 });
    }

    console.error('[support-admin][resend] failed to list received emails', error);
    return NextResponse.json({ error: 'Failed to load received emails' }, { status: 502 });
  }
}

function parseLimit(value: string | null): number | null {
  if (value === null) return 25;
  if (!/^\d+$/.test(value)) return null;

  const parsed = Number(value);
  return parsed >= 1 && parsed <= 100 ? parsed : null;
}

function resendErrorResponse(error: { message: string; statusCode: number | null }) {
  const status = error.statusCode && error.statusCode >= 400 && error.statusCode <= 599
    ? error.statusCode
    : 502;

  return NextResponse.json({ error: error.message }, { status });
}
