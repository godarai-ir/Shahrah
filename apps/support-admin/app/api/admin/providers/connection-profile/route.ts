import { NextResponse } from 'next/server';

export async function POST() {
  return NextResponse.json({
    error: {
      code: 'provider_connection_profile_retired',
      message: 'Use the platform provider connections workflow.',
    },
  }, { status: 410 });
}
