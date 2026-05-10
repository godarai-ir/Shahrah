"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

import { getBudgetLimitRule } from "./ruleApi";
import { BudgetLimitRuleForm } from "./BudgetLimitRuleForm";
import type { BudgetLimitRule } from "./ruleTypes";

export function EditBudgetLimitRulePage({ ruleId }: { ruleId: string }) {
  const t = useTranslations("budgetLimitRules");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rule, setRule] = useState<BudgetLimitRule | null>(null);

  useEffect(() => {
    let active = true;
    void getBudgetLimitRule(ruleId)
      .then((response) => {
        if (!active) return;
        setRule(response);
      })
      .catch((err) => {
        if (!active) return;
        setError(err instanceof Error ? err.message : t("messages.loadRuleError"));
      })
      .finally(() => {
        if (!active) return;
        setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [ruleId, t]);

  if (isLoading) return <p className="text-sm text-muted-foreground">{t("messages.loading")}</p>;
  if (error || !rule) {
    return (
      <Alert variant="destructive">
        <AlertTitle>{t("alerts.error")}</AlertTitle>
        <AlertDescription>{error ?? t("messages.ruleNotFound")}</AlertDescription>
      </Alert>
    );
  }
  return <BudgetLimitRuleForm initialRule={rule} />;
}
