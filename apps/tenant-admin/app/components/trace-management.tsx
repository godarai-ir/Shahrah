"use client";

import Link from "next/link";
import { useEffect, useMemo, useState, type ReactNode } from "react";
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
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts?: number;
};

type TraceListResponse = {
  data?: RequestTrace[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type FacetBucket = {
  value: string;
  count: number;
};

type StatusFacetBucket = {
  status_code: number;
  count: number;
};

type TraceFacets = {
  providers?: FacetBucket[];
  models?: FacetBucket[];
  routes?: FacetBucket[];
  error_codes?: FacetBucket[];
  status_codes?: StatusFacetBucket[];
};

type FacetsResponse = {
  data?: TraceFacets;
};

type TraceFilters = {
  traceID: string;
  requestID: string;
  model: string;
  provider: string;
  route: string;
  statusCode: string;
  errorCode: string;
  from: string;
  to: string;
  sortBy: string;
  sortOrder: "asc" | "desc";
};

type TraceManagementProps = {
  initialData: RequestTrace[];
  initialTotal: number;
  initialError: string | null;
};

const PAGE_SIZE = 25;

const DEFAULT_FILTERS: TraceFilters = {
  traceID: "",
  requestID: "",
  model: "",
  provider: "",
  route: "",
  statusCode: "",
  errorCode: "",
  from: "",
  to: "",
  sortBy: "started_at",
  sortOrder: "desc",
};

export default function TraceManagement(props: TraceManagementProps) {
  const t = useTranslations("traces");
  const common = useTranslations("common");

  const [items, setItems] = useState<RequestTrace[]>(props.initialData);
  const [total, setTotal] = useState(props.initialTotal);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(props.initialError);
  const [loading, setLoading] = useState(false);
  const [facets, setFacets] = useState<TraceFacets>({});

  const [draft, setDraft] = useState<TraceFilters>(DEFAULT_FILTERS);
  const [applied, setApplied] = useState<TraceFilters>(DEFAULT_FILTERS);

  const canPrev = offset > 0;
  const canNext = offset + items.length < total;
  const showingFrom = total === 0 ? 0 : offset + 1;
  const showingTo = Math.min(offset + items.length, total);

  const filtersQueryString = useMemo(() => {
    const params = new URLSearchParams();
    if (applied.traceID) params.set("trace_id", applied.traceID);
    if (applied.requestID) params.set("request_id", applied.requestID);
    if (applied.model) params.set("model", applied.model);
    if (applied.provider) params.set("provider", applied.provider);
    if (applied.route) params.set("route", applied.route);
    if (applied.statusCode) params.set("status_code", applied.statusCode);
    if (applied.errorCode) params.set("error_code", applied.errorCode);
    if (applied.from) params.set("from", applied.from);
    if (applied.to) params.set("to", applied.to);
    if (applied.sortBy) params.set("sort_by", applied.sortBy);
    if (applied.sortOrder) params.set("sort_order", applied.sortOrder);
    return params.toString();
  }, [applied]);

  const listQueryString = useMemo(() => {
    const params = new URLSearchParams(filtersQueryString);
    params.set("limit", String(PAGE_SIZE));
    params.set("offset", String(offset));
    return params.toString();
  }, [filtersQueryString, offset]);

  const activeFilters = useMemo(() => {
    const out: Array<{ key: string; label: string; value: string }> = [];
    if (applied.traceID) out.push({ key: "traceID", label: t("activeFilters.traceId"), value: applied.traceID });
    if (applied.requestID) out.push({ key: "requestID", label: t("activeFilters.requestId"), value: applied.requestID });
    if (applied.model) out.push({ key: "model", label: t("activeFilters.model"), value: applied.model });
    if (applied.provider) out.push({ key: "provider", label: t("activeFilters.provider"), value: applied.provider });
    if (applied.route) out.push({ key: "route", label: t("activeFilters.route"), value: applied.route });
    if (applied.statusCode) out.push({ key: "statusCode", label: t("activeFilters.status"), value: applied.statusCode });
    if (applied.errorCode) out.push({ key: "errorCode", label: t("activeFilters.errorCode"), value: applied.errorCode });
    if (applied.from) out.push({ key: "from", label: t("activeFilters.from"), value: applied.from });
    if (applied.to) out.push({ key: "to", label: t("activeFilters.to"), value: applied.to });
    return out;
  }, [applied, t]);

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listQueryString]);

  async function load() {
    setLoading(true);
    try {
      const [listPayload, facetsPayload] = await Promise.all([
        requestJSON<TraceListResponse>(`/api/traces?${listQueryString}`),
        requestJSON<FacetsResponse>(`/api/traces/facets${filtersQueryString ? `?${filtersQueryString}` : ""}`),
      ]);
      setItems(listPayload.data ?? []);
      setTotal(listPayload.total ?? 0);
      setFacets(facetsPayload.data ?? {});
      setError(null);
    } catch (err) {
      setError(normalizeError(err, common("requestError")));
    } finally {
      setLoading(false);
    }
  }

  function applyFilters() {
    setApplied(normalizeFilters(draft));
    setOffset(0);
  }

  function clearFilters() {
    setDraft(DEFAULT_FILTERS);
    setApplied(DEFAULT_FILTERS);
    setOffset(0);
  }

  function removeFilter(key: keyof TraceFilters) {
    const nextDraft = { ...draft, [key]: DEFAULT_FILTERS[key] };
    const nextApplied = { ...applied, [key]: DEFAULT_FILTERS[key] };
    setDraft(nextDraft);
    setApplied(nextApplied);
    setOffset(0);
  }

  function selectFacet(update: Partial<TraceFilters>) {
    const nextDraft = normalizeFilters({ ...draft, ...update });
    const nextApplied = normalizeFilters({ ...applied, ...update });
    setDraft(nextDraft);
    setApplied(nextApplied);
    setOffset(0);
  }

  function setSort(sortBy: string) {
    const nextOrder: "asc" | "desc" =
      applied.sortBy === sortBy && applied.sortOrder === "desc" ? "asc" : "desc";
    const next = normalizeFilters({ ...applied, sortBy, sortOrder: nextOrder });
    setDraft(next);
    setApplied(next);
    setOffset(0);
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-4">
          <Field label={t("filters.traceId")}>
            <Input value={draft.traceID} onChange={(e) => setDraft((prev) => ({ ...prev, traceID: e.target.value }))} />
          </Field>
          <Field label={t("filters.requestId")}>
            <Input value={draft.requestID} onChange={(e) => setDraft((prev) => ({ ...prev, requestID: e.target.value }))} />
          </Field>
          <Field label={t("filters.model")}>
            <Input value={draft.model} onChange={(e) => setDraft((prev) => ({ ...prev, model: e.target.value }))} />
          </Field>
          <Field label={t("filters.provider")}>
            <Input value={draft.provider} onChange={(e) => setDraft((prev) => ({ ...prev, provider: e.target.value }))} />
          </Field>
          <Field label={t("filters.route")}>
            <Input value={draft.route} onChange={(e) => setDraft((prev) => ({ ...prev, route: e.target.value }))} />
          </Field>
          <Field label={t("filters.statusCode")}>
            <Input value={draft.statusCode} onChange={(e) => setDraft((prev) => ({ ...prev, statusCode: e.target.value }))} placeholder={t("filters.placeholders.statusCode")} />
          </Field>
          <Field label={t("filters.errorCode")}>
            <Input value={draft.errorCode} onChange={(e) => setDraft((prev) => ({ ...prev, errorCode: e.target.value }))} placeholder={t("filters.placeholders.errorCode")} />
          </Field>
          <Field label={t("filters.from")}>
            <Input value={draft.from} onChange={(e) => setDraft((prev) => ({ ...prev, from: e.target.value }))} placeholder={t("filters.placeholders.from")} />
          </Field>
          <Field label={t("filters.to")}>
            <Input value={draft.to} onChange={(e) => setDraft((prev) => ({ ...prev, to: e.target.value }))} placeholder={t("filters.placeholders.to")} />
          </Field>
          <div className="md:col-span-4 flex flex-wrap gap-2">
            <Button onClick={applyFilters}>{t("actions.apply")}</Button>
            <Button variant="outline" onClick={clearFilters}>
              {t("actions.clear")}
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                const now = new Date();
                const from = new Date(now.getTime() - 60 * 60 * 1000).toISOString();
                setDraft((prev) => ({ ...prev, from, to: now.toISOString() }));
              }}
            >
              {t("actions.last1h")}
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                const now = new Date();
                const from = new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString();
                setDraft((prev) => ({ ...prev, from, to: now.toISOString() }));
              }}
            >
              {t("actions.last24h")}
            </Button>
          </div>
          <div className="md:col-span-4 flex flex-wrap gap-2">
            {activeFilters.length === 0 ? (
              <span className="text-xs text-muted-foreground">{t("activeFilters.none")}</span>
            ) : (
              activeFilters.map((item) => (
                <Badge key={item.key} variant="secondary" className="cursor-pointer" onClick={() => removeFilter(item.key as keyof TraceFilters)}>
                  {item.label}={item.value} {t("actions.removeFilter")}
                </Badge>
              ))
            )}
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
        <CardHeader>
          <CardTitle>{t("facets.title")}</CardTitle>
          <CardDescription>{t("facets.description")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          <FacetSection
            title={t("facets.providers")}
            buckets={facets.providers}
            activeValue={applied.provider}
            onSelect={(value) => selectFacet({ provider: value })}
            noDataLabel={t("facets.noData")}
            emptyValueLabel={t("facets.empty")}
          />
          <FacetSection
            title={t("facets.models")}
            buckets={facets.models}
            activeValue={applied.model}
            onSelect={(value) => selectFacet({ model: value })}
            noDataLabel={t("facets.noData")}
            emptyValueLabel={t("facets.empty")}
          />
          <FacetSection
            title={t("facets.routes")}
            buckets={facets.routes}
            activeValue={applied.route}
            onSelect={(value) => selectFacet({ route: value })}
            noDataLabel={t("facets.noData")}
            emptyValueLabel={t("facets.empty")}
          />
          <FacetSection
            title={t("facets.errorCodes")}
            buckets={facets.error_codes}
            activeValue={applied.errorCode}
            onSelect={(value) => selectFacet({ errorCode: value })}
            noDataLabel={t("facets.noData")}
            emptyValueLabel={t("facets.empty")}
          />
          <StatusFacetSection
            title={t("facets.statusCodes")}
            buckets={facets.status_codes}
            activeValue={applied.statusCode}
            onSelect={(value) => selectFacet({ statusCode: String(value) })}
            noDataLabel={t("facets.noData")}
          />
        </CardContent>
      </Card>

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
                <SortableHead
                  label={t("list.columns.requestId")}
                  isActive={applied.sortBy === "request_id"}
                  sortOrder={applied.sortOrder}
                  onClick={() => setSort("request_id")}
                />
                <TableHead>{t("list.columns.route")}</TableHead>
                <TableHead>{t("list.columns.model")}</TableHead>
                <TableHead>{t("list.columns.provider")}</TableHead>
                <SortableHead
                  label={t("list.columns.status")}
                  isActive={applied.sortBy === "status_code"}
                  sortOrder={applied.sortOrder}
                  onClick={() => setSort("status_code")}
                />
                <SortableHead
                  label={t("list.columns.duration")}
                  isActive={applied.sortBy === "duration_ms"}
                  sortOrder={applied.sortOrder}
                  onClick={() => setSort("duration_ms")}
                />
                <SortableHead
                  label={t("list.columns.startedAt")}
                  isActive={applied.sortBy === "started_at"}
                  sortOrder={applied.sortOrder}
                  onClick={() => setSort("started_at")}
                />
                <TableHead>{t("list.columns.attempts")}</TableHead>
                <TableHead>{t("list.columns.error")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={10}>{t("list.empty")}</TableCell>
                </TableRow>
              ) : (
                items.map((item, index) => (
                  <TableRow key={`${item.trace_id}:${item.started_at}:${item.request_id ?? ""}:${index}`}>
                    <TableCell className="font-mono text-xs">
                      <Link className="text-primary underline-offset-2 hover:underline" href={`/traces/${encodeURIComponent(item.trace_id)}`}>
                        {item.trace_id}
                      </Link>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{item.request_id || "-"}</TableCell>
                    <TableCell>{item.route}</TableCell>
                    <TableCell>{item.model || common("notSet")}</TableCell>
                    <TableCell>{item.provider || common("notSet")}</TableCell>
                    <TableCell>
                      <Badge variant={item.status_code >= 500 ? "destructive" : "secondary"}>{item.status_code}</Badge>
                    </TableCell>
                    <TableCell>
                      {item.duration_ms} {t("list.units.ms")}
                    </TableCell>
                    <TableCell>{formatDateTime(item.started_at, t("list.units.utc"))}</TableCell>
                    <TableCell>{item.attempts ?? 0}</TableCell>
                    <TableCell className="text-xs">{item.error_code || "-"}</TableCell>
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

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="grid gap-1.5">
      <Label>{label}</Label>
      {children}
    </div>
  );
}

function SortableHead({
  label,
  isActive,
  sortOrder,
  onClick,
}: {
  label: string;
  isActive: boolean;
  sortOrder: "asc" | "desc";
  onClick: () => void;
}) {
  return (
    <TableHead>
      <button type="button" className="inline-flex items-center gap-1 text-left hover:underline" onClick={onClick}>
        {label}
        <span className="text-xs text-muted-foreground">{isActive ? (sortOrder === "asc" ? "^" : "v") : ""}</span>
      </button>
    </TableHead>
  );
}

function FacetSection({
  title,
  buckets,
  activeValue,
  onSelect,
  noDataLabel,
  emptyValueLabel,
}: {
  title: string;
  buckets?: FacetBucket[];
  activeValue?: string;
  onSelect: (value: string) => void;
  noDataLabel: string;
  emptyValueLabel: string;
}) {
  return (
    <div className="space-y-2">
      <p className="text-sm font-medium">{title}</p>
      <div className="flex flex-wrap gap-2">
        {(buckets ?? []).length === 0 ? (
          <span className="text-xs text-muted-foreground">{noDataLabel}</span>
        ) : (
          (buckets ?? []).slice(0, 8).map((item) => (
            <Badge
              key={`${title}:${item.value}`}
              variant={activeValue === (item.value || "") ? "secondary" : "outline"}
              className="cursor-pointer hover:bg-accent"
              onClick={() => onSelect(item.value || "")}
            >
              {item.value || emptyValueLabel} - {item.count}
            </Badge>
          ))
        )}
      </div>
    </div>
  );
}

function StatusFacetSection({
  title,
  buckets,
  activeValue,
  onSelect,
  noDataLabel,
}: {
  title: string;
  buckets?: StatusFacetBucket[];
  activeValue?: string;
  onSelect: (statusCode: number) => void;
  noDataLabel: string;
}) {
  return (
    <div className="space-y-2">
      <p className="text-sm font-medium">{title}</p>
      <div className="flex flex-wrap gap-2">
        {(buckets ?? []).length === 0 ? (
          <span className="text-xs text-muted-foreground">{noDataLabel}</span>
        ) : (
          (buckets ?? []).slice(0, 8).map((item) => (
            <Badge
              key={`${title}:${item.status_code}`}
              variant={activeValue === String(item.status_code) ? "secondary" : "outline"}
              className="cursor-pointer hover:bg-accent"
              onClick={() => onSelect(item.status_code)}
            >
              {item.status_code} - {item.count}
            </Badge>
          ))
        )}
      </div>
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

function formatDateTime(value: string, utcLabel: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toISOString().replace("T", " ").replace("Z", ` ${utcLabel}`);
}

function normalizeFilters(filters: TraceFilters): TraceFilters {
  return {
    ...filters,
    traceID: filters.traceID.trim(),
    requestID: filters.requestID.trim(),
    model: filters.model.trim(),
    provider: filters.provider.trim(),
    route: filters.route.trim(),
    statusCode: filters.statusCode.trim(),
    errorCode: filters.errorCode.trim(),
    from: filters.from.trim(),
    to: filters.to.trim(),
    sortBy: (filters.sortBy || "started_at").trim(),
    sortOrder: filters.sortOrder === "asc" ? "asc" : "desc",
  };
}

