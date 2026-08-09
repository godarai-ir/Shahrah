"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Inbox, Plus } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DataListPanel } from "@/components/app-ui/DataListPanel";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { PageContainer } from "@/components/app-ui/PageContainer";
import { TablePagination } from "@/components/app-ui/TablePagination";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { TicketStatusBadge } from "./TicketStatusBadge";
import { readTicketResponse, type TicketListResponse, type TicketStatus } from "./types";

export function TicketListPage({ canReadAll }: { canReadAll: boolean }) {
  const t = useTranslations("support");
  const locale = useLocale();
  const router = useRouter();
  const search = useSearchParams();
  const [result, setResult] = useState<TicketListResponse>({ data: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const status = search.get("status") ?? "all";
  const query = search.get("query") ?? "";
  const cursor = search.get("cursor") ?? "";

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    const controller = new AbortController();
    const params = new URLSearchParams();
    if (status !== "all") params.set("status", status);
    if (query) params.set("query", query);
    if (cursor) params.set("cursor", cursor);
    params.set("limit", "25");
    setLoading(true);
    setError("");
    fetch(`/api/support/tickets?${params}`, { cache: "no-store", signal: controller.signal })
      .then(readTicketResponse<TicketListResponse>)
      .then(setResult)
      .catch((cause: unknown) => {
        if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : t("errors.load"));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [cursor, query, status, t]);
  /* eslint-enable react-hooks/set-state-in-effect */

  function replaceFilters(next: Record<string, string>) {
    const params = new URLSearchParams(search.toString());
    Object.entries(next).forEach(([key, value]) => {
      if (value && value !== "all") params.set(key, value);
      else params.delete(key);
    });
    params.delete("cursor");
    router.replace(`/support/tickets${params.size ? `?${params}` : ""}`);
  }

  const formatDate = (value: string) =>
    new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));

  return (
    <PageContainer>
      <div className="flex justify-end">
        <Button asChild>
          <Link href="/support/tickets/new"><Plus aria-hidden="true" />{t("actions.new")}</Link>
        </Button>
      </div>
      {error ? <Alert variant="destructive"><AlertTitle>{t("states.errorTitle")}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <DataListPanel
        aria-label={t("list.title")}
        loading={loading}
        loadingLabel={t("states.loading")}
        empty={!loading && result.data.length === 0}
        emptyState={<EmptyState icon={Inbox} title={t("states.emptyTitle")} description={t("states.emptyDescription")} primaryAction={<Button asChild><Link href="/support/tickets/new">{t("actions.new")}</Link></Button>} />}
        toolbar={
          <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_14rem]">
            <Input
              defaultValue={query}
              aria-label={t("filters.query")}
              placeholder={t("filters.queryPlaceholder")}
              onKeyDown={(event) => {
                if (event.key === "Enter") replaceFilters({ query: event.currentTarget.value.trim() });
              }}
            />
            <AppSelectField
              value={status}
              ariaLabel={t("filters.status")}
              onValueChange={(value) => replaceFilters({ status: value })}
              options={(["all", "open", "in_progress", "waiting_for_customer", "resolved", "closed"] as const).map((value) => ({
                value,
                label: value === "all" ? t("filters.all") : t(`statuses.${value}`),
              }))}
            />
          </div>
        }
        pagination={
          <TablePagination
            label={t("pagination.label")}
            summary={t("pagination.summary", { count: result.data.length })}
            previousLabel={t("pagination.previous")}
            nextLabel={t("pagination.next")}
            onPrevious={() => replaceFilters({ cursor: "" })}
            onNext={() => {
              const params = new URLSearchParams(search.toString());
              params.set("cursor", result.next_cursor ?? "");
              router.push(`/support/tickets?${params}`);
            }}
            canPrevious={Boolean(cursor)}
            canNext={Boolean(result.next_cursor)}
          />
        }
      >
        <Table>
          <TableHeader><TableRow>
            <TableHead>{t("fields.id")}</TableHead><TableHead>{t("fields.subject")}</TableHead>
            <TableHead>{t("fields.category")}</TableHead><TableHead>{t("fields.status")}</TableHead>
            {canReadAll ? <TableHead>{t("fields.requester")}</TableHead> : null}
            <TableHead>{t("fields.lastActivity")}</TableHead>
          </TableRow></TableHeader>
          <TableBody>
            {result.data.map((ticket) => (
              <TableRow key={ticket.id}>
                <TableCell><Link className="font-mono text-primary hover:underline" dir="ltr" href={`/support/tickets/${ticket.public_id}`}>{ticket.public_id}</Link></TableCell>
                <TableCell className="max-w-80 font-medium">{ticket.subject}</TableCell>
                <TableCell>{t(`categories.${ticket.category}`)}</TableCell>
                <TableCell><TicketStatusBadge status={ticket.status as TicketStatus} label={t(`statuses.${ticket.status}`)} /></TableCell>
                {canReadAll ? <TableCell>{ticket.requester_display_name || ticket.requester_email || ticket.requester_user_id}</TableCell> : null}
                <TableCell className="whitespace-nowrap">{formatDate(ticket.last_message_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataListPanel>
    </PageContainer>
  );
}
