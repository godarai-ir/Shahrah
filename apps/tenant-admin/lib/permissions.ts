export const PERMISSIONS = {
  BILLING_DOCUMENTS_READ: "billing:documents_read",
  BILLING_MANAGE: "billing:manage",
  BILLING_PROFILE_MANAGE: "billing:profile_manage",
  BILLING_READ: "billing:read",
  LOGS_READ_CONTENT: "logs:read_content",
  LOGS_READ_METADATA: "logs:read_metadata",
  PLAYGROUND_READ: "playground:read",
  PLAYGROUND_RUN: "playground:run",
  POLICY_LIST: "policy:list",
  POLICY_MANAGE: "policy:manage",
  PROMPT_LIST: "prompt:list",
  PROMPT_MANAGE: "prompt:manage",
  PROVIDER_ACCOUNT_CREATE: "provider_account:create",
  PROVIDER_ACCOUNT_DELETE: "provider_account:delete",
  PROVIDER_ACCOUNT_MANAGE: "provider_account:manage",
  PROVIDER_ACCOUNT_READ: "provider_account:read",
  PROVIDER_ACCOUNT_USE_INTEGRATIONS: "provider_account:use_integrations",
  ROLE_LIST: "role:list",
  ROLE_MANAGE: "role:manage",
  ROLE_READ: "role:read",
  SETTINGS_LIST: "settings:list",
  SETTINGS_MANAGE: "settings:manage",
  SUPPORT_TICKET_CREATE: "support_ticket:create",
  SUPPORT_TICKET_READ_ALL: "support_ticket:read_all",
  SUPPORT_TICKET_READ_OWN: "support_ticket:read_own",
  TEAM_CREATE: "team:create",
  TEAM_DELETE: "team:delete",
  TEAM_MANAGE: "team:manage",
  TEAM_READ: "team:read",
  USER_LIST: "user:list",
  USER_MANAGE: "user:manage",
  USER_READ: "user:read",
  VIRTUAL_ACCOUNT_CREATE: "virtual_account:create",
  VIRTUAL_ACCOUNT_MANAGE: "virtual_account:manage",
  VIRTUAL_ACCOUNT_READ: "virtual_account:read",
} as const;

export type Permission = (typeof PERMISSIONS)[keyof typeof PERMISSIONS];

const permissionSet = new Set<string>(Object.values(PERMISSIONS));

export function isPermission(value: string): value is Permission {
  return permissionSet.has(value.trim().toLowerCase());
}
