import { type NextRequest } from "next/server";

import { proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET(request: NextRequest, { params }: { params: Promise<{ attachmentID: string }> }) {
  const { attachmentID } = await params;
  return proxyToControlPlaneBodyStream({
    request,
    method: "GET",
    path: `/api/v1/support/ticket-attachments/${encodeURIComponent(attachmentID)}/download`,
  });
}
