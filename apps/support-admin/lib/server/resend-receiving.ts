import 'server-only';

import { Resend } from 'resend';

let cachedClient: Resend | null = null;

export class ResendReceivingConfigurationError extends Error {
  constructor() {
    super('RESEND_RECEIVING_API_KEY or RESEND_API_KEY is required');
    this.name = 'ResendReceivingConfigurationError';
  }
}

export function getResendReceivingClient(): Resend {
  if (cachedClient) return cachedClient;

  const apiKey = process.env.RESEND_RECEIVING_API_KEY?.trim()
    || process.env.RESEND_API_KEY?.trim();
  if (!apiKey) throw new ResendReceivingConfigurationError();

  cachedClient = new Resend(apiKey);
  return cachedClient;
}
