import { NextRequest, NextResponse } from "next/server";

import {
  readAuthFromRequest,
  type AuthenticatedRequestContext,
} from "@/lib/auth/request-auth";

export async function withRequestAuth(
  request: NextRequest,
  handler: (auth: AuthenticatedRequestContext) => Response | Promise<Response>,
): Promise<Response> {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  return handler(auth);
}
