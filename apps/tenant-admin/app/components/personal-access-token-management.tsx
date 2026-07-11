"use client";

import { FormEvent, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Check, Copy, KeyRound, Plus, RefreshCcw, Trash2 } from "lucide-react";

import { AppCard, EmptyState, SectionHeader, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
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

const stableDateTimeFormatter = new Intl.DateTimeFormat("fa-IR", {
  dateStyle: "short",
  timeStyle: "medium",
  timeZone: "Asia/Tehran",
});

export default function PersonalAccessTokenManagement(props: {
  initialTokens: PAT[];
  initialTotal: number;
  initialError: string | null;
}) {
  const t = useTranslations("personalAccessTokens");
  const [tokens, setTokens] = useState<PAT[]>(props.initialTokens);
  const [total, setTotal] = useState<number>(props.initialTotal);
  const [query, setQuery] = useState("");
  const [name, setName] = useState("");
  const [expiresIn, setExpiresIn] = useState("30d");
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isCopyingSecret, setIsCopyingSecret] = useState(false);
  const [secretCopied, setSecretCopied] = useState(false);
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteIntent, setDeleteIntent] = useState<DeleteIntent | null>(null);
  const [banner, setBanner] = useState<Banner | null>(props.initialError ? { kind: "error", message: props.initialError } : null);
  const createHint = useMemo(() => t("create.accessHint"), [t]);

  async function load() {
    setIsLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("limit", "25");
      params.set("offset", "0");
      if (query.trim()) params.set("query", query.trim());
      const data = await requestJSON<ListResponse>(`/api/personal-access-tokens?${params.toString()}`);
      setTokens(data.data ?? []);
      setTotal(data.total ?? 0);
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
      setSecretCopied(false);
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
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeError")) });
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
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeAllError")) });
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
    if (deleteIntent.type === "single") {
      await revokeOne(deleteIntent.token.id);
    } else {
      await revokeAll();
    }
    setDeleteDialogOpen(false);
    setDeleteIntent(null);
  }

  async function copySecret() {
    if (!newSecret || isCopyingSecret) return;
    setIsCopyingSecret(true);
    try {
      if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(newSecret);
      } else if (typeof document !== "undefined") {
        const textarea = document.createElement("textarea");
        textarea.value = newSecret;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
      }
      setSecretCopied(true);
    } catch {
      setBanner({ kind: "error", message: t("oneTime.copyFailed") });
    } finally {
      setIsCopyingSecret(false);
    }
  }

  return (
    <div className="flex flex-col gap-5">
      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <AppCard
        title={t("title")}
        description={t("create.description")}
        actions={
          <Sheet>
            <SheetTrigger asChild>
              <Button><Plus data-icon="inline-start" />{t("actions.create")}</Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[94vw] max-w-md">
              <SheetHeader>
                <SheetTitle>{t("create.title")}</SheetTitle>
                <SheetDescription>{t("create.description")}</SheetDescription>
              </SheetHeader>
              <form className="mt-6 flex flex-col gap-4" onSubmit={createPAT}>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">{t("fields.name")}</label>
                  <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t("fields.namePlaceholder")} />
                </div>
                <p className="text-xs text-muted-foreground">{createHint}</p>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">{t("fields.expiresIn")}</label>
                  <Input value={expiresIn} onChange={(e) => setExpiresIn(e.target.value)} placeholder={t("fields.expiresInPlaceholder")} />
                </div>
                <Button type="submit" className="w-full" disabled={isCreating}>{isCreating ? t("actions.creating") : t("actions.create")}</Button>
              </form>
            </SheetContent>
          </Sheet>
        }
      >
        <div className="flex flex-col gap-4">
          <SectionHeader
            title={t("summary", { total })}
            description={t("searchPlaceholder")}
            action={
              <div className="flex flex-wrap items-center gap-2">
                <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("searchPlaceholder")} className="w-64" />
                <Button variant="outline" onClick={() => void load()} disabled={isLoading}><RefreshCcw data-icon="inline-start" />{t("actions.refresh")}</Button>
                <Button variant="destructive" onClick={() => openDeleteDialog({ type: "all" })}><Trash2 data-icon="inline-start" />{t("actions.revokeAll")}</Button>
              </div>
            }
          />

          {tokens.length === 0 ? (
            <EmptyState icon={KeyRound} title={t("table.empty")} description={t("create.accessHint")} />
          ) : (
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
                    <TableCell>{formatDate(item.created_at)}</TableCell>
                    <TableCell>{formatDate(item.last_used_at)}</TableCell>
                    <TableCell>{formatDate(item.expires_at)}</TableCell>
                    <TableCell><StatusBadge status={item.revoked_at ? "inactive" : "active"}>{item.revoked_at ? t("status.revoked") : t("status.active")}</StatusBadge></TableCell>
                    <TableCell><Button variant="outline" size="sm" onClick={() => openDeleteDialog({ type: "single", token: item })}>{t("actions.revoke")}</Button></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </AppCard>

      <Dialog
        open={deleteDialogOpen}
        onOpenChange={(open) => {
          if (isDeleting) return;
          setDeleteDialogOpen(open);
          if (!open) setDeleteIntent(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>
              {deleteIntent?.type === "single"
                ? t("deleteDialog.singleDescription", { name: deleteIntent.token.key_alias })
                : t("deleteDialog.allDescription")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setDeleteDialogOpen(false);
                setDeleteIntent(null);
              }}
              disabled={isDeleting}
            >
              {t("deleteDialog.cancel")}
            </Button>
            <Button type="button" variant="destructive" onClick={() => void confirmDelete()} disabled={isDeleting}>
              {isDeleting ? t("deleteDialog.deleting") : t("deleteDialog.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={secretDialogOpen && Boolean(newSecret)}
        onOpenChange={(open) => {
          setSecretDialogOpen(open);
          if (!open) {
            setSecretCopied(false);
            setNewSecret(null);
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("oneTime.title")}</DialogTitle>
            <DialogDescription>{t("oneTime.description")}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            <Input value={newSecret ?? ""} readOnly className="font-mono text-xs" />
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>{t("oneTime.copyHint")}</span>
              {secretCopied ? <span className="text-foreground">{t("oneTime.copied")}</span> : null}
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setSecretDialogOpen(false)}>
              {t("oneTime.close")}
            </Button>
            <Button type="button" onClick={() => void copySecret()} disabled={isCopyingSecret || !newSecret}>
              {secretCopied ? <Check data-icon="inline-start" /> : <Copy data-icon="inline-start" />}
              {isCopyingSecret ? t("oneTime.copying") : secretCopied ? t("oneTime.copied") : t("oneTime.copy")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function formatDate(input?: string) {
  if (!input) return "-";
  const date = new Date(input);
  if (Number.isNaN(date.getTime())) return "-";
  return stableDateTimeFormatter.format(date);
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
