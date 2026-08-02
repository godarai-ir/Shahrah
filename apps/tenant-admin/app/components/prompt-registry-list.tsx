"use client";

import { Plus, ScrollText } from "lucide-react";
import Link from "next/link";

import {
  AppCard,
  AppSelectField,
  CopyAction,
  DateTimeValue,
  RefreshButton,
  RowActionsMenu,
  SearchField,
  TableEmptyRow,
  TableLoadingRows,
  TablePagination,
} from "@/components/app-ui";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

import { formatNumber, type Prompt, type PromptTranslationFn } from "./prompt-management-shared";

type PromptRegistryListProps = {
  isLoading: boolean;
  locale: string;
  onNextPage: () => void;
  onPreviousPage: () => void;
  onDelete: (prompt: Prompt) => void;
  onClearSearch: () => void;
  onRefresh: () => void;
  onRowClick: (prompt: Prompt) => void;
  onRowsPerPageChange: (next: number) => void;
  onSearchKeyDown: (key: string) => void;
  onSearchValueChange: (value: string) => void;
  pageOffset: number;
  pageSize: number;
  promptTotal: number;
  prompts: Prompt[];
  query: string;
  t: PromptTranslationFn;
  userLabelByID: Record<string, string>;
};

export function PromptRegistryList(props: PromptRegistryListProps) {
  const pageCountEnd = Math.min(props.pageOffset + props.pageSize, props.promptTotal);
  const currentPage = Math.floor(props.pageOffset / props.pageSize) + 1;

  return (
    <AppCard
      className="min-w-0 shadow-none"
      contentClassName="flex min-w-0 flex-col gap-4"
    >
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <SearchField
          containerClassName="min-w-0 flex-1"
          value={props.query}
          onChange={(event) => props.onSearchValueChange(event.target.value)}
          onKeyDown={(event) => props.onSearchKeyDown(event.key)}
          placeholder={props.t("list.searchPlaceholder")}
        />
        <div className="flex items-center gap-2">
          <RefreshButton label={props.t("actions.refresh")} onClick={props.onRefresh} isLoading={props.isLoading} />
          <Button asChild className="flex-1 sm:flex-none">
            <Link href="/playground">
              <Plus data-icon="inline-start" />
              {props.t("actions.createPrompt")}
            </Link>
          </Button>
        </div>
      </div>

      <div className="@container overflow-x-auto rounded-lg border">
        <Table className="table-fixed">
          <TableHeader>
            <TableRow>
              <TableHead>{props.t("list.columns.name")}</TableHead>
              <TableHead className="w-24">{props.t("list.columns.latestVersion")}</TableHead>
              <TableHead className="hidden w-72 @lg:table-cell">{props.t("list.columns.technicalIdentifier")}</TableHead>
              <TableHead className="hidden w-40 @5xl:table-cell">{props.t("list.columns.description")}</TableHead>
              <TableHead className="hidden w-44 @4xl:table-cell">{props.t("list.columns.updatedAt")}</TableHead>
              <TableHead className="hidden w-52 @5xl:table-cell">{props.t("list.columns.createdBy")}</TableHead>
              <TableHead className="w-16 text-end">{props.t("list.columns.actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {props.isLoading && props.prompts.length === 0 ? (
              <TableLoadingRows colSpan={7} label={props.t("list.loadingRows")} />
            ) : props.prompts.length === 0 ? (
              <TableEmptyRow
                colSpan={7}
                icon={ScrollText}
                title={props.query ? props.t("list.noResultsTitle") : props.t("list.emptyRows")}
                description={props.query ? props.t("list.noResultsDescription") : props.t("list.emptyDescription")}
                primaryAction={
                  props.query ? (
                    <Button type="button" variant="outline" onClick={props.onClearSearch}>
                      {props.t("actions.clearSearch")}
                    </Button>
                  ) : (
                    <Button asChild>
                      <Link href="/playground">
                        <Plus data-icon="inline-start" />
                        {props.t("actions.createPrompt")}
                      </Link>
                    </Button>
                  )
                }
              />
            ) : (
              props.prompts.map((item) => (
                <TableRow
                  key={item.id}
                  className={cn("cursor-pointer")}
                  onClick={() => props.onRowClick(item)}
                >
                  <TableCell className="whitespace-nowrap font-medium">{item.name}</TableCell>
                  <TableCell className="whitespace-nowrap">{formatNumber(item.latest_version, props.locale)}</TableCell>
                  <TableCell className="hidden @lg:table-cell" onClick={(event) => event.stopPropagation()}>
                    {item.latest_version_detail?.fqn ? (
                      <div dir="ltr" className="flex min-w-0 items-center justify-start gap-1 text-start">
                        <code className="block min-w-0 truncate font-mono text-xs" title={item.latest_version_detail.fqn}>
                          {item.latest_version_detail.fqn}
                        </code>
                        <CopyAction
                          value={item.latest_version_detail.fqn}
                          copyLabel={props.t("actions.copyFqn")}
                          copiedLabel={props.t("messages.fqnCopied")}
                          className="shrink-0"
                        />
                      </div>
                    ) : "-"}
                  </TableCell>
                  <TableCell className="hidden @5xl:table-cell"><span className="block truncate">{item.description || "-"}</span></TableCell>
                  <TableCell className="hidden whitespace-nowrap @4xl:table-cell"><DateTimeValue value={item.updated_at} locale={props.locale} /></TableCell>
                  <TableCell className="hidden truncate @5xl:table-cell">{props.userLabelByID[item.created_by ?? ""] || item.created_by || "-"}</TableCell>
                  <TableCell className="text-end">
                    <RowActionsMenu
                      label={props.t("actions.openMenu")}
                      dir={props.locale === "fa" ? "rtl" : "ltr"}
                      stopPropagation
                      actions={[
                        {
                          label: props.t("actions.tryPrompt"),
                          href: `/playground?promptID=${encodeURIComponent(item.id)}&versionID=${encodeURIComponent(item.latest_version_detail?.id ?? "")}`,
                        },
                        { label: props.t("actions.delete"), destructive: true, separatorBefore: true, onSelect: () => props.onDelete(item) },
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <TablePagination
        label={props.t("list.title")}
        summary={props.t("pagination.showing", {
          from: props.prompts.length === 0 ? 0 : props.pageOffset + 1,
          to: pageCountEnd,
          total: props.promptTotal,
        })}
        previousLabel={props.t("pagination.previous")}
        nextLabel={props.t("pagination.next")}
        onPrevious={props.onPreviousPage}
        onNext={props.onNextPage}
        canPrevious={props.pageOffset > 0}
        canNext={props.pageOffset + props.pageSize < props.promptTotal}
        disabled={props.isLoading}
        page={currentPage}
        pageSizeLabel={
          <span className="flex items-center gap-2">
            <span>{props.t("pagination.rowsPerPage")}</span>
            <AppSelectField
              ariaLabel={props.t("pagination.rowsPerPage")}
              value={String(props.pageSize)}
              onValueChange={(value) => props.onRowsPerPageChange(Number(value))}
              size="sm"
              className="w-auto"
              triggerClassName="min-w-16"
              options={[
                { value: "10", label: "10" },
                { value: "25", label: "25" },
                { value: "50", label: "50" },
              ]}
            />
          </span>
        }
      />
    </AppCard>
  );
}
