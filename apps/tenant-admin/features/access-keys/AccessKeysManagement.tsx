"use client";

import { FormEvent, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";
import { KeyRound, Plus } from "lucide-react";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { DataTableToolbar } from "@/components/app-ui/DataTableToolbar";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { ExpirationDurationField } from "@/components/app-ui/ExpirationDurationField";
import { FilterField } from "@/components/app-ui/FilterField";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
  createPersonalAccessToken,
  listPersonalAccessTokens,
  revokePersonalAccessToken,
} from "@/lib/api/personal-access-tokens";
import { guessDisplayName, resolveIdentity } from "@/lib/access/identity";
import type { TenantAdminUser } from "@/lib/contracts/tenant-users";
import type { VirtualUser } from "@/lib/contracts/virtual-users";

import { listTenantKeys, type TenantKey } from "@/lib/api/access-keys";
import { isExpired, resolveKeyState, resolveKeyType, type KeyType } from "@/lib/access/key-status";

type Banner = { kind: "success" | "error"; message: string };
type TypeFilter = "all" | KeyType;

export type AccessKeysManagementProps = {
  initialKeys: TenantKey[];
  initialError: string | null;
  members: TenantAdminUser[];
  /** Names the owner column for service keys. */
  virtualUsers: VirtualUser[];
  currentUserID: string;
  /** True when the caller may read every key; false when this is only their own. */
  isTenantWide: boolean;
};

export default function AccessKeysManagement(props: AccessKeysManagementProps) {
  const t = useTranslations("accessManagement");
  const tPAT = useTranslations("personalAccessTokens");
  const locale = useLocale();

  const [keys, setKeys] = useState<TenantKey[]>(props.initialKeys);
  const [query, setQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");
  const [name, setName] = useState("");
  const [expiresIn, setExpiresIn] = useState("30d");
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<TenantKey | null>(null);
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError ? { kind: "error", message: props.initialError } : null,
  );
  const [createNotice, setCreateNotice] = useState<Banner | null>(null);
  const [revokeNotice, setRevokeNotice] = useState<Banner | null>(null);

  const noticeLabels = {
    success: t("alerts.success"),
    error: t("alerts.error"),
    warning: t("alerts.error"),
    info: t("alerts.success"),
  };

  const ownerNames = useMemo(() => {
    const names = new Map<string, string>();
    for (const member of props.members) {
      const identity = resolveIdentity(member);
      names.set(member.id, guessDisplayName(identity) || identity);
    }
    return names;
  }, [props.members]);

  const serviceNames = useMemo(() => {
    const names = new Map<string, string>();
    for (const virtualUser of props.virtualUsers) {
      names.set(virtualUser.id, virtualUser.display_name?.trim() || virtualUser.alias);
    }
    return names;
  }, [props.virtualUsers]);

  function resolveOwner(key: TenantKey): string {
    if (resolveKeyType(key) === "vat") {
      // A service key belongs to a virtual user, not a person.
      const serviceID = key.virtual_user_id?.trim();
      return (serviceID && serviceNames.get(serviceID)) || t("keys.service");
    }
    const owner = key.created_by_user_id?.trim();
    return (owner && ownerNames.get(owner)) || t("keys.unknownOwner");
  }

  const visibleKeys = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase();
    return keys.filter((key) => {
      if (typeFilter !== "all" && resolveKeyType(key) !== typeFilter) return false;
      if (!needle) return true;
      return [key.key_alias, resolveOwner(key)].some((value) =>
        value.toLocaleLowerCase().includes(needle),
      );
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [keys, query, typeFilter, ownerNames, serviceNames]);

  const columns: ColumnDef<TenantKey>[] = [
    {
      id: "name",
      header: t("keys.name"),
      meta: { label: t("keys.name"), headerClassName: "w-[26%]" },
      cell: ({ row }) => (
        <span className="block truncate font-medium" title={row.original.key_alias}>
          {row.original.key_alias || "-"}
        </span>
      ),
    },
    {
      id: "type",
      header: t("keys.type"),
      meta: { label: t("keys.type") },
      cell: ({ row }) => {
        const type = resolveKeyType(row.original);
        return (
          <Badge variant="outline" dir="ltr" className="font-mono text-xs uppercase">
            {type === "vat" ? t("keys.typeVat") : t("keys.typePat")}
          </Badge>
        );
      },
    },
    ...(props.isTenantWide
      ? [{
          id: "owner",
          header: t("keys.owner"),
          meta: { label: t("keys.owner"), headerClassName: "w-[22%]" },
          cell: ({ row }) => {
            const owner = resolveOwner(row.original);
            return <span className="block truncate" title={owner}>{owner}</span>;
          },
        } satisfies ColumnDef<TenantKey>]
      : []),
    {
      id: "lastUsedAt",
      header: t("keys.lastUsedAt"),
      meta: { optional: true, label: t("keys.lastUsedAt") },
      cell: ({ row }) => (
        <DateTimeValue
          value={row.original.last_used_at}
          locale={locale}
          fallback="-"
          className="text-muted-foreground"
        />
      ),
    },
    {
      id: "expiresAt",
      header: t("keys.expiresAt"),
      meta: { label: t("keys.expiresAt") },
      cell: ({ row }) => {
        if (!row.original.expires_at?.trim()) {
          return <span className="text-muted-foreground">{t("keys.noExpiry")}</span>;
        }
        const expired = isExpired(row.original.expires_at);
        return (
          <DateTimeValue
            value={row.original.expires_at}
            locale={locale}
            className={expired ? "text-destructive" : undefined}
          />
        );
      },
    },
    {
      id: "status",
      header: t("keys.status"),
      meta: { label: t("keys.status") },
      cell: ({ row }) => {
        const state = resolveKeyState(row.original);
        const tone = state === "active" ? "active" : state === "expired" ? "error" : "inactive";
        return <StatusBadge status={tone}>{t(`keys.state.${state}`)}</StatusBadge>;
      },
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{tPAT("actions.revoke")}</span>,
      meta: { headerClassName: "w-24", cellClassName: "w-24 text-end" },
      cell: ({ row }) => {
        const key = row.original;
        const isOwn =
          resolveKeyType(key) === "pat" && key.created_by_user_id?.trim() === props.currentUserID;
        const alreadyGone = resolveKeyState(key) === "revoked";
        if (alreadyGone) return null;
        if (isOwn) {
          return (
            <Button variant="outline" size="sm" onClick={() => openRevoke(key)}>
              {tPAT("actions.revoke")}
            </Button>
          );
        }
        // The control plane scopes PAT revocation to the key's own holder, so an
        // admin cannot revoke someone else's key from here.
        return (
          <TooltipProvider delayDuration={150}>
            <Tooltip>
              <TooltipTrigger asChild>
                <span tabIndex={0} className="inline-block">
                  <Button variant="outline" size="sm" disabled>
                    {tPAT("actions.revoke")}
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">{t("keys.revokeOthersBlocked")}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        );
      },
    },
  ];

  async function load() {
    setIsLoading(true);
    try {
      const data = props.isTenantWide
        ? await listTenantKeys()
        : await listPersonalAccessTokens(new URLSearchParams({ limit: "200", offset: "0" }));
      setKeys(data.data ?? []);
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, tPAT("messages.loadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function createKey(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!name.trim()) {
      setCreateNotice({ kind: "error", message: tPAT("messages.validation") });
      return;
    }
    setCreateNotice(null);
    setIsCreating(true);
    try {
      const created = await createPersonalAccessToken({
        name: name.trim(),
        expiresIn: expiresIn.trim() || undefined,
      });
      setNewSecret(created.key);
      setCreateOpen(false);
      setSecretDialogOpen(true);
      setName("");
      setExpiresIn("30d");
      showFeedbackToast({
        kind: "success",
        title: tPAT("alerts.success"),
        description: tPAT("messages.createSuccess"),
      });
      await load();
    } catch (error) {
      setCreateNotice({ kind: "error", message: normalizeError(error, tPAT("messages.createError")) });
    } finally {
      setIsCreating(false);
    }
  }

  function openRevoke(key: TenantKey) {
    setRevokeNotice(null);
    setRevokeTarget(key);
  }

  async function confirmRevoke() {
    if (!revokeTarget || isDeleting) return;
    setIsDeleting(true);
    try {
      await revokePersonalAccessToken(revokeTarget.id);
      showFeedbackToast({
        kind: "success",
        title: tPAT("alerts.success"),
        description: tPAT("messages.revokeSuccess"),
      });
      setRevokeTarget(null);
      await load();
    } catch (error) {
      setRevokeNotice({ kind: "error", message: normalizeError(error, tPAT("messages.revokeError")) });
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <div className="flex min-w-0 flex-col gap-4">
      <OperationNotice notice={banner} labels={noticeLabels} />

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
        <Button type="button" className="w-full sm:w-auto" onClick={() => setCreateOpen(true)}>
          <Plus data-icon="inline-start" />
          {tPAT("actions.create")}
        </Button>
      </div>

      <DataTable
        tableId="access-keys"
        columns={columns}
        data={visibleKeys}
        state={isLoading && keys.length === 0 ? "loading" : "ready"}
        skeletonRows={4}
        emptyLabel={t("keys.empty")}
        noResultsSlot={
          <EmptyState
            icon={KeyRound}
            variant="table"
            title={t("keys.empty")}
            description={t("keys.emptyDescription")}
          />
        }
        getRowId={(key) => key.id}
        toolbar={
          <DataTableToolbar
            search={
              <SearchField
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={t("keys.searchPlaceholder")}
              />
            }
            filters={
              <FilterField label={t("keys.filterType")} hideLabel>
                <AppSelectField
                  value={typeFilter}
                  onValueChange={(value) => setTypeFilter(value as TypeFilter)}
                  ariaLabel={t("keys.filterType")}
                  triggerClassName="h-8 w-36 text-xs"
                  options={[
                    { value: "all", label: t("keys.allTypes") },
                    { value: "pat", label: t("keys.typePat") },
                    { value: "vat", label: t("keys.typeVat") },
                  ]}
                />
              </FilterField>
            }
          />
        }
      />

      <FormSheet
        open={createOpen}
        onOpenChange={(open) => {
          setCreateOpen(open);
          if (!open) setCreateNotice(null);
        }}
        title={tPAT("create.title")}
        description={tPAT("create.description")}
        size="sm"
        pending={isCreating}
        closeLabel={tPAT("oneTime.close")}
        footer={
          <FormSheetActions
            secondary={
              <Button
                type="button"
                variant="outline"
                disabled={isCreating}
                onClick={() => {
                  setCreateOpen(false);
                  setCreateNotice(null);
                }}
              >
                {tPAT("actions.cancel")}
              </Button>
            }
            primary={
              <Button type="submit" form="create-access-key-form" disabled={isCreating}>
                {isCreating ? tPAT("actions.creating") : tPAT("actions.create")}
              </Button>
            }
          />
        }
      >
        <OperationNotice notice={createNotice} labels={noticeLabels} />
        <form id="create-access-key-form" className="flex flex-col gap-4" onSubmit={createKey}>
          <div className="grid gap-2">
            <label className="text-sm font-medium" htmlFor="access-key-name">{tPAT("fields.name")}</label>
            <Input
              id="access-key-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={tPAT("fields.namePlaceholder")}
            />
          </div>
          <ExpirationDurationField
            value={expiresIn}
            onChange={(event) => setExpiresIn(event.target.value)}
            placeholder={tPAT("fields.expiresInPlaceholder")}
          />
        </form>
      </FormSheet>

      <ConfirmDialog
        open={Boolean(revokeTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setRevokeTarget(null);
            setRevokeNotice(null);
          }
        }}
        title={tPAT("deleteDialog.title")}
        description={tPAT("deleteDialog.singleDescription", { name: revokeTarget?.key_alias ?? "" })}
        cancelLabel={tPAT("deleteDialog.cancel")}
        confirmLabel={tPAT("deleteDialog.confirm")}
        pendingLabel={tPAT("deleteDialog.deleting")}
        pending={isDeleting}
        closeOnConfirm={false}
        onConfirm={confirmRevoke}
        feedback={<OperationNotice notice={revokeNotice} labels={noticeLabels} />}
      />

      <OneTimeSecretDialog
        open={secretDialogOpen}
        onOpenChange={setSecretDialogOpen}
        value={newSecret}
        title={tPAT("oneTime.title")}
        description={tPAT("oneTime.description")}
        warning={tPAT("oneTime.copyHint")}
        closeLabel={tPAT("oneTime.close")}
        copyLabels={{
          valueAriaLabel: tPAT("oneTime.title"),
          copy: tPAT("oneTime.copy"),
          copying: tPAT("oneTime.copying"),
          copied: tPAT("oneTime.copied"),
        }}
        onClear={() => setNewSecret(null)}
        onCopyError={() =>
          showFeedbackToast({
            kind: "error",
            title: tPAT("alerts.error"),
            description: tPAT("oneTime.copyFailed"),
          })
        }
      />
    </div>
  );
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) {
    const message = cleanErrorMessage(error.message);
    return message?.startsWith("Request failed (") ? fallback : message ?? fallback;
  }
  return fallback;
}

function cleanErrorMessage(message: unknown): string | null {
  if (typeof message !== "string") return null;
  const value = message.trim();
  if (!value || /<(?:!doctype|html|head|body|script|div)\b/i.test(value)) return null;
  return value.length > 300 ? `${value.slice(0, 300)}...` : value;
}
