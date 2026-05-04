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
  const subject = "کد تایید فعال‌سازی تننت شاهراه";
  const text = [
    `کد تایید فعال‌سازی تننت «${input.tenantSlug}»: ${input.otpCode}`,
    `این کد تا ${input.expiresInMinutes} دقیقه معتبر است.`,
    "اگر شما این درخواست را ثبت نکرده‌اید، این ایمیل را نادیده بگیرید.",
  ].join("\n");

  const html = `
<div dir="rtl" style="font-family: Tahoma, Arial, sans-serif; line-height: 1.8; color: #111827;">
  <h2>کد تایید فعال‌سازی تننت شاهراه</h2>
  <p>کد تایید یک‌بارمصرف برای تننت <strong>${input.tenantSlug}</strong>:</p>
  <p style="font-size: 24px; font-weight: bold; letter-spacing: 3px;">${input.otpCode}</p>
  <p>این کد تا ${input.expiresInMinutes} دقیقه معتبر است.</p>
  <p>اگر شما این درخواست را ثبت نکرده‌اید، این ایمیل را نادیده بگیرید.</p>
</div>`;

  await sendEmail({
    to: input.to,
    subject,
    text,
    html,
  });
}
