export type SupportTicketStatus = 'open' | 'in_progress' | 'waiting_for_customer' | 'resolved' | 'closed';
export type SupportTicket = {
  id: string;
  public_id: string;
  tenant_id: string;
  tenant_slug?: string;
  requester_user_id: string;
  requester_display_name?: string;
  requester_email?: string;
  subject: string;
  category: string;
  priority: string;
  status: SupportTicketStatus;
  assignee_id?: string;
  last_message_author_type: string;
  last_message_at: string;
  first_support_response_at?: string;
  resolved_at?: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  version: number;
  rating?: { score: number; comment?: string; created_at: string; updated_at: string };
};
export type SupportAttachment = {
  id: string; original_filename: string; detected_content_type: string; size_bytes: number; scan_status: string;
};
export type SupportMessage = {
  id: string; author_type: string; author_id: string; body: string; attachments: SupportAttachment[]; created_at: string;
};
export type SupportTicketDetail = { ticket: SupportTicket; messages: SupportMessage[] };
export type SupportTicketList = { data: SupportTicket[]; next_cursor?: string };
export type SupportTicketCounts = {
  needs_support_reply: number; open: number; in_progress: number; waiting_for_customer: number; resolved: number; closed: number; all: number;
  average_rating?: number;
};
export type SupportTicketEvent = {
  id: string; type: string; actor_type: string; actor_id?: string; actor_role?: string; request_id?: string;
  payload?: Record<string, unknown>; created_at: string;
};

export async function supportTicketRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/support/tickets${path}`, { ...init, cache: 'no-store' });
  const payload = await response.json().catch(() => ({})) as {
    error?: string | { message?: string };
    message?: string;
  };
  if (!response.ok) {
    const message = typeof payload.error === 'string' ? payload.error : payload.error?.message;
    throw new Error(message || payload.message || `Request failed (${response.status})`);
  }
  return payload as T;
}

export function ticketDate(value?: string): string {
  if (!value) return '—';
  return new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

export function idempotencyKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}
