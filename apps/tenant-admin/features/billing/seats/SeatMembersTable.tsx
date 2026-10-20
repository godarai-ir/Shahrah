"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";

import { RowActionsMenu, type RowAction } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { DataTable } from "@/components/ui/data-table";
import { Progress } from "@/components/ui/progress";

import { useBillingFormat } from "../use-billing-format";
import type { SeatMember } from "./api";
import { personalUsagePercent, SEAT_TONE } from "./domain";

type SeatMembersTableProps = {
  members: SeatMember[];
  canManage: boolean;
  pendingUserID: string | null;
  onChangeSeat: (member: SeatMember) => void;
  onCancelPending: (member: SeatMember) => void;
};

/** Members with their seat, personal credits, and pool usage this month. */
export function SeatMembersTable({ members, canManage, pendingUserID, onChangeSeat, onCancelPending }: SeatMembersTableProps) {
  const t = useTranslations("billingSeats");
  const format = useBillingFormat();
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return members;
    return members.filter((member) => `${member.name} ${member.email}`.toLowerCase().includes(needle));
  }, [members, query]);

  const columns = useMemo<ColumnDef<SeatMember>[]>(
    () => [
      {
        id: "member",
        header: t("members.columns.member"),
        cell: ({ row }) => (
          <div className="flex min-w-0 flex-col">
            <span className="truncate font-medium">{row.original.name || row.original.email}</span>
            {row.original.name ? <span className="truncate text-xs text-muted-foreground" dir="ltr">{row.original.email}</span> : null}
          </div>
        ),
      },
      {
        accessorKey: "seat",
        header: t("members.columns.seat"),
        cell: ({ row }) => (
          <div className="flex flex-col items-start gap-1">
            <StatusBadge status={SEAT_TONE[row.original.seat]}>{t(`seats.${row.original.seat}`)}</StatusBadge>
            {row.original.pending_seat ? (
              <span className="text-xs text-muted-foreground">
                {t("members.pending", {
                  seat: t(`seats.${row.original.pending_seat === "pro" ? "pro" : "unassigned"}`),
                  date: format.date(row.original.pending_effective_at),
                })}
              </span>
            ) : null}
          </div>
        ),
      },
      {
        id: "personal",
        header: t("members.columns.personal"),
        cell: ({ row }) => {
          const member = row.original;
          if (member.seat === "unassigned" && Number(member.personal_granted_credits) === 0) {
            return <span className="text-muted-foreground">—</span>;
          }
          return (
            <div className="flex min-w-36 flex-col gap-1">
              <span className="text-sm tabular-nums">
                {t("members.remainingOf", { remaining: format.credits(member.personal_remaining_credits), granted: format.credits(member.personal_granted_credits) })}
              </span>
              <Progress value={personalUsagePercent(member)} aria-label={t("members.columns.personal")} />
            </div>
          );
        },
      },
      {
        id: "pool",
        header: t("members.columns.pool"),
        cell: ({ row }) => (
          <span className="tabular-nums">
            {row.original.pool_limit_credits
              ? t("members.poolOfLimit", { used: format.credits(row.original.pool_used_credits), limit: format.credits(row.original.pool_limit_credits) })
              : format.credits(row.original.pool_used_credits)}
          </span>
        ),
      },
      {
        id: "actions",
        header: () => <span className="sr-only">{t("members.columns.actions")}</span>,
        cell: ({ row }) => {
          const member = row.original;
          const actions: RowAction[] = [{ key: "change", label: t("members.changeSeat"), onSelect: () => onChangeSeat(member), disabled: !canManage }];
          if (member.pending_seat) {
            actions.push({
              key: "cancel",
              label: t("members.cancelPending"),
              onSelect: () => onCancelPending(member),
              disabled: !canManage,
              pending: pendingUserID === member.user_id,
            });
          }
          return <RowActionsMenu label={t("members.actionsFor", { member: member.name || member.email })} actions={actions} />;
        },
      },
    ],
    [canManage, format, onCancelPending, onChangeSeat, pendingUserID, t],
  );

  return (
    <div className="flex flex-col gap-3">
      <SearchField
        containerClassName="max-w-sm"
        placeholder={t("members.search")}
        value={query}
        onChange={(event) => setQuery(event.target.value)}
      />
      <DataTable
        tableId="seat-members"
        columns={columns}
        data={filtered}
        getRowId={(member) => member.user_id}
        emptyLabel={query ? t("members.noResults") : t("members.empty")}
      />
    </div>
  );
}
