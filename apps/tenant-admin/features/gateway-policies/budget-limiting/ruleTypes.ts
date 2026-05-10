export type BudgetLimitConditionField = "subject" | "model" | "metadata";

export type BudgetLimitConditionOperator =
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

export type BudgetLimitCondition = {
  field: BudgetLimitConditionField;
  operator: BudgetLimitConditionOperator;
  values?: string[];
  value?: string | number | boolean;
  metadataKey?: string;
  subjectKind?: "user" | "team" | "virtual_account" | "organization_unit";
};

export type BudgetLimitConditionGroup = {
  combinator: "AND";
  conditions: BudgetLimitCondition[];
};

export type BudgetLimitUnit = "cost_per_day" | "cost_per_week" | "cost_per_month";

export type BudgetLimitAlertTarget = {
  type: "email" | "slack-webhook" | "slack-bot";
  notification_channel?: string;
  to_emails?: string[];
  channels?: string[];
};

export type BudgetLimitAlertConfig = {
  enabled: boolean;
  thresholds?: number[];
  notification_target?: BudgetLimitAlertTarget[];
};

export type BudgetLimitRule = {
  id: string;
  rule_id: string;
  name?: string;
  description?: string;
  enabled: boolean;
  priority: number;
  conditions: BudgetLimitConditionGroup;
  limit_to: number;
  unit: BudgetLimitUnit;
  budget_applies_per?: string;
  block_on_budget_exceed: boolean;
  alerts?: BudgetLimitAlertConfig;
  created_at: string;
  updated_at: string;
};
