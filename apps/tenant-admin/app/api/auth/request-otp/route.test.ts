import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

const normalizeIdentifierMock = vi.hoisted(() => vi.fn());
const consumeCaptchaMock = vi.hoisted(() => vi.fn());
const createOTPRequestMock = vi.hoisted(() => vi.fn());
const getAuthRepositoryMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/factory", () => ({
  getAuthRepository: getAuthRepositoryMock,
}));

vi.mock("@/lib/auth/debug-otp", () => ({
  debugOTPResponseField: (debugOTP?: string) =>
    debugOTP ? { debug_otp: debugOTP } : {},
}));

describe("auth/request-otp route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    normalizeIdentifierMock.mockImplementation((raw: string) => {
      const normalized = raw.trim().toLowerCase();
      if (!normalized.includes("@")) return null;
      return { channel: "email", normalized };
    });
    consumeCaptchaMock.mockResolvedValue({ ok: true });
    createOTPRequestMock.mockResolvedValue({
      request_id: "otp-request-1",
      channel: "email",
      destination_masked: "us***@example.com",
      expires_in_seconds: 300,
    });
    getAuthRepositoryMock.mockReturnValue({
      normalizeIdentifier: normalizeIdentifierMock,
      consumeCaptcha: consumeCaptchaMock,
      createOTPRequest: createOTPRequestMock,
    });
  });

  it("does not expose account existence when login OTP creation is denied", async () => {
    createOTPRequestMock.mockRejectedValueOnce(
      new Error(
        'Auth API error: 403 Forbidden - {"error":"You do not have access to this tenant."}',
      ),
    );

    const res = await POST(
      buildRequest({
        identifier: "missing@example.com",
        captcha_id: "captcha-missing",
        captcha_answer: "abcde",
        ip: "198.51.100.10",
      }),
    );
    const body = await res.json();

    expect(res.status).toBe(200);
    expect(body).toEqual({
      request_id: expect.any(String),
      channel: "email",
      destination_masked: "mi***@example.com",
      expires_in_seconds: 300,
    });
    expect(JSON.stringify(body)).not.toContain("access");
    expect(JSON.stringify(body)).not.toContain("not found");
  });

  it("rate limits repeated login OTP requests before creating more OTPs", async () => {
    const responses = [];

    for (let attempt = 1; attempt <= 6; attempt += 1) {
      responses.push(
        await POST(
          buildRequest({
            identifier: "ratelimited@example.com",
            captcha_id: `captcha-rate-${attempt}`,
            captcha_answer: "abcde",
            ip: "198.51.100.20",
          }),
        ),
      );
    }

    expect(responses.slice(0, 5).map((res) => res.status)).toEqual([
      200,
      200,
      200,
      200,
      200,
    ]);
    expect(responses[5]?.status).toBe(429);
    expect(responses[5]?.headers.get("Retry-After")).toMatch(/^\d+$/);
    expect(createOTPRequestMock).toHaveBeenCalledTimes(5);
  });
});

function buildRequest(input: {
  identifier: string;
  captcha_id: string;
  captcha_answer: string;
  ip: string;
}) {
  return new NextRequest("http://tenant1.example.com/api/auth/request-otp", {
    method: "POST",
    headers: {
      "content-type": "application/json",
      host: "tenant1.example.com",
      "x-forwarded-for": input.ip,
    },
    body: JSON.stringify({
      identifier: input.identifier,
      captcha_id: input.captcha_id,
      captcha_answer: input.captcha_answer,
    }),
  });
}
