import { type NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest, { params }: { params: Promise<{ ticketID: string }> }) {
  const { ticketID } = await params;
  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/support/tickets/${encodeURIComponent(ticketID)}`,
  });
}
