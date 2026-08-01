import { NextResponse } from "next/server";

import { socialAuth } from "@/app/lib/social-auth/config";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const session = await socialAuth.api.getSession({ headers: request.headers });
  if (!session?.user) return NextResponse.json({ user: null });
  return NextResponse.json({
    user: {
      id: session.user.id,
      name: session.user.name,
      email: session.user.email,
      email_verified: session.user.emailVerified,
      image: session.user.image,
    },
  });
}
