import { z } from "zod";

const CREDIT_AMOUNT_PATTERN = /^\d+(\.\d{1,2})?$/;

/** The gateway's thresholds when a rule sets none (credits.DefaultAlertThresholds). */
export const DEFAULT_ALERT_THRESHOLDS = [80, 100] as const;

const operatorWithValue = new Set([
  "equals",
  "not_equals",
  "contains",
  "starts_with",
  "ends_with",
  "gt",
  "gte",
  "lt",
  "lte",
]);

export const budgetLimitConditionSchema = z
  .object({
    field: z.enum(["subject", "model", "metadata"]),
    operator: z.enum([
      "in",
      "not_in",
      "equals",
      "not_equals",
      "contains",
      "starts_with",
      "ends_with",
      "exists",
      "not_exists",
      "gt",
      "gte",
      "lt",
      "lte",
    ]),
    values: z.array(z.string().min(1)).optional(),
    value: z.union([z.string(), z.number(), z.boolean()]).optional(),
    metadataKey: z.string().optional(),
    subjectKind: z.enum(["user", "team", "virtual_account"]).optional(),
    /** Display names captured when the value was picked; the gateway ignores this key. */
    valueLabels: z.record(z.string(), z.string()).optional(),
  })
  .superRefine((val, ctx) => {
    if ((val.field === "subject" || val.field === "model") && val.operator !== "in") {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["operator"], message: "Operator must be in." });
    }
    if ((val.field === "subject" || val.field === "model") && (!val.values || val.values.length === 0 || !val.values.some((v) => v.trim()))) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["values"], message: "At least one value is required." });
    }
    if (val.field === "metadata") {
      if (!val.metadataKey?.trim()) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["metadataKey"], message: "Metadata key is required." });
      }
      if (operatorWithValue.has(val.operator) && (val.value === undefined || `${val.value}`.trim() === "")) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["value"], message: "Metadata value is required." });
      }
    }
  });

export const budgetLimitRuleSchema = z.object({
  rule_id: z
    .string()
    .min(1, "Rule ID is required.")
    .regex(/^[a-z0-9_-]+$/, "Rule ID must contain lowercase letters, numbers, underscore or hyphen."),
  name: z.string().optional(),
  description: z.string().optional(),
  enabled: z.boolean().default(true),
  priority: z.number().int().nonnegative().default(100),
  conditions: z.object({
    combinator: z.literal("AND").default("AND"),
    conditions: z.array(budgetLimitConditionSchema).default([]),
  }),
  /** Credits as a decimal string with at most two places. */
  limit: z
    .string()
    .trim()
    .regex(CREDIT_AMOUNT_PATTERN, "Enter credits with at most two decimals.")
    .refine((value) => Number(value) > 0, "Budget must be greater than zero."),
  unit: z.enum(["cost_per_day", "cost_per_week", "cost_per_month"]),
  budget_applies_per: z
    .string()
    .optional()
    .refine(
      (value) =>
        !value ||
        value === "tenant" ||
        value === "user" ||
        value === "team" ||
        value === "model" ||
        value === "virtualaccount" ||
        value.startsWith("metadata."),
      "Invalid apply-per value.",
    ),
  /** block refuses requests past the limit; alert only counts them and logs the crossing. */
  on_exceed: z.enum(["block", "alert"]).default("block"),
  /** Percent of the limit at which the gateway records an alert, 1–1000. */
  alert_thresholds: z.array(z.number().int().min(1).max(1000)).max(10).default([...DEFAULT_ALERT_THRESHOLDS]),
});

export const budgetLimitRuleResponseSchema = budgetLimitRuleSchema.extend({
  // Rules written before the credit cutover may still lack a credit limit.
  limit: z.string().optional(),
  limit_to: z.number().optional(),
  on_exceed: z.enum(["block", "alert"]).optional(),
  alert_thresholds: z.array(z.number().int()).optional(),
  block_on_budget_exceed: z.boolean().optional(),
  id: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type BudgetLimitRuleFormInput = z.input<typeof budgetLimitRuleSchema>;
