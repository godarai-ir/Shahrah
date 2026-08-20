import { type NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane, proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = { params: Promise<{ path: string[] }> };

export async function GET(request: NextRequest, context: RouteContext) {
  return forwardJSON(request, context, "GET");
}

export async function POST(request: NextRequest, context: RouteContext) {
  const path = await controlPlanePath(request, context);
  if (!path.ok) return path.response;

  const contentType = request.headers.get("content-type") ?? "";
  if (contentType.includes("multipart/form-data")) {
    return proxyToControlPlaneBodyStream({ request, method: "POST", path: path.value });
  }
  return forwardJSON(request, context, "POST", path.value);
}

export async function PUT(request: NextRequest, context: RouteContext) {
  return forwardJSON(request, context, "PUT");
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  return forwardJSON(request, context, "PATCH");
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  return forwardJSON(request, context, "DELETE");
}

async function forwardJSON(
  request: NextRequest,
  context: RouteContext,
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
  resolvedPath?: string,
) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const path = resolvedPath ? { ok: true as const, value: resolvedPath } : await controlPlanePath(request, context);
  if (!path.ok) return path.response;

  let body: unknown;
  if (["POST", "PUT", "PATCH"].includes(method) && request.headers.get("content-length") !== "0") {
    body = await request.json().catch(() => undefined);
    if (body === undefined) {
      return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
    }
  }

  return proxyToControlPlane({
    request,
    method,
    path: path.value,
    body,
    userID: auth.user.id,
  });
}

async function controlPlanePath(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  if (!path.length || !["knowledge-bases", "chatbots"].includes(path[0] ?? "")) {
    return {
      ok: false as const,
      response: NextResponse.json({ error: { message: "Unsupported managed RAG path." } }, { status: 404 }),
    };
  }

  const safeSegments = path.map((segment) => encodeURIComponent(segment));
  return {
    ok: true as const,
    value: `/api/v1/${safeSegments.join("/")}${request.nextUrl.search}`,
  };
}
