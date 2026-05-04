import nodemailer from "nodemailer";

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

export async function sendOnboardingOTPEmail(input: {
  to: string;
  otpCode: string;
  expiresInMinutes: number;
  tenantSlug: string;
}) {
  const subject = "Shahrah Tenant Onboarding OTP";
  const text = [
    `Your one-time onboarding code for tenant "${input.tenantSlug}" is: ${input.otpCode}`,
    `This code expires in ${input.expiresInMinutes} minutes.`,
    "If you did not request this tenant, ignore this email.",
  ].join("\n");

  const html = `
<div style="font-family: sans-serif; line-height: 1.6;">
  <h2>Shahrah Tenant Onboarding OTP</h2>
  <p>Your one-time onboarding code for tenant <strong>${input.tenantSlug}</strong> is:</p>
  <p style="font-size: 24px; font-weight: bold; letter-spacing: 3px;">${input.otpCode}</p>
  <p>This code expires in ${input.expiresInMinutes} minutes.</p>
  <p>If you did not request this tenant, ignore this email.</p>
</div>`;

  await sendEmail({
    to: input.to,
    subject,
    text,
    html,
  });
}
