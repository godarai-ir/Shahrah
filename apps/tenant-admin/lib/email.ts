import nodemailer from "nodemailer";
import { Resend } from "resend";

import {
  renderInvitationEmailTemplate,
  renderOTPEmailTemplate,
} from "@/lib/email-templates";

type SendEmailInput = {
  to: string;
  subject: string;
  text: string;
  html?: string;
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
      replyTo: config.replyTo ? [config.replyTo] : undefined,
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
  });
}

export async function sendOTPEmail(input: {
  to: string;
  otpCode: string;
  expiresInMinutes: number;
  tenantSlug: string;
}) {
  const workspaceURL = resolveWorkspaceURL(input.tenantSlug);
  const template = renderOTPEmailTemplate({
    tenantSlug: input.tenantSlug,
    otpCode: input.otpCode,
    expiresInMinutes: input.expiresInMinutes,
    workspaceURL,
    logoURL: resolveEmailLogoURL(workspaceURL),
  });

  await sendEmail({
    to: input.to,
    subject: template.subject,
    text: template.text,
    html: template.html,
    tags: [
      { name: "app", value: "tenant-admin" },
      { name: "flow", value: "login-otp" },
      { name: "tenant", value: input.tenantSlug },
    ],
  });
}

export async function sendInvitationEmail(input: {
  to: string;
  link: string;
  tenantSlug: string;
  tenantName?: string;
}) {
  const workspaceURL = resolveWorkspaceURL(input.tenantSlug, input.link);
  const template = renderInvitationEmailTemplate({
    tenantName: input.tenantName?.trim() || input.tenantSlug,
    tenantSlug: input.tenantSlug,
    invitationLink: input.link,
    workspaceURL,
    logoURL: resolveEmailLogoURL(workspaceURL),
  });

  await sendEmail({
    to: input.to,
    subject: template.subject,
    text: template.text,
    html: template.html,
    tags: [
      { name: "app", value: "tenant-admin" },
      { name: "flow", value: "invitation" },
      { name: "tenant", value: input.tenantSlug },
    ],
  });
}

function resolveEmailLogoURL(baseURL?: string) {
  const configured =
    process.env.GODARAI_EMAIL_LOGO_URL?.trim() ||
    process.env.SHAHRAH_EMAIL_LOGO_URL?.trim();
  if (configured) {
    return configured;
  }

  return resolvePublicAssetURL("/brand/mini_logo.png", baseURL);
}

function resolveWorkspaceURL(tenantSlug: string, fallbackURL?: string) {
  const fallbackOrigin = originFromURL(fallbackURL);
  if (fallbackOrigin) {
    return fallbackOrigin;
  }

  const configuredAppURL =
    process.env.NEXT_PUBLIC_TENANT_ADMIN_URL?.trim() ||
    process.env.SHAHRAH_TENANT_ADMIN_URL?.trim();
  if (configuredAppURL) {
    return configuredAppURL;
  }

  const baseDomain =
    process.env.SHAHRAH_TENANT_BASE_DOMAIN?.trim() ||
    process.env.TENANT_BASE_DOMAIN?.trim() ||
    "godarai.ir";
  const normalizedDomain = baseDomain.replace(/^https?:\/\//, "").replace(/\/.*$/, "");
  const normalizedTenant = tenantSlug.trim().toLowerCase();

  return normalizedTenant
    ? `https://${normalizedTenant}.${normalizedDomain}`
    : `https://${normalizedDomain}`;
}

function resolvePublicAssetURL(path: string, baseURL?: string) {
  if (!baseURL) {
    return undefined;
  }

  try {
    return new URL(path, baseURL).toString();
  } catch {
    return undefined;
  }
}

function originFromURL(value?: string) {
  if (!value) {
    return undefined;
  }

  try {
    return new URL(value).origin;
  } catch {
    return undefined;
  }
}
