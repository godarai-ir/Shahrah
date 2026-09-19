"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { Activity, KeyRound, ShieldCheck } from "lucide-react";

import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { useCalendarLocale } from "@/hooks/use-locale-format";
import { collectEffectivePermissions } from "@/features/roles/effective-permissions";
import { resolveRoleDisplayName } from "@/features/roles/role-display";
import { translatePermissionGroup } from "@/features/roles/permissions";
import type { PermissionItem } from "@/features/roles/contracts";
import { countActiveKeys, listKeysOwnedByMember, resolveLastKeyUse } from "@/features/access-keys/key-status";
import type { TenantKey } from "@/features/access-keys/api";

import type { RoleDefinition, TenantAdminUser } from "./api";
import { guessDisplayName, isUserActive, listUserRoleNames, resolveIdentity } from "./member-filters";

export type MemberDetailSheetProps = {
  member: TenantAdminUser | null;
  roles: RoleDefinition[];
  permissionCatalog: PermissionItem[];
  keys: TenantKey[];
  /** False while the tenant-wide key list is still loading, or when it was refused. */
  hasKeyData: boolean;
  canManageUsers: boolean;
  isSelf: boolean;
  onOpenChange: (open: boolean) => void;
  onChangeRole: (member: TenantAdminUser) => void;
  onToggleActive: (member: TenantAdminUser, activate: boolean) => void;
};

export function MemberDetailSheet(props: MemberDetailSheetProps) {
  const t = useTranslations("accessManagement");
  const calendarLocale = useCalendarLocale();
  const member = props.member;

  const identity = member ? resolveIdentity(member) : "";
  const displayName = guessDisplayName(identity) || identity;
  const roleNames = useMemo(() => (member ? listUserRoleNames(member) : []), [member]);
  const effective = useMemo(
    () => collectEffectivePermissions(roleNames, props.roles, props.permissionCatalog),
    [roleNames, props.roles, props.permissionCatalog],
  );
  const ownedKeys = useMemo(
    () => (member ? listKeysOwnedByMember(props.keys, member.id) : []),
    [member, props.keys],
  );

  const active = member ? isUserActive(member) : false;
  const lastKeyUse = resolveLastKeyUse(ownedKeys);
  // Permissions come from the role, so the "edit role" link points at the first one.
  const primaryRole = roleNames[0] ?? "";
  const primaryRoleID = props.roles.find((role) => role.name.trim().toLowerCase() === primaryRole)?.id ?? "";

  return (
    <Sheet open={Boolean(member)} onOpenChange={props.onOpenChange}>
      <SheetContent
        closeLabel={t("actions.close")}
        className="flex w-full min-w-0 flex-col overflow-hidden p-0 sm:max-w-[440px]"
      >
        <SheetHeader className="shrink-0 gap-2 border-b border-border px-5 pb-4 pe-12 pt-5 text-start">
          <SheetTitle className="min-w-0 truncate" title={displayName}>{displayName || "-"}</SheetTitle>
          <SheetDescription dir="ltr" className="truncate text-start font-mono text-xs" title={identity}>
            {identity}
          </SheetDescription>
          <div className="flex flex-wrap items-center gap-2">
            <StatusBadge status={active ? "active" : "warning"}>
              {active ? t("status.active") : t("status.deactivated")}
            </StatusBadge>
            {roleNames.length === 0 ? (
              <Badge variant="outline">{t("member.noAccess")}</Badge>
            ) : (
              roleNames.map((role) => (
                <Badge key={role} variant="secondary">{resolveRoleDisplayName(role, props.roles, t)}</Badge>
              ))
            )}
          </div>
        </SheetHeader>

        <div className="flex min-w-0 flex-1 flex-col gap-5 overflow-y-auto px-5 py-5">
          <div className="grid grid-cols-2 gap-3">
            <MetricCard
              density="compact"
              icon={KeyRound}
              label={t("member.activeKeys")}
              value={props.hasKeyData ? countActiveKeys(ownedKeys) : "—"}
              description={props.hasKeyData ? undefined : t("member.keysUnavailable")}
            />
            <MetricCard
              density="compact"
              icon={Activity}
              label={t("member.lastActivity")}
              value={(
                <DateTimeValue
                  value={member?.last_login_at}
                  locale={calendarLocale}
                  fallback={t("table.neverActive")}
                />
              )}
            />
          </div>

          <section className="flex min-w-0 flex-col gap-2">
            <div className="flex items-center justify-between gap-2">
              <h3 className="flex items-center gap-2 text-sm font-semibold">
                <ShieldCheck className="size-4 text-muted-foreground" aria-hidden="true" />
                {t("member.effectivePerms")}
              </h3>
              <span className="text-xs tabular-nums text-muted-foreground">
                {t("perms.grantCount", { count: effective.total })}
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {primaryRole
                ? t("member.permsSource", { role: resolveRoleDisplayName(primaryRole, props.roles, t) })
                : t("member.noAccess")}
            </p>
            {effective.groups.length === 0 ? (
              <p className="rounded-lg border border-dashed border-border px-3 py-4 text-center text-xs text-muted-foreground">
                {t("member.noAccess")}
              </p>
            ) : (
              <DefinitionList className="gap-y-2">
                {effective.groups.map((group) => (
                  <DefinitionListItem
                    key={group.group}
                    surface
                    wrap
                    label={translatePermissionGroup(group.group, t)}
                    value={(
                      <span className="flex flex-wrap gap-1">
                        {group.grants.map((grant) => (
                          <code key={grant.key} dir="ltr" className="rounded bg-muted px-1.5 py-0.5 text-xs">
                            {grant.key}
                          </code>
                        ))}
                      </span>
                    )}
                  />
                ))}
              </DefinitionList>
            )}
            {primaryRoleID && props.canManageUsers ? (
              <Button asChild type="button" variant="outline" size="sm" className="self-start">
                <Link href={`/roles/${encodeURIComponent(primaryRoleID)}/edit`}>{t("member.editRole")}</Link>
              </Button>
            ) : null}
          </section>

          {props.hasKeyData && ownedKeys.length > 0 ? (
            <section className="flex min-w-0 flex-col gap-2">
              <h3 className="text-sm font-semibold">{t("member.recentActivity")}</h3>
              <DefinitionList className="gap-y-2">
                <DefinitionListItem
                  label={t("member.lastKeyUse")}
                  value={(
                    <DateTimeValue
                      value={lastKeyUse ?? undefined}
                      locale={calendarLocale}
                      fallback={t("table.neverActive")}
                    />
                  )}
                />
              </DefinitionList>
            </section>
          ) : null}
        </div>

        {props.canManageUsers && member && !props.isSelf ? (
          <SheetFooter className="shrink-0 flex-row gap-2 border-t border-border px-5 py-4">
            <Button type="button" variant="outline" size="sm" onClick={() => props.onChangeRole(member)}>
              {t("actions.changeRole")}
            </Button>
            <Button
              type="button"
              variant={active ? "destructive" : "outline"}
              size="sm"
              onClick={() => props.onToggleActive(member, !active)}
            >
              {active ? t("actions.deactivate") : t("actions.activate")}
            </Button>
          </SheetFooter>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
