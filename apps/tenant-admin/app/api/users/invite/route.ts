import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { sendInvitationEmail } from "@/lib/email";

export const dynamic = "force-dynamic";

type InvitePayload = {
  email?: unknown;
  role?: unknown;
  accept_invite_client_url?: unknown;
};

const inviteUpstreamResponseSchema = z.object({
  link: z.string().optional(),
  invitation: z.object({ role: z.string().optional() }).optional(),
});

function resolveInviteBaseURL(request: NextRequest) {
  const configured = process.env.SHAHRAH_INVITE_BASE_URL?.trim();
  if (configured) {
    return configured.endsWith("/") ? configured : `${configured}/`;
  }

  const forwardedProto = request.headers.get("x-forwarded-proto")?.trim();
  const host = request.headers.get("x-forwarded-host")?.trim() || request.headers.get("host")?.trim();
  const protocol = forwardedProto || request.nextUrl.protocol.replace(":", "") || "https";

  if (!host) {
    return "";
  }

  return `${protocol}://${host}/invite/`;
}

export async function POST(request: NextRequest) {
  try {
    const auth = await readAuthFromRequest(request);
    if (!auth.ok) {
      return NextResponse.json(
        { error: { message: auth.message } },
        { status: auth.status },
      );
    }

    const body = (await request.json().catch(() => null)) as InvitePayload | null;
    if (!body || typeof body !== "object") {
      return NextResponse.json(
        { error: { message: "Invalid request body." } },
        { status: 400 },
      );
    }

    const email = typeof body.email === "string" ? body.email.trim() : "";
    const role = typeof body.role === "string" ? body.role.trim() : "member";
    const acceptInviteClientURL =
      typeof body.accept_invite_client_url === "string"
        ? body.accept_invite_client_url.trim()
        : resolveInviteBaseURL(request);

    if (!email) {
      return NextResponse.json(
        { error: { message: "Email is required." } },
        { status: 400 },
      );
    }

    const upstream = await fetchControlPlaneJSON({
      tenantSlug: auth.tenantSlug,
      userID: auth.user.id,
      method: "POST",
      path: "/api/v1/users/invite",
      body: {
        email,
        role: role || "member",
        accept_invite_client_url: acceptInviteClientURL || undefined,
      },
    });

    if (!upstream.ok) {
      console.error(
        `[tenant-admin][invite-upstream] status=${upstream.status} message=${upstream.errorMessage ?? "n/a"} payload=${JSON.stringify(upstream.payload)}`,
      );
      return NextResponse.json(
        {
          error: {
            message:
              upstream.errorMessage ??
              `Control-plane request failed (${upstream.status})`,
          },
          details: upstream.payload,
        },
        { status: upstream.status },
      );
    }

    const parsedPayload = inviteUpstreamResponseSchema.safeParse(upstream.payload);
    const payload = parsedPayload.success ? parsedPayload.data : undefined;

  if (payload?.link) {
    try {
      await sendInvitationEmail({
        to: email,
        link: payload.link,
        tenantSlug: auth.tenantSlug,
        tenantName: auth.tenantSlug,
      });
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        console.error(
          `[tenant-admin][invite-email] failed tenant=${auth.tenantSlug} to=${email}: ${message}`,
        );
      }
    }

    return NextResponse.json(upstream.payload, { status: upstream.status });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unexpected invite route error";
    console.error(`[tenant-admin][invite-route] ${message}`);
    return NextResponse.json(
      { error: { message: "Internal invite route error." } },
      { status: 500 },
    );
  }
}


