import nodemailer from "nodemailer";
import { Resend } from "resend";

type SendEmailInput = {
  to: string;
  subject: string;
  text: string;
  html?: string;
  replyTo?: string;
  tags?: Array<{ name: string; value: string }>;
};

let cachedTransporter: nodemailer.Transporter | null = null;
let cachedResendClient: Resend | null = null;

function isProductionRuntime() {
  return process.env.NODE_ENV === "production";
}

function getSMTPConfig() {
  return {
    host: process.env.SMTP_HOST ?? "127.0.0.1",
    port: Number(process.env.SMTP_PORT ?? "1025"),
    secure: (process.env.SMTP_SECURE ?? "false").toLowerCase() === "true",
    user: process.env.SMTP_USER?.trim() || undefined,
    pass: process.env.SMTP_PASS?.trim() || undefined,
    from: process.env.SMTP_FROM?.trim() || "no-reply@shahrah.local",
  };
}

function getResendConfig() {
  const apiKey = process.env.RESEND_API_KEY?.trim() || "";
  const fromEmail = process.env.RESEND_FROM_EMAIL?.trim() || "";
  const fromName = process.env.RESEND_FROM_NAME?.trim() || "GodarAI";
  const replyTo = process.env.RESEND_REPLY_TO?.trim() || undefined;

  if (isProductionRuntime() && (!apiKey || !fromEmail)) {
    throw new Error(
      "RESEND_API_KEY and RESEND_FROM_EMAIL are required when NODE_ENV=production.",
    );
  }

  return { apiKey, fromEmail, fromName, replyTo };
}

function getTransporter() {
  if (cachedTransporter) {
    return cachedTransporter;
  }

  const config = getSMTPConfig();
  cachedTransporter = nodemailer.createTransport({
    host: config.host,
    port: config.port,
    secure: config.secure,
    auth: config.user && config.pass ? { user: config.user, pass: config.pass } : undefined,
  });
  return cachedTransporter;
}

function getResendClient() {
  if (cachedResendClient) {
    return cachedResendClient;
  }

  const config = getResendConfig();
  cachedResendClient = new Resend(config.apiKey);
  return cachedResendClient;
}

export function isEmailSendingEnabled() {
  return (process.env.EMAIL_SENDING_ENABLED ?? "true").toLowerCase() === "true";
}

export async function sendEmail(input: SendEmailInput) {
  if (!isEmailSendingEnabled()) {
    return;
  }

  if (isProductionRuntime()) {
    const config = getResendConfig();
    const resend = getResendClient();
    const { error } = await resend.emails.send({
      from: `${config.fromName} <${config.fromEmail}>`,
      to: [input.to],
      subject: input.subject,
      text: input.text,
      html: input.html,
      replyTo: input.replyTo ? [input.replyTo] : config.replyTo ? [config.replyTo] : undefined,
      tags: input.tags,
    });

    if (error) {
      throw new Error(`Resend send failed: ${error.message}`);
    }
    return;
  }

  const config = getSMTPConfig();
  const transporter = getTransporter();
  await transporter.sendMail({
    from: config.from,
    to: input.to,
    subject: input.subject,
    text: input.text,
    html: input.html,
    replyTo: input.replyTo,
  });
}

export async function sendContactEmail(input: {
  name: string;
  email: string;
  subject: string;
  message: string;
}) {
  const text = [
    `Name: ${input.name}`,
    `Email: ${input.email}`,
    `Subject: ${input.subject}`,
    "",
    input.message,
  ].join("\n");

  const html = `
<div style="font-family: Tahoma, Arial, sans-serif; line-height: 1.8; color: #111827;">
  <h2>${escapeHTML(input.subject)}</h2>
  <p><strong>Name:</strong> ${escapeHTML(input.name)}</p>
  <p><strong>Email:</strong> ${escapeHTML(input.email)}</p>
  <hr style="border: 0; border-top: 1px solid #e5e7eb; margin: 20px 0;" />
  <p style="white-space: pre-wrap;">${escapeHTML(input.message)}</p>
</div>`;

  await sendEmail({
    to: "support@godarai.ir",
    subject: `[Website contact] ${input.subject}`,
    text,
    html,
    replyTo: input.email,
    tags: [
      { name: "app", value: "website" },
      { name: "flow", value: "contact" },
    ],
  });
}

function escapeHTML(value: string) {
  return value.replace(/[&<>"']/g, (character) => {
    const entities: Record<string, string> = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    };
    return entities[character];
  });
}

export async function sendOnboardingOTPEmail(input: {
  to: string;
  otpCode: string;
  expiresInMinutes: number;
  tenantSlug: string;
}) {
  const subject = "کد تایید فعال‌سازی فضای کار گُدار AI";
  const text = [
    `کد تایید فعال‌سازی فضای کار «${input.tenantSlug}»: ${input.otpCode}`,
    `این کد تا ${input.expiresInMinutes} دقیقه معتبر است.`,
    "اگر شما این درخواست را ثبت نکرده‌اید، این ایمیل را نادیده بگیرید.",
  ].join("\n");

  const html = `
<div dir="rtl" style="font-family: Tahoma, Arial, sans-serif; line-height: 1.8; color: #111827;">
  <h2>کد تایید فعال‌سازی فضای کار گُدار AI</h2>
  <p>کد تایید یک‌بارمصرف برای فضای کار <strong>${input.tenantSlug}</strong>:</p>
  <p style="font-size: 24px; font-weight: bold; letter-spacing: 3px;">${input.otpCode}</p>
  <p>این کد تا ${input.expiresInMinutes} دقیقه معتبر است.</p>
  <p>اگر شما این درخواست را ثبت نکرده‌اید، این ایمیل را نادیده بگیرید.</p>
</div>`;

  await sendEmail({
    to: input.to,
    subject,
    text,
    html,
    tags: [
      { name: "app", value: "website" },
      { name: "flow", value: "onboarding-otp" },
      { name: "tenant", value: input.tenantSlug },
    ],
  });
}
