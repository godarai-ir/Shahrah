import { StatusBadge } from "@/components/app-ui/StatusBadge";
import type { TicketStatus } from "./types";

const tone: Record<TicketStatus, "active" | "pending" | "warning" | "success" | "neutral"> = {
  open: "active",
  in_progress: "pending",
  waiting_for_customer: "warning",
  resolved: "success",
  closed: "neutral",
};

export function TicketStatusBadge({ status, label }: { status: TicketStatus; label: string }) {
  return <StatusBadge status={tone[status]}>{label}</StatusBadge>;
}
