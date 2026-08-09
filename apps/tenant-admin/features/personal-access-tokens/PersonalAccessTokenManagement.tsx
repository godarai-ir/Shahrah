"use client";

import { FormEvent, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { KeyRound, Plus, RefreshCcw, Trash2 } from "lucide-react";

import { AppCard } from "@/components/app-ui/AppCard";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { ExpirationDurationField } from "@/components/app-ui/ExpirationDurationField";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  createPersonalAccessToken,
  listPersonalAccessTokens,
  revokeAllPersonalAccessTokens,
  revokePersonalAccessToken,
  type PersonalAccessToken as PAT,
} from "./api";

type Banner = { kind: "success" | "error"; message: string };
type DeleteIntent =
  | { type: "single"; token: PAT }
  | { type: "all" };

const PAT_COLUMN_CLASS = {
  name: "w-52",
  masked: "hidden w-32 @md:table-cell",
  createdAt: "hidden w-40 @lg:table-cell",
  lastUsedAt: "hidden w-40 @lg:table-cell",
  expiresAt: "hidden w-40 @lg:table-cell",
  status: "w-24 whitespace-nowrap",
  actions: "w-24 whitespace-nowrap text-end",
} as const;

export default function PersonalAccessTokenManagement(props: {
  initialTokens: PAT[];
  initialError: string | null;
}) {
  const t = useTranslations("personalAccessTokens");
  const locale = useLocale();
  const [tokens, setTokens] = useState<PAT[]>(props.initialTokens);
  const [query, setQuery] = useState("");
  const [name, setName] = useState("");
  const [expiresIn, setExpiresIn] = useState("30d");
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteIntent, setDeleteIntent] = useState<DeleteIntent | null>(null);
  const [banner, setBanner] = useState<Banner | null>(props.initialError ? { kind: "error", message: props.initialError } : null);
  const [createNotice, setCreateNotice] = useState<Banner | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<Banner | null>(null);

  async function load() {
    setIsLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", "25");
      params.set("offset", "0");
      if (query.trim()) params.set("query", query.trim());
      const data = await listPersonalAccessTokens(params);
      setTokens(data.data ?? []);
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.loadError")) });
    } finally {
      setIsLoading(false);
    }
  }

  async function createPAT(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!name.trim()) {
      setCreateNotice({ kind: "error", message: t("messages.validation") });
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
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.createSuccess") });
      await load();
    } catch (error) {
      setCreateNotice({ kind: "error", message: normalizeError(error, t("messages.createError")) });
    } finally {
      setIsCreating(false);
    }
  }

  async function revokeOne(id: string) {
    setIsDeleting(true);
    try {
      await revokePersonalAccessToken(id);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.revokeSuccess") });
      await load();
      return true;
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, t("messages.revokeError")) });
      return false;
    } finally {
      setIsDeleting(false);
    }
  }

  async function revokeAll() {
    setIsDeleting(true);
    try {
      await revokeAllPersonalAccessTokens();
      setTokens([]);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.revokeAllSuccess") });
      return true;
    } catch (error) {
      setDeleteNotice({ kind: "error", message: normalizeError(error, t("messages.revokeAllError")) });
      return false;
    } finally {
      setIsDeleting(false);
    }
  }

  function openDeleteDialog(intent: DeleteIntent) {
    setDeleteNotice(null);
    setDeleteIntent(intent);
    setDeleteDialogOpen(true);
  }

  async function confirmDelete() {
    if (!deleteIntent || isDeleting) return;
    const succeeded = deleteIntent.type === "single"
      ? await revokeOne(deleteIntent.token.id)
      : await revokeAll();
    if (!succeeded) return;
    setDeleteDialogOpen(false);
    setDeleteIntent(null);
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("title")}
        description={t("create.description")}
      />

      <FormSheet open={createOpen} onOpenChange={(open) => { setCreateOpen(open); if (!open) setCreateNotice(null); }} title={t("create.title")} description={t("create.description")} size="sm" pending={isCreating} closeLabel={t("oneTime.close")}>
        <OperationNotice notice={createNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />
        <form className="flex flex-col gap-4" onSubmit={createPAT}>
          <div className="grid gap-2">
            <label className="text-sm font-medium">{t("fields.name")}</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t("fields.namePlaceholder")} />
          </div>
          <ExpirationDurationField value={expiresIn} onChange={(e) => setExpiresIn(e.target.value)} placeholder={t("fields.expiresInPlaceholder")} />
          <Button type="submit" className="w-full" disabled={isCreating}>{isCreating ? t("actions.creating") : t("actions.create")}</Button>
        </form>
      </FormSheet>

      <OperationNotice notice={banner} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

      <AppCard>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <SearchField containerClassName="w-full sm:min-w-64 sm:flex-1" value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("searchPlaceholder")} />
            <div className="flex flex-wrap items-center gap-2 sm:shrink-0 sm:flex-nowrap">
              <Button onClick={() => setCreateOpen(true)}><Plus data-icon="inline-start" />{t("actions.create")}</Button>
              <Button variant="destructive" className="w-full sm:w-auto" onClick={() => openDeleteDialog({ type: "all" })}><Trash2 data-icon="inline-start" />{t("actions.revokeAll")}</Button>
              <Button variant="outline" size="icon" aria-label={t("actions.refresh")} title={t("actions.refresh")} onClick={() => void load()} disabled={isLoading}><RefreshCcw /></Button>
            </div>
          </div>

          {tokens.length === 0 ? (
            <EmptyState icon={KeyRound} title={t("table.empty")} description={t("create.accessHint")} />
          ) : (
            <div className="@container/pats overflow-hidden rounded-lg border border-border">
              <Table className="w-max min-w-full table-fixed text-xs lg:text-sm">
                <TableHeader>
                  <TableRow>
                    <TableHead className={PAT_COLUMN_CLASS.name}>{t("table.name")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.masked}>{t("table.masked")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.createdAt}>{t("table.createdAt")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.lastUsedAt}>{t("table.lastUsedAt")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.expiresAt}>{t("table.expiresAt")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.status}>{t("table.status")}</TableHead>
                    <TableHead className={PAT_COLUMN_CLASS.actions}>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className={PAT_COLUMN_CLASS.name}><span className="block truncate" title={item.key_alias}>{item.key_alias}</span></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.masked}><span className="block truncate" title={item.key_name ?? "-"}>{item.key_name ?? "-"}</span></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.createdAt}><DateTimeValue value={item.created_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.lastUsedAt}><DateTimeValue value={item.last_used_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.expiresAt}><DateTimeValue value={item.expires_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.status}><StatusBadge status={item.revoked_at ? "inactive" : "active"}>{item.revoked_at ? t("status.revoked") : t("status.active")}</StatusBadge></TableCell>
                      <TableCell className={PAT_COLUMN_CLASS.actions}><Button variant="outline" size="sm" onClick={() => openDeleteDialog({ type: "single", token: item })}>{t("actions.revoke")}</Button></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </AppCard>

      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={(open) => {
          setDeleteDialogOpen(open);
          if (!open) {
            setDeleteIntent(null);
            setDeleteNotice(null);
          }
        }}
        title={t("deleteDialog.title")}
        description={deleteIntent?.type === "single"
          ? t("deleteDialog.singleDescription", { name: deleteIntent.token.key_alias })
          : t("deleteDialog.allDescription")}
        cancelLabel={t("deleteDialog.cancel")}
        confirmLabel={t("deleteDialog.confirm")}
        pendingLabel={t("deleteDialog.deleting")}
        pending={isDeleting}
        closeOnConfirm={false}
        onConfirm={confirmDelete}
        feedback={<OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />}
      />

      <OneTimeSecretDialog
        open={secretDialogOpen}
        onOpenChange={setSecretDialogOpen}
        value={newSecret}
        title={t("oneTime.title")}
        description={t("oneTime.description")}
        warning={t("oneTime.copyHint")}
        closeLabel={t("oneTime.close")}
        copyLabels={{
          valueAriaLabel: t("oneTime.title"),
          copy: t("oneTime.copy"),
          copying: t("oneTime.copying"),
          copied: t("oneTime.copied"),
        }}
        onClear={() => setNewSecret(null)}
        onCopyError={() => showFeedbackToast({ kind: "error", title: t("alerts.error"), description: t("oneTime.copyFailed") })}
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
