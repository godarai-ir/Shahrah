import nodemailer from "nodemailer";

import { renderInvitationEmailTemplate } from "@/lib/email-templates";

type SendEmailInput = {
  to: string;
  subject: string;
  text: string;
  html?: string;
};

let cachedTransporter: nodemailer.Transporter | null = null;

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

export function isEmailSendingEnabled() {
  return (process.env.EMAIL_SENDING_ENABLED ?? "true").toLowerCase() === "true";
}

export async function sendEmail(input: SendEmailInput) {
  if (!isEmailSendingEnabled()) {
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
  });
}

export async function sendOTPEmail(input: {
  to: string;
  otpCode: string;
  expiresInMinutes: number;
  tenantSlug: string;
}) {
  const subject = "Shahrah OTP Code";
  const text = [
    `Your one-time code for tenant "${input.tenantSlug}" is: ${input.otpCode}`,
    `This code expires in ${input.expiresInMinutes} minutes.`,
    "If you did not request this code, ignore this email.",
  ].join("\n");

  const html = `
<div style="font-family: sans-serif; line-height: 1.6;">
  <h2>Shahrah OTP Code</h2>
  <p>Your one-time code for tenant <strong>${input.tenantSlug}</strong> is:</p>
  <p style="font-size: 24px; font-weight: bold; letter-spacing: 3px;">${input.otpCode}</p>
  <p>This code expires in ${input.expiresInMinutes} minutes.</p>
  <p>If you did not request this code, ignore this email.</p>
</div>`;

  await sendEmail({
    to: input.to,
    subject,
    text,
    html,
  });
}

export async function sendInvitationEmail(input: {
  to: string;
  link: string;
  tenantSlug: string;
  tenantName?: string;
}) {
  const template = renderInvitationEmailTemplate({
    tenantName: input.tenantName?.trim() || input.tenantSlug,
    tenantSlug: input.tenantSlug,
    invitationLink: input.link,
  });

  await sendEmail({
    to: input.to,
    subject: template.subject,
    text: template.text,
    html: template.html,
  });
}
