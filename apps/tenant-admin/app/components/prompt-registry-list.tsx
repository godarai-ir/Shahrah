"use client";

import { MoreVertical } from "lucide-react";
import { Search } from "lucide-react";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { cn } from "@/lib/utils";

import { formatDate, type Prompt, type PromptTranslationFn } from "./prompt-management-shared";

type PromptRegistryListProps = {
  isLoading: boolean;
  locale: string;
  onNextPage: () => void;
  onPreviousPage: () => void;
  onCopyFqn: (prompt: Prompt) => void;
  onCreateVersion: (prompt: Prompt) => void;
  onDelete: (prompt: Prompt) => void;
  onEditLatestVersion: (prompt: Prompt) => void;
  onRefresh: () => void;
  onRowClick: (prompt: Prompt) => void;
  onRowsPerPageChange: (next: number) => void;
  onSearchKeyDown: (key: string) => void;
  onSearchValueChange: (value: string) => void;
  onTryPrompt: (prompt: Prompt) => void;
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

  return (
    <Card className="shadow-none">
      <CardHeader className="border-b">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <CardTitle>{props.t("list.title")}</CardTitle>
            <CardDescription>{props.t("list.description")}</CardDescription>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <div className="relative min-w-[240px]">
              <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={props.query}
                onChange={(event) => props.onSearchValueChange(event.target.value)}
                onKeyDown={(event) => props.onSearchKeyDown(event.key)}
                placeholder={props.t("list.searchPlaceholder")}
                className="ps-9"
              />
            </div>
            <Button type="button" variant="outline" onClick={props.onRefresh}>
              {props.t("actions.refresh")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="p-0">
        {props.isLoading ? (
          <div className="space-y-2 p-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : props.prompts.length === 0 ? (
          <div className="p-10 text-center text-sm text-muted-foreground">{props.t("list.emptyRows")}</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{props.t("list.columns.name")}</TableHead>
                <TableHead>{props.t("list.columns.latestVersion")}</TableHead>
                <TableHead>{props.t("list.columns.model")}</TableHead>
                <TableHead>{props.t("list.columns.description")}</TableHead>
                <TableHead>{props.t("list.columns.updatedAt")}</TableHead>
                <TableHead>{props.t("list.columns.createdBy")}</TableHead>
                <TableHead className="w-16 text-right">{props.t("list.columns.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {props.prompts.map((item) => (
                <TableRow
                  key={item.id}
                  className={cn("cursor-pointer")}
                  onClick={() => props.onRowClick(item)}
                >
                  <TableCell className="font-medium">{item.name}</TableCell>
                  <TableCell>{props.t("list.versionValue", { version: item.latest_version })}</TableCell>
                  <TableCell className="font-mono text-xs">{item.latest_version_detail?.model || "-"}</TableCell>
                  <TableCell>{item.description || "-"}</TableCell>
                  <TableCell>{formatDate(item.updated_at, props.locale)}</TableCell>
                  <TableCell>{props.userLabelByID[item.created_by ?? ""] || item.created_by || "-"}</TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={(event) => event.stopPropagation()}
                          aria-label={props.t("actions.openMenu")}
                        >
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" onClick={(event) => event.stopPropagation()}>
                        <DropdownMenuItem onClick={() => props.onTryPrompt(item)}>{props.t("actions.tryPrompt")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => props.onCreateVersion(item)}>{props.t("actions.createVersion")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => props.onEditLatestVersion(item)}>{props.t("actions.editLatestVersion")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => props.onCopyFqn(item)}>{props.t("actions.copyFqn")}</DropdownMenuItem>
                        <DropdownMenuItem onClick={() => props.onDelete(item)} className="text-destructive">
                          {props.t("actions.delete")}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        <div className="flex flex-col gap-3 border-t px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between">
          <div className="text-muted-foreground">
            {props.t("pagination.showing", {
              from: props.prompts.length === 0 ? 0 : props.pageOffset + 1,
              to: pageCountEnd,
              total: props.promptTotal,
            })}
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground">{props.t("pagination.rowsPerPage")}</span>
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
            <Button type="button" variant="outline" disabled={props.pageOffset === 0} onClick={props.onPreviousPage}>
              {props.t("pagination.previous")}
            </Button>
            <Button type="button" variant="outline" disabled={props.pageOffset + props.pageSize >= props.promptTotal} onClick={props.onNextPage}>
              {props.t("pagination.next")}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
