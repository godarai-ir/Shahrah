export type InboundEmailAttachment = {
  id: string;
  filename: string | null;
  size: number;
  content_type: string;
  content_id: string | null;
  content_disposition: string | null;
};

export type InboundEmailSummary = {
  id: string;
  to: string[];
  from: string;
  created_at: string;
  subject: string;
  bcc: string[] | null;
  cc: string[] | null;
  reply_to: string[] | null;
  received_for: string[];
  message_id: string;
  attachments: InboundEmailAttachment[];
};

export type InboundEmailDetail = InboundEmailSummary & {
  object: 'email';
  html: string | null;
  text: string | null;
  headers: Record<string, string> | null;
  raw?: {
    download_url: string;
    expires_at: string;
  } | null;
};

export type InboundEmailList = {
  object: 'list';
  has_more: boolean;
  data: InboundEmailSummary[];
};

export async function resendReceivingRequest<T>(path: string): Promise<T> {
  const response = await fetch(`/api/admin/resend/receiving${path}`, { cache: 'no-store' });
  const payload = await response.json().catch(() => ({})) as {
    error?: string;
  };

  if (!response.ok) {
    throw new Error(payload.error || `Request failed (${response.status})`);
  }

  return payload as T;
}
