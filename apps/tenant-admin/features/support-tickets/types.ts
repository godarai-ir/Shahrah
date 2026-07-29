export type TicketStatus = "open" | "in_progress" | "waiting_for_customer" | "resolved" | "closed";
export type TicketCategory = "technical" | "billing" | "account" | "feature_request" | "other";
export type TicketPriority = "low" | "normal" | "high" | "urgent";

export type TicketAttachment = {
  id: string;
  original_filename: string;
  detected_content_type: string;
  size_bytes: number;
  scan_status: "quarantined" | "clean" | "rejected";
};

export type TicketMessage = {
  id: string;
  author_type: "tenant_user" | "support_agent" | "system";
  author_id: string;
  author_label?: string;
  body: string;
  attachments: TicketAttachment[];
  created_at: string;
};

export type TicketRating = {
  score: number;
  comment?: string;
  created_at: string;
  updated_at: string;
};

export type Ticket = {
  id: string;
  public_id: string;
  tenant_id: string;
  tenant_slug?: string;
  requester_user_id: string;
  requester_display_name?: string;
  requester_email?: string;
  subject: string;
  category: TicketCategory;
  priority: TicketPriority;
  status: TicketStatus;
  assignee_id?: string;
  last_message_author_type: string;
  last_message_at: string;
  first_support_response_at?: string;
  resolved_at?: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  version: number;
  rating?: TicketRating;
};

export type TicketDetail = {
  ticket: Ticket;
  messages: TicketMessage[];
};

export type TicketListResponse = {
  data: Ticket[];
  next_cursor?: string;
};

export async function readTicketResponse<T>(response: Response): Promise<T> {
  const payload = (await response.json().catch(() => null)) as
    | T
    | { error?: { message?: string }; message?: string; details?: { error?: { message?: string } } }
    | null;
  if (!response.ok) {
    const errorPayload = payload as { error?: { message?: string }; message?: string; details?: { error?: { message?: string } } } | null;
    throw new Error(
      errorPayload?.error?.message ??
        errorPayload?.details?.error?.message ??
        errorPayload?.message ??
        `Request failed (${response.status})`,
    );
  }
  return payload as T;
}

export function createIdempotencyKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
