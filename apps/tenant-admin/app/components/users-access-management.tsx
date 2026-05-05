"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import {
  BadgeCheck,
  BriefcaseBusiness,
  KeyRound,
  MoreVertical,
  Plus,
  Search,
  ShieldCheck,
  UserCircle2,
  Users,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type TenantAdminUser = {
  id: string;
  tenant_slug: string;
  email?: string;
  mobile?: string;
  role: string;
  active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

type UserListResponse = {
  data?: TenantAdminUser[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

type InviteResponse = {
  link?: string;
  invitation?: {
    email?: string;
    role?: string;
  };
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type UsersAccessManagementProps = {
  initialUsers: TenantAdminUser[];
  initialTotal: number;
  initialError: string | null;
};

const PAGE_SIZE = 25;

export default function UsersAccessManagement(props: UsersAccessManagementProps) {
  const t = useTranslations("accessManagement");

  const [users, setUsers] = useState<TenantAdminUser[]>(props.initialUsers);
  const [total, setTotal] = useState(props.initialTotal);
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isInviting, setIsInviting] = useState(false);
  const [isUpdatingUserID, setIsUpdatingUserID] = useState<string | null>(null);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("tenant_member");
  const [banner, setBanner] = useState<Banner | null>(
    props.initialError
      ? {
          kind: "error",
          message: props.initialError,
        }
      : null,
  );
  const [lastInviteLink, setLastInviteLink] = useState<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setOffset(0);
      void loadUsers(0, query.trim());
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + users.length, total);
  const canPrev = offset > 0;
  const canNext = offset + users.length < total;

  const accessTabs = useMemo(
    () => [
      { href: "/users", label: t("tabs.users"), icon: UserCircle2, active: true },
      { href: "/teams", label: t("tabs.teams"), icon: Users, active: false },
      { href: "#", label: t("tabs.externalAuth"), icon: ShieldCheck, active: false },
      { href: "#", label: t("tabs.personalAccessTokens"), icon: KeyRound, active: false },
      { href: "#", label: t("tabs.virtualAccounts"), icon: BriefcaseBusiness, active: false },
      { href: "#", label: t("tabs.roles"), icon: BadgeCheck, active: false },
      { href: "#", label: t("tabs.auditLogs"), icon: ShieldCheck, active: false },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [t],
  );

  async function loadUsers(nextOffset: number, q: string) {
    setIsLoading(true);
    try {
      const search = new URLSearchParams();
      search.set("limit", String(PAGE_SIZE));
      search.set("offset", String(nextOffset));
      if (q) search.set("query", q);

      const data = await requestJSON<UserListResponse>(`/api/users?${search.toString()}`);
      setUsers(data.data ?? []);
      setTotal(data.total ?? 0);
      setOffset(nextOffset);
      setBanner(null);
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.loadUsersError")),
      });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleInvite(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);
    setLastInviteLink(null);

    const email = inviteEmail.trim();
    if (!email) {
      setBanner({
        kind: "error",
        message: t("messages.provideEmail"),
      });
      return;
    }

    setIsInviting(true);
    try {
      const response = await requestJSON<InviteResponse>("/api/users/invite", {
        method: "POST",
        body: JSON.stringify({
          email,
          role: inviteRole,
        }),
      });
      setLastInviteLink(response.link ?? null);
      setBanner({
        kind: "success",
        message: t("messages.invitationSuccess"),
      });
      setInviteEmail("");
      setInviteRole("tenant_member");
      await loadUsers(0, query.trim());
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.inviteError")),
      });
    } finally {
      setIsInviting(false);
    }
  }

  async function changeUserStatus(userID: string, activate: boolean) {
    setIsUpdatingUserID(userID);
    setBanner(null);
    try {
      await requestJSON(`/api/users/${encodeURIComponent(userID)}/${activate ? "activate" : "deactivate"}`, {
        method: "POST",
      });
      setBanner({
        kind: "success",
        message: activate ? t("messages.activateSuccess") : t("messages.deactivateSuccess"),
      });
      await loadUsers(offset, query.trim());
    } catch (error) {
      setBanner({
        kind: "error",
        message: normalizeError(error, t("messages.updateStatusError")),
      });
    } finally {
      setIsUpdatingUserID(null);
    }
  }

  return (
    <div className="space-y-5">
      <header className="space-y-4">
        <h1 className="text-4xl font-bold tracking-tight">{t("title")}</h1>
        <nav className="border-b">
          <ul className="flex flex-wrap items-center gap-5 pb-2">
            {accessTabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <li key={tab.label}>
                  <Link
                    href={tab.href}
                    className={`inline-flex items-center gap-2 border-b-2 px-1 py-2 ${
                      tab.active
                        ? "border-indigo-500 text-foreground"
                        : "border-transparent text-muted-foreground hover:text-foreground"
                    }`}
                    aria-current={tab.active ? "page" : undefined}
                  >
                    <Icon className="h-4 w-4" />
                    <span className="text-sm font-semibold">{tab.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </header>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      {lastInviteLink ? (
        <Alert className="border-primary/40 bg-primary/5">
          <AlertTitle>{t("alerts.invitationLink")}</AlertTitle>
          <AlertDescription>
            <code className="block overflow-x-auto rounded bg-muted px-3 py-2 text-xs">
              {lastInviteLink}
            </code>
          </AlertDescription>
        </Alert>
      ) : null}

      <section className="rounded-xl border bg-card px-4 py-5">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold tracking-wide text-slate-700">{t("allUsers")}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={t("searchPlaceholder")}
                className="w-64 pl-9"
              />
            </div>

            <Sheet>
              <SheetTrigger asChild>
                <Button className="gap-2">
                  <Plus className="h-4 w-4" />
                  {t("inviteUser")}
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-[94vw] max-w-md">
                <SheetHeader>
                  <SheetTitle>{t("inviteTitle")}</SheetTitle>
                  <SheetDescription>
                    {t("inviteDescription")}
                  </SheetDescription>
                </SheetHeader>

                <form className="mt-6 space-y-4" onSubmit={handleInvite}>
                  <div className="grid gap-2">
                    <label className="text-sm font-medium" htmlFor="invite-email">
                      {t("form.email")}
                    </label>
                    <Input
                      id="invite-email"
                      value={inviteEmail}
                      onChange={(event) => setInviteEmail(event.target.value)}
                      placeholder={t("form.emailPlaceholder")}
                    />
                  </div>
                  <div className="grid gap-2">
                    <label className="text-sm font-medium" htmlFor="invite-role">
                      {t("form.role")}
                    </label>
                    <select
                      id="invite-role"
                      value={inviteRole}
                      onChange={(event) => setInviteRole(event.target.value)}
                      className="h-10 rounded-md border border-input bg-background px-3 text-sm"
                    >
                      <option value="tenant_member">{t("roles.member")}</option>
                      <option value="tenant_admin">{t("roles.tenantAdmin")}</option>
                    </select>
                  </div>

                  <Button type="submit" disabled={isInviting} className="w-full">
                    {isInviting ? t("button.inviting") : t("button.createInvite")}
                  </Button>
                </form>
              </SheetContent>
            </Sheet>
          </div>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("table.emailMobile")}</TableHead>
              <TableHead>{t("table.displayName")}</TableHead>
              <TableHead>{t("table.role")}</TableHead>
              <TableHead className="w-16 text-right"> </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-muted-foreground">
                  {t("table.noUsersFound")}
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => {
                const identity = user.email || user.mobile || user.id;
                const displayName = guessDisplayName(identity);
                const active = user.active && user.role !== "deactivated";

                return (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{identity}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Avatar className="h-8 w-8">
                          <AvatarFallback>{displayName.slice(0, 1).toUpperCase()}</AvatarFallback>
                        </Avatar>
                        <span>{displayName || "-"}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={active ? "secondary" : "outline"}>
                        {active
                          ? user.role === "tenant_admin"
                            ? t("roles.tenantAdmin")
                            : t("roles.member")
                          : t("status.deactivated")}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            disabled={isUpdatingUserID === user.id}
                          >
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {active ? (
                            <DropdownMenuItem onClick={() => void changeUserStatus(user.id, false)}>
                              {t("actions.deactivate")}
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem onClick={() => void changeUserStatus(user.id, true)}>
                              {t("actions.activate")}
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
          <p>
            {t("pagination.showingResults", { from, to, total })}
          </p>
          <div className="flex items-center gap-2">
            <span>{t("pagination.rowsPerPage")}</span>
            <Badge variant="outline">{PAGE_SIZE}</Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void loadUsers(Math.max(0, offset - PAGE_SIZE), query.trim())}
              disabled={!canPrev || isLoading}
            >
              {t("pagination.prev")}
            </Button>
            <Badge variant="secondary">{Math.floor(offset / PAGE_SIZE) + 1}</Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void loadUsers(offset + PAGE_SIZE, query.trim())}
              disabled={!canNext || isLoading}
            >
              {t("pagination.next")}
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}

function guessDisplayName(identity: string) {
  const value = identity.trim();
  if (!value) return "";
  if (value.includes("@")) {
    return value.split("@")[0];
  }
  return value;
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, {
    ...options,
    headers,
    cache: "no-store",
  });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  }
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
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
