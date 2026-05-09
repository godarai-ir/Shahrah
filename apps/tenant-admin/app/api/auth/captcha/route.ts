import { NextResponse } from "next/server";

import { getAuthService } from "@/lib/auth/factory";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const authService = getAuthService();
    const challenge = await authService.createCaptcha();
    return NextResponse.json(challenge);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Authentication failed.";
    const statusMatch =
      error instanceof Error ? error.message.match(/Auth API error:\s+(\d{3})/) : null;
    const status = statusMatch ? Number(statusMatch[1]) : 500;
    return NextResponse.json({ error: { message } }, { status });
  }
}
