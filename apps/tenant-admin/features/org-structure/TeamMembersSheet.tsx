"use client";

import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CheckboxOptionList } from "@/components/app-ui/CheckboxOptionList";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { FilterCombobox } from "@/components/app-ui/FilterCombobox";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { hideMobileIdentity } from "@/lib/auth/mobile-auth-ui";
import type { AccessPrincipalCandidate, TeamMember } from "@/lib/api/teams";

export const PERMISSIONS = ["/key/info", "/key/list", "/key/generate", "/key/update", "/key/delete", "/key/regenerate", "/key/block", "/key/unblock", "/key/health"] as const;
export const DEFAULT_PERMISSIONS = ["/key/info", "/key/health"];

type Banner = { kind: "success" | "error" | "warning"; message: string } | null;
type MemberCandidateOption = { userID: string; label: string };

export function TeamMembersSheet({
  open,
  onOpenChange,
  members,
  candidateOptions,
  memberLabelByUserID,
  canManage,
  isSaving,
  notice,
  onAddMember,
  onUpdateMember,
  onRemoveMember,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  members: TeamMember[];
  candidateOptions: MemberCandidateOption[];
  memberLabelByUserID: Map<string, string>;
  canManage: boolean;
  isSaving: boolean;
  notice: Banner;
  onAddMember: (userID: string, role: string, permissions: string[]) => Promise<void>;
  onUpdateMember: (member: TeamMember, role: string, permissions: string[]) => Promise<void>;
  onRemoveMember: (userID: string) => Promise<void>;
}) {
  const t = useTranslations("teams");
  const tShared = useTranslations("orgStructure");
  const [userInput, setUserInput] = useState("");
  const [role, setRole] = useState("team_member");
  const [permissions, setPermissions] = useState<string[]>(DEFAULT_PERMISSIONS);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const resolvedUserID = resolveMemberUserID(userInput, candidateOptions);
    if (!resolvedUserID) return;
    await onAddMember(resolvedUserID, role, role === "team_member" ? permissions : [...PERMISSIONS]);
    setUserInput("");
    setRole("team_member");
    setPermissions(DEFAULT_PERMISSIONS);
  }

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      title={t("members.title")}
      description={t("members.description")}
      pending={isSaving}
      closeLabel={tShared("actions.cancel")}
      footer={<FormSheetActions secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => onOpenChange(false)}>{tShared("actions.cancel")}</Button>} primary={<Button type="submit" form="org-structure-add-team-member-form" disabled={!canManage || isSaving}>{t("members.add")}</Button>} />}
    >
      <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.warning"), info: t("alerts.warning") }} />

      <form id="org-structure-add-team-member-form" onSubmit={submit} className="flex flex-col gap-3 border-b border-border pb-5">
        <FieldGroup className="gap-3">
          <Field>
            <FieldLabel htmlFor="org-structure-member-user">{t("members.user")}</FieldLabel>
            <FilterCombobox
              id="org-structure-member-user"
              value={userInput}
              onValueChange={setUserInput}
              options={candidateOptions.map((user) => ({ value: user.userID, label: user.label }))}
              placeholder={t("members.userPlaceholder")}
              emptyText={t("members.userEmpty")}
              className="h-10"
            />
            <FieldDescription>{t("members.userHelp")}</FieldDescription>
          </Field>
          <div className="flex flex-col gap-2">
            <Button
              type="button"
              variant={role === "team_member" ? "default" : "outline"}
              className="justify-start"
              onClick={() => setRole((current) => (current === "team_member" ? "team_admin" : "team_member"))}
            >
              {role === "team_member" ? t("members.teamMember") : t("members.teamAdmin")}
            </Button>
            <p className="text-xs text-muted-foreground">
              {role === "team_member" ? t("members.userHelp") : t("members.teamAdminAccessDescription")}
            </p>
          </div>
        </FieldGroup>
        {role === "team_member" ? (
          <details className="rounded-md border border-border">
            <summary className="cursor-pointer px-3 py-2 text-sm font-medium">{t("members.permissions")}</summary>
            <div className="px-3 pb-3">
              <div className="flex justify-end pb-2">
                <Button type="button" variant="ghost" size="sm" onClick={() => setPermissions(DEFAULT_PERMISSIONS)}>{t("members.resetPermissions")}</Button>
              </div>
              <CheckboxOptionList
                label={t("members.permissions")}
                description={t("members.permissionsHelp")}
                value={permissions}
                onValueChange={setPermissions}
                emptyText={t("members.permissions")}
                options={PERMISSIONS.map((perm) => ({ value: perm, label: t(`members.permissionOptions.${permissionTranslationKey(perm)}.label`), description: t(`members.permissionOptions.${permissionTranslationKey(perm)}.description`) }))}
              />
            </div>
          </details>
        ) : null}
      </form>

      <div className="mt-4 flex flex-col gap-2">
        {members.map((member) => (
          <div key={member.user_id} className="rounded-md border border-border p-3">
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-sm" dir="auto">{getMemberDisplayLabel(member, memberLabelByUserID, t("members.unknownUser"))}</span>
              <Badge variant={member.role === "team_admin" ? "default" : "secondary"}>{member.role}</Badge>
            </div>
            <div className="mt-2 flex flex-wrap gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => void onUpdateMember(member, member.role === "team_admin" ? "team_member" : "team_admin", member.team_member_permissions ?? DEFAULT_PERMISSIONS)}
                disabled={!canManage || isSaving}
              >
                {member.role === "team_admin" ? t("members.makeTeamMember") : t("members.makeTeamAdmin")}
              </Button>
              <Button size="sm" variant="outline" className="text-destructive" onClick={() => void onRemoveMember(member.user_id)} disabled={!canManage || isSaving}>
                {t("members.remove")}
              </Button>
            </div>
          </div>
        ))}
      </div>
    </FormSheet>
  );
}

function resolveMemberUserID(input: string, options: MemberCandidateOption[]) {
  const normalizedInput = input.trim();
  if (!normalizedInput) return "";
  const byID = options.find((option) => option.userID === normalizedInput);
  if (byID) return byID.userID;
  const byLabel = options.find((option) => option.label === normalizedInput);
  if (byLabel) return byLabel.userID;
  if (looksLikeUserID(normalizedInput)) return normalizedInput;
  return "";
}

function looksLikeUserID(value: string) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value.trim());
}

function getMemberDisplayLabel(member: TeamMember, memberLabelByUserID: Map<string, string>, unknownUserLabel: string) {
  const userLabel = member.user_email?.trim() || hideMobileIdentity(member.user_display_name) || memberLabelByUserID.get(member.user_id);
  if (userLabel) return userLabel;
  if (looksLikeUserID(member.user_id)) return unknownUserLabel;
  return member.user_id;
}

function permissionTranslationKey(permission: (typeof PERMISSIONS)[number]) {
  return permission.slice("/key/".length);
}

export function formatMemberCandidateLabel(candidate: AccessPrincipalCandidate) {
  const secondary = hideMobileIdentity(candidate.secondary);
  const display = hideMobileIdentity(candidate.display_name);
  if (secondary && !looksLikeUserID(secondary)) return secondary;
  if (!looksLikeUserID(display)) return display;
  if (secondary) return secondary;
  return display || candidate.principal_id;
}
