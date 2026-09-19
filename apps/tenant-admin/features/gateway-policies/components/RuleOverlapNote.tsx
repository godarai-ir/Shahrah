"use client";

import { useTranslations } from "next-intl";
import { TriangleAlert } from "lucide-react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { FieldDescription } from "@/components/ui/field";

import { policyRuleName, type PolicyRule } from "../policy-rules";

type RuleOverlapNoteProps = {
  overlapping: PolicyRule[];
  /** `row` sits under a list sentence and stays silent when there is no overlap; `editor` always answers. */
  variant: "row" | "editor";
};

export function RuleOverlapNote({ overlapping, variant }: RuleOverlapNoteProps) {
  const t = useTranslations("gatewayPolicies");
  const message = t("overlapWarning", {
    count: overlapping.length,
    names: overlapping.map(policyRuleName).join(t("sentence.listSeparator")),
  });

  if (variant === "row") {
    if (overlapping.length === 0) return null;
    return (
      <p className="flex items-start gap-1.5 text-xs leading-5 text-warning-soft-foreground">
        <TriangleAlert aria-hidden="true" className="mt-0.5 size-3.5 shrink-0" />
        <span>{message}</span>
      </p>
    );
  }

  if (overlapping.length === 0) return <FieldDescription>{t("overlapNone")}</FieldDescription>;
  return (
    <Alert variant="warning">
      <AlertDescription>
        {message} {t("overlapExplanation")}
      </AlertDescription>
    </Alert>
  );
}
