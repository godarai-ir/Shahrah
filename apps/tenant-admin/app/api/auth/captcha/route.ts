import { NextResponse } from "next/server";

import { createCaptchaChallenge } from "@/lib/auth/store";

export const dynamic = "force-dynamic";

export async function GET() {
  const challenge = await createCaptchaChallenge();
  return NextResponse.json(challenge);
}
