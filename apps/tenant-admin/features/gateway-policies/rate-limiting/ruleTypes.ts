export type RateLimitConditionField = "subject" | "model" | "metadata";

export type RateLimitConditionOperator =
  | "in"
  | "not_in"
  | "equals"
  | "not_equals"
  | "contains"
  | "starts_with"
  | "ends_with"
  | "exists"
  | "not_exists"
  | "gt"
  | "gte"
  | "lt"
  | "lte";

export type RateLimitCondition = {
  field: RateLimitConditionField;
  operator: RateLimitConditionOperator;
  values?: string[];
  value?: string | number | boolean;
  metadataKey?: string;
  subjectKind?: "user" | "team" | "virtual_account" | "organization_unit";
};

export type RateLimitConditionGroup = {
  combinator: "AND";
  conditions: RateLimitCondition[];
};

export type RateLimitConfig = {
  dimension: "requests" | "tokens";
  max: number;
  window: {
    value: number;
    unit: "second" | "minute" | "hour" | "day" | "week" | "month";
  };
};

export type RateLimitConflictBehavior = "first_match" | "most_restrictive" | "apply_all";
export type RateLimitResponseBehavior = "return_429";

export type RateLimitRule = {
  id: string;
  rule_id: string;
  name?: string;
  description?: string;
  enabled: boolean;
  priority: number;
  conditions: RateLimitConditionGroup;
  limit: RateLimitConfig;
  conflict_behavior: RateLimitConflictBehavior;
  response_behavior: RateLimitResponseBehavior;
  soft_limit: boolean;
  retry_after_header: boolean;
  effective_from?: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
};

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
