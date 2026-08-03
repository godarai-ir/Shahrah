import * as React from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export type OperationNoticeKind = "success" | "error" | "warning" | "info";
export type OperationNoticeValue = { kind: OperationNoticeKind; message: React.ReactNode };
export type OperationNoticeLabels = Record<OperationNoticeKind, React.ReactNode>;

export type OperationNoticeProps = {
  notice?: OperationNoticeValue | null;
  labels: OperationNoticeLabels;
  action?: React.ReactNode;
  className?: string;
};

const variants = { success: "success", error: "destructive", warning: "warning", info: "info" } as const;

export function OperationNotice({ notice, labels, action, className }: OperationNoticeProps) {
  if (!notice) return null;

  return (
    <Alert
      variant={variants[notice.kind]}
      role={notice.kind === "error" ? "alert" : "status"}
      aria-live={notice.kind === "error" ? "assertive" : "polite"}
      className={className}
    >
      <AlertTitle>{labels[notice.kind]}</AlertTitle>
      <AlertDescription>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0 flex-1">{notice.message}</div>
          {action ? <div className="shrink-0">{action}</div> : null}
        </div>
      </AlertDescription>
    </Alert>
  );
}
