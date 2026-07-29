import { type NextRequest } from "next/server";

import { proxyToControlPlane, proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET(request: NextRequest, { params }: { params: Promise<{ ticketID: string }> }) {
  const { ticketID } = await params;
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/support/tickets/${encodeURIComponent(ticketID)}/messages${request.nextUrl.search}`,
  });
}

export async function POST(request: NextRequest, { params }: { params: Promise<{ ticketID: string }> }) {
  const { ticketID } = await params;
  return proxyToControlPlaneBodyStream({
    request,
    method: "POST",
    path: `/api/v1/support/tickets/${encodeURIComponent(ticketID)}/messages`,
  });
}
