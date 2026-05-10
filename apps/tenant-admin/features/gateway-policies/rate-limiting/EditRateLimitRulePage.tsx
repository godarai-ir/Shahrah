"use client";

import { useEffect, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { getRateLimitRule } from "./ruleApi";
import { RateLimitRuleForm } from "./RateLimitRuleForm";
import type { RateLimitRule } from "./ruleTypes";

export function EditRateLimitRulePage({ ruleId }: { ruleId: string }) {
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rule, setRule] = useState<RateLimitRule | null>(null);

  useEffect(() => {
    let active = true;
    void getRateLimitRule(ruleId)
      .then((response) => {
        if (!active) return;
        setRule(response);
      })
      .catch((err) => {
        if (!active) return;
        setError(err instanceof Error ? err.message : "Failed to load rule.");
      })
      .finally(() => {
        if (!active) return;
        setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [ruleId]);

  if (isLoading) return <p className="text-sm text-muted-foreground">Loading...</p>;
  if (error || !rule) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{error ?? "Rule not found."}</AlertDescription>
      </Alert>
    );
  }
  return <RateLimitRuleForm initialRule={rule} />;
}
