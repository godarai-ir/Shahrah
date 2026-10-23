import Link from 'next/link';
import { Section, StepList } from './patterns';
import { sectionHeading, textLink } from './sectionType';

const steps = [
  { title: 'فضای کاری بسازید', text: 'ثبت‌نام می‌کنید و فضای کاری سازمانتان همان لحظه آماده است.' },
  { title: 'یک دستیار و یک اتصال تعریف کنید', text: 'مثلاً دستیار پشتیبانی که به CRM شما وصل است. برای شروع همین کافی است.' },
  { title: 'تیم را دعوت کنید', text: 'سقف مصرف را بگذارید و بگذارید همکارها امتحان کنند.' },
];

const code = `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_GODARAI_TOKEN",
    base_url="https://YOUR_WORKSPACE.godarai.ir/v1",
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "یک پیام کوتاه برای پیگیری سفارش بنویس."}],
)`;

export function GetStarted() {
  return (
    <Section id="start" tone="panel">
      <div>
        <h2 className={`mb-8 max-w-3xl ${sectionHeading}`}>سه قدم تا اولین دستیار. بدون شیرجه.</h2>
      </div>

      <div className="mb-6">
        <StepList steps={steps} />
      </div>
      <p className="mb-10 text-label text-site-ink-soft lg:mb-16">
        اگر وسط راه گیر کردید، به ما بگویید.{' '}
        <Link href="/contact" className={textLink}>
          تماس با ما
        </Link>
      </p>

      <div className="grid items-center gap-8 rounded-card border border-site-line bg-site-surface p-6 sm:p-10 lg:grid-cols-2 lg:gap-16">
        <div>
          <h3 className="mb-2 text-h3 font-semibold text-site-ink">برای تیم فنی</h3>
          <p className="mb-4 text-body text-site-ink-soft">
            اگر خودتان محصول می‌سازید، همان درگاه را مستقیم هم می‌توانید بردارید. API سازگار با OpenAI است؛ فقط نشانی و کلید را عوض می‌کنید.
          </p>
          <Link href="/docs" className={textLink}>
            مشاهده‌ی مستندات
          </Link>
        </div>
        <div dir="ltr" className="min-w-0 overflow-hidden rounded-card bg-site-night">
          <pre className="overflow-x-auto p-5 text-caption text-site-night-soft sm:p-6 sm:text-label">
            <code>{code}</code>
          </pre>
        </div>
      </div>
    </Section>
  );
}
