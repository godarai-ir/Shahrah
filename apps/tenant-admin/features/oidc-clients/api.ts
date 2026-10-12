import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

import type { OIDCClientInput } from "./input";

const oidcClientSchema = z.object({
  id: z.string(),
  client_id: z.string(),
  name: z.string(),
  redirect_uris: z.array(z.string()).nullish().transform((value) => value ?? []),
  post_logout_redirect_uris: z.array(z.string()).nullish().transform((value) => value ?? []),
  active: z.boolean(),
  credit_product: z.string().optional(),
  created_by: z.string().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
  secret_rotated_at: z.string().optional(),
});

const oidcClientWithSecretSchema = oidcClientSchema.extend({ client_secret: z.string() });
const oidcClientListSchema = z.object({ data: z.array(oidcClientSchema).optional() });

export type OIDCClient = z.infer<typeof oidcClientSchema>;
export type OIDCClientWithSecret = z.infer<typeof oidcClientWithSecretSchema>;

export async function listOIDCClients(): Promise<OIDCClient[]> {
  const payload = await requestParsedJSON(oidcClientListSchema, "/api/oidc-clients");
  return payload.data ?? [];
}

export function createOIDCClient(input: OIDCClientInput) {
  return requestParsedJSON(oidcClientWithSecretSchema, "/api/oidc-clients", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateOIDCClient(id: string, input: OIDCClientInput) {
  return requestParsedJSON(oidcClientSchema, `/api/oidc-clients/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function rotateOIDCClientSecret(id: string) {
  return requestParsedJSON(oidcClientWithSecretSchema, `/api/oidc-clients/${encodeURIComponent(id)}/rotate-secret`, {
    method: "POST",
  });
}

export async function deleteOIDCClient(id: string): Promise<void> {
  await requestJSON(`/api/oidc-clients/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export function parseOIDCClientListPayload(payload: unknown): OIDCClient[] {
  const parsed = oidcClientListSchema.safeParse(payload);
  return parsed.success ? parsed.data.data ?? [] : [];
}
