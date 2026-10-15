"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { Activity, KeyRound, ShieldCheck } from "lucide-react";

import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { useCalendarLocale } from "@/hooks/use-locale-format";
import { collectEffectivePermissions } from "@/lib/access/effective-permissions";
import { resolveRoleDisplayName } from "@/lib/access/role-display";
import { translatePermissionGroup } from "@/lib/access/permission-groups";
import type { PermissionItem } from "@/lib/contracts/roles";
import { countActiveKeys, listKeysOwnedByMember, resolveLastKeyUse } from "@/lib/access/key-status";
import type { TenantKey } from "@/lib/api/access-keys";

import { memberDisplayName, resolveIdentity } from "@/lib/access/identity";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";

import type { RoleDefinition } from "./api";
import { isUserActive, listUserRoleNames } from "./member-filters";

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
  onSaveName: (member: TenantAdminUser, firstName: string, lastName: string) => Promise<void>;
  onChangeRole: (member: TenantAdminUser) => void;
  onToggleActive: (member: TenantAdminUser, activate: boolean) => void;
};

/** Edits a member's first and last name; remounted by its key whenever the saved names change. */
function MemberNameForm(props: {
  member: TenantAdminUser;
  canManageUsers: boolean;
  onSave: MemberDetailSheetProps["onSaveName"];
}) {
  const t = useTranslations("accessManagement");
  const [firstName, setFirstName] = useState(props.member.first_name ?? "");
  const [lastName, setLastName] = useState(props.member.last_name ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const changed =
    firstName.trim() !== (props.member.first_name ?? "") || lastName.trim() !== (props.member.last_name ?? "");

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      await props.onSave(props.member, firstName.trim(), lastName.trim());
    } catch {
      setError(t("messages.nameUpdateError"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="flex min-w-0 flex-col gap-3">
      <h3 className="text-sm font-semibold">{t("member.nameSection")}</h3>
      <div className="grid grid-cols-2 gap-3">
        <AppTextField
          label={t("form.firstName")}
          value={firstName}
          onChange={(event) => setFirstName(event.target.value)}
          disabled={!props.canManageUsers || saving}
          autoComplete="off"
          maxLength={100}
        />
        <AppTextField
          label={t("form.lastName")}
          value={lastName}
          onChange={(event) => setLastName(event.target.value)}
          disabled={!props.canManageUsers || saving}
          autoComplete="off"
          maxLength={100}
        />
      </div>
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
      {props.canManageUsers ? (
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="self-start"
          disabled={!changed || saving}
          onClick={() => void handleSave()}
        >
          {saving ? t("button.savingName") : t("button.saveName")}
        </Button>
      ) : null}
    </section>
  );
}

export function MemberDetailSheet(props: MemberDetailSheetProps) {
  const t = useTranslations("accessManagement");
  const calendarLocale = useCalendarLocale();
  const member = props.member;

  const identity = member ? resolveIdentity(member) : "";
  const displayName = member ? memberDisplayName(member) : "";
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

          {member ? (
            <MemberNameForm
              key={`${member.id}:${member.first_name ?? ""}:${member.last_name ?? ""}`}
              member={member}
              canManageUsers={props.canManageUsers}
              onSave={props.onSaveName}
            />
          ) : null}

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
