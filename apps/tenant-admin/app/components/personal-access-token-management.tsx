"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { KeyRound, Plus, RefreshCcw, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

type PAT = {
  id: string;
  key_alias: string;
  key_name?: string;
  key_type?: string;
  models?: string[];
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
  models?: string[];
  created_at?: string;
  expires_at?: string;
  last_used_at?: string;
};

type ListResponse = {
  data?: PAT[];
  total?: number;
};

type Banner = { kind: "success" | "error"; message: string };
const stableDateTimeFormatter = new Intl.DateTimeFormat("fa-IR", {
  dateStyle: "short",
  timeStyle: "medium",
  timeZone: "Asia/Tehran",
});

export default function PersonalAccessTokenManagement(props: {
  initialTokens: PAT[];
  initialTotal: number;
  initialModels: string[];
  initialError: string | null;
}) {
  const t = useTranslations("personalAccessTokens");
  const [tokens, setTokens] = useState<PAT[]>(props.initialTokens);
  const [total, setTotal] = useState<number>(props.initialTotal);
  const [query, setQuery] = useState("");
  const [modelQuery, setModelQuery] = useState("");
  const [name, setName] = useState("");
  const [models, setModels] = useState<string[]>([]);
  const [expiresIn, setExpiresIn] = useState("30d");
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [banner, setBanner] = useState<Banner | null>(props.initialError ? { kind: "error", message: props.initialError } : null);
  const selectableModels = useMemo(() => {
    const q = modelQuery.trim().toLowerCase();
    if (!q) return props.initialModels;
    return props.initialModels.filter((item) => item.toLowerCase().includes(q));
  }, [modelQuery, props.initialModels]);

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
    if (!name.trim() || models.length === 0) {
      setBanner({ kind: "error", message: t("messages.validation") });
      return;
    }
    setIsCreating(true);
    try {
      const created = await requestJSON<PATResponse>("/api/personal-access-tokens", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), models, expires_in: expiresIn.trim() || undefined }),
      });
      setNewSecret(created.key);
      setName("");
      setModels([]);
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
    try {
      await requestJSON(`/api/personal-access-tokens/${encodeURIComponent(id)}`, { method: "DELETE" });
      setBanner({ kind: "success", message: t("messages.revokeSuccess") });
      await load();
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeError")) });
    }
  }

  async function revokeAll() {
    try {
      await requestJSON("/api/personal-access-tokens/revoke-all", { method: "DELETE" });
      setBanner({ kind: "success", message: t("messages.revokeAllSuccess") });
      await load();
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.revokeAllError")) });
    }
  }

  return (
    <div className="space-y-5">
      <header className="space-y-4">
        <h1 className="text-4xl font-bold tracking-tight">{t("title")}</h1>
        <nav className="border-b">
          <ul className="flex flex-wrap items-center gap-5 pb-2">
            <li><Link href="/users" className="text-muted-foreground hover:text-foreground">{t("tabs.users")}</Link></li>
            <li><span className="inline-flex items-center gap-2 border-b-2 border-indigo-500 px-1 py-2 text-sm font-semibold"><KeyRound className="h-4 w-4" />{t("tabs.pat")}</span></li>
          </ul>
        </nav>
      </header>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      {newSecret ? (
        <Alert className="border-primary/40 bg-primary/5">
          <AlertTitle>{t("oneTime.title")}</AlertTitle>
          <AlertDescription>
            <p className="mb-2 text-sm">{t("oneTime.description")}</p>
            <code className="block overflow-x-auto rounded bg-muted px-3 py-2 text-xs">{newSecret}</code>
          </AlertDescription>
        </Alert>
      ) : null}

      <section className="rounded-xl border bg-card px-4 py-5 space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t("searchPlaceholder")} className="w-64" />
          <Button variant="outline" onClick={() => void load()} disabled={isLoading}><RefreshCcw className="h-4 w-4" />{t("actions.refresh")}</Button>
          <Button variant="destructive" onClick={() => void revokeAll()}><Trash2 className="h-4 w-4" />{t("actions.revokeAll")}</Button>

          <Sheet>
            <SheetTrigger asChild>
              <Button><Plus className="h-4 w-4" />{t("actions.create")}</Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[94vw] max-w-md">
              <SheetHeader>
                <SheetTitle>{t("create.title")}</SheetTitle>
                <SheetDescription>{t("create.description")}</SheetDescription>
              </SheetHeader>
              <form className="mt-6 space-y-4" onSubmit={createPAT}>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">{t("fields.name")}</label>
                  <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t("fields.namePlaceholder")} />
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">{t("fields.models")}</label>
                  <Input value={modelQuery} onChange={(e) => setModelQuery(e.target.value)} placeholder={t("searchPlaceholder")} />
                  <div className="max-h-56 overflow-auto rounded-md border p-2">
                    {selectableModels.length === 0 ? (
                      <p className="text-xs text-muted-foreground">{t("table.empty")}</p>
                    ) : selectableModels.map((modelID) => (
                      <label key={modelID} className="flex items-center gap-2 py-1 text-sm">
                        <input
                          type="checkbox"
                          checked={models.includes(modelID)}
                          onChange={(event) => {
                            if (event.target.checked) {
                              setModels((prev) => Array.from(new Set([...prev, modelID])));
                              return;
                            }
                            setModels((prev) => prev.filter((item) => item !== modelID));
                          }}
                        />
                        <span className="font-mono text-xs">{modelID}</span>
                      </label>
                    ))}
                  </div>
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">{t("fields.expiresIn")}</label>
                  <Input value={expiresIn} onChange={(e) => setExpiresIn(e.target.value)} placeholder={t("fields.expiresInPlaceholder")} />
                </div>
                <Button type="submit" className="w-full" disabled={isCreating}>{isCreating ? t("actions.creating") : t("actions.create")}</Button>
              </form>
            </SheetContent>
          </Sheet>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("table.name")}</TableHead>
              <TableHead>{t("table.masked")}</TableHead>
              <TableHead>{t("table.models")}</TableHead>
              <TableHead>{t("table.createdAt")}</TableHead>
              <TableHead>{t("table.lastUsedAt")}</TableHead>
              <TableHead>{t("table.expiresAt")}</TableHead>
              <TableHead>{t("table.status")}</TableHead>
              <TableHead>{t("table.actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tokens.length === 0 ? <TableRow><TableCell colSpan={8}>{t("table.empty")}</TableCell></TableRow> : tokens.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.key_alias}</TableCell>
                <TableCell>{item.key_name ?? "-"}</TableCell>
                <TableCell>{(item.models ?? []).join(", ") || "-"}</TableCell>
                <TableCell>{formatDate(item.created_at)}</TableCell>
                <TableCell>{formatDate(item.last_used_at)}</TableCell>
                <TableCell>{formatDate(item.expires_at)}</TableCell>
                <TableCell><Badge variant="secondary">{item.revoked_at ? t("status.revoked") : t("status.active")}</Badge></TableCell>
                <TableCell><Button variant="outline" size="sm" onClick={() => void revokeOne(item.id)}>{t("actions.revoke")}</Button></TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>

        <p className="text-sm text-muted-foreground">{t("summary", { total })}</p>
      </section>
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
