"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Copy, Pencil, Power, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { createRateLimitRule, deleteRateLimitRule, listRateLimitRules, updateRateLimitRuleStatus } from "./ruleApi";
import type { RateLimitRule } from "./ruleTypes";

const PERSIAN_DATE_TIME = new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
});

function formatJalaliDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return PERSIAN_DATE_TIME.format(date);
}

function windowUnitLabel(unit: RateLimitRule["limit"]["window"]["unit"]) {
  if (unit === "second") return "ثانیه";
  if (unit === "minute") return "دقیقه";
  if (unit === "hour") return "ساعت";
  if (unit === "day") return "روز";
  if (unit === "week") return "هفته";
  return "ماه";
}

export function RateLimitRulesPage() {
  const t = useTranslations("rateLimitRules");
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [items, setItems] = useState<RateLimitRule[]>([]);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const response = await listRateLimitRules(query);
      setItems(response.data);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.loadError") });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const filtered = useMemo(() => items, [items]);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("description")}</CardDescription>
          </div>
          <Button asChild>
            <Link href="/gateway-policies/rate-limiting-rules/new">{t("actions.new")}</Link>
          </Button>
        </div>
        <Input placeholder={t("searchPlaceholder")} value={query} onChange={(event) => setQuery(event.target.value)} />
      </CardHeader>
      <CardContent className="space-y-3">
        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}
        {filtered.length === 0 && !isLoading ? (
          <div className="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">{t("empty")}</div>
        ) : (
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.ruleId")}</TableHead>
                  <TableHead>{t("table.limit")}</TableHead>
                  <TableHead>{t("table.window")}</TableHead>
                  <TableHead>{t("table.status")}</TableHead>
                  <TableHead>{t("table.updatedAt")}</TableHead>
                  <TableHead>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((rule) => (
                  <TableRow
                    key={rule.id}
                    className={rule.limit.dimension === "tokens" ? "bg-amber-50/50 dark:bg-amber-950/20" : "bg-sky-50/40 dark:bg-sky-950/20"}
                  >
                    <TableCell className="font-medium">{rule.rule_id}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span>{rule.limit.max}</span>
                        <Badge variant="outline">{rule.limit.dimension === "tokens" ? "توکن" : "درخواست"}</Badge>
                      </div>
                    </TableCell>
                    <TableCell>{`${rule.limit.window.value} ${windowUnitLabel(rule.limit.window.unit)}`}</TableCell>
                    <TableCell>
                      <Badge variant={rule.enabled ? "default" : "secondary"}>{rule.enabled ? "فعال" : "غیرفعال"}</Badge>
                    </TableCell>
                    <TableCell>{formatJalaliDateTime(rule.updated_at)}</TableCell>
                    <TableCell className="space-x-2">
                      <Button asChild size="icon" variant="outline" title={t("actions.edit")}>
                        <Link href={`/gateway-policies/rate-limiting-rules/${encodeURIComponent(rule.rule_id)}`} aria-label={t("actions.edit")}>
                          <Pencil className="h-4 w-4" />
                        </Link>
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        title={rule.enabled ? t("actions.disable") : t("actions.enable")}
                        aria-label={rule.enabled ? t("actions.disable") : t("actions.enable")}
                        onClick={() => void updateRateLimitRuleStatus(rule.rule_id, !rule.enabled).then(load)}
                      >
                        <Power className="h-4 w-4" />
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        title={t("actions.duplicate")}
                        aria-label={t("actions.duplicate")}
                        onClick={() =>
                          void createRateLimitRule({
                            ...rule,
                            rule_id: `${rule.rule_id}-copy`,
                            name: `${rule.name ?? rule.rule_id} Copy`,
                          } as never).then(load)
                        }
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        title={t("actions.delete")}
                        aria-label={t("actions.delete")}
                        onClick={() => void deleteRateLimitRule(rule.rule_id).then(load)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
