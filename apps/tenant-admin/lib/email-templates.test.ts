import { describe, expect, it } from "vitest";

import {
  renderInvitationEmailTemplate,
  renderOTPEmailTemplate,
} from "./email-templates";

describe("email templates", () => {
  it("renders branded RTL invitation email with escaped tenant content", () => {
    const template = renderInvitationEmailTemplate({
      tenantName: `Acme <script>alert("x")</script>`,
      tenantSlug: "acme",
      invitationLink: "https://acme.godarai.ir/invite/?invitetoken=abc",
      logoURL: "https://acme.godarai.ir/brand/mini_logo.png",
      workspaceURL: "https://acme.godarai.ir",
    });

    expect(template.subject).toContain("گدارAI");
    expect(template.html).toContain(`dir="rtl"`);
    expect(template.html).toContain("https://acme.godarai.ir/brand/mini_logo.png");
    expect(template.html).toContain("پذیرش دعوت‌نامه");
    expect(template.html).toContain("&lt;script&gt;");
    expect(template.html).not.toContain("<script>alert");
    expect(template.html).toContain('role="presentation"');
    expect(template.html).toContain('bgcolor="#0238e1"');
    expectUnsupportedEmailStylesToBeAbsent(template.html);
  });

  it("renders OTP email with an LTR code block and Persian text fallback", () => {
    const template = renderOTPEmailTemplate({
      tenantSlug: "acme",
      otpCode: "123456",
      expiresInMinutes: 5,
      workspaceURL: "https://acme.godarai.ir",
    });

    expect(template.subject).toBe("کد ورود یک‌بارمصرف گدارAI");
    expect(template.text).toContain("کد ورود یک‌بارمصرف شما");
    expect(template.text).toContain("123456");
    expect(template.html).toContain(`dir="ltr"`);
    expect(template.html).toContain("123456");
    expect(template.html).toContain("برای امنیت حساب");
    expect(template.html).toContain('bgcolor="#f8fbff"');
    expectUnsupportedEmailStylesToBeAbsent(template.html);
  });
});

function expectUnsupportedEmailStylesToBeAbsent(html: string) {
  expect(html).not.toMatch(
    /box-shadow|opacity|max-height|overflow|word-break|rgba\(|linear-gradient|border-radius|text-decoration|background:/i,
  );
}
