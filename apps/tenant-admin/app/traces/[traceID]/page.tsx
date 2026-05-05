import Link from "next/link";

import AdminShell from "@/app/components/admin-shell";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RequestTrace = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
};

type TraceSpan = {
  span_id: string;
  trace_id: string;
  parent_span_id?: string;
  name: string;
  kind: string;
  started_at: string;
  duration_ms: number;
  status: string;
  attributes: Record<string, unknown>;
};

type SpanResponse = { data?: TraceSpan[] };

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ traceID: string }> };

export default async function TraceDetailPage({ params }: Params) {
  const auth = await requireAuth("/login?next=/traces");
  const tenantSlug = auth.tenantSlug;
  const { traceID } = await params;
  const normalized = traceID.trim();

  const [traceResult, spansResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v1/traces/${encodeURIComponent(normalized)}`,
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: `/api/v1/traces/${encodeURIComponent(normalized)}/spans`,
      userID: auth.user.id,
    }),
  ]);

  const trace = (traceResult.ok ? traceResult.payload : null) as RequestTrace | null;
  const spansPayload = (spansResult.payload ?? {}) as SpanResponse;
  const spans = spansResult.ok ? spansPayload.data ?? [] : [];

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">Trace Detail</h1>
          <Link href="/traces" className="text-sm text-primary underline-offset-2 hover:underline">
            Back to traces
          </Link>
        </div>

        {!trace ? (
          <Card>
            <CardHeader>
              <CardTitle>Trace Not Found</CardTitle>
              <CardDescription>{traceResult.errorMessage ?? "The requested trace is not available."}</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle className="font-mono text-sm">{trace.trace_id}</CardTitle>
                <CardDescription>Request summary</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-2 text-sm md:grid-cols-2">
                <p><strong>Route:</strong> {trace.route}</p>
                <p><strong>Model:</strong> {trace.model || "-"}</p>
                <p><strong>Status:</strong> <Badge variant={trace.status_code >= 500 ? "destructive" : "secondary"}>{trace.status_code}</Badge></p>
                <p><strong>Duration:</strong> {trace.duration_ms} ms</p>
                <p><strong>Started:</strong> {new Date(trace.started_at).toLocaleString()}</p>
                <p><strong>Sampled:</strong> {trace.sampled ? "yes" : "no"}</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Span Timeline</CardTitle>
                <CardDescription>{spansResult.ok ? `${spans.length} span(s)` : spansResult.errorMessage ?? "Failed to load spans."}</CardDescription>
              </CardHeader>
              <CardContent>
                {spans.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No spans available for this trace.</p>
                ) : (
                  <div className="space-y-3">
                    {spans.map((span) => (
                      <div key={span.span_id} className="rounded-lg border p-3">
                        <div className="flex items-center justify-between">
                          <p className="font-medium">{span.name}</p>
                          <Badge variant={span.status === "error" ? "destructive" : "secondary"}>{span.status}</Badge>
                        </div>
                        <p className="text-xs text-muted-foreground font-mono mt-1">{span.span_id}</p>
                        <div className="mt-2 grid gap-1 text-sm md:grid-cols-3">
                          <p><strong>Kind:</strong> {span.kind}</p>
                          <p><strong>Duration:</strong> {span.duration_ms} ms</p>
                          <p><strong>Started:</strong> {new Date(span.started_at).toLocaleString()}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </AdminShell>
  );
}

