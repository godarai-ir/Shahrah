import { NextResponse } from 'next/server';
import { getAdminSession, roleHasPermission } from '@/lib/admin-auth';
import {
  getResendReceivingClient,
  ResendReceivingConfigurationError,
} from '@/lib/server/resend-receiving';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ emailId: string }> },
) {
  const session = await getAdminSession();
  if (!session) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  if (!roleHasPermission(session.role, 'inbound_email:read')) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const { emailId } = await params;
  if (!/^[A-Za-z0-9_-]{1,128}$/.test(emailId)) {
    return NextResponse.json({ error: 'Invalid received email ID' }, { status: 400 });
  }

  try {
    const resend = getResendReceivingClient();
    const { data, error } = await resend.emails.receiving.get(emailId);

    if (error) return resendErrorResponse(error);

    return NextResponse.json(data, { headers: { 'cache-control': 'no-store' } });
  } catch (error) {
    if (error instanceof ResendReceivingConfigurationError) {
      return NextResponse.json({ error: error.message }, { status: 503 });
    }

    console.error('[support-admin][resend] failed to get received email', error);
    return NextResponse.json({ error: 'Failed to load received email' }, { status: 502 });
  }
}

function resendErrorResponse(error: { message: string; statusCode: number | null }) {
  const status = error.statusCode && error.statusCode >= 400 && error.statusCode <= 599
    ? error.statusCode
    : 502;

  return NextResponse.json({ error: error.message }, { status });
}
