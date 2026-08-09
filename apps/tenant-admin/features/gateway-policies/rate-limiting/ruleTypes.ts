import type { z } from "zod";

import type {
  rateLimitConditionSchema,
  rateLimitRuleResponseSchema,
  rateLimitRuleSchema,
} from "./ruleValidation";

export type RateLimitCondition = z.infer<typeof rateLimitConditionSchema>;
export type RateLimitConditionField = RateLimitCondition["field"];
export type RateLimitConditionOperator = RateLimitCondition["operator"];
export type RateLimitConditionGroup = z.infer<typeof rateLimitRuleSchema>["conditions"];
export type RateLimitConfig = z.infer<typeof rateLimitRuleSchema>["limit"];
export type RateLimitConflictBehavior = z.infer<typeof rateLimitRuleSchema>["conflict_behavior"];
export type RateLimitResponseBehavior = z.infer<typeof rateLimitRuleSchema>["response_behavior"];
export type RateLimitRule = z.infer<typeof rateLimitRuleResponseSchema>;

export type GatewayRequestContext = {
  subject?: string;
  model?: string;
  provider?: string;
  virtualAccount?: string;
  userId?: string;
  apiKeyId?: string;
  organizationUnitId?: string;
  teamId?: string;
  metadata?: Record<string, string | number | boolean>;
};
