"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Copy, Pencil, Power, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { createBudgetLimitRule, deleteBudgetLimitRule, listBudgetLimitRules, updateBudgetLimitRuleStatus } from "./ruleApi";
import type { BudgetLimitRule } from "./ruleTypes";

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

function unitKey(unit: BudgetLimitRule["unit"]) {
  if (unit === "cost_per_day") return "cost_per_day";
  if (unit === "cost_per_week") return "cost_per_week";
  return "cost_per_month";
}

export function BudgetLimitRulesPage() {
  const t = useTranslations("budgetLimitRules");
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [items, setItems] = useState<BudgetLimitRule[]>([]);
  const [pendingDeleteRule, setPendingDeleteRule] = useState<BudgetLimitRule | null>(null);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const response = await listBudgetLimitRules(query);
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

  async function confirmDeleteRule() {
    if (!pendingDeleteRule || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteBudgetLimitRule(pendingDeleteRule.rule_id);
      setPendingDeleteRule(null);
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: error instanceof Error ? error.message : t("messages.deleteError") });
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("description")}</CardDescription>
          </div>
          <Button asChild>
            <Link href="/gateway-policies/budget-limiting-rules/new">{t("actions.new")}</Link>
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
                  <TableHead>{t("table.unit")}</TableHead>
                  <TableHead>{t("table.status")}</TableHead>
                  <TableHead>{t("table.updatedAt")}</TableHead>
                  <TableHead>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-medium">{rule.rule_id}</TableCell>
                    <TableCell>{rule.limit_to}</TableCell>
                    <TableCell>{t(`units.${unitKey(rule.unit)}`)}</TableCell>
                    <TableCell>
                      <Badge variant={rule.enabled ? "default" : "secondary"}>
                        {rule.enabled ? t("status.enabled") : t("status.disabled")}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatJalaliDateTime(rule.updated_at)}</TableCell>
                    <TableCell className="space-x-2">
                      <Button asChild size="icon" variant="outline" title={t("actions.edit")}>
                        <Link href={`/gateway-policies/budget-limiting-rules/${encodeURIComponent(rule.rule_id)}`} aria-label={t("actions.edit")}>
                          <Pencil className="h-4 w-4" />
                        </Link>
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        title={rule.enabled ? t("actions.disable") : t("actions.enable")}
                        onClick={() => void updateBudgetLimitRuleStatus(rule.rule_id, !rule.enabled).then(load)}
                      >
                        <Power className="h-4 w-4" />
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        title={t("actions.duplicate")}
                        onClick={() =>
                          void createBudgetLimitRule({
                            ...rule,
                            rule_id: `${rule.rule_id}-copy`,
                            name: `${rule.name ?? rule.rule_id} ${t("actions.copySuffix")}`,
                          } as never).then(load)
                        }
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                      <Button size="icon" variant="outline" title={t("actions.delete")} onClick={() => setPendingDeleteRule(rule)}>
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

      <Dialog open={pendingDeleteRule !== null} onOpenChange={(open) => !open && setPendingDeleteRule(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>{t("deleteDialog.description", { ruleId: pendingDeleteRule?.rule_id ?? "-" })}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" disabled={isDeleting} onClick={() => setPendingDeleteRule(null)}>
              {t("deleteDialog.cancel")}
            </Button>
            <Button type="button" variant="destructive" onClick={() => void confirmDeleteRule()} disabled={isDeleting}>
              {isDeleting ? t("deleteDialog.deleting") : t("deleteDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

