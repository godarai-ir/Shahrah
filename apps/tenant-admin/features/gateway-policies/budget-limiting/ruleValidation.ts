import { z } from "zod";

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
  limit_to: z.number().positive("Budget must be greater than zero."),
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
        value === "org_unit" ||
        value === "model" ||
        value === "virtualaccount" ||
        value.startsWith("metadata."),
      "Invalid apply-per value.",
    ),
  block_on_budget_exceed: z.boolean().default(true),
  alerts: z
    .object({
      enabled: z.boolean().default(false),
      thresholds: z.array(z.number().int()).optional(),
      notification_target: z
        .array(
          z.object({
            type: z.enum(["email", "slack-webhook", "slack-bot"]),
            notification_channel: z.string().optional(),
            to_emails: z.array(z.string().email()).optional(),
            channels: z.array(z.string()).optional(),
          }),
        )
        .optional(),
    })
    .default({ enabled: false }),
});

export const budgetLimitRuleResponseSchema = budgetLimitRuleSchema.extend({
  id: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type BudgetLimitRuleFormInput = z.input<typeof budgetLimitRuleSchema>;
