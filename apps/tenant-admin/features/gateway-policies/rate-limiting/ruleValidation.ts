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

export const rateLimitConditionSchema = z
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
    subjectKind: z.enum(["user", "team", "virtual_account", "organization_unit"]).optional(),
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

export const rateLimitRuleSchema = z.object({
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
    conditions: z.array(rateLimitConditionSchema).default([]),
  }),
  limit: z.object({
    dimension: z.enum(["requests", "tokens"]).default("requests"),
    max: z.number().int().min(1, "Limit must be greater than zero."),
    window: z.object({
      value: z.number().int().min(1, "Time window value must be greater than zero."),
      unit: z.enum(["second", "minute", "hour", "day", "week", "month"]),
    }),
  }),
  conflict_behavior: z.enum(["first_match", "most_restrictive", "apply_all"]).default("most_restrictive"),
  response_behavior: z.literal("return_429").default("return_429"),
  soft_limit: z.boolean().default(false),
  retry_after_header: z.boolean().default(true),
  effective_from: z.string().optional(),
  expires_at: z.string().optional(),
});

export type RateLimitRuleFormInput = z.input<typeof rateLimitRuleSchema>;
