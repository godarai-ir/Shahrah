"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { memo } from "react";
import { useTranslations } from "next-intl";

import { CompactValue } from "@/components/app-ui/CompactValue";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { MetricCard } from "@/components/app-ui/MetricCard";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import type { LogDetail, LogsFormatter, RequestLog } from "./logs-types";

const JsonViewer = dynamic(
  () => import("@/components/app-ui/JsonViewer").then((module) => module.JsonViewer),
  {
    ssr: false,
    loading: () => <div className="min-h-40 animate-pulse rounded-xl border bg-muted/35" aria-hidden="true" />,
  },
);

export const LogDetailSheet = memo(function LogDetailSheet({
  selectedLog,
  detail,
  traceAvailable,
  onClose,
  formatters,
  closeLabel,
  t,
}: {
  selectedLog: RequestLog | null;
  detail: LogDetail | null;
  traceAvailable: boolean;
  onClose: () => void;
  formatters: LogsFormatter;
  closeLabel: string;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <Sheet open={selectedLog !== null} onOpenChange={(open) => { if (!open) onClose(); }}>
      <SheetContent side="left" closeLabel={closeLabel} className="w-full overflow-y-auto sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>{t("detailTitle")}</SheetTitle>
          <SheetDescription asChild>
            <span className="flex min-w-0 items-center justify-end gap-1" dir="ltr">
              <span className="truncate font-mono">{selectedLog?.log_id}</span>
              {selectedLog?.log_id ? (
                <CopyAction
                  value={selectedLog.log_id}
                  copyLabel={t("actions.copyLogId")}
                  copyingLabel={t("actions.copyingLogId")}
                  copiedLabel={t("actions.copiedLogId")}
                  size="icon-xs"
                />
              ) : null}
            </span>
          </SheetDescription>
        </SheetHeader>
        {detail ? (
          <div className="flex flex-col gap-4 pt-4">
            <div className="grid grid-cols-3 gap-2">
              <MetricCard label={t("table.status")} value={detail.log.status_code} />
              <MetricCard label={t("table.latency")} value={formatters.ms(detail.log.duration_ms)} />
              <MetricCard label={t("table.cost")} value={<CompactValue display={formatters.usdDisplay(detail.log.cost_usd ?? 0)} />} />
            </div>
            {traceAvailable && detail.log.trace_id ? (
              <div className="flex flex-wrap justify-end gap-2">
                <Button asChild size="sm" variant="outline">
                  <Link href={`/traces/${detail.log.trace_id}`}>{t("actions.openTrace")}</Link>
                </Button>
              </div>
            ) : null}
            <Tabs defaultValue="summary" className="min-w-0">
              <TabsList variant="line" aria-label={t("detailTitle")} className="w-full justify-start overflow-x-auto">
                <TabsTrigger value="summary">{t("tabs.summary")}</TabsTrigger>
                <TabsTrigger value="request">{t("content.request")}</TabsTrigger>
                <TabsTrigger value="response">{t("content.response")}</TabsTrigger>
                <TabsTrigger value="metadata">{t("tabs.metadata")}</TabsTrigger>
              </TabsList>
              <TabsContent value="summary"><JSONBlock label={t("tabs.summary")} value={{ log: detail.log, actor: detail.actor, scope: detail.scope, content_visibility: detail.content_visibility }} /></TabsContent>
              <TabsContent value="request"><PayloadBlock label={t("content.request")} available={detail.content_visibility.request_available} value={detail.request_payload} reason={detail.content_visibility.request_reason} hidden={t("content.hidden")} /></TabsContent>
              <TabsContent value="response"><PayloadBlock label={t("content.response")} available={detail.content_visibility.response_available} value={detail.response_payload} reason={detail.content_visibility.response_reason} hidden={t("content.hidden")} /></TabsContent>
              <TabsContent value="metadata"><JSONBlock label={t("tabs.metadata")} value={{ metadata: detail.metadata, request_payload_meta: detail.request_payload_meta, response_payload_meta: detail.response_payload_meta }} /></TabsContent>
            </Tabs>
          </div>
        ) : (
          <EmptyState title={t("detailLoading")} description={t("detailLoadingDescription")} />
        )}
      </SheetContent>
    </Sheet>
  );
});

function JSONBlock({ label, value }: { label: string; value: unknown }) {
  return <JsonViewer aria-label={label} value={value} maxHeight="520px" />;
}

function PayloadBlock({ label, available, value, reason, hidden }: { label: string; available: boolean; value?: unknown; reason?: string; hidden: string }) {
  if (!available || value === undefined || value === null) {
    return <Alert><AlertTitle>{hidden}</AlertTitle><AlertDescription>{reason}</AlertDescription></Alert>;
  }
  return <JSONBlock label={label} value={value} />;
}
