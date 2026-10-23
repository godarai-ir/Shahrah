import type { z } from "zod";

import type {
  budgetLimitConditionSchema,
  budgetLimitRuleResponseSchema,
  budgetLimitRuleSchema,
} from "./ruleValidation";

export type BudgetLimitCondition = z.infer<typeof budgetLimitConditionSchema>;
export type BudgetLimitConditionField = BudgetLimitCondition["field"];
export type BudgetLimitConditionOperator = BudgetLimitCondition["operator"];
export type BudgetLimitConditionGroup = z.infer<typeof budgetLimitRuleSchema>["conditions"];
export type BudgetLimitUnit = z.infer<typeof budgetLimitRuleSchema>["unit"];
export type BudgetLimitRule = z.infer<typeof budgetLimitRuleResponseSchema>;
