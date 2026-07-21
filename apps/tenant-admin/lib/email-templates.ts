type InvitationTemplateInput = {
  tenantName: string;
  tenantSlug: string;
  invitationLink: string;
  logoURL?: string;
  workspaceURL?: string;
};

type OTPTemplateInput = {
  tenantSlug: string;
  otpCode: string;
  expiresInMinutes: number;
  logoURL?: string;
  workspaceURL?: string;
};

type BrandedEmailInput = {
  subject: string;
  preheader: string;
  eyebrow: string;
  title: string;
  intro: string;
  contentHTML: string;
  logoURL?: string;
  primaryAction?: {
    label: string;
    href: string;
  };
  footerDetail?: string;
};

const brandName = "گدارAI";
const productURL = "https://godarai.ir";
const supportEmail = "support@godarai.ir";

export function renderInvitationEmailTemplate(input: InvitationTemplateInput) {
  const tenantName = normalizeTextLine(input.tenantName.trim() || input.tenantSlug);
  const workspaceLabel = "فضای کاری";
  const workspaceURL = input.workspaceURL ?? originFromURL(input.invitationLink);
  const safeInvitationLink = normalizeHTTPURL(input.invitationLink);
  const safeWorkspaceURL = normalizeHTTPURL(workspaceURL);
  const subject = normalizeTextLine(`دعوت به فضای کاری ${brandName} - ${tenantName}`);
  const preheader = normalizeTextLine(`برای پیوستن به فضای کاری ${tenantName} در ${brandName} دعوت شده‌اید.`);

  const text = [
    `سلام،`,
    "",
    `شما برای پیوستن به فضای کاری «${tenantName}» در ${brandName} دعوت شده‌اید.`,
    "برای پذیرش دعوت‌نامه و فعال‌سازی دسترسی، از لینک زیر استفاده کنید:",
    input.invitationLink,
    "",
    safeWorkspaceURL ? `${workspaceLabel}: ${safeWorkspaceURL}` : "",
    `وب‌سایت ${brandName}: ${productURL}`,
    "",
    "اگر با این دعوت‌نامه آشنا نیستید، می‌توانید این ایمیل را نادیده بگیرید.",
    "",
    `تیم ${brandName}`,
    supportEmail,
  ].filter(Boolean).join("\n");

  const contentHTML = `
    <p style="${paragraphStyle()}">
      شما برای پیوستن به فضای کاری <strong>${escapeHTML(tenantName)}</strong> در ${brandName} دعوت شده‌اید.
    </p>
    <p style="${paragraphStyle()}">
      با پذیرش این دعوت‌نامه، دسترسی شما به پنل مدیریت و ابزارهای مرتبط با همین فضای کاری فعال می‌شود.
    </p>
    ${safeWorkspaceURL ? renderInfoRow(workspaceLabel, safeWorkspaceURL) : ""}
    ${safeInvitationLink ? renderPlainLink("لینک دعوت‌نامه", safeInvitationLink) : ""}
    ${renderNotice("اگر انتظار دریافت این دعوت‌نامه را نداشتید، نیازی به انجام کاری نیست و می‌توانید این پیام را نادیده بگیرید.")}
  `;

  return {
    subject,
    text,
    html: renderBrandedEmail({
      subject,
      preheader,
      eyebrow: "دعوت‌نامه عضویت",
      title: `به ${tenantName} بپیوندید`,
      intro: `دعوت‌نامه شما برای ورود به ${brandName} آماده است.`,
      logoURL: input.logoURL,
      primaryAction: safeInvitationLink
        ? { label: "پذیرش دعوت‌نامه", href: safeInvitationLink }
        : undefined,
      contentHTML,
      footerDetail: "این پیام برای مدیریت دسترسی کاربران در پنل tenant-admin ارسال شده است.",
    }),
  };
}

export function renderOTPEmailTemplate(input: OTPTemplateInput) {
  const tenantSlug = normalizeTextLine(input.tenantSlug);
  const subject = `کد ورود یک‌بارمصرف ${brandName}`;
  const preheader = `کد ورود شما به فضای کاری ${tenantSlug} تا ${input.expiresInMinutes} دقیقه معتبر است.`;
  const safeWorkspaceURL = normalizeHTTPURL(input.workspaceURL);

  const text = [
    `کد ورود یک‌بارمصرف شما برای فضای کاری «${tenantSlug}»: ${input.otpCode}`,
    `این کد تا ${input.expiresInMinutes} دقیقه معتبر است.`,
    "برای امنیت حساب، این کد را در اختیار هیچ‌کس قرار ندهید.",
    safeWorkspaceURL ? `ورود به فضای کاری: ${safeWorkspaceURL}` : "",
    "",
    "اگر شما درخواست ورود نداده‌اید، این ایمیل را نادیده بگیرید.",
    "",
    `تیم ${brandName}`,
    supportEmail,
  ].filter(Boolean).join("\n");

  const contentHTML = `
    <p style="${paragraphStyle()}">
      از کد زیر برای ورود به فضای کاری <strong dir="ltr">${escapeHTML(tenantSlug)}</strong> استفاده کنید.
    </p>
    ${renderOTPCode(input.otpCode)}
    ${renderInfoRow("اعتبار کد", `${input.expiresInMinutes} دقیقه`)}
    ${safeWorkspaceURL ? renderPlainLink("ورود به فضای کاری", safeWorkspaceURL) : ""}
    ${renderNotice("برای امنیت حساب، این کد را در اختیار هیچ‌کس قرار ندهید. اگر شما درخواست ورود نداده‌اید، این پیام را نادیده بگیرید.")}
  `;

  return {
    subject,
    text,
    html: renderBrandedEmail({
      subject,
      preheader,
      eyebrow: "ورود امن",
      title: "کد ورود یک‌بارمصرف",
      intro: "این کد فقط برای همین درخواست ورود صادر شده است.",
      logoURL: input.logoURL,
      contentHTML,
      footerDetail: "این پیام به درخواست شما برای ورود به پنل کاربری گدارAI ارسال شده است.",
    }),
  };
}

function renderBrandedEmail(input: BrandedEmailInput) {
  const logoURL = normalizeHTTPURL(input.logoURL);
  const actionHref = normalizeHTTPURL(input.primaryAction?.href);
  const action = input.primaryAction && actionHref
    ? { label: input.primaryAction.label, href: actionHref }
    : undefined;

  return `<!doctype html>
<html lang="fa" dir="rtl">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>${escapeHTML(input.subject)}</title>
  </head>
  <body dir="rtl" bgcolor="#f3f7fb" style="margin: 0; padding: 0; background-color: #f3f7fb; color: #0b1220; font-family: Tahoma, Arial, sans-serif;">
    <div aria-hidden="true" style="display: none; color: #f3f7fb; font-size: 1px; line-height: 1px;">
      ${escapeHTML(input.preheader)}
    </div>
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" bgcolor="#f3f7fb" style="width: 100%; border-collapse: collapse; background-color: #f3f7fb;">
      <tr>
        <td align="center" style="padding: 32px 16px;">
          <table role="presentation" width="640" cellspacing="0" cellpadding="0" border="0" style="width: 100%; max-width: 640px; border-collapse: collapse;">
            <tr>
              <td style="padding: 0 0 14px; text-align: right;">
                ${renderLogo(logoURL)}
              </td>
            </tr>
            <tr>
              <td bgcolor="#ffffff" style="border: 1px solid #d6e2f0; background-color: #ffffff;">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width: 100%; border-collapse: collapse;">
                  <tr>
                    <td width="50%" height="5" bgcolor="#04b8a9" style="height: 5px; background-color: #04b8a9; font-size: 0; line-height: 0;">&nbsp;</td>
                    <td width="50%" height="5" bgcolor="#0238e1" style="height: 5px; background-color: #0238e1; font-size: 0; line-height: 0;">&nbsp;</td>
                  </tr>
                  <tr>
                    <td colspan="2" style="padding: 34px 34px 28px;">
                      <p style="margin: 0 0 10px; color: #0238e1; font-size: 13px; font-weight: 700;">${escapeHTML(input.eyebrow)}</p>
                      <h1 style="margin: 0; color: #0b1220; font-size: 25px; line-height: 1.6; font-weight: 800;">${escapeHTML(input.title)}</h1>
                      <p style="margin: 10px 0 0; color: #475569; font-size: 15px; line-height: 1.9;">${escapeHTML(input.intro)}</p>
                      <div style="margin-top: 26px;">
                        ${input.contentHTML}
                      </div>
                      ${action ? renderPrimaryAction(action.label, action.href) : ""}
                    </td>
                  </tr>
                  <tr>
                    <td colspan="2" bgcolor="#fbfdff" style="border-top: 1px solid #e2e8f0; padding: 22px 34px 28px; background-color: #fbfdff;">
                      <p style="margin: 0; color: #64748b; font-size: 12px; line-height: 1.9;">
                        ${escapeHTML(input.footerDetail ?? `این پیام توسط ${brandName} ارسال شده است.`)}
                      </p>
                      <p style="margin: 8px 0 0; color: #64748b; font-size: 12px; line-height: 1.9;">
                        <a href="${productURL}" style="color: #0238e1;">${productURL}</a>
                        <span style="padding: 0 6px; color: #cbd5e1;">|</span>
                        <a href="mailto:${supportEmail}" style="color: #0238e1;">${supportEmail}</a>
                      </p>
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`;
}

function renderLogo(logoURL: string | null) {
  if (logoURL) {
    return `<img src="${escapeAttribute(logoURL)}" width="52" height="52" alt="${brandName}" style="display: block; width: 52px; height: 52px; border: 0;" />`;
  }

  return `<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="border-collapse: collapse;">
    <tr>
      <td bgcolor="#ffffff" style="border: 1px solid #d6e2f0; background-color: #ffffff; padding: 10px 14px; color: #0238e1; font-size: 18px; font-weight: 800;">${brandName}</td>
    </tr>
  </table>`;
}

function renderPrimaryAction(label: string, href: string) {
  return `<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin-top: 28px; border-collapse: collapse;">
    <tr>
      <td bgcolor="#0238e1" style="background-color: #0238e1;">
        <a href="${escapeAttribute(href)}" style="display: inline-block; padding: 13px 22px; color: #ffffff; font-size: 15px; font-weight: 700; line-height: 1.4;">${escapeHTML(label)}</a>
      </td>
    </tr>
  </table>`;
}

function renderPlainLink(label: string, href: string) {
  return `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width: 100%; margin-top: 18px; border-collapse: collapse;">
    <tr>
      <td bgcolor="#f8fbff" style="border: 1px solid #d6e2f0; background-color: #f8fbff; padding: 14px 16px;">
        <p style="margin: 0 0 6px; color: #64748b; font-size: 12px; font-weight: 700;">${escapeHTML(label)}</p>
        <a dir="ltr" href="${escapeAttribute(href)}" style="display: block; color: #0238e1; font-family: Arial, sans-serif; font-size: 13px; line-height: 1.7; text-align: left;">${renderBreakableURL(href)}</a>
      </td>
    </tr>
  </table>`;
}

function renderInfoRow(label: string, value: string) {
  return `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width: 100%; margin-top: 16px; border-collapse: collapse;">
    <tr>
      <td bgcolor="#ffffff" style="border: 1px solid #d6e2f0; background-color: #ffffff; padding: 13px 16px;">
        <p style="margin: 0; color: #64748b; font-size: 12px; line-height: 1.7;">${escapeHTML(label)}</p>
        <p style="margin: 3px 0 0; color: #0b1220; font-size: 14px; font-weight: 700; line-height: 1.7;">${escapeHTML(value)}</p>
      </td>
    </tr>
  </table>`;
}

function renderNotice(value: string) {
  return `<p style="margin: 22px 0 0; border-right: 3px solid #04b8a9; padding: 10px 14px 10px 0; color: #475569; font-size: 13px; line-height: 1.9; background-color: #f8fbff;">
    ${escapeHTML(value)}
  </p>`;
}

function renderOTPCode(value: string) {
  return `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width: 100%; margin: 22px 0; border-collapse: collapse;">
    <tr>
      <td dir="ltr" align="center" bgcolor="#f8fbff" style="border: 1px solid #d6e2f0; background-color: #f8fbff; padding: 18px 20px; color: #0b1220; font-family: Consolas, 'Liberation Mono', Menlo, monospace; font-size: 30px; font-weight: 800; letter-spacing: 8px; text-align: center;">
        ${escapeHTML(value)}
      </td>
    </tr>
  </table>`;
}

function renderBreakableURL(value: string) {
  return value.replace(/[\/?&=._-]/g, (character) => `${escapeHTML(character)}&#8203;`).replace(/[^\/?&=._-]+/g, escapeHTML);
}

function paragraphStyle() {
  return "margin: 0 0 14px; color: #334155; font-size: 15px; line-height: 2;";
}

function normalizeTextLine(value: string) {
  return value.replace(/[\r\n]+/g, " ").replace(/\s+/g, " ").trim();
}

function originFromURL(value: string) {
  try {
    return new URL(value).origin;
  } catch {
    return undefined;
  }
}

function normalizeHTTPURL(value: string | undefined) {
  if (!value) {
    return null;
  }

  try {
    const url = new URL(value);
    if (url.protocol !== "http:" && url.protocol !== "https:") {
      return null;
    }
    return url.toString();
  } catch {
    return null;
  }
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

function escapeAttribute(value: string) {
  return escapeHTML(value);
}
