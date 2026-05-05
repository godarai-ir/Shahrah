"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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

type TraceListResponse = {
  data?: RequestTrace[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type TraceManagementProps = {
  initialData: RequestTrace[];
  initialTotal: number;
  initialError: string | null;
};

const PAGE_SIZE = 25;

export default function TraceManagement(props: TraceManagementProps) {
  const t = useTranslations("traces");
  const common = useTranslations("common");

  const [items, setItems] = useState<RequestTrace[]>(props.initialData);
  const [total, setTotal] = useState(props.initialTotal);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(props.initialError);
  const [loading, setLoading] = useState(false);

  const [traceID, setTraceID] = useState("");
  const [model, setModel] = useState("");
  const [route, setRoute] = useState("");
  const [statusCode, setStatusCode] = useState("");

  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const canPrev = offset > 0;
  const canNext = offset + items.length < total;
  const showingFrom = total === 0 ? 0 : offset + 1;
  const showingTo = Math.min(offset + items.length, total);

  const queryString = useMemo(() => {
    const params = new URLSearchParams();
    params.set("limit", String(PAGE_SIZE));
    params.set("offset", String(offset));
    if (traceID.trim()) params.set("trace_id", traceID.trim());
    if (model.trim()) params.set("model", model.trim());
    if (route.trim()) params.set("route", route.trim());
    if (statusCode.trim()) params.set("status_code", statusCode.trim());
    if (from.trim()) params.set("from", from.trim());
    if (to.trim()) params.set("to", to.trim());
    return params.toString();
  }, [offset, traceID, model, route, statusCode, from, to]);

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [queryString]);

  async function load() {
    setLoading(true);
    try {
      const payload = await requestJSON<TraceListResponse>(`/api/traces?${queryString}`);
      setItems(payload.data ?? []);
      setTotal(payload.total ?? 0);
      setError(null);
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
    } finally {
      setLoading(false);
    }
  }

  function applyFilters() {
    setOffset(0);
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="traceID">{t("filters.traceId")}</Label>
            <Input id="traceID" value={traceID} onChange={(e) => setTraceID(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="model">{t("filters.model")}</Label>
            <Input id="model" value={model} onChange={(e) => setModel(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="route">{t("filters.route")}</Label>
            <Input id="route" value={route} onChange={(e) => setRoute(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="status">{t("filters.statusCode")}</Label>
            <Input id="status" value={statusCode} onChange={(e) => setStatusCode(e.target.value)} placeholder="200" />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="from">{t("filters.from")}</Label>
            <Input id="from" value={from} onChange={(e) => setFrom(e.target.value)} placeholder="2026-01-01T00:00:00Z" />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="to">{t("filters.to")}</Label>
            <Input id="to" value={to} onChange={(e) => setTo(e.target.value)} placeholder="2026-01-01T23:59:59Z" />
          </div>
          <div className="md:col-span-3 flex gap-2">
            <Button onClick={applyFilters}>{t("actions.apply")}</Button>
            <Button
              variant="outline"
              onClick={() => {
                setTraceID("");
                setModel("");
                setRoute("");
                setStatusCode("");
                setFrom("");
                setTo("");
                setOffset(0);
              }}
            >
              {t("actions.clear")}
            </Button>
          </div>
        </CardContent>
      </Card>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t("list.title")}</CardTitle>
          <Button variant="outline" onClick={() => void load()} disabled={loading}>
            {loading ? common("loading") : common("refresh")}
          </Button>
        </CardHeader>
        <CardContent className="space-y-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("list.columns.traceId")}</TableHead>
                <TableHead>{t("list.columns.route")}</TableHead>
                <TableHead>{t("list.columns.model")}</TableHead>
                <TableHead>{t("list.columns.status")}</TableHead>
                <TableHead>{t("list.columns.duration")}</TableHead>
                <TableHead>{t("list.columns.startedAt")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>{t("list.empty")}</TableCell>
                </TableRow>
              ) : (
                items.map((item, index) => (
                  <TableRow key={`${item.trace_id}:${item.started_at}:${item.request_id ?? ""}:${index}`}>
                    <TableCell className="font-mono text-xs">
                      <Link className="text-primary underline-offset-2 hover:underline" href={`/traces/${encodeURIComponent(item.trace_id)}`}>
                        {item.trace_id}
                      </Link>
                    </TableCell>
                    <TableCell>{item.route}</TableCell>
                    <TableCell>{item.model || common("notSet")}</TableCell>
                    <TableCell>
                      <Badge variant={item.status_code >= 500 ? "destructive" : "secondary"}>{item.status_code}</Badge>
                    </TableCell>
                    <TableCell>{item.duration_ms} ms</TableCell>
                    <TableCell>{new Date(item.started_at).toLocaleString()}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <p>{t("pagination.showing", { from: showingFrom, to: showingTo, total })}</p>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" disabled={!canPrev || loading} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}>
                {t("pagination.prev")}
              </Button>
              <Button size="sm" variant="outline" disabled={!canNext || loading} onClick={() => setOffset(offset + PAGE_SIZE)}>
                {t("pagination.next")}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
