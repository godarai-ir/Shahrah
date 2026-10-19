"use client";

import { Copy, Edit3, Trash2 } from "lucide-react";
import type { useTranslations } from "next-intl";

import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { VirtualModel } from "@/features/virtual-models/api";
import { cn } from "@/lib/utils";

import { formatStrategyLabel } from "./helpers";

export function VirtualModelTable({
  items,
  selectedID,
  locale,
  userLabelByID,
  isLoading,
  actingID,
  onDuplicate,
  onEdit,
  onDelete,
  t,
}: {
  items: VirtualModel[];
  selectedID: string | null;
  locale: string;
  userLabelByID: Record<string, string>;
  isLoading: boolean;
  actingID: string | null;
  onDuplicate: (item: VirtualModel) => void;
  onEdit: (item: VirtualModel) => void;
  onDelete: (item: VirtualModel) => void;
  t: ReturnType<typeof useTranslations<"virtualModels">>;
}) {
  return (
    <Table className="table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead>{t("list.alias")}</TableHead>
          <TableHead className="hidden w-24 @sm:table-cell">{t("list.strategy")}</TableHead>
          <TableHead className="hidden w-20 @md:table-cell">{t("list.targets")}</TableHead>
          <TableHead className="hidden w-40 @4xl:table-cell">{t("list.updatedAt")}</TableHead>
          <TableHead className="hidden w-48 @4xl:table-cell">{t("list.createdBy")}</TableHead>
          <TableHead className="w-24">{t("list.status")}</TableHead>
          <TableHead className="w-16 whitespace-nowrap text-center">{t("list.actions")}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {items.map((item) => (
          <TableRow
            key={item.id}
            className={cn(selectedID === item.id && "bg-muted/60")}
          >
            <TableCell>
              <ModelNameValue
                value={item.alias}
                className="w-full flex-row-reverse"
                onCopyError={() => showFeedbackToast({ kind: "error", title: t("alerts.error"), description: t("messages.copyAliasError") })}
              />
            </TableCell>
            <TableCell className="hidden @sm:table-cell">{formatStrategyLabel(item.routing_strategy, t)}</TableCell>
            <TableCell className="hidden @md:table-cell">{item.targets?.length ?? 0}</TableCell>
            <TableCell className="hidden text-sm text-muted-foreground @4xl:table-cell"><DateTimeValue value={item.updated_at} locale={locale} format={{ year: "numeric", month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" }} /></TableCell>
            <TableCell className="hidden truncate text-sm text-muted-foreground @4xl:table-cell">{userLabelByID[item.created_by ?? ""] || item.created_by || "-"}</TableCell>
            <TableCell>
              <StatusBadge status={item.enabled ? "active" : "inactive"}>
                {item.enabled ? t("status.enabled") : t("status.disabled")}
              </StatusBadge>
            </TableCell>
            <TableCell onClick={(event) => event.stopPropagation()}>
              <RowActionsMenu
                label={t("actions.openMenu", { alias: item.alias })}
                dir={locale.startsWith("fa") ? "rtl" : "ltr"}
                disabled={isLoading}
                actions={[
                  {
                    key: "copy",
                    label: t("actions.copy"),
                    pendingLabel: t("actions.copying"),
                    pending: actingID === item.id,
                    icon: <Copy />,
                    onSelect: () => onDuplicate(item),
                  },
                  {
                    key: "edit",
                    label: t("actions.edit"),
                    disabled: actingID === item.id,
                    icon: <Edit3 />,
                    onSelect: () => onEdit(item),
                  },
                  {
                    key: "delete",
                    label: t("actions.delete"),
                    disabled: actingID === item.id,
                    destructive: true,
                    separatorBefore: true,
                    icon: <Trash2 />,
                    onSelect: () => onDelete(item),
                  },
                ]}
              />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
