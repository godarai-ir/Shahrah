import { proxyToControlPlane } from "@/lib/control-plane";

export async function POST(request: Request) {
  let body: unknown = {};
  try {
    body = await request.json();
  } catch {
    body = {};
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/metrics/query",
    body,
  });
}
