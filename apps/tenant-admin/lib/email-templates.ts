type InvitationTemplateInput = {
  tenantName: string;
  tenantSlug: string;
  invitationLink: string;
};

const invitationSubjectTemplate = "دعوت به پلتفرم گُدارAI - {{tenant_name}}";

const invitationTextTemplate = `کاربر گرامی،

با سلام،

شما از سوی همکار خود دعوت شده‌اید تا به پلتفرم گُدارAI بپیوندید و به مجموعه خدمات و امکانات آن در محدوده {{tenant_name}} دسترسی داشته باشید.

برای تکمیل فرآیند عضویت و فعال‌سازی حساب کاربری خود، لطفاً از طریق لینک زیر اقدام فرمایید:

{{invitation_link}}

پس از ورود، می‌توانید با استفاده از اطلاعات کاربری خود، فعالیت خود را در محیط گُدارAI آغاز کنید.

در صورتی که با این دعوت آشنایی ندارید یا تصور می‌کنید این ایمیل به اشتباه برای شما ارسال شده است، می‌توانید این پیام را نادیده بگیرید. در این صورت هیچ اقدامی از سوی شما لازم نخواهد بود.

برای اطلاعات بیشتر درباره گُدارAI، می‌توانید به آدرس زیر مراجعه نمایید:
https://sharah.ir

همچنین دسترسی به Tenant اختصاصی شما از طریق لینک زیر امکان‌پذیر است:
https://{{tenant_slug}}.sharah.ir

با احترام
تیم پشتیبانی گُدارAI
Email: salam@sharah.ir`;

const invitationHTMLTemplate = `<div dir="rtl" style="font-family: Tahoma, Arial, sans-serif; line-height: 1.9; color: #111827;">
  <p>کاربر گرامی،</p>
  <p>با سلام،</p>
  <p>
    شما از سوی همکار خود دعوت شده‌اید تا به پلتفرم <strong>گُدارAI</strong> بپیوندید و به مجموعه خدمات و امکانات آن در محدوده
    <strong>{{tenant_name}}</strong> دسترسی داشته باشید.
  </p>
  <p>برای تکمیل فرآیند عضویت و فعال‌سازی حساب کاربری خود، لطفاً از طریق لینک زیر اقدام فرمایید:</p>
  <p><a href="{{invitation_link}}">{{invitation_link}}</a></p>
  <p>پس از ورود، می‌توانید با استفاده از اطلاعات کاربری خود، فعالیت خود را در محیط گُدارAI آغاز کنید.</p>
  <p>
    در صورتی که با این دعوت آشنایی ندارید یا تصور می‌کنید این ایمیل به اشتباه برای شما ارسال شده است، می‌توانید این پیام را نادیده بگیرید.
    در این صورت هیچ اقدامی از سوی شما لازم نخواهد بود.
  </p>
  <p>
    برای اطلاعات بیشتر درباره گُدارAI، می‌توانید به آدرس زیر مراجعه نمایید:
    <a href="https://sharah.ir">https://sharah.ir</a>
  </p>
  <p>
    همچنین دسترسی به Tenant اختصاصی شما از طریق لینک زیر امکان‌پذیر است:
    <a href="https://{{tenant_slug}}.sharah.ir">https://{{tenant_slug}}.sharah.ir</a>
  </p>
  <p>
    با احترام<br />
    تیم پشتیبانی گُدارAI<br />
    Email:
    <a href="mailto:salam@sahrah.ir">salam@sahrah.ir</a>
  </p>
</div>`;

function renderTemplate(template: string, replacements: Record<string, string>) {
  let output = template;
  for (const [key, value] of Object.entries(replacements)) {
    output = output.replaceAll(`{{${key}}}`, value);
  }
  return output;
}

export function renderInvitationEmailTemplate(input: InvitationTemplateInput) {
  const replacements = {
    tenant_name: input.tenantName,
    tenant_slug: input.tenantSlug,
    invitation_link: input.invitationLink,
  };

  return {
    subject: renderTemplate(invitationSubjectTemplate, replacements),
    text: renderTemplate(invitationTextTemplate, replacements),
    html: renderTemplate(invitationHTMLTemplate, replacements),
  };
}
