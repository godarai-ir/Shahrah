import { describe, expect, it } from "vitest";

import { APIClientError } from "@/lib/http/api-client";

import { personalUsagePercent, previewSeatChange, seatErrorMessage } from "./seats";

describe("previewSeatChange", () => {
  it.each([
    [{ seat: "unassigned", free_seat_used: false }, "free", { allowed: true, timing: "immediate", needsPaidSeat: false }],
    [{ seat: "unassigned", free_seat_used: true }, "free", { allowed: false, reason: "free_used" }],
    [{ seat: "free", free_seat_used: true }, "pro", { allowed: true, timing: "immediate", needsPaidSeat: true }],
    [{ seat: "pro", free_seat_used: true }, "max", { allowed: true, timing: "immediate", needsPaidSeat: true }],
    [{ seat: "max", free_seat_used: true }, "pro", { allowed: true, timing: "period_end", needsPaidSeat: false }],
    [{ seat: "pro", free_seat_used: true }, "unassigned", { allowed: true, timing: "period_end", needsPaidSeat: false }],
    [{ seat: "free", free_seat_used: true }, "unassigned", { allowed: true, timing: "immediate", needsPaidSeat: false }],
    [{ seat: "pro", free_seat_used: true }, "free", { allowed: false, reason: "back_to_free" }],
    [{ seat: "max", free_seat_used: true }, "max", { allowed: false, reason: "same" }],
  ] as const)("%o to %s", (member, target, expected) => {
    expect(previewSeatChange(member, target)).toEqual(expected);
  });
});

describe("seat helpers", () => {
  it("measures personal credit use", () => {
    expect(personalUsagePercent({ personal_granted_credits: "2000.00", personal_remaining_credits: "500.00" })).toBe(75);
    expect(personalUsagePercent({ personal_granted_credits: "0.00", personal_remaining_credits: "0.00" })).toBe(0);
  });

  it("translates known error codes and keeps the server text otherwise", () => {
    const t = (key: string) => `t:${key}`;
    expect(seatErrorMessage(t, new APIClientError("x", { status: 409, code: "free_seat_used", payload: {} }))).toBe("t:errors.free_seat_used");
    expect(seatErrorMessage(t, new APIClientError("Server said no", { status: 500, code: "other", payload: {} }))).toBe("Server said no");
  });
});
