import { NextRequest } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export async function DELETE(request: NextRequest) {

  return proxyToControlPlane({
    request,
    method: "DELETE",
    path: "/api/v1/personal-access-tokens/revoke/all",
  });
}
