import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

const ticketStatusSchema = z.enum(["open", "in_progress", "waiting_for_customer", "resolved", "closed"]);
const ticketCategorySchema = z.enum(["technical", "billing", "account", "feature_request", "other"]);
const ticketPrioritySchema = z.enum(["low", "normal", "high", "urgent"]);
const ticketAttachmentSchema = z.object({
  id: z.string(), original_filename: z.string(), detected_content_type: z.string(), size_bytes: z.number(),
  scan_status: z.enum(["quarantined", "clean", "rejected"]),
});
const ticketMessageSchema = z.object({
  id: z.string(), author_type: z.enum(["tenant_user", "support_agent", "system"]), author_id: z.string(),
  author_label: z.string().optional(), body: z.string(), attachments: z.array(ticketAttachmentSchema), created_at: z.string(),
});
const ticketRatingSchema = z.object({
  score: z.number(), comment: z.string().optional(), created_at: z.string(), updated_at: z.string(),
});
const ticketSchema = z.object({
  id: z.string(), public_id: z.string(), tenant_id: z.string(), tenant_slug: z.string().optional(),
  requester_user_id: z.string(), requester_display_name: z.string().optional(), requester_email: z.string().optional(),
  subject: z.string(), category: ticketCategorySchema, priority: ticketPrioritySchema, status: ticketStatusSchema,
  assignee_id: z.string().optional(), last_message_author_type: z.string(), last_message_at: z.string(),
  first_support_response_at: z.string().optional(), resolved_at: z.string().optional(), closed_at: z.string().optional(),
  created_at: z.string(), updated_at: z.string(), version: z.number(), rating: ticketRatingSchema.optional(),
});
const ticketDetailSchema = z.object({ ticket: ticketSchema, messages: z.array(ticketMessageSchema) });
const ticketListSchema = z.object({ data: z.array(ticketSchema), next_cursor: z.string().optional() });

export function listTickets(query: URLSearchParams, signal?: AbortSignal) {
  return requestParsedJSON(ticketListSchema, `/api/support/tickets?${query.toString()}`, { signal });
}

export function getTicket(ticketID: string) {
  return requestParsedJSON(ticketDetailSchema, `/api/support/tickets/${encodeURIComponent(ticketID)}`);
}

export function createTicket(form: FormData, idempotencyKey: string) {
  return requestParsedJSON(ticketDetailSchema, "/api/support/tickets", {
    method: "POST", body: form, headers: { "Idempotency-Key": idempotencyKey },
  });
}

export function replyToTicket(ticketID: string, form: FormData, idempotencyKey: string) {
  return requestParsedJSON(ticketDetailSchema, `/api/support/tickets/${encodeURIComponent(ticketID)}/messages`, {
    method: "POST", body: form, headers: { "Idempotency-Key": idempotencyKey },
  });
}

export function rateTicket(ticketID: string, input: { score: number; comment: string }, idempotencyKey: string) {
  return requestParsedJSON(ticketSchema, `/api/support/tickets/${encodeURIComponent(ticketID)}/rating`, {
    method: "PUT", body: JSON.stringify(input), headers: { "Idempotency-Key": idempotencyKey },
  });
}
