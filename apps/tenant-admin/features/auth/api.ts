import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const captchaChallengeSchema = z.object({
  captcha_id: z.string(),
  image_data_url: z.string(),
  expires_in_seconds: z.number(),
});
const otpRequestResponseSchema = z.object({
  request_id: z.string(),
  channel: z.enum(["email", "mobile"]),
  destination_masked: z.string(),
  expires_in_seconds: z.number(),
  debug_otp: z.string().optional(),
});

export type CaptchaChallenge = z.infer<typeof captchaChallengeSchema>;
export type OTPRequestResponse = z.infer<typeof otpRequestResponseSchema>;

export function refreshLoginCaptcha() {
  return requestParsedJSON(captchaChallengeSchema, "/api/auth/captcha");
}

export function requestLoginOTP(input: {
  identifier: string;
  captchaID: string;
  captchaAnswer: string;
}) {
  return requestParsedJSON(otpRequestResponseSchema, "/api/auth/request-otp", {
    method: "POST",
    body: JSON.stringify({
      identifier: input.identifier,
      captcha_id: input.captchaID,
      captcha_answer: input.captchaAnswer,
    }),
  });
}

export async function verifyLoginOTP(requestID: string, otpCode: string): Promise<void> {
  await requestJSON("/api/auth/verify-otp", {
    method: "POST",
    body: JSON.stringify({ request_id: requestID, otp_code: otpCode }),
  });
}

export async function acceptInvitation(token: string): Promise<void> {
  await requestJSON("/api/users/invitations/accept", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
}
