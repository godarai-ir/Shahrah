import { type NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function PUT(request: NextRequest, { params }: { params: Promise<{ ticketID: string }> }) {
  const { ticketID } = await params;
  const body = await request.json().catch(() => null);
  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/support/tickets/${encodeURIComponent(ticketID)}/rating`,
    body,
    upstreamHeaders: {
      "idempotency-key": request.headers.get("idempotency-key") ?? "",
    },
  });
}
