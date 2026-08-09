import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

const userProfileSchema = z.object({
  id: z.string(),
  email: z.string().optional(),
  mobile: z.string().optional(),
  socialProviders: z.array(z.string()).optional(),
  role: z.string(),
});

const socialLinkSchema = z.object({ redirect_url: z.string().url() });
const otpRequestSchema = z.object({
  request_id: z.string(),
  destination_masked: z.string(),
  debug_otp: z.string().optional(),
});
const verifyContactSchema = z.object({ user: userProfileSchema });

export type UserProfile = z.infer<typeof userProfileSchema>;

export function createSocialLink(provider: "google" | "github" | "microsoft") {
  return requestParsedJSON(socialLinkSchema, "/api/auth/social/link-ticket", {
    method: "POST",
    body: JSON.stringify({ provider }),
  });
}

export function requestContactChange(field: "email" | "mobile", newValue: string) {
  return requestParsedJSON(otpRequestSchema, "/api/auth/profile/request-contact-change", {
    method: "POST",
    body: JSON.stringify({ field, new_value: newValue }),
  });
}

export function verifyContactChange(input: {
  field: "email" | "mobile";
  requestID: string;
  otpCode: string;
}) {
  return requestParsedJSON(verifyContactSchema, "/api/auth/profile/verify-contact-change", {
    method: "POST",
    body: JSON.stringify({
      field: input.field,
      request_id: input.requestID,
      otp_code: input.otpCode,
    }),
  });
}
