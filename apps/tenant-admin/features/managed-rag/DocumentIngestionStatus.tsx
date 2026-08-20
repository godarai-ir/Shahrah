"use client";

import {
  CheckCircle2,
  Clock3,
  DatabaseZap,
  FileSearch,
  Orbit,
  ScanSearch,
  ScissorsLineDashed,
  Trash2,
  XCircle,
  type LucideIcon,
} from "lucide-react";

import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Progress } from "@/components/ui/progress";
import type { KnowledgeBaseDocument } from "@/lib/contracts/managed-rag";
import { cn } from "@/lib/utils";

type DocumentStatus = KnowledgeBaseDocument["status"];

const statusPresentation: Record<DocumentStatus, {
  icon: LucideIcon;
  progress: number;
  tone: "pending" | "success" | "warning" | "error";
}> = {
  queued: { icon: Clock3, progress: 5, tone: "pending" },
  scanning: { icon: ScanSearch, progress: 15, tone: "pending" },
  parsing: { icon: FileSearch, progress: 30, tone: "pending" },
  chunking: { icon: ScissorsLineDashed, progress: 50, tone: "pending" },
  embedding: { icon: Orbit, progress: 70, tone: "pending" },
  indexing: { icon: DatabaseZap, progress: 90, tone: "pending" },
  ready: { icon: CheckCircle2, progress: 100, tone: "success" },
  failed: { icon: XCircle, progress: 100, tone: "error" },
  deleting: { icon: Trash2, progress: 85, tone: "warning" },
};

export const processingDocumentStatuses = new Set<DocumentStatus>([
  "queued", "scanning", "parsing", "chunking", "embedding", "indexing", "deleting",
]);

export function DocumentIngestionStatus({
  status,
  label,
  changed = false,
}: {
  status: DocumentStatus;
  label: string;
  changed?: boolean;
}) {
  const presentation = statusPresentation[status];
  const Icon = presentation.icon;
  const processing = processingDocumentStatuses.has(status);

  return (
    <div
      className={cn(
        "flex min-w-40 flex-col gap-2 rounded-md p-2 transition-colors duration-500 motion-reduce:transition-none",
        changed && "bg-primary-soft",
      )}
      data-status={status}
      data-status-changed={changed || undefined}
    >
      <StatusBadge
        key={status}
        status={presentation.tone}
        className="w-fit animate-in fade-in-0 zoom-in-95 duration-300 motion-reduce:animate-none"
      >
        <Icon
          data-icon="inline-start"
          aria-hidden="true"
          className={cn(processing && "motion-safe:animate-pulse")}
        />
        {label}
      </StatusBadge>
      <Progress
        value={presentation.progress}
        aria-label={label}
        aria-valuenow={presentation.progress}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuetext={label}
        className="h-1.5"
      />
    </div>
  );
}
