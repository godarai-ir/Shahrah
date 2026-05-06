"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;

type Row = Record<string, unknown>;

export default function ModelMetricsDashboard() {
  const t = useTranslations("metrics");
  const common = useTranslations("common");

  const [viewBy, setViewBy] = useState<(typeof VIEW_BY)[number]>("modelName");
  const [modelFilter, setModelFilter] = useState("");
  const [status, setStatus] = useState("all");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [rows, setRows] = useState<Row[]>([]);

  const [from, setFrom] = useState(() => new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString());
  const [to, setTo] = useState(() => new Date().toISOString());

  const summary = useMemo(() => {
    const totalRequests = rows.reduce((acc, row) => acc + Number(row.countModelName ?? row.countAll ?? 0), 0);
    const inputTokens = rows.reduce((acc, row) => acc + Number(row.sumInputTokens ?? 0), 0);
    const outputTokens = rows.reduce((acc, row) => acc + Number(row.sumOutputTokens ?? 0), 0);
    const cost = rows.reduce((acc, row) => acc + Number(row.sumCostInUSD ?? 0), 0);
    return { totalRequests, inputTokens, outputTokens, cost };
  }, [rows]);

  async function runQuery() {
    setLoading(true);
    setError(null);
    try {
      const filters: Array<Record<string, unknown>> = [];
      if (modelFilter.trim()) {
        filters.push({ fieldName: "modelName", operator: "STRING_CONTAINS", value: modelFilter.trim() });
      }
      if (status === "success") {
        filters.push({ fieldName: "statusCode", operator: "LESS_THAN", value: 400 });
      } else if (status === "error") {
        filters.push({ fieldName: "statusCode", operator: "GREATER_THAN_EQUAL", value: 400 });
      }
      const res = await fetch("/api/metrics/query", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          startTs: from,
          endTs: to,
          datasource: "modelMetrics",
          type: "distribution",
          aggregations: [
            { type: "count", column: "modelName" },
            { type: "sum", column: "inputTokens" },
            { type: "sum", column: "outputTokens" },
            { type: "sum", column: "costInUSD" },
            { type: "p99", column: "latencyMs" },
          ],
          groupBy: [viewBy],
          filters,
        }),
      });
      const payload = await res.json();
      if (!res.ok) throw new Error(payload?.error?.message || common("requestError"));
      setRows(Array.isArray(payload?.data) ? payload.data : []);
    } catch (e) {
      setError(e instanceof Error ? e.message : common("requestError"));
      setRows([]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-3">
          <div className="grid gap-1.5">
            <Label>{t("filters.from")}</Label>
            <Input value={from} onChange={(e) => setFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.to")}</Label>
            <Input value={to} onChange={(e) => setTo(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.model")}</Label>
            <Input value={modelFilter} onChange={(e) => setModelFilter(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.viewBy")}</Label>
            <select className="h-10 rounded-md border bg-background px-3 text-sm" value={viewBy} onChange={(e) => setViewBy(e.target.value as (typeof VIEW_BY)[number])}>
              {VIEW_BY.map((v) => (
                <option key={v} value={v}>{v}</option>
              ))}
            </select>
          </div>
          <div className="grid gap-1.5">
            <Label>{t("filters.status")}</Label>
            <select className="h-10 rounded-md border bg-background px-3 text-sm" value={status} onChange={(e) => setStatus(e.target.value)}>
              <option value="all">all</option>
              <option value="success">success</option>
              <option value="error">error</option>
            </select>
          </div>
          <div className="flex items-end gap-2">
            <Button onClick={runQuery} disabled={loading}>{loading ? common("loading") : common("refresh")}</Button>
          </div>
        </CardContent>
      </Card>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-3 md:grid-cols-4">
        <MetricCard title={t("cards.totalRequests")} value={String(summary.totalRequests)} />
        <MetricCard title={t("cards.inputTokens")} value={String(summary.inputTokens)} />
        <MetricCard title={t("cards.outputTokens")} value={String(summary.outputTokens)} />
        <MetricCard title={t("cards.totalCost")} value={summary.cost.toFixed(6)} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("table.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{viewBy}</TableHead>
                <TableHead>count</TableHead>
                <TableHead>input</TableHead>
                <TableHead>output</TableHead>
                <TableHead>cost</TableHead>
                <TableHead>p99 latency</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>{t("table.empty")}</TableCell>
                </TableRow>
              ) : rows.map((row, idx) => (
                <TableRow key={idx}>
                  <TableCell>{String(row[viewBy] ?? common("notSet"))}</TableCell>
                  <TableCell>{String(row.countModelName ?? row.countAll ?? 0)}</TableCell>
                  <TableCell>{String(row.sumInputTokens ?? 0)}</TableCell>
                  <TableCell>{String(row.sumOutputTokens ?? 0)}</TableCell>
                  <TableCell>{String(row.sumCostInUSD ?? 0)}</TableCell>
                  <TableCell>{String(row.p99LatencyMs ?? 0)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold">{value}</p>
      </CardContent>
    </Card>
  );
}
