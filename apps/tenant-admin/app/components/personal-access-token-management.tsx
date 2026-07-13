"use client";

import { FormEvent, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { KeyRound, Plus, RefreshCcw, Trash2 } from "lucide-react";

import { AppCard, ConfirmDialog, DateTimeValue, EmptyState, ExpirationDurationField, FormSheet, HelpLink, OneTimeSecretDialog, OperationNotice, PageHeader, SearchField, StatusBadge } from "@/components/app-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type PAT = {
  id: string;
  key_alias: string;
  key_name?: string;
  key_type?: string;
  created_at?: string;
  expires_at?: string;
  last_used_at?: string;
  revoked_at?: string;
};

type PATResponse = {
  id: string;
  key: string;
  key_alias: string;
  key_name?: string;
  created_at?: string;
  expires_at?: string;
  last_used_at?: string;
};

type ListResponse = {
  data?: PAT[];
  total?: number;
};

type Banner = { kind: "success" | "error"; message: string };
type DeleteIntent =
  | { type: "single"; token: PAT }
  | { type: "all" };


export default function PersonalAccessTokenManagement(props: {
  initialTokens: PAT[];
  initialError: string | null;
  docsHref: string;
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

  async function load() {
    setIsLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", "25");
      params.set("offset", "0");
      if (query.trim()) params.set("query", query.trim());
      const data = await requestJSON<ListResponse>(`/api/personal-access-tokens?${params.toString()}`);
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
      setBanner({ kind: "error", message: t("messages.validation") });
      return;
    }
    setIsCreating(true);
    try {
      const created = await requestJSON<PATResponse>("/api/personal-access-tokens", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), expires_in: expiresIn.trim() || undefined }),
      });
      setNewSecret(created.key);
      setCreateOpen(false);
      setSecretDialogOpen(true);
      setName("");
      setExpiresIn("30d");
      setBanner({ kind: "success", message: t("messages.createSuccess") });
      await load();
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.createError")) });
    } finally {
      setIsCreating(false);
    }
  }

  async function revokeOne(id: string) {
    setIsDeleting(true);
    try {
      await requestJSON(`/api/personal-access-tokens/${encodeURIComponent(id)}`, { method: "DELETE" });
      setBanner({ kind: "success", message: t("messages.revokeSuccess") });
      await load();
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeError")) });
      return false;
    } finally {
      setIsDeleting(false);
    }
  }

  async function revokeAll() {
    setIsDeleting(true);
    try {
      await requestJSON("/api/personal-access-tokens/revoke-all", { method: "DELETE" });
      setBanner({ kind: "success", message: t("messages.revokeAllSuccess") });
      await load();
      return true;
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeAllError")) });
      return false;
    } finally {
      setIsDeleting(false);
    }
  }

  function openDeleteDialog(intent: DeleteIntent) {
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
        actions={
          <>
            <Button onClick={() => setCreateOpen(true)}><Plus data-icon="inline-start" />{t("actions.create")}</Button>
            <HelpLink href={props.docsHref} />
          </>
        }
      />

      <FormSheet open={createOpen} onOpenChange={setCreateOpen} title={t("create.title")} description={t("create.description")} size="sm" pending={isCreating}>
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
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <SearchField containerClassName="w-full sm:w-72" value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("searchPlaceholder")} />
            <div className="flex items-center gap-2">
              <Button variant="outline" size="icon" aria-label={t("actions.refresh")} title={t("actions.refresh")} onClick={() => void load()} disabled={isLoading}><RefreshCcw /></Button>
              <Button variant="destructive" onClick={() => openDeleteDialog({ type: "all" })}><Trash2 data-icon="inline-start" />{t("actions.revokeAll")}</Button>
            </div>
          </div>

          {tokens.length === 0 ? (
            <EmptyState icon={KeyRound} title={t("table.empty")} description={t("create.accessHint")} />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("table.name")}</TableHead>
                    <TableHead>{t("table.masked")}</TableHead>
                    <TableHead>{t("table.createdAt")}</TableHead>
                    <TableHead>{t("table.lastUsedAt")}</TableHead>
                    <TableHead>{t("table.expiresAt")}</TableHead>
                    <TableHead>{t("table.status")}</TableHead>
                    <TableHead>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{item.key_alias}</TableCell>
                      <TableCell>{item.key_name ?? "-"}</TableCell>
                      <TableCell><DateTimeValue value={item.created_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell><DateTimeValue value={item.last_used_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell><DateTimeValue value={item.expires_at} locale={locale} format={{ dateStyle: "short", timeStyle: "medium", timeZone: "Asia/Tehran" }} /></TableCell>
                      <TableCell><StatusBadge status={item.revoked_at ? "inactive" : "active"}>{item.revoked_at ? t("status.revoked") : t("status.active")}</StatusBadge></TableCell>
                      <TableCell><Button variant="outline" size="sm" onClick={() => openDeleteDialog({ type: "single", token: item })}>{t("actions.revoke")}</Button></TableCell>
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
          if (!open) setDeleteIntent(null);
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
        onCopyError={() => setBanner({ kind: "error", message: t("oneTime.copyFailed") })}
      />
    </div>
  );
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try { return JSON.parse(text) as unknown; } catch { return { message: text }; }
}

function extractMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
